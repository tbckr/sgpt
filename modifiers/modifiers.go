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
	"log/slog"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"text/template"

	"github.com/spf13/viper"

	"github.com/tbckr/sgpt/v2/fs"
)

const (
	envKeyShell = "SHELL"
)

var (
	personaNameRegex   = "^[a-zA-Z0-9-_]+$"
	personaNameMatcher = regexp.MustCompile(personaNameRegex)

	ErrUnsupportedModifier = errors.New("unsupported modifier")
)

func GetChatModifier(config *viper.Viper, modifier string) (string, error) {
	persona, err := getPersonasModifier(config, modifier)
	if err != nil {
		return "", err
	}
	// if a persona is found, render the prompt
	// this overrides the default personas
	if persona != "" {
		return renderPrompt(persona)
	}
	// if no persona is found, try to load the default prompt
	var loadedDefaultPrompts map[string]Prompt
	loadedDefaultPrompts, err = loadDefaultPrompts()
	if err != nil {
		return "", err
	}
	switch modifier {
	case "sh":
		return renderPrompt(loadedDefaultPrompts[modifier].Messages[0].Text)
	case "code":
		return renderPrompt(loadedDefaultPrompts[modifier].Messages[0].Text)
	case "txt":
		return "", nil
	default:
		return "", ErrUnsupportedModifier
	}
}

func getPersonasModifier(config *viper.Viper, modifier string) (string, error) {
	var err error

	var personasPath string
	if config.IsSet("personas") {
		personasPath = config.GetString("personas")
	} else {
		personasPath, err = fs.GetPersonasPath()
		if err != nil {
			return "", err
		}
	}

	var personaPath string
	err = filepath.WalkDir(personasPath, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if personasPath == path {
			return nil
		}
		if d.IsDir() {
			return filepath.SkipDir
		}
		if !personaNameMatcher.MatchString(d.Name()) {
			return nil
		}
		if modifier == d.Name() {
			personaPath = path
			return filepath.SkipAll
		}
		return nil
	})
	if err != nil {
		return "", err
	}

	if personaPath == "" {
		slog.Debug("could not find custom persona")
		return "", nil
	}

	var data []byte
	data, err = os.ReadFile(personaPath)
	if err != nil {
		return "", err
	}
	return string(data), nil
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
