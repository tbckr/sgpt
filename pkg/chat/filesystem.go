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
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/sashabaranov/go-openai"
	"github.com/spf13/viper"
)

type FilesystemChatSessionManager struct {
	config *viper.Viper
}

func NewFilesystemChatSessionManager(config *viper.Viper) (SessionManager, error) {
	return FilesystemChatSessionManager{
		config: config,
	}, nil
}

func (m FilesystemChatSessionManager) getFilepathForSession(sessionName string) (string, error) {
	cacheDir := m.config.GetString("cacheDir")
	sessionDir := filepath.Join(cacheDir, sessionName)
	return sessionDir, nil
}

func (m FilesystemChatSessionManager) getMessagesFilepath(sessionDir string) string {
	return filepath.Join(sessionDir, "messages.json")
}

func (m FilesystemChatSessionManager) fileExists(filePath string) (bool, error) {
	info, err := os.Stat(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}
	// For chat sessions, we expect a directory
	if !info.IsDir() {
		return false, fmt.Errorf("%q is not a directory", filePath)
	}
	return true, nil
}

func (m FilesystemChatSessionManager) SessionExists(sessionName string) (bool, error) {
	if err := validateSessionName(sessionName); err != nil {
		return false, err
	}
	sessionFilepath, err := m.getFilepathForSession(sessionName)
	if err != nil {
		return false, err
	}
	var exists bool
	exists, err = m.fileExists(sessionFilepath)
	if err != nil {
		return false, err
	}
	return exists, nil
}

func (m FilesystemChatSessionManager) GetSession(sessionName string) ([]openai.ChatCompletionMessage, error) {
	// Validate session name
	if err := validateSessionName(sessionName); err != nil {
		return nil, err
	}

	// Get path to file
	sessionFilepath, err := m.getFilepathForSession(sessionName)
	if err != nil {
		return nil, err
	}
	slog.Debug("Using session file at: " + sessionFilepath)

	// Check, if session exists
	var exists bool
	exists, err = m.fileExists(sessionFilepath)
	if err != nil {
		return nil, err
	}
	if !exists {
		return []openai.ChatCompletionMessage{}, ErrChatSessionDoesNotExist
	}
	slog.Debug("Session exists")

	// Open messages file
	messagesFile := m.getMessagesFilepath(sessionFilepath)
	var file *os.File
	file, err = os.Open(messagesFile)
	if err != nil {
		if os.IsNotExist(err) {
			// Create empty messages file if it doesn't exist
			if err = os.MkdirAll(sessionFilepath, 0755); err != nil {
				return nil, err
			}
			if err = os.WriteFile(messagesFile, []byte("[]"), 0644); err != nil {
				return nil, err
			}
			return []openai.ChatCompletionMessage{}, nil
		}
		return nil, err
	}
	defer file.Close()

	slog.Debug("Reading messages from session file")
	// Read messages array
	var messages []openai.ChatCompletionMessage
	decoder := json.NewDecoder(file)
	if err = decoder.Decode(&messages); err != nil {
		return nil, fmt.Errorf("failed to decode messages: %w", err)
	}
	slog.Debug("Messages from session file imported")
	return messages, nil
}

func (m FilesystemChatSessionManager) SaveSession(sessionName string, messages []openai.ChatCompletionMessage) error {
	// Validate session name
	if err := validateSessionName(sessionName); err != nil {
		return err
	}

	// Get path to file
	sessionFilepath, err := m.getFilepathForSession(sessionName)
	if err != nil {
		return err
	}
	slog.Debug("Using session file at: " + sessionFilepath)

	// Check, if session exists
	var exists bool
	exists, err = m.fileExists(sessionFilepath)
	if err != nil {
		return err
	}
	slog.Debug("Session exists")

	// Create session directory if it doesn't exist
	if !exists {
		if err = os.MkdirAll(sessionFilepath, 0755); err != nil {
			return err
		}
		slog.Debug("Created new session directory")
	}

	// Open messages file
	messagesFile := m.getMessagesFilepath(sessionFilepath)
	var file *os.File
	if _, err = os.Stat(messagesFile); err == nil {
		// Open and truncate existing file
		file, err = os.OpenFile(messagesFile, os.O_WRONLY|os.O_TRUNC, defaultFilePermissions)
		if err != nil {
			return err
		}
		slog.Debug("Existing messages file opened and truncated")
	} else {
		// Create file
		file, err = os.Create(messagesFile)
		if err != nil {
			return err
		}
		slog.Debug("New messages file created")
	}
	defer file.Close()

	// Save messages as JSON array
	encoder := json.NewEncoder(file)
	if err = encoder.Encode(messages); err != nil {
		return fmt.Errorf("failed to encode messages: %w", err)
	}
	slog.Debug("Messages saved to session file")
	return nil
}

func (m FilesystemChatSessionManager) ListSessions() ([]string, error) {
	cacheDir := m.config.GetString("cacheDir")
	slog.Debug("Listing files in cache directory: " + cacheDir)
	dirFiles, err := os.ReadDir(cacheDir)
	if err != nil {
		return nil, err
	}
	slog.Debug("Iterating files in cache directory")
	var files []string
	for _, file := range dirFiles {
		files = append(files, file.Name())
	}
	return files, nil
}

func (m FilesystemChatSessionManager) DeleteSession(sessionName string) error {
	sessionFilepath, err := m.getFilepathForSession(sessionName)
	slog.Debug("Deleting session file at: " + sessionFilepath)
	if err != nil {
		return err
	}
	var exists bool
	exists, err = m.fileExists(sessionFilepath)
	if err != nil {
		return err
	}
	if !exists {
		slog.Debug("Session does not exist - nothing to delete")
		return nil
	}
	err = os.RemoveAll(sessionFilepath)
	if err != nil {
		return err
	}
	slog.Debug("Session deleted")
	return nil
}
