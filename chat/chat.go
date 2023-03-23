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

	"github.com/sashabaranov/go-openai"
	"github.com/tbckr/sgpt/filesystem"
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
	dir, err := filesystem.GetAppCacheDir(appDirName)
	if err != nil {
		return "", err
	}
	filePath := path.Join(dir, sessionName)
	return filePath, nil
}

func validateSession(sessionName string) error {
	if !sessionNameMatcher.Match([]byte(sessionName)) {
		return ErrChatSessionNameInvalid
	}
	if utf8.RuneCountInString(sessionName) > sessionNameMaxLength {
		return ErrChatSessionNameTooLong
	}
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
	return filesystem.FileExists(filepath)
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
	exists, err = filesystem.FileExists(filepath)
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
	exists, err = filesystem.FileExists(filepath)
	if err != nil {
		return err
	}

	// Open file
	var file *os.File
	if exists {
		// Open and truncate existing file
		file, err = os.OpenFile(filepath, os.O_WRONLY|os.O_TRUNC, defaultFilePermissions)
		if err != nil {
			return err
		}
	} else {
		// Create file
		file, err = os.Create(filepath)
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
	return nil
}

func ListSessions() ([]string, error) {
	dir, err := filesystem.GetAppCacheDir(appDirName)
	if err != nil {
		return nil, err
	}
	var dirFiles []os.DirEntry
	dirFiles, err = os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	var files []string
	for _, file := range dirFiles {
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
	exists, err = filesystem.FileExists(filepath)
	if err != nil {
		return err
	}
	if !exists {
		return nil
	}
	return os.Remove(filepath)
}
