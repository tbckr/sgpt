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
	"testing"

	"github.com/jarcoal/httpmock"
	"github.com/sashabaranov/go-openai"
	"github.com/stretchr/testify/require"
	"github.com/tbckr/sgpt/v2/internal/testlib"
	"github.com/tbckr/sgpt/v2/pkg/chat"
)

func TestCreateClient(t *testing.T) {
	// Set the api key
	err := os.Setenv("OPENAI_API_KEY", "test")
	require.NoError(t, err)

	var client *OpenAIClient
	client, err = CreateClient(nil, nil)
	require.NoError(t, err)
	require.NotNil(t, client)
}

func TestCreateClientMissingApiKey(t *testing.T) {
	err := os.Unsetenv("OPENAI_API_KEY")
	require.NoError(t, err)

	var client *OpenAIClient
	client, err = CreateClient(nil, nil)
	require.Error(t, err)
	require.ErrorIs(t, err, ErrMissingAPIKey)
	require.Nil(t, client)
}

func TestSimplePrompt(t *testing.T) {
	testCtx := testlib.NewTestCtx(t)
	testlib.SetApiKey(t)
	client, err := CreateClient(testCtx.Config, nil)
	require.NoError(t, err)

	prompt := "Say: Hello World!"
	expected := "Hello World!"

	httpmock.ActivateNonDefault(client.HTTPClient)
	t.Cleanup(httpmock.DeactivateAndReset)
	testlib.RegisterExpectedChatResponse(expected)

	var result string
	result, err = client.CreateCompletion(context.Background(), "", prompt, "txt")
	require.NoError(t, err)
	require.Equal(t, expected, result)

	// Cache dir should be empty
	cacheDir := testCtx.Config.GetString("cacheDir")
	err = filepath.Walk(cacheDir, func(path string, info os.FileInfo, err error) error {
		if path == cacheDir {
			// Skip the root dir
			return nil
		}
		require.NoError(t, err)
		require.Fail(t, "Cache dir should be empty")
		return nil
	})
	require.NoError(t, err)
}

func TestPromptSaveAsChat(t *testing.T) {
	testCtx := testlib.NewTestCtx(t)
	testlib.SetApiKey(t)
	client, err := CreateClient(testCtx.Config, nil)
	require.NoError(t, err)

	prompt := "Say: Hello World!"
	expected := "Hello World!"

	httpmock.ActivateNonDefault(client.HTTPClient)
	t.Cleanup(httpmock.DeactivateAndReset)
	testlib.RegisterExpectedChatResponse(expected)

	var result string
	result, err = client.CreateCompletion(context.Background(), "test_chat", prompt, "txt")
	require.NoError(t, err)
	require.Equal(t, expected, result)

	require.FileExists(t, filepath.Join(testCtx.Config.GetString("cacheDir"), "test_chat"))

	var manager chat.SessionManager
	manager, err = chat.NewFilesystemChatSessionManager(testCtx.Config)
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
	testCtx := testlib.NewTestCtx(t)
	testlib.SetApiKey(t)
	client, err := CreateClient(testCtx.Config, nil)
	require.NoError(t, err)

	prompt := "Repeat last message"
	expected := "World!"

	httpmock.ActivateNonDefault(client.HTTPClient)
	t.Cleanup(httpmock.DeactivateAndReset)
	testlib.RegisterExpectedChatResponse(expected)

	var manager chat.SessionManager
	manager, err = chat.NewFilesystemChatSessionManager(testCtx.Config)
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
	result, err = client.CreateCompletion(context.Background(), "test_chat", prompt, "txt")
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
	testCtx := testlib.NewTestCtx(t)
	testlib.SetApiKey(t)
	client, err := CreateClient(testCtx.Config, nil)
	require.NoError(t, err)

	prompt := "Print Hello World"
	response := `echo \"Hello World\"`
	expected := `echo "Hello World"`

	httpmock.ActivateNonDefault(client.HTTPClient)
	t.Cleanup(httpmock.DeactivateAndReset)
	testlib.RegisterExpectedChatResponse(response)

	err = os.Setenv("SHELL", "/bin/bash")
	require.NoError(t, err)
	t.Cleanup(func() {
		require.NoError(t, os.Unsetenv("SHELL"))
	})

	testCtx.Config.Set("chat", "test_chat")

	var result string
	result, err = client.CreateCompletion(context.Background(), "test_chat", prompt, "sh")
	require.NoError(t, err)
	require.Equal(t, expected, result)

	require.FileExists(t, filepath.Join(testCtx.Config.GetString("cacheDir"), "test_chat"))

	var manager chat.SessionManager
	manager, err = chat.NewFilesystemChatSessionManager(testCtx.Config)
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
