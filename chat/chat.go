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
	"errors"
	"fmt"
	"os"
	"path"
	"regexp"
	"unicode/utf8"

	jww "github.com/spf13/jwalterweatherman"

	"github.com/sashabaranov/go-openai"
	"github.com/tbckr/sgpt/fs"
)

const (
	appDirName             = "sgpt"
	defaultFilePermissions = 0755
	sessionNameMaxLength   = 65
)

var (
	ErrChatSessionIsNotExist  = errors.New("chat session does not exist")
	ErrChatSessionNameInvalid = fmt.Errorf("chat session name does not match the regex %s", sessionNameRegex)
	ErrChatSessionNameTooLong = fmt.Errorf("chat session name is greater than %d", sessionNameMaxLength)

	sessionNameRegex   = "^([-a-zA-Z0-9]*[a-zA-Z0-9])?"
	sessionNameMatcher = regexp.MustCompile(sessionNameRegex)
)

func getFilepathForSession(sessionName string) (string, error) {
	dir, err := fs.GetAppCacheDir(appDirName)
	if err != nil {
		return "", err
	}
	filePath := path.Join(dir, sessionName)
	jww.DEBUG.Printf("file path for session %s: %s", sessionName, filePath)
	return filePath, nil
}

func validateSession(sessionName string) error {
	if !sessionNameMatcher.Match([]byte(sessionName)) {
		jww.ERROR.Printf("session name %s does not match regex %s", sessionName, sessionNameRegex)
		return ErrChatSessionNameInvalid
	}
	if utf8.RuneCountInString(sessionName) > sessionNameMaxLength {
		jww.ERROR.Printf("session name is too long (max length is %d)", sessionNameMaxLength)
		return ErrChatSessionNameTooLong
	}
	jww.DEBUG.Println("session name is valid")
	return nil
}

func SessionExists(sessionName string) (bool, error) {
	if err := validateSession(sessionName); err != nil {
		return false, err
	}
	filepath, err := getFilepathForSession(sessionName)
	if err != nil {
		return false, err
	}
	var exists bool
	exists, err = fs.FileExists(filepath)
	if err != nil {
		return false, err
	}
	if exists {
		jww.DEBUG.Println("session exists")
		return true, nil
	}
	jww.DEBUG.Println("session does not exist")
	return false, nil
}

func GetSession(sessionName string) ([]openai.ChatCompletionMessage, error) {
	// Validate session name
	if err := validateSession(sessionName); err != nil {
		return nil, err
	}

	// Get path to file
	filepath, err := getFilepathForSession(sessionName)
	if err != nil {
		return nil, err
	}

	// Check, if session exists
	var exists bool
	exists, err = fs.FileExists(filepath)
	if err != nil {
		return nil, err
	}
	if !exists {
		return []openai.ChatCompletionMessage{}, ErrChatSessionIsNotExist
	}

	// Open file
	var file *os.File
	file, err = os.Open(filepath)
	if err != nil {
		jww.ERROR.Println("could not open file")
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
			jww.ERROR.Println("could not unmarshal message")
			return nil, err
		}
		jww.DEBUG.Printf("read message: %+v\n", readMessage)
		messages = append(messages, readMessage)
	}
	return messages, nil
}

func SaveSession(sessionName string, messages []openai.ChatCompletionMessage) error {
	// Validate session name
	if err := validateSession(sessionName); err != nil {
		return err
	}

	// Get path to file
	filepath, err := getFilepathForSession(sessionName)
	if err != nil {
		return err
	}

	// Check, if session exists
	var exists bool
	exists, err = fs.FileExists(filepath)
	if err != nil {
		return err
	}

	// Open file
	var file *os.File
	if exists {
		jww.DEBUG.Println("session already exists, opening file for truncation")
		// Open and truncate existing file
		file, err = os.OpenFile(filepath, os.O_WRONLY|os.O_TRUNC, defaultFilePermissions)
		if err != nil {
			jww.ERROR.Println("could not open file for truncation")
			return err
		}
	} else {
		jww.DEBUG.Println("session does not exist, creating file")
		// Create file
		file, err = os.Create(filepath)
		if err != nil {
			jww.ERROR.Println("could not create file")
			return err
		}
	}
	defer file.Close()

	// Save messages to file
	var data []byte
	for _, message := range messages {
		jww.DEBUG.Printf("writing following message to file: %+v\n", message)
		data, err = json.Marshal(message)
		if err != nil {
			jww.ERROR.Println("could not marshal message")
			return err
		}
		_, err = file.WriteString(string(data) + "\n")
		if err != nil {
			jww.ERROR.Println("could not write message to file")
			return err
		}
	}
	jww.DEBUG.Println("session saved")
	return nil
}

func ListSessions() ([]string, error) {
	dir, err := fs.GetAppCacheDir(appDirName)
	if err != nil {
		return nil, err
	}
	var dirFiles []os.DirEntry
	dirFiles, err = os.ReadDir(dir)
	if err != nil {
		jww.ERROR.Println("could not read directory")
		return nil, err
	}
	jww.DEBUG.Println("listing sessions")
	var files []string
	for _, file := range dirFiles {
		jww.DEBUG.Printf("found file: %s\n", file.Name())
		files = append(files, file.Name())
	}
	return files, nil
}

func DeleteSession(sessionName string) error {
	filepath, err := getFilepathForSession(sessionName)
	if err != nil {
		return err
	}
	var exists bool
	exists, err = fs.FileExists(filepath)
	if err != nil {
		return err
	}
	if !exists {
		jww.DEBUG.Println("session does not exist, skipping deletion")
		return nil
	}
	err = os.Remove(filepath)
	if err != nil {
		jww.ERROR.Println("could not remove file")
		return err
	}
	jww.DEBUG.Println("session deleted")
	return nil
}
