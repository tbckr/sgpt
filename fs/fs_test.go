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

package fs

import (
	"io"
	"os"
	"path/filepath"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGetAppCacheDir(t *testing.T) {
	existingEnv := os.Getenv("XDG_CACHE_HOME")
	defer require.NoError(t, os.Setenv("XDG_CACHE_HOME", existingEnv))

	tempDir, err := createTempDir()
	require.NoError(t, err)
	defer require.NoError(t, removeTempDir(tempDir))

	err = os.Setenv("XDG_CACHE_HOME", tempDir)
	require.NoError(t, err)

	var cacheDir string
	cacheDir, err = GetAppCacheDir()
	require.NoError(t, err)
	require.Equal(t, filepath.Join(tempDir, "sgpt"), cacheDir)
	require.DirExists(t, cacheDir)
}

func TestGetAppConfigDir(t *testing.T) {
	existingEnv := os.Getenv("XDG_CONFIG_HOME")
	defer require.NoError(t, os.Setenv("XDG_CONFIG_HOME", existingEnv))

	tempDir, err := createTempDir()
	require.NoError(t, err)
	defer require.NoError(t, removeTempDir(tempDir))

	err = os.Setenv("XDG_CONFIG_HOME", tempDir)
	require.NoError(t, err)

	var cacheDir string
	cacheDir, err = GetAppConfigPath()
	require.NoError(t, err)
	require.Equal(t, filepath.Join(tempDir, "sgpt"), cacheDir)
	require.DirExists(t, cacheDir)
}

func TestReadString(t *testing.T) {
	reader, writer := io.Pipe()
	defer require.NoError(t, reader.Close())
	defer require.NoError(t, writer.Close())

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		_, errWrite := writer.Write([]byte("Hello World"))
		require.NoError(t, writer.Close())
		require.NoError(t, errWrite)
	}()

	out, err := ReadString(reader)
	require.NoError(t, err)
	require.Equal(t, "Hello World", out)

	wg.Wait()
}

func createTempDir() (string, error) {
	return os.MkdirTemp("", "sgpt_temp_*")
}

func removeTempDir(fullConfigPath string) error {
	return os.RemoveAll(fullConfigPath)
}
