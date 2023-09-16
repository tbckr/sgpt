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

package api

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/sashabaranov/go-openai"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"
	"github.com/tbckr/sgpt/chat"
)

func createTestConfig(t *testing.T) *viper.Viper {
	cacheDir, err := createTempDir("cache")
	require.NoError(t, err)

	var configDir string
	configDir, err = createTempDir("config")
	require.NoError(t, err)

	config := viper.New()
	config.AddConfigPath(configDir)
	config.SetConfigFile("config")
	config.SetConfigType("yaml")
	config.Set("cacheDir", cacheDir)

	return config
}

func teardownTestDirs(t *testing.T, config *viper.Viper) {
	cacheDir := config.GetString("cacheDir")
	require.NoError(t, removeTempDir(cacheDir))

	configDir := config.ConfigFileUsed()
	require.NoError(t, removeTempDir(configDir))
}

func createTempDir(suffix string) (string, error) {
	if suffix != "" {
		suffix = "_" + suffix
	}
	return os.MkdirTemp("", strings.Join([]string{"sgpt_temp_*", suffix}, ""))
}

func removeTempDir(fullConfigPath string) error {
	return os.RemoveAll(fullConfigPath)
}

func TestCreateClient(t *testing.T) {
	// Set the api key
	err := os.Setenv("OPENAI_API_KEY", "test")
	require.NoError(t, err)

	var client *OpenAIClient
	client, err = CreateClient()
	require.NoError(t, err)
	require.NotNil(t, client)
}

func TestCreateClientMissingApiKey(t *testing.T) {
	err := os.Unsetenv("OPENAI_API_KEY")
	require.NoError(t, err)

	var client *OpenAIClient
	client, err = CreateClient()
	require.Error(t, err)
	require.ErrorIs(t, err, ErrMissingAPIKey)
	require.Nil(t, client)
}

func TestSimplePrompt(t *testing.T) {
	prompt := "Say: Hello World!"
	expected := "Hello World!"

	client, err := MockClient(strings.Clone(expected), nil)()
	require.NoError(t, err)
	config := createTestConfig(t)
	defer teardownTestDirs(t, config)

	var result string
	result, err = client.GetChatCompletion(context.Background(), config, prompt, "txt")
	require.NoError(t, err)
	require.Equal(t, expected, result)

	// Cache dir should be empty
	cacheDir := config.GetString("cacheDir")
	err = filepath.Walk(cacheDir, func(path string, info os.FileInfo, err error) error {
		if path == cacheDir {
			// Skip the root dir
			return nil
		}
		require.NoError(t, err)
		require.Fail(t, "Cache dir should be empty")
		return nil
	})
}

func TestPromptSaveAsChat(t *testing.T) {
	prompt := "Say: Hello World!"
	expected := "Hello World!"

	client, err := MockClient(strings.Clone(expected), nil)()
	require.NoError(t, err)
	config := createTestConfig(t)
	defer teardownTestDirs(t, config)

	config.Set("chat", "test_chat")

	var result string
	result, err = client.GetChatCompletion(context.Background(), config, prompt, "txt")
	require.NoError(t, err)
	require.Equal(t, expected, result)

	require.FileExists(t, filepath.Join(config.GetString("cacheDir"), "test_chat"))

	var manager chat.SessionManager
	manager, err = chat.NewFilesystemChatSessionManager(config)
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
	require.Equal(t, expected, messages[1].Content)
}

func TestPromptLoadChat(t *testing.T) {
	prompt := "Repeat last message"
	expected := "World!"

	client, err := MockClient(strings.Clone(expected), nil)()
	require.NoError(t, err)
	config := createTestConfig(t)
	defer teardownTestDirs(t, config)

	config.Set("chat", "test_chat")

	var manager chat.SessionManager
	manager, err = chat.NewFilesystemChatSessionManager(config)
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

	var result string
	result, err = client.GetChatCompletion(context.Background(), config, prompt, "txt")
	require.NoError(t, err)
	require.Equal(t, expected, result)

	var messages []openai.ChatCompletionMessage
	messages, err = manager.GetSession("test_chat")
	require.NoError(t, err)
	require.Len(t, messages, 4)

	// Check if the prompt was added
	require.Equal(t, openai.ChatMessageRoleUser, messages[2].Role)
	require.Equal(t, prompt, messages[2].Content)

	// Check if the response was added
	require.Equal(t, openai.ChatMessageRoleAssistant, messages[3].Role)
	require.Equal(t, expected, messages[3].Content)
}

func TestPromptWithModifier(t *testing.T) {
	prompt := "Print Hello World"
	expected := `echo "Hello World"`

	client, err := MockClient(strings.Clone(expected), nil)()
	require.NoError(t, err)
	config := createTestConfig(t)
	defer teardownTestDirs(t, config)

	config.Set("chat", "test_chat")

	var result string
	result, err = client.GetChatCompletion(context.Background(), config, prompt, "sh")
	require.NoError(t, err)
	require.Equal(t, expected, result)

	require.FileExists(t, filepath.Join(config.GetString("cacheDir"), "test_chat"))

	var manager chat.SessionManager
	manager, err = chat.NewFilesystemChatSessionManager(config)
	require.NoError(t, err)

	var messages []openai.ChatCompletionMessage
	messages, err = manager.GetSession("test_chat")
	require.NoError(t, err)
	require.Len(t, messages, 3)

	// Check if the modifier message was added
	require.Equal(t, openai.ChatMessageRoleSystem, messages[0].Role)

	// Check if the prompt was added
	require.Equal(t, openai.ChatMessageRoleUser, messages[1].Role)
	require.Equal(t, prompt, messages[1].Content)

	// Check if the response was added
	require.Equal(t, openai.ChatMessageRoleAssistant, messages[2].Role)
	require.Equal(t, expected, messages[2].Content)
}
