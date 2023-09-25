// Copyright (c) 2023 Tim <tbckr>
//
// Permission is hereby granted, free of charge, to any person obtaining a copy of
// this software and associated documentation files (the "Software"), to deal in
// the Software without restriction, including without limitation the rights to
// use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of
// the Software, and to permit persons to whom the Software is furnished to do so,
// subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS
// FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR
// COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER
// IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN
// CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
//
// SPDX-License-Identifier: MIT

package cli

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/sashabaranov/go-openai"
	"github.com/stretchr/testify/require"
	"github.com/tbckr/sgpt/v2/api"
	"github.com/tbckr/sgpt/v2/chat"
)

func TestCreateViperConfig(t *testing.T) {
	config, err := createViperConfig()
	require.NoError(t, err)
	require.NotNil(t, config)
}

func TestRootCmd_SimplePrompt(t *testing.T) {
	prompt := "Say: Hello World!"
	expected := "Hello World!\n"

	mem := &exitMemento{}
	var wg sync.WaitGroup
	reader, writer := io.Pipe()

	config := createTestConfig(t)

	root := newRootCmd(mem.Exit, config, api.MockClient(strings.Clone(expected), nil))
	root.cmd.SetOut(writer)

	wg.Add(1)
	go func() {
		defer wg.Done()
		var buf bytes.Buffer
		_, err := io.Copy(&buf, reader)
		require.NoError(t, err)
		require.NoError(t, reader.Close())
		require.Equal(t, expected, buf.String())
	}()

	root.Execute([]string{"txt", prompt})
	require.Equal(t, 0, mem.code)
	require.NoError(t, writer.Close())

	wg.Wait()
}

func TestRootCmd_SimplePromptOnly(t *testing.T) {
	prompt := "Say: Hello World!"
	expected := "Hello World!\n"

	mem := &exitMemento{}
	var wg sync.WaitGroup
	reader, writer := io.Pipe()

	config := createTestConfig(t)

	root := newRootCmd(mem.Exit, config, api.MockClient(strings.Clone(expected), nil))
	root.cmd.SetOut(writer)

	wg.Add(1)
	go func() {
		defer wg.Done()
		var buf bytes.Buffer
		_, err := io.Copy(&buf, reader)
		require.NoError(t, err)
		require.NoError(t, reader.Close())
		require.Equal(t, expected, buf.String())
	}()

	root.Execute([]string{prompt})
	require.Equal(t, 0, mem.code)
	require.NoError(t, writer.Close())

	wg.Wait()
}

func TestRootCmd_SimplePromptOverrideValuesWithConfigFile(t *testing.T) {
	prompt := "Say: Hello World!"
	mem := &exitMemento{}

	configDir := t.TempDir()

	config, err := createViperConfig()
	config.SetConfigFile(filepath.Join(configDir, "config.yaml"))
	config.Set("TESTING", 1)

	var configFile *os.File
	configFile, err = os.Create(filepath.Join(configDir, "config.yaml"))
	require.NoError(t, err)

	_, err = configFile.WriteString(fmt.Sprintf("model: \"%s\"\n", openai.GPT4))

	root := newRootCmd(mem.Exit, config, api.MockClient("Hello World", nil))

	root.Execute([]string{"txt", prompt})
	require.Equal(t, 0, mem.code)

	require.Equal(t, openai.GPT4, config.GetString("model"))
}

func TestRootCmd_SimplePromptNoPrompt(t *testing.T) {
	mem := &exitMemento{}

	config := createTestConfig(t)

	root := newRootCmd(mem.Exit, config, api.MockClient("", nil))

	root.Execute([]string{})
	require.Equal(t, 1, mem.code)
}

