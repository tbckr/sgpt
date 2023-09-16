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

package chat

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/sashabaranov/go-openai"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"
)

func TestValidateSessionNameInvalid(t *testing.T) {
	tests := []string{
		"test!123",               // special chars
		strings.Repeat("a", 128), // too long
		strings.Repeat("b", 69),  // too long
		"test 123",               // spaces
		"test!",                  // contains special character (!)
		"Test Space",             // contains space
		"Mixed Spaces _",         // contains spaces and underscore
		"mixed_!Chars",
		"Mixed Spaces _",
	}
	for _, tt := range tests {
		t.Run(tt, func(t *testing.T) {
			require.Error(t, validateSessionName(tt))
		})
	}
}

func TestValidateSessionNameValid(t *testing.T) {
	tests := []string{
		"a", "test", "my-chat",
		"123", "a1", "a-1", "a-1-2",
		"a-1-2-3", "a-1-2-3-4", "a-1-2-3-4-5",
		"a-1-2-3-4-5-6", "a-1-2-3-4-5-6-7", "a-1-2-3-4-5-6-7-8",
		"a-1-2-3-4-5-6-7-8-9", "a-1-2-3-4-5-6-7-8-9-0",
		"test_123", "test-123",
		"test_123",     // underscores
		"test-123",     // dashes
		"test-123-456", // multiple dashes{"Test"}
		"1234567890",   // only numbers
		"_test",        // leading underscore
		"test_",        // trailing underscore
		"-test",        // leading hyphen
		"test-",        // trailing hyphen
		"test__123",    // consecutive underscores
		"test--123",    // consecutive hyphens
		"test_123_",    // underscore at the end
		"test-123-",    // hyphen at the end
		"test_123-",    // underscore and hyphen at the end
		"test_123-456", // underscore, hyphen, and numbers
	}
	for _, tt := range tests {
		t.Run(tt, func(t *testing.T) {
			require.NoError(t, validateSessionName(tt))
		})
	}
}

func TestFilesystemChatSessionManager_SessionExists(t *testing.T) {
	config := createTestConfig(t)

	manager, err := NewFilesystemChatSessionManager(config)
	require.NoError(t, err)

	messages := createTestMessages()
	err = manager.SaveSession("test", messages)
	require.NoError(t, err)

	var exists bool
	exists, err = manager.SessionExists("test")
	require.NoError(t, err)
	require.True(t, exists)
	require.FileExists(t, filepath.Join(config.GetString("cacheDir"), "test"))
}

func TestFilesystemChatSessionManager_SessionDoesNotExist(t *testing.T) {
	config := createTestConfig(t)

	manager, err := NewFilesystemChatSessionManager(config)
	require.NoError(t, err)

	var exists bool
	exists, err = manager.SessionExists("test")
	require.NoError(t, err)
	require.False(t, exists)
	require.NoFileExists(t, filepath.Join(config.GetString("cacheDir"), "test"))
}

func TestFilesystemChatSessionManager_SaveExistingSession(t *testing.T) {
	config := createTestConfig(t)

	manager, err := NewFilesystemChatSessionManager(config)
	require.NoError(t, err)

	messages := createTestMessages()
	err = manager.SaveSession("test", messages)
	require.NoError(t, err)

	var exists bool
	exists, err = manager.SessionExists("test")
	require.NoError(t, err)
	require.True(t, exists)
	require.FileExists(t, filepath.Join(config.GetString("cacheDir"), "test"))

	messages = append(messages, openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleUser,
		Content: "Please stop talking.",
	})

	err = manager.SaveSession("test", messages)
	require.NoError(t, err)

	var loadedMessages []openai.ChatCompletionMessage
	loadedMessages, err = manager.GetSession("test")
	require.NoError(t, err)
	require.Equal(t, 3, len(loadedMessages))
}

