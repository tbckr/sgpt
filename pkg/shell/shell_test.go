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
	"bytes"
	"context"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestIsPipedShell(t *testing.T) {
	// Create a mock Stdin that is a pipe
	reader, writer, _ := os.Pipe()

	existingStdin := os.Stdin
	t.Cleanup(func() {
		os.Stdin = existingStdin
	})

	os.Stdin = reader
	require.NoError(t, writer.Close())

	isPiped, err := IsPipedShell()
	require.NoError(t, err)
	require.True(t, isPiped)
	require.NoError(t, os.Stdin.Close())
}

func TestIsPipedShell_NotPiped(t *testing.T) {
	existingStdin := os.Stdin
	t.Cleanup(func() {
		os.Stdin = existingStdin
	})

	tempDir := t.TempDir()

	testIn, err := os.Create(filepath.Join(tempDir, "pipe"))
	require.NoError(t, err)

	os.Stdin = testIn

	var isPiped bool
	isPiped, err = IsPipedShell()
	require.NoError(t, err)
	require.False(t, isPiped)
	require.NoError(t, os.Stdin.Close())
}

func TestGetUserConfirmationYes(t *testing.T) {
	stdinReader, stdinWriter := io.Pipe()
	stdoutReader, stdoutWriter := io.Pipe()

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		_, errWrite := stdinWriter.Write([]byte("Y\n"))
		require.NoError(t, stdinWriter.Close())
		require.NoError(t, errWrite)
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		var buf bytes.Buffer
		_, errRead := io.Copy(&buf, stdoutReader)
		require.NoError(t, stdoutReader.Close())
		require.NoError(t, errRead)
		require.Equal(t, "Do you want to execute this command? (Y/n) ", buf.String())
	}()

	var ok bool
	var err error

	ok, err = getUserConfirmation(stdinReader, stdoutWriter)

	require.NoError(t, stdinReader.Close())
	require.NoError(t, stdoutWriter.Close())

	require.NoError(t, err)
	require.Equal(t, true, ok)

	wg.Wait()
}

func TestGetUserConfirmationLinuxEnter(t *testing.T) {
	stdinReader, stdinWriter := io.Pipe()
	stdoutReader, stdoutWriter := io.Pipe()

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		_, errWrite := stdinWriter.Write([]byte("\n"))
		require.NoError(t, stdinWriter.Close())
		require.NoError(t, errWrite)
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		var buf bytes.Buffer
		_, errRead := io.Copy(&buf, stdoutReader)
		require.NoError(t, stdoutReader.Close())
		require.NoError(t, errRead)
		require.Equal(t, "Do you want to execute this command? (Y/n) ", buf.String())
	}()

	var ok bool
	var err error

	ok, err = getUserConfirmation(stdinReader, stdoutWriter)

	require.NoError(t, stdinReader.Close())
	require.NoError(t, stdoutWriter.Close())

	require.NoError(t, err)
	require.Equal(t, true, ok)

	wg.Wait()
}

func TestGetUserConfirmationWindowsEnter(t *testing.T) {
	stdinReader, stdinWriter := io.Pipe()
	stdoutReader, stdoutWriter := io.Pipe()

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		_, errWrite := stdinWriter.Write([]byte("\r"))
		require.NoError(t, stdinWriter.Close())
		require.NoError(t, errWrite)
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		var buf bytes.Buffer
		_, errRead := io.Copy(&buf, stdoutReader)
		require.NoError(t, stdoutReader.Close())
		require.NoError(t, errRead)
		require.Equal(t, "Do you want to execute this command? (Y/n) ", buf.String())
	}()

	var ok bool
	var err error

	ok, err = getUserConfirmation(stdinReader, stdoutWriter)

	require.NoError(t, stdinReader.Close())
	require.NoError(t, stdoutWriter.Close())

	require.NoError(t, err)
	require.Equal(t, true, ok)

	wg.Wait()
}

