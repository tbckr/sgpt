package shell

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
)

var ErrMissingInput = errors.New("no input was provided")

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

func GetInput(args []string) (string, error) {
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
			return "", ErrMissingInput
		}
		prompt = args[0]
	}
	return prompt, nil
}

func ExecuteCommandWithConfirmation(ctx context.Context, bashCommand string) error {
	// Print out the command to be executed
	if _, err := fmt.Fprintln(os.Stdout, bashCommand); err != nil {
		return err
	}
	// Require user confirmation
	ok, err := GetUserConfirmation()
	if err != nil {
		return err
	}
	if ok {
		return ExecuteShellCommand(ctx, bashCommand)
	}
	return nil
}

func GetUserConfirmation() (bool, error) {
	for {
		if _, err := fmt.Fprint(os.Stdout, "Do you want to execute this command? (Y/n) "); err != nil {
			return false, err
		}
		reader := bufio.NewReader(os.Stdin)
		char, _, err := reader.ReadRune()
		if err != nil {
			return false, err
		}
		// 10 = enter
		if char == 10 || char == 'Y' || char == 'y' {
			return true, nil
		} else if char == 'N' || char == 'n' {
			return false, nil
		}
	}
}

func ExecuteShellCommand(ctx context.Context, response string) error {
	// Execute cmd from response text
	cmd := exec.CommandContext(ctx, "bash", "-c", response)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
