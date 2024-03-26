// Copyright (c) 2024 Tim <tbckr>
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

package testlib

import (
	"github.com/stretchr/testify/require"
	"io"
	"sync"
	"testing"

	"github.com/spf13/viper"
	"github.com/tbckr/sgpt/v2/pkg/cli/root"
)

type CommandExecutor struct {
	cacheDir  string
	configDir string
	envVars   map[string]string

	config *viper.Viper

	args         []string
	prompt       string
	modifier     string
	isPipedShell bool
	stdinInput   string
}

type CommandResult struct {
	Stdout   string
	Stderr   string
	ExitCode int
}

func NewCommandExecutor(t *testing.T) *CommandExecutor {
	cacheDir := t.TempDir()
	configDir := t.TempDir()

	config := viper.New()
	config.AddConfigPath(configDir)
	config.SetConfigName("config")
	config.SetConfigType("yaml")

	return &CommandExecutor{
		cacheDir:  cacheDir,
		configDir: configDir,
		envVars:   make(map[string]string),

		config: config,
	}
}

func (c *CommandExecutor) WithEnvVar(key, value string) *CommandExecutor {
	c.envVars[key] = value
	return c
}

func (c *CommandExecutor) WithArgs(args ...string) *CommandExecutor {
	c.args = args
	return c
}

func (c *CommandExecutor) WithPrompt(prompt string) *CommandExecutor {
	c.prompt = prompt
	return c
}

func (c *CommandExecutor) WithModifier(modifier string) *CommandExecutor {
	c.modifier = modifier
	return c
}

func (c *CommandExecutor) WithPipedShell(isPipedShell bool) *CommandExecutor {
	c.isPipedShell = isPipedShell
	return c
}

func (c *CommandExecutor) WithStdinInput(stdinInput string) *CommandExecutor {
	c.stdinInput = stdinInput
	return c
}

func (c *CommandExecutor) Run(t *testing.T) (int, error) {
	var wg sync.WaitGroup
	stdinReader, stdinWriter := io.Pipe()
	stdoutReader, stdoutWriter := io.Pipe()
	stderrReader, stderrWriter := io.Pipe()

	cmd, err := root.NewRootCmd(
		stdinReader,
		stdoutWriter,
		stderrWriter,
		func(key string) (string, bool) {
			value, ok := c.envVars[key]
			return value, ok
		},
		func() (string, error) {
			return c.configDir, nil
		},
		func() (string, error) {
			return c.cacheDir, nil
		},
		c.config,
		func() (bool, error) {
			return c.isPipedShell, nil
		},
		nil,
	)
	require.NoError(t, err)

	var args []string
	if len(c.args) > 0 {
		args = c.args
	} else {
		args = []string{c.modifier, c.prompt}
	}

	returnCode, err := cmd.Execute(args)
	require.NoError(t, err)

	return 0, nil
}
