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
	"bytes"
	"io"
	"strings"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/tbckr/sgpt/api"
)

func TestLicensesCmd(t *testing.T) {
	mem := &exitMemento{}
	expected := `To see the open source packages included in SGPT and
their respective license information, visit:
`
	var wg sync.WaitGroup
	reader, writer := io.Pipe()

	config := createTestConfig(t)

	root := newRootCmd(mem.Exit, config, api.MockClient("", nil))
	root.cmd.SetOut(writer)

	wg.Add(1)
	go func() {
		defer wg.Done()
		var buf bytes.Buffer
		_, err := io.Copy(&buf, reader)
		require.NoError(t, err)
		require.NoError(t, reader.Close())
		require.True(t, strings.HasPrefix(buf.String(), expected))
	}()

	root.Execute([]string{"licenses"})
	require.Equal(t, 0, mem.code)
	require.NoError(t, writer.Close())

	wg.Wait()
}
