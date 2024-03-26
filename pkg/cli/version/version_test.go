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

package version

import (
	"bytes"
	"github.com/tbckr/sgpt/v2/pkg/cli"
	root2 "github.com/tbckr/sgpt/v2/pkg/cli/root"
	"io"
	"strings"
	"testing"

	"github.com/tbckr/sgpt/v2/internal/testlib"

	"github.com/stretchr/testify/require"
)

func TestVersionCmd(t *testing.T) {
	testCtx := testlib.NewTestCtx(t)
	mem := &cli.exitMemento{}

	root := root2.NewRootCmd(mem.Exit, testCtx.Config, cli.mockIsPipedShell(false, nil), nil)
	cmd := root.cmd

	outBytes := bytes.NewBufferString("")
	cmd.SetOut(outBytes)

	root.Execute([]string{"version"})
	require.Equal(t, 0, mem.code)

	out, err := io.ReadAll(outBytes)
	if err != nil {
		t.Fatal(err)
	}
	printedVersion := string(out)
	require.NotEmpty(t, printedVersion)
	require.True(t, strings.HasPrefix(printedVersion, "v"))
}

func TestVersionCmdFull(t *testing.T) {
	testCtx := testlib.NewTestCtx(t)
	mem := &cli.exitMemento{}

	root := root2.NewRootCmd(mem.Exit, testCtx.Config, cli.mockIsPipedShell(false, nil), nil)
	cmd := root.cmd

	outBytes := bytes.NewBufferString("")
	cmd.SetOut(outBytes)

	root.Execute([]string{"version", "--full"})
	require.Equal(t, 0, mem.code)

	out, err := io.ReadAll(outBytes)
	if err != nil {
		t.Fatal(err)
	}
	printedVersion := string(out)
	require.NotEmpty(t, printedVersion)
	require.True(t, strings.HasPrefix(printedVersion, "version: "))
}

func TestVersionCmdUnknowArg(t *testing.T) {
	testCtx := testlib.NewTestCtx(t)
	mem := &cli.exitMemento{}

	root2.NewRootCmd(mem.Exit, testCtx.Config, cli.mockIsPipedShell(false, nil), nil).Execute([]string{"version", "abcd"})
	require.Equal(t, 1, mem.code)
}
