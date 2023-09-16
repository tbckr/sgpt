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

package cli

import (
	"io"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCompletionCmd(t *testing.T) {
	mem := &exitMemento{}

	config := createTestConfig(t)

	root := newRootCmd(mem.Exit, config, nil)
	root.cmd.SetOut(io.Discard)
	root.Execute([]string{"completion"})

	require.Equal(t, 0, mem.code)
}

func TestCompletionCmdBash(t *testing.T) {
	mem := &exitMemento{}

	config := createTestConfig(t)

	root := newRootCmd(mem.Exit, config, nil)
	root.cmd.SetOut(io.Discard)
	root.Execute([]string{"completion", "bash"})

	require.Equal(t, 0, mem.code)
}

func TestCompletionCmdFish(t *testing.T) {
	mem := &exitMemento{}

	config := createTestConfig(t)

	root := newRootCmd(mem.Exit, config, nil)
	root.cmd.SetOut(io.Discard)
	root.Execute([]string{"completion", "fish"})

	require.Equal(t, 0, mem.code)
}

func TestCompletionCmdPowershell(t *testing.T) {
	mem := &exitMemento{}

	config := createTestConfig(t)

	root := newRootCmd(mem.Exit, config, nil)
	root.cmd.SetOut(io.Discard)
	root.Execute([]string{"completion", "powershell"})

	require.Equal(t, 0, mem.code)
}

func TestCompletionCmdZsh(t *testing.T) {
	mem := &exitMemento{}

	config := createTestConfig(t)

	root := newRootCmd(mem.Exit, config, nil)
	root.cmd.SetOut(io.Discard)
	root.Execute([]string{"completion", "zsh"})

	require.Equal(t, 0, mem.code)
}

func TestCompletionCmdUnknownCompletion(t *testing.T) {
	mem := &exitMemento{}

	config := createTestConfig(t)

	root := newRootCmd(mem.Exit, config, nil)
	root.cmd.SetOut(io.Discard)
	root.Execute([]string{"completion", "abcd"})

	require.Equal(t, 0, mem.code)
}

func TestCompletionCmdTooManyArgs(t *testing.T) {
	mem := &exitMemento{}

	config := createTestConfig(t)

	root := newRootCmd(mem.Exit, config, nil)
	root.cmd.SetOut(io.Discard)
	root.Execute([]string{"completion", "abcd", "efgh"})

	require.Equal(t, 0, mem.code)
}
