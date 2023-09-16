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
	"io"
	"path/filepath"
	"sync"
	"testing"

	"github.com/sashabaranov/go-openai"
	"github.com/tbckr/sgpt/chat"

	"github.com/stretchr/testify/require"
	"github.com/tbckr/sgpt/api"
)

func TestChatCmd(t *testing.T) {
	mem := &exitMemento{}

	config := createTestConfig(t)

	root := newRootCmd(mem.Exit, config, api.MockClient("", nil))

	root.Execute([]string{"chat"})
	require.Equal(t, 0, mem.code)
}

func TestChatCmdListEmptySessions(t *testing.T) {
	expected := ""

	mem := &exitMemento{}
	var wg sync.WaitGroup
	reader, writer := io.Pipe()

	config := createTestConfig(t)

	root := newRootCmd(mem.Exit, config, api.MockClient("", nil))
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

	root.Execute([]string{"chat", "ls"})
	require.Equal(t, 0, mem.code)
	require.NoError(t, writer.Close())

	wg.Wait()
}

func TestChatCmdListOneSession(t *testing.T) {
	expected := "test\n"

	mem := &exitMemento{}
	var wg sync.WaitGroup
	reader, writer := io.Pipe()

	config := createTestConfig(t)

	manager, err := chat.NewFilesystemChatSessionManager(config)
	require.NoError(t, err)

	messages := createTestMessages()
	err = manager.SaveSession("test", messages)
	require.NoError(t, err)

	root := newRootCmd(mem.Exit, config, api.MockClient("", nil))
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

	root.Execute([]string{"chat", "ls"})
	require.Equal(t, 0, mem.code)
	require.NoError(t, writer.Close())

	wg.Wait()
}

func TestChatCmdListTwoSessions(t *testing.T) {
	expected := "test\ntest2\n"

	mem := &exitMemento{}
	var wg sync.WaitGroup
	reader, writer := io.Pipe()

	config := createTestConfig(t)

	manager, err := chat.NewFilesystemChatSessionManager(config)
	require.NoError(t, err)

	messages := createTestMessages()
	err = manager.SaveSession("test", messages)
	require.NoError(t, err)
	err = manager.SaveSession("test2", messages)
	require.NoError(t, err)

	root := newRootCmd(mem.Exit, config, api.MockClient("", nil))
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

	root.Execute([]string{"chat", "ls"})
	require.Equal(t, 0, mem.code)
	require.NoError(t, writer.Close())

	wg.Wait()
}

func TestChatCmdShowSession(t *testing.T) {
	mem := &exitMemento{}
	var wg sync.WaitGroup
	reader, writer := io.Pipe()

	config := createTestConfig(t)

	manager, err := chat.NewFilesystemChatSessionManager(config)
	require.NoError(t, err)

	messages := createTestMessages()
	err = manager.SaveSession("test", messages)
	require.NoError(t, err)

	root := newRootCmd(mem.Exit, config, api.MockClient("", nil))
	root.cmd.SetOut(writer)

	wg.Add(1)
	go func() {
		defer wg.Done()
		var buf bytes.Buffer
		_, err := io.Copy(&buf, reader)
		require.NoError(t, err)
		require.NoError(t, reader.Close())
		require.Contains(t, buf.String(), "You are a chat bot.")
		require.Contains(t, buf.String(), "I am a chat bot.")
	}()

	root.Execute([]string{"chat", "show", "test"})
	require.Equal(t, 0, mem.code)
	require.NoError(t, writer.Close())

	wg.Wait()
}

func TestChatCmdShowSessionMissingName(t *testing.T) {
	mem := &exitMemento{}

	config := createTestConfig(t)

	root := newRootCmd(mem.Exit, config, api.MockClient("", nil))

	root.Execute([]string{"chat", "show"})
	require.Equal(t, 1, mem.code)
}

func TestChatCmdShowSessionNonExistent(t *testing.T) {
	mem := &exitMemento{}

	config := createTestConfig(t)

	manager, err := chat.NewFilesystemChatSessionManager(config)
	require.NoError(t, err)

	messages := createTestMessages()
	err = manager.SaveSession("test", messages)
	require.NoError(t, err)

	root := newRootCmd(mem.Exit, config, api.MockClient("", nil))

	root.Execute([]string{"chat", "show", "test2"})
	require.Equal(t, 1, mem.code)
}