func TestGetUserConfirmationNo(t *testing.T) {
	stdinReader, stdinWriter := io.Pipe()
	stdoutReader, stdoutWriter := io.Pipe()

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		_, errWrite := stdinWriter.Write([]byte("n\n"))
		require.NoError(t, stdinWriter.Close())
		require.NoError(t, errWrite)
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		var buf bytes.Buffer
		_, errRead := io.Copy(&buf, stdoutReader)
		require.NoError(t, stdoutReader.Close())
		require.NoError(t, errRead)
		require.Equal(t, "Do you want to execute this command? (Y/n) ", buf.String())
	}()

	var ok bool
	var err error

	ok, err = getUserConfirmation(stdinReader, stdoutWriter)

	require.NoError(t, stdinReader.Close())
	require.NoError(t, stdoutWriter.Close())

	require.NoError(t, err)
	require.Equal(t, false, ok)

	wg.Wait()
}

func TestGetUserConfirmationAsked2Times(t *testing.T) {
	stdinReader, stdinWriter := io.Pipe()
	stdoutReader, stdoutWriter := io.Pipe()

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		_, errWrite := stdinWriter.Write([]byte("abcd\n"))
		require.NoError(t, errWrite)
		_, errWrite = stdinWriter.Write([]byte("Y\n"))
		require.NoError(t, errWrite)
		require.NoError(t, stdinWriter.Close())
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		var buf bytes.Buffer
		_, errRead := io.Copy(&buf, stdoutReader)
		require.NoError(t, stdoutReader.Close())
		require.NoError(t, errRead)
		require.Equal(t, "Do you want to execute this command? (Y/n) Do you want to execute this command? (Y/n) ", buf.String())
	}()

	var ok bool
	var err error

	ok, err = getUserConfirmation(stdinReader, stdoutWriter)

	require.NoError(t, stdinReader.Close())
	require.NoError(t, stdoutWriter.Close())

	require.NoError(t, err)
	require.Equal(t, true, ok)

	wg.Wait()
}

func TestExecuteShellCommandEcho(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("shell execution tests require bash and are not supported on Windows")
	}
	stdoutReader, stdoutWriter := io.Pipe()

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		var buf bytes.Buffer
		_, errRead := io.Copy(&buf, stdoutReader)
		require.NoError(t, errRead)
		require.NoError(t, stdoutReader.Close())
		require.Equal(t, "test\n", buf.String())
	}()

	err := executeShellCommand(context.Background(), stdoutWriter, "echo test")

	require.NoError(t, stdoutWriter.Close())

	require.NoError(t, err)

	wg.Wait()
}

func TestExecuteCommandWithConfirmation(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("shell execution tests require bash and are not supported on Windows")
	}
	stdinReader, stdinWriter := io.Pipe()
	stdoutReader, stdoutWriter := io.Pipe()

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		_, errWrite := stdinWriter.Write([]byte("Y\n"))
		require.NoError(t, errWrite)
		require.NoError(t, stdinWriter.Close())
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		var buf bytes.Buffer
		_, errRead := io.Copy(&buf, stdoutReader)
		require.NoError(t, errRead)
		require.NoError(t, stdoutReader.Close())
		require.Equal(t, "Do you want to execute this command? (Y/n) test\n", buf.String())
	}()

	require.NoError(t, ExecuteCommandWithConfirmation(context.Background(), stdinReader, stdoutWriter, "echo test"))

	require.NoError(t, stdinReader.Close())
	require.NoError(t, stdoutWriter.Close())

	wg.Wait()
}

func TestSanitizeCommand_RemovesBidiOverrides(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "LRM (U+200E)",
			input:    "echo\u200Etest",
			expected: "echotest",
		},
		{
			name:     "RLM (U+200F)",
			input:    "echo\u200Ftest",
			expected: "echotest",
		},
		{
			name:     "LRE (U+202A)",
			input:    "echo\u202Atest",
			expected: "echotest",
		},
		{
			name:     "RLE (U+202B)",
			input:    "echo\u202Btest",
			expected: "echotest",
		},
		{
			name:     "PDF (U+202C)",
			input:    "echo\u202Ctest",
			expected: "echotest",
		},
		{
			name:     "LRO (U+202D)",
			input:    "echo\u202Dtest",
			expected: "echotest",
		},
		{
			name:     "RLO (U+202E)",
			input:    "echo\u202Etest",
			expected: "echotest",
		},
		{
			name:     "LRI (U+2066)",
			input:    "echo\u2066test",
			expected: "echotest",
		},
		{
			name:     "RLI (U+2067)",
			input:    "echo\u2067test",
			expected: "echotest",
		},
		{
			name:     "FSI (U+2068)",
			input:    "echo\u2068test",
			expected: "echotest",
		},
		{
			name:     "PDI (U+2069)",
			input:    "echo\u2069test",
			expected: "echotest",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := SanitizeCommand(tt.input)
			require.NoError(t, err)
			require.Equal(t, tt.expected, result)
		})
	}
}

