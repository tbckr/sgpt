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
	"bufio"
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/sashabaranov/go-openai"
	jww "github.com/spf13/jwalterweatherman"
	"github.com/spf13/viper"
)

type FilesystemChatSessionManager struct {
	config *viper.Viper
}

func NewFilesystemChatSessionManager(config *viper.Viper) (ChatSessionManager, error) {
	return FilesystemChatSessionManager{
		config: config,
	}, nil
}

func (m FilesystemChatSessionManager) getFilepathForSession(sessionName string) (string, error) {
	cacheDir := m.config.GetString("cacheDir")
	filePath := filepath.Join(cacheDir, sessionName)
	return filePath, nil
}

func (m FilesystemChatSessionManager) fileExists(filePath string) (bool, error) {
	if _, err := os.Stat(filePath); err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
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

	// Check, if session exists
	var exists bool
	exists, err = m.fileExists(sessionFilepath)
	if err != nil {
		return nil, err
	}
	if !exists {
		return []openai.ChatCompletionMessage{}, ErrChatSessionDoesNotExist
	}

	// Open file
	var file *os.File
	file, err = os.Open(sessionFilepath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	// Read messages
	scanner := bufio.NewScanner(file)
	scanner.Split(bufio.ScanLines)

	var messages []openai.ChatCompletionMessage
	var data []byte
	var readMessage openai.ChatCompletionMessage

	for scanner.Scan() {
		data = scanner.Bytes()
		readMessage = openai.ChatCompletionMessage{}
		if err = json.Unmarshal(data, &readMessage); err != nil {
			return nil, err
		}
		messages = append(messages, readMessage)
	}
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

	// Check, if session exists
	var exists bool
	exists, err = m.fileExists(sessionFilepath)
	if err != nil {
		return err
	}

	// Open file
	var file *os.File
	if exists {
		// Open and truncate existing file
		file, err = os.OpenFile(sessionFilepath, os.O_WRONLY|os.O_TRUNC, defaultFilePermissions)
		if err != nil {
			return err
		}
	} else {
		// Create file
		file, err = os.Create(sessionFilepath)
		if err != nil {
			return err
		}
	}
	defer file.Close()

	// Save messages to file
	var data []byte
	for _, message := range messages {
		data, err = json.Marshal(message)
		if err != nil {
			return err
		}
		_, err = file.WriteString(string(data) + "\n")
		if err != nil {
			return err
		}
	}
	jww.DEBUG.Println("session saved")
	return nil
}

func (m FilesystemChatSessionManager) ListSessions() ([]string, error) {
	cacheDir := m.config.GetString("cacheDir")
	dirFiles, err := os.ReadDir(cacheDir)
	if err != nil {
		return nil, err
	}
	var files []string
	for _, file := range dirFiles {
		files = append(files, file.Name())
	}
	return files, nil
}

func (m FilesystemChatSessionManager) DeleteSession(sessionName string) error {
	sessionFilepath, err := m.getFilepathForSession(sessionName)
	if err != nil {
		return err
	}
	var exists bool
	exists, err = m.fileExists(sessionFilepath)
	if err != nil {
		return err
	}
	if !exists {
		return nil
	}
	err = os.Remove(sessionFilepath)
	if err != nil {
		return err
	}
	return nil
}