func TestRootCmd_SimplePromptVerbose(t *testing.T) {
	prompt := "Say: Hello World!"
	expected := "Hello World!\n"

	mem := &exitMemento{}
	var wg sync.WaitGroup
	reader, writer := io.Pipe()

	config := createTestConfig(t)

	root := newRootCmd(mem.Exit, config, api.MockClient(strings.Clone(expected), nil))
	root.cmd.SetOut(writer)

	wg.Add(1)
	go func() {
		defer wg.Done()
		var buf bytes.Buffer
		_, err := io.Copy(&buf, reader)
		require.NoError(t, err)
		require.NoError(t, reader.Close())
		require.Equal(t, expected, buf.String())
	}()

	root.Execute([]string{"txt", prompt, "--verbose"})
	require.Equal(t, 0, mem.code)
	require.NoError(t, writer.Close())

	wg.Wait()
}

func TestRootCmd_SimplePromptViaStdin(t *testing.T) {
	prompt := "Say: Hello World!"
	expected := "Hello World!\n"

	mem := &exitMemento{}
	var wg sync.WaitGroup
	stdinReader, stdinWriter := io.Pipe()
	stdoutReader, stdoutWriter := io.Pipe()

	config := createTestConfig(t)

	root := newRootCmd(mem.Exit, config, api.MockClient(strings.Clone(expected), nil))
	root.cmd.SetIn(stdinReader)
	root.cmd.SetOut(stdoutWriter)

	wg.Add(1)
	go func() {
		defer wg.Done()
		_, errWrite := stdinWriter.Write([]byte(prompt))
		require.NoError(t, stdinWriter.Close())
		require.NoError(t, errWrite)
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		var buf bytes.Buffer
		_, err := io.Copy(&buf, stdoutReader)
		require.NoError(t, err)
		require.NoError(t, stdoutReader.Close())
		require.Equal(t, expected, buf.String())
	}()

	root.Execute([]string{})
	require.Equal(t, 0, mem.code)
	require.NoError(t, stdinReader.Close())
	require.NoError(t, stdoutWriter.Close())

	wg.Wait()
}

func TestRootCmd_SimplePromptViaStdinAndModifier(t *testing.T) {
	prompt := "Say: Hello World!"
	expected := "Hello World!\n"

	mem := &exitMemento{}
	var wg sync.WaitGroup
	stdinReader, stdinWriter := io.Pipe()
	stdoutReader, stdoutWriter := io.Pipe()

	config := createTestConfig(t)

	root := newRootCmd(mem.Exit, config, api.MockClient(strings.Clone(expected), nil))
	root.cmd.SetIn(stdinReader)
	root.cmd.SetOut(stdoutWriter)

	wg.Add(1)
	go func() {
		defer wg.Done()
		_, errWrite := stdinWriter.Write([]byte(prompt))
		require.NoError(t, stdinWriter.Close())
		require.NoError(t, errWrite)
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		var buf bytes.Buffer
		_, err := io.Copy(&buf, stdoutReader)
		require.NoError(t, err)
		require.NoError(t, stdoutReader.Close())
		require.Equal(t, expected, buf.String())
	}()

	root.Execute([]string{"txt"})
	require.Equal(t, 0, mem.code)
	require.NoError(t, stdinReader.Close())
	require.NoError(t, stdoutWriter.Close())

	wg.Wait()
}

func TestRootCmd_SimpleShellPrompt(t *testing.T) {
	prompt := `echo "Hello World"`
	expected := "Hello World\n"

	mem := &exitMemento{}
	var wg sync.WaitGroup
	reader, writer := io.Pipe()

	config := createTestConfig(t)

	err := os.Setenv("SHELL", "/bin/bash")
	require.NoError(t, err)
	t.Cleanup(func() {
		require.NoError(t, os.Unsetenv("SHELL"))
	})

	root := newRootCmd(mem.Exit, config, api.MockClient(strings.Clone(expected), nil))
	root.cmd.SetOut(writer)

	wg.Add(1)
	go func() {
		defer wg.Done()
		var buf bytes.Buffer
		_, err := io.Copy(&buf, reader)
		require.NoError(t, err)
		require.NoError(t, reader.Close())
		require.Equal(t, expected, buf.String())
	}()

	root.Execute([]string{"sh", prompt})
	require.Equal(t, 0, mem.code)
	require.NoError(t, writer.Close())

	wg.Wait()
}

