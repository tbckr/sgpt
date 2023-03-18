package chat

import (
	"bufio"
	"encoding/json"
	"errors"
	"os"
	"path"

	"github.com/sashabaranov/go-openai"
)

const (
	appConfigDir    = "sgpt"
	dirPermissions  = 0750
	filePermissions = 0755
)

var ErrChatSessionNotExist = errors.New("chat session does not exist")

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

func validateSession(_ string) error {
	// TODO
	return nil
}

func fileExists(filepath string) (bool, error) {
	_, err := os.Stat(filepath)
	if os.IsExist(err) || os.IsNotExist(err) {
		return os.IsExist(err), nil
	}
	return false, err
}

func SessionExists(sessionName string) (bool, error) {
	if err := validateSession(sessionName); err != nil {
		return false, err
	}
	file, err := getFilepath(sessionName)
	if err != nil {
		return false, err
	}
	_, err = os.Stat(file)
	return os.IsExist(err), nil
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
	// TODO
	return nil, nil
}

func DeleteSession(sessionName string) error {
	// TODO
	return nil
}
