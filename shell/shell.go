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