func TestRootCmd_SimpleShellPromptWithExecution(t *testing.T) {
	prompt := `Print: Hello World`
	expected := "echo \"Hello World\"\n"

	mem := &exitMemento{}
	var wg sync.WaitGroup
	stdinReader, stdinWriter := io.Pipe()
	stdoutReader, stdoutWriter := io.Pipe()

	config := createTestConfig(t)

	err := os.Setenv("SHELL", "/bin/bash")
	require.NoError(t, err)
	t.Cleanup(func() {
		require.NoError(t, os.Unsetenv("SHELL"))
	})

	root := newRootCmd(mem.Exit, config, api.MockClient(strings.Clone(expected), nil))
	root.cmd.SetIn(stdinReader)
	root.cmd.SetOut(stdoutWriter)

	wg.Add(1)
	go func() {
		defer wg.Done()
		_, errWrite := stdinWriter.Write([]byte("\n"))
		require.NoError(t, stdinWriter.Close())
		require.NoError(t, errWrite)
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		var buf bytes.Buffer
		_, err := io.Copy(&buf, stdoutReader)
		require.NoError(t, err)
		require.NoError(t, stdoutReader.Close())
		stdoutOutput := expected + "Do you want to execute this command? (Y/n) Hello World\n"
		require.Equal(t, stdoutOutput, buf.String())
	}()

	root.Execute([]string{"sh", prompt, "--execute"})
	require.Equal(t, 0, mem.code)

	require.NoError(t, stdinReader.Close())
	require.NoError(t, stdoutWriter.Close())

	wg.Wait()
}

func TestRootCmd_SimplePromptWithChat(t *testing.T) {
	prompt := "Say: Hello World!"
	expected := "Hello World!\n"

	mem := &exitMemento{}
	var wg sync.WaitGroup
	reader, writer := io.Pipe()

	config := createTestConfig(t)

	root := newRootCmd(mem.Exit, config, api.MockClient(strings.Clone(expected), nil))
	root.cmd.SetOut(writer)

	wg.Add(1)
	go func() {
		defer wg.Done()
		var buf bytes.Buffer
		_, err := io.Copy(&buf, reader)
		require.NoError(t, err)
		require.NoError(t, reader.Close())
		require.Equal(t, expected, buf.String())
	}()

	root.Execute([]string{"txt", prompt, "--chat", "test_chat"})
	require.Equal(t, 0, mem.code)
	require.NoError(t, writer.Close())

	require.FileExists(t, filepath.Join(config.GetString("cacheDir"), "test_chat"))

	manager, err := chat.NewFilesystemChatSessionManager(config)
	require.NoError(t, err)

	var messages []openai.ChatCompletionMessage
	messages, err = manager.GetSession("test_chat")
	require.NoError(t, err)
	require.Len(t, messages, 2)

	// Check if the prompt was added
	require.Equal(t, openai.ChatMessageRoleUser, messages[0].Role)
	require.Equal(t, prompt, messages[0].Content)

	// Check if the response was added
	require.Equal(t, openai.ChatMessageRoleAssistant, messages[1].Role)
	require.Equal(t, strings.TrimSpace(expected), messages[1].Content)

	wg.Wait()
}

