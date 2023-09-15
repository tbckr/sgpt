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
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
)

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

func TestGetUserConfirmationEnter(t *testing.T) {
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
