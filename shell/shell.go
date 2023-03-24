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

	jww "github.com/spf13/jwalterweatherman"
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
		jww.DEBUG.Println("Input via stdin was provided")
	} else {
		// Check, if prompt was provided via command line
		if len(args) != 1 {
			jww.ERROR.Println("No input via args was provided")
			return "", ErrMissingInput
		}
		prompt = args[0]
		jww.DEBUG.Println("Input via args was provided")
	}
	return prompt, nil
}

func IsPipedShell() (bool, error) {
	fi, err := os.Stdin.Stat()
	if err != nil {
		jww.ERROR.Println("Could not get stdin info")
		return false, err
	}
	if fi.Mode()&os.ModeNamedPipe == 0 {
		jww.DEBUG.Println("Input via stdin was not provided")
		return false, nil
	}
	jww.DEBUG.Println("Input via stdin was provided")
	return true, nil
}

func GetPipedData() (string, error) {
	var buf []byte
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		buf = append(buf, scanner.Bytes()...)
	}
	if err := scanner.Err(); err != nil {
		jww.ERROR.Println("Could not read data from stdin")
		return "", err
	}
	input := string(buf)
	jww.DEBUG.Printf("Input via stdin was: %s", input)
	return input, nil
}

func ExecuteCommandWithConfirmation(ctx context.Context, bashCommand string) error {
	// Print out the command to be executed
	if _, err := fmt.Fprintln(os.Stdout, bashCommand); err != nil {
		jww.ERROR.Println("Could not print command to be executed")
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
		jww.DEBUG.Println("Waiting for user confirmation")
		if _, err := fmt.Fprint(os.Stdout, "Do you want to execute this command? (Y/n) "); err != nil {
			return false, err
		}
		reader := bufio.NewReader(os.Stdin)
		char, _, err := reader.ReadRune()
		if err != nil {
			jww.ERROR.Println("Could not read user input for confirmation")
			return false, err
		}
		// 10 = enter
		if char == 10 || char == 'Y' || char == 'y' {
			jww.DEBUG.Println("User confirmed execution")
			return true, nil
		} else if char == 'N' || char == 'n' {
			jww.DEBUG.Println("User denied execution")
			return false, nil
		}
	}
}

func ExecuteShellCommand(ctx context.Context, response string) error {
	jww.DEBUG.Println("Executing command: ", response)
	// Execute cmd from response text
	cmd := exec.CommandContext(ctx, "bash", "-c", response)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		jww.ERROR.Println("Error while executing command: ", err)
		return err
	}
	return nil
}