func TestRootCmd_SimplePromptWithChatAndCustomPersona(t *testing.T) {
	persona := "This is my custom persona"
	prompt := "Say: Hello World!"
	expected := "Hello World!\n"

	mem := &exitMemento{}
	var wg sync.WaitGroup
	reader, writer := io.Pipe()

	config := createTestConfig(t)

	err := os.Setenv("SHELL", "/bin/bash")
	require.NoError(t, err)
	t.Cleanup(func() {
		require.NoError(t, os.Unsetenv("SHELL"))
	})

	fileHandler, err := os.Create(filepath.Join(config.GetString("personas"), "my-persona"))
	require.NoError(t, err)
	_, err = fileHandler.WriteString(persona)
	require.NoError(t, err)
	require.NoError(t, fileHandler.Close())

	root := newRootCmd(mem.Exit, config, api.MockClient(strings.Clone(expected), nil))
	root.cmd.SetOut(writer)

	wg.Add(1)
	go func() {
		defer wg.Done()
		var buf bytes.Buffer
		_, err := io.Copy(&buf, reader)
		require.NoError(t, err)
		require.NoError(t, reader.Close())
		require.Equal(t, expected, buf.String())
	}()

	root.Execute([]string{"my-persona", prompt, "--chat", "test_chat"})
	require.Equal(t, 0, mem.code)
	require.NoError(t, writer.Close())

	require.FileExists(t, filepath.Join(config.GetString("cacheDir"), "test_chat"))

	var manager chat.SessionManager
	manager, err = chat.NewFilesystemChatSessionManager(config)
	require.NoError(t, err)

	var messages []openai.ChatCompletionMessage
	messages, err = manager.GetSession("test_chat")
	require.NoError(t, err)
	require.Len(t, messages, 3)

	// Check if the persona was added
	require.Equal(t, openai.ChatMessageRoleSystem, messages[0].Role)
	require.Equal(t, persona, messages[0].Content)

	// Check if the prompt was added
	require.Equal(t, openai.ChatMessageRoleUser, messages[1].Role)
	require.Equal(t, prompt, messages[1].Content)

	// Check if the response was added
	require.Equal(t, openai.ChatMessageRoleAssistant, messages[2].Role)
	require.Equal(t, strings.TrimSpace(expected), messages[2].Content)

	wg.Wait()
}

func TestRootCmd_ChatConversation(t *testing.T) {
	prompt := "Repeat last message"
	expected := "World!\n"

	mem := &exitMemento{}
	var wg sync.WaitGroup
	reader, writer := io.Pipe()

	config := createTestConfig(t)

	// Create an existing chat session
	manager, err := chat.NewFilesystemChatSessionManager(config)
	require.NoError(t, err)
	err = manager.SaveSession("test_chat", []openai.ChatCompletionMessage{
		{
			Role:    openai.ChatMessageRoleUser,
			Content: "Hello",
		},
		{
			Role:    openai.ChatMessageRoleAssistant,
			Content: "World!",
		},
	})
	require.NoError(t, err)

	root := newRootCmd(mem.Exit, config, api.MockClient(strings.Clone(expected), nil))
	root.cmd.SetOut(writer)

	wg.Add(1)
	go func() {
		defer wg.Done()
		var buf bytes.Buffer
		_, err := io.Copy(&buf, reader)
		require.NoError(t, err)
		require.NoError(t, reader.Close())
		require.Equal(t, expected, buf.String())
	}()

	root.Execute([]string{"txt", prompt, "--chat", "test_chat"})
	require.Equal(t, 0, mem.code)
	require.NoError(t, writer.Close())

	require.FileExists(t, filepath.Join(config.GetString("cacheDir"), "test_chat"))

	var messages []openai.ChatCompletionMessage
	messages, err = manager.GetSession("test_chat")
	require.NoError(t, err)
	require.Len(t, messages, 4)

	// Check if the prompt was added
	require.Equal(t, openai.ChatMessageRoleUser, messages[2].Role)
	require.Equal(t, prompt, messages[2].Content)

	// Check if the response was added
	require.Equal(t, openai.ChatMessageRoleAssistant, messages[3].Role)
	require.Equal(t, strings.TrimSpace(expected), messages[3].Content)

	wg.Wait()
}