func TestSanitizeCommand_RemovesANSIEscapes(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "CSI reset",
			input:    "echo \x1b[0mtest",
			expected: "echo test",
		},
		{
			name:     "CSI color",
			input:    "echo \x1b[31mred\x1b[0m",
			expected: "echo red",
		},
		{
			name:     "CSI bold",
			input:    "\x1b[1mbold\x1b[0m text",
			expected: "bold text",
		},
		{
			name:     "OSC sequence",
			input:    "echo \x1b]0;title\x07test",
			expected: "echo test",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := SanitizeCommand(tt.input)
			require.NoError(t, err)
			require.Equal(t, tt.expected, result)
		})
	}
}

func TestSanitizeCommand_RemovesControlChars(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "NULL byte",
			input:    "echo\x00test",
			expected: "echotest",
		},
		{
			name:     "Bell",
			input:    "echo\x07test",
			expected: "echotest",
		},
		{
			name:     "Backspace",
			input:    "echo\x08test",
			expected: "echotest",
		},
		{
			name:     "Vertical tab",
			input:    "echo\x0btest",
			expected: "echotest",
		},
		{
			name:     "Form feed",
			input:    "echo\x0ctest",
			expected: "echotest",
		},
		{
			name:     "Escape",
			input:    "echo\x1btest",
			expected: "echotest",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := SanitizeCommand(tt.input)
			require.NoError(t, err)
			require.Equal(t, tt.expected, result)
		})
	}
}

func TestSanitizeCommand_RejectsMultilineInjection(t *testing.T) {
	// A prompt-injected `ls\nrm -rf ~` would execute both commands under
	// `bash -c` once the user confirms. SanitizeCommand must refuse any
	// command containing \n or \r so the caller can abort before the
	// confirmation prompt (#360).
	tests := []string{
		"echo\ntest",
		"ls\nrm -rf ~",
		"echo hi\r\necho bye",
		"echo\r",
	}
	for _, cmd := range tests {
		t.Run(cmd, func(t *testing.T) {
			result, err := SanitizeCommand(cmd)
			require.ErrorIs(t, err, ErrMultilineCommand)
			require.Empty(t, result)
		})
	}
}

func TestSanitizeCommand_PreservesTab(t *testing.T) {
	// Tabs are a legitimate part of shell invocations (printf args etc)
	// and must pass through.
	result, err := SanitizeCommand("printf 'a\\tb'")
	require.NoError(t, err)
	require.Equal(t, "printf 'a\\tb'", result)

	result, err = SanitizeCommand("echo\ttest")
	require.NoError(t, err)
	require.Equal(t, "echo\ttest", result)
}

func TestSanitizeCommand_CleanCommandUnchanged(t *testing.T) {
	cleanCommands := []string{
		"echo hello",
		"ls -la",
		"git status",
		"rm -rf /tmp/test",
		"cat file.txt | grep pattern",
	}

	for _, cmd := range cleanCommands {
		t.Run(cmd, func(t *testing.T) {
			result, err := SanitizeCommand(cmd)
			require.NoError(t, err)
			require.Equal(t, cmd, result)
		})
	}
}

func TestSanitizeCommand_ZeroWidthChars(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Zero-width no-break space (U+FEFF)",
			input:    "echo\uFEFFtest",
			expected: "echotest",
		},
		{
			name:     "Zero-width space (U+200B)",
			input:    "echo\u200Btest",
			expected: "echotest",
		},
		{
			name:     "Zero-width non-joiner (U+200C)",
			input:    "echo\u200Ctest",
			expected: "echotest",
		},
		{
			name:     "Zero-width joiner (U+200D)",
			input:    "echo\u200Dtest",
			expected: "echotest",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := SanitizeCommand(tt.input)
			require.NoError(t, err)
			require.Equal(t, tt.expected, result)
		})
	}
}
