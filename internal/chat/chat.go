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
)

const (
	appConfigDir         = "sgpt"
	dirPermissions       = 0750
	filePermissions      = 0755
	sessionNameMaxLength = 65
)

var (
	ErrChatSessionNotExist    = errors.New("chat session does not exist")
	ErrChatSessionNameInvalid = fmt.Errorf("chat session name does not match the regex %s", sessionNameRegex)
	ErrChatSessionNameTooLong = fmt.Errorf("chat session name is greater than %d", sessionNameMaxLength)
	sessionNameRegex          = "^([-a-zA-Z0-9]*[a-zA-Z0-9])?"
	sessionNameMatcher        = regexp.MustCompile(sessionNameRegex)
)

// TODO refactor functions: move to file package, rm duplicates

func getAppCacheDir() (string, error) {
	// Get user specific config dir
	baseConfigDir, err := os.UserCacheDir()
	if err != nil {
		return "", err
	}
	// Application specific cache dir
	configPath := path.Join(baseConfigDir, appConfigDir)
	_, err = os.Stat(configPath)
	// Check, if application cache dir exists
	if os.IsNotExist(err) {
		// Create application cache dir
		if err = os.Mkdir(configPath, dirPermissions); err != nil {
			return "", err
		}
	}
	return configPath, nil
}

func getFilepath(sessionName string) (string, error) {
	dir, err := getAppCacheDir()
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

func fileExists(filepath string) (bool, error) {
	fileInfo, err := os.Stat(filepath)
	if os.IsNotExist(err) {
		return false, nil
	} else if err != nil {
		return false, err
	}
	return fileInfo.Name() != "", nil
}

func SessionExists(sessionName string) (bool, error) {
	if err := validateSession(sessionName); err != nil {
		return false, err
	}
	file, err := getFilepath(sessionName)
	if err != nil {
		return false, err
	}
	return fileExists(file)
}

func GetSession(sessionName string) ([]openai.ChatCompletionMessage, error) {
	// Validate session name
	if err := validateSession(sessionName); err != nil {
		return nil, err
	}

	// Get path to file
	filepath, err := getFilepath(sessionName)
	if err != nil {
		return nil, err
	}

	// Check, if session exists
	var exists bool
	exists, err = fileExists(filepath)
	if err != nil {
		return nil, err
	}
	if !exists {
		return []openai.ChatCompletionMessage{}, ErrChatSessionNotExist
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
	filepath, err := getFilepath(sessionName)
	if err != nil {
		return err
	}

	// Check, if session exists
	var exists bool
	exists, err = fileExists(filepath)
	if err != nil {
		return err
	}

	// Open file
	var file *os.File
	if exists {
		// Open and truncate existing file
		file, err = os.OpenFile(filepath, os.O_RDWR|os.O_TRUNC, filePermissions)
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
	dir, err := getAppCacheDir()
	if err != nil {
		return nil, err
	}
	var files []os.DirEntry
	files, err = os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	var fileList []string
	for _, file := range files {
		fileList = append(fileList, file.Name())
	}
	return fileList, nil
}

func DeleteSession(sessionName string) error {
	filepath, err := getFilepath(sessionName)
	if err != nil {
		return err
	}
	var exists bool
	exists, err = fileExists(filepath)
	if err != nil {
		return err
	}
	if !exists {
		return nil
	}
	return os.Remove(filepath)
}
