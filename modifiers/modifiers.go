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
	"os"
	"runtime"
	"strings"
	"text/template"
)

const (
	envKeyShell = "SHELL"
)

// TODO: read custom modifiers from os app config dir

var (
	ErrUnsupportedModifier = errors.New("unsupported modifier")
)

func GetChatModifier(modifier string) (string, error) {
	loadedPrompts, err := loadDefaultPrompts()
	if err != nil {
		return "", err
	}
	switch modifier {
	case "sh":
		return renderPrompt(loadedPrompts[modifier].Messages[0].Text)
	case "code":
		return renderPrompt(loadedPrompts[modifier].Messages[0].Text)
	case "txt":
		return "", nil
	default:
		return "", ErrUnsupportedModifier
	}
}

func renderPrompt(promptText string) (string, error) {
	t, err := template.New("prompt").Parse(promptText)
	if err != nil {
		return "", err
	}

	var vars map[string]string
	var renderedPrompt strings.Builder

	vars, err = loadVariables()
	if err != nil {
		return "", err
	}

	err = t.Execute(&renderedPrompt, vars)
	if err != nil {
		return "", err
	}
	return renderedPrompt.String(), nil
}

func loadVariables() (map[string]string, error) {
	operatingSystem := runtime.GOOS
	shell, ok := os.LookupEnv(envKeyShell)
	if !ok {
		return map[string]string{}, errors.New("could not determine shell")
	}

	return map[string]string{
		"Shell": shell,
		"Os":    operatingSystem,
	}, nil
}