func TestChatCmdShowSessionWithAlias(t *testing.T) {
	mem := &exitMemento{}
	var wg sync.WaitGroup
	reader, writer := io.Pipe()

	config := createTestConfig(t)

	manager, err := chat.NewFilesystemChatSessionManager(config)
	require.NoError(t, err)

	messages := createTestMessages()
	err = manager.SaveSession("test", messages)
	require.NoError(t, err)

	root := newRootCmd(mem.Exit, config, api.MockClient("", nil))
	root.cmd.SetOut(writer)

	wg.Add(1)
	go func() {
		defer wg.Done()
		var buf bytes.Buffer
		_, err := io.Copy(&buf, reader)
		require.NoError(t, err)
		require.NoError(t, reader.Close())
		require.Contains(t, buf.String(), "You are a chat bot.")
		require.Contains(t, buf.String(), "I am a chat bot.")
	}()

	root.Execute([]string{"chat", "cat", "test"})
	require.Equal(t, 0, mem.code)
	require.NoError(t, writer.Close())

	wg.Wait()
}

func TestChatCmdRmSession(t *testing.T) {
	mem := &exitMemento{}

	config := createTestConfig(t)

	manager, err := chat.NewFilesystemChatSessionManager(config)
	require.NoError(t, err)

	messages := createTestMessages()
	err = manager.SaveSession("test", messages)
	require.NoError(t, err)
	require.FileExists(t, filepath.Join(config.GetString("cacheDir"), "test"))

	root := newRootCmd(mem.Exit, config, api.MockClient("", nil))

	root.Execute([]string{"chat", "rm", "test"})
	require.Equal(t, 0, mem.code)
	require.NoFileExists(t, filepath.Join(config.GetString("cacheDir"), "test"))
}

func TestChatCmdRmSessionNonExistent(t *testing.T) {
	mem := &exitMemento{}

	config := createTestConfig(t)

	manager, err := chat.NewFilesystemChatSessionManager(config)
	require.NoError(t, err)

	messages := createTestMessages()
	err = manager.SaveSession("test", messages)
	require.NoError(t, err)
	require.FileExists(t, filepath.Join(config.GetString("cacheDir"), "test"))

	root := newRootCmd(mem.Exit, config, api.MockClient("", nil))

	root.Execute([]string{"chat", "rm", "test2"})
	require.Equal(t, 0, mem.code)
	require.FileExists(t, filepath.Join(config.GetString("cacheDir"), "test"))
}

func TestChatCmdRmSessionAll(t *testing.T) {
	mem := &exitMemento{}

	config := createTestConfig(t)

	manager, err := chat.NewFilesystemChatSessionManager(config)
	require.NoError(t, err)

	messages := createTestMessages()
	err = manager.SaveSession("test", messages)
	require.NoError(t, err)
	require.FileExists(t, filepath.Join(config.GetString("cacheDir"), "test"))
	err = manager.SaveSession("test2", messages)
	require.NoError(t, err)
	require.FileExists(t, filepath.Join(config.GetString("cacheDir"), "test2"))

	root := newRootCmd(mem.Exit, config, api.MockClient("", nil))

	root.Execute([]string{"chat", "rm", "--all"})
	require.Equal(t, 0, mem.code)
	require.NoFileExists(t, filepath.Join(config.GetString("cacheDir"), "test"))
	require.NoFileExists(t, filepath.Join(config.GetString("cacheDir"), "test2"))
}

func TestChatCmdRmSessionMissingName(t *testing.T) {
	mem := &exitMemento{}

	config := createTestConfig(t)

	root := newRootCmd(mem.Exit, config, api.MockClient("", nil))

	root.Execute([]string{"chat", "rm"})
	require.Equal(t, 1, mem.code)
}

func createTestMessages() []openai.ChatCompletionMessage {
	return []openai.ChatCompletionMessage{
		{
			Role:    openai.ChatMessageRoleUser,
			Content: "You are a chat bot.",
		},
		{
			Role:    openai.ChatMessageRoleAssistant,
			Content: "I am a chat bot.",
		},
	}
}
