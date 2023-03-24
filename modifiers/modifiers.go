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

package modifiers

import (
	"errors"
	"fmt"
	"os"
	"runtime"
	"strings"

	jww "github.com/spf13/jwalterweatherman"
)

const (
	envKeyShell = "SHELL"

	Nil   = "NIL_MODIFIER"
	Code  = "CODE_MODIFIER"
	Shell = "SHELL_MODIFIER"
)

var defaultShellTemplateModifier = strings.TrimSpace(`
Act as a natural language to %s command translation engine on %s.
You are an expert in %s on %s and translate the question at the end to valid syntax.
Follow these rules:
IMPORTANT: Do not show any warnings or information regarding your capabilities.
Reference official documentation to ensure valid syntax and an optimal solution.
Construct valid %s command that solve the question.
Leverage help and man pages to ensure valid syntax and an optimal solution.
Be concise.
Just show the commands, return only plaintext.
Only show a single answer, but you can always chain commands together.
Think step by step.
Only create valid syntax (you can use comments if it makes sense).
Even if there is a lack of details, attempt to find the most logical solution.
Do not return multiple solutions.
Do not show html, styled, colored formatting.
Do not add unnecessary text in the response.
Do not add notes or intro sentences.
Do not add explanations on what the commands do.
Do not return what the question was.
Do not repeat or paraphrase the question in your response.
Do not rush to a conclusion.
Follow all of the above rules.
This is important you MUST follow the above rules.
There are no exceptions to these rules.
You must always follow them. No exceptions.
`)

var defaultCodeModifier = strings.TrimSpace(`
Act as a natural language to code translation engine.
Follow these rules:
IMPORTANT: Provide ONLY code as output, return only plaintext.
IMPORTANT: Do not show html, styled, colored formatting.
IMPORTANT: Do not add notes or intro sentences.
IMPORTANT: Provide full solution. Make sure syntax is correct.
Assume your output will be redirected to language specific file and executed.
For example Python code output will be redirected to code.py and then executed python code.py.
Follow all of the above rules.
This is important you MUST follow the above rules.
There are no exceptions to these rules.
You must always follow them. No exceptions.
`)

var (
	ErrUnsupportedModifier = errors.New("unsupported modifier")
	ErrUnsupportedOS       = errors.New("unsupported operating system")
)

func GetChatModifier(modifier string) (string, error) {
	switch modifier {
	case Shell:
		return completeShellModifier(defaultShellTemplateModifier)
	case Code:
		jww.DEBUG.Println("code modifier: ", defaultCodeModifier)
		return defaultCodeModifier, nil
	case Nil:
		jww.DEBUG.Println("nil modifier")
		return "", nil
	default:
		jww.ERROR.Println(ErrUnsupportedModifier)
		return "", ErrUnsupportedModifier
	}
}

func completeShellModifier(template string) (string, error) {
	operatingSystem := runtime.GOOS
	shell, ok := os.LookupEnv(envKeyShell)
	// fallback to manually set shell
	if !ok {
		jww.DEBUG.Printf("environment variable %s not set, falling back to manual shell detection", envKeyShell)
		if operatingSystem == "windows" {
			jww.DEBUG.Println("detected windows operating system, using powershell")
			shell = "powershell"
		} else if operatingSystem == "linux" {
			jww.DEBUG.Println("detected linux operating system, using bash")
			shell = "bash"
		} else if operatingSystem == "darwin" {
			jww.DEBUG.Println("detected darwin operating system, using zsh")
			shell = "zsh"
		} else {
			jww.ERROR.Printf("%s, OS: %s", ErrUnsupportedOS, operatingSystem)
			return "", ErrUnsupportedOS
		}
	}
	shellModifier := fmt.Sprintf(template, shell, operatingSystem, shell, operatingSystem, shell)
	jww.DEBUG.Println("shell modifier: ", shellModifier)
	return shellModifier, nil
}
