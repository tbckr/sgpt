package shell

import (
	"bufio"
	"errors"
	"os"
)

var ErrMissingPrompt = errors.New("exactly one prompt must be provided")

func IsPipedShell() (bool, error) {
	fi, err := os.Stdin.Stat()
	if err != nil {
		return false, err
	}
	if fi.Mode()&os.ModeNamedPipe == 0 {
		return false, nil
	}
	return true, nil
}

func GetPipedData() (string, error) {
	var buf []byte
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		buf = append(buf, scanner.Bytes()...)
	}
	if err := scanner.Err(); err != nil {
		return "", err
	}
	return string(buf), nil
}

func GetPrompt(args []string) (string, error) {
	var prompt string
	// Check, if prompt was provided via stdin
	pipedShell, err := IsPipedShell()
	if err != nil {
		return "", err
	}
	if pipedShell {
		prompt, err = GetPipedData()
		if err != nil {
			return "", err
		}
	} else {
		// Check, if prompt was provided via command line
		if len(args) != 1 {
			return "", ErrMissingPrompt
		}
		prompt = args[0]
	}
	return prompt, nil
}