func TestFilesystemChatSessionManager_GetSession(t *testing.T) {
	config := createTestConfig(t)

	manager, err := NewFilesystemChatSessionManager(config)
	require.NoError(t, err)

	messages := createTestMessages()
	err = manager.SaveSession("test", messages)
	require.NoError(t, err)

	var exists bool
	exists, err = manager.SessionExists("test")
	require.NoError(t, err)
	require.True(t, exists)
	require.FileExists(t, filepath.Join(config.GetString("cacheDir"), "test"))

	var loadedMessages []openai.ChatCompletionMessage
	loadedMessages, err = manager.GetSession("test")
	require.NoError(t, err)
	require.Equal(t, messages, loadedMessages)
}

func TestFilesystemChatSessionManager_GetSessionInvalid(t *testing.T) {
	config := createTestConfig(t)

	manager, err := NewFilesystemChatSessionManager(config)
	require.NoError(t, err)

	var messages []openai.ChatCompletionMessage
	messages, err = manager.GetSession("abcd")
	require.Error(t, err)
	require.Empty(t, messages)
}

func TestFilesystemChatSessionManager_ListSessions(t *testing.T) {
	config := createTestConfig(t)

	manager, err := NewFilesystemChatSessionManager(config)
	require.NoError(t, err)

	messages := createTestMessages()
	err = manager.SaveSession("test", messages)
	require.NoError(t, err)

	var exists bool
	exists, err = manager.SessionExists("test")
	require.NoError(t, err)
	require.True(t, exists)
	require.FileExists(t, filepath.Join(config.GetString("cacheDir"), "test"))

	var sessions []string
	sessions, err = manager.ListSessions()
	require.NoError(t, err)
	require.Equal(t, []string{"test"}, sessions)
}

func TestFilesystemChatSessionManager_DeleteSession(t *testing.T) {
	config := createTestConfig(t)

	manager, err := NewFilesystemChatSessionManager(config)
	require.NoError(t, err)

	messages := createTestMessages()
	err = manager.SaveSession("test", messages)
	require.NoError(t, err)

	var exists bool
	exists, err = manager.SessionExists("test")
	require.NoError(t, err)
	require.True(t, exists)
	require.FileExists(t, filepath.Join(config.GetString("cacheDir"), "test"))

	var sessions []string
	sessions, err = manager.ListSessions()
	require.NoError(t, err)
	require.Equal(t, []string{"test"}, sessions)

	err = manager.DeleteSession("test")
	require.NoError(t, err)

	exists, err = manager.SessionExists("test")
	require.NoError(t, err)
	require.False(t, exists)

	sessions, err = manager.ListSessions()
	require.NoError(t, err)
	require.Empty(t, sessions)
}

func TestFilesystemChatSessionManager_DeleteNotExistingSession(t *testing.T) {
	config := createTestConfig(t)

	manager, err := NewFilesystemChatSessionManager(config)
	require.NoError(t, err)

	messages := createTestMessages()
	err = manager.SaveSession("test", messages)
	require.NoError(t, err)

	var exists bool
	exists, err = manager.SessionExists("test")
	require.NoError(t, err)
	require.True(t, exists)
	require.FileExists(t, filepath.Join(config.GetString("cacheDir"), "test"))

	var sessions []string
	sessions, err = manager.ListSessions()
	require.NoError(t, err)
	require.Equal(t, []string{"test"}, sessions)

	err = manager.DeleteSession("test2")
	require.NoError(t, err)

	exists, err = manager.SessionExists("test")
	require.NoError(t, err)
	require.True(t, exists)

	exists, err = manager.SessionExists("test2")
	require.NoError(t, err)
	require.False(t, exists)
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

func createTestConfig(t *testing.T) *viper.Viper {
	cacheDir := createTempDir(t, "cache")
	configDir := createTempDir(t, "config")

	config := viper.New()
	config.AddConfigPath(configDir)
	config.SetConfigName("config")
	config.SetConfigType("yaml")
	config.Set("cacheDir", cacheDir)
	config.Set("TESTING", 1)

	return config
}

func createTempDir(t *testing.T, suffix string) string {
	if suffix != "" {
		suffix = "_" + suffix
	}
	tempFilepath, err := os.MkdirTemp("", strings.Join([]string{"sgpt_temp_*", suffix}, ""))
	require.NoError(t, err)
	t.Cleanup(func() {
		require.NoError(t, os.RemoveAll(tempFilepath))
	})
	return tempFilepath
}
