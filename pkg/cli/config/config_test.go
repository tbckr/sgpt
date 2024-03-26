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

package config

import (
	"github.com/tbckr/sgpt/v2/pkg/cli"
	"github.com/tbckr/sgpt/v2/pkg/cli/root"
	"path/filepath"
	"testing"

	"github.com/tbckr/sgpt/v2/internal/testlib"

	"github.com/stretchr/testify/require"
)

func TestConfigCmd(t *testing.T) {
	testCtx := testlib.NewTestCtx(t)
	mem := &cli.exitMemento{}

	root := root.NewRootCmd(mem.Exit, testCtx.Config, nil, nil)

	root.Execute([]string{"config"})
	require.Equal(t, 0, mem.code)
}

func TestConfigCmdInit(t *testing.T) {
	testCtx := testlib.NewTestCtx(t)
	mem := &cli.exitMemento{}

	root.NewRootCmd(mem.Exit, testCtx.Config, nil, nil).Execute([]string{"config", "init"})
	require.Equal(t, 0, mem.code)

	require.FileExists(t, filepath.Join(testCtx.ConfigDir, "config.yaml"))
	// config must only contain values for model, maxtokens, temperature, topp
	require.NoError(t, testCtx.Config.ReadInConfig())
	// TESTING may be in the config, because this is a test
	require.Equal(t, 8, len(testCtx.Config.AllSettings()))
	for _, key := range []string{"model", "maxtokens", "temperature", "topp", "cachedir", "personas", "stream", "testing"} {
		require.Contains(t, testCtx.Config.AllSettings(), key)
	}
}

func TestConfigCmdInitAlreadyExists(t *testing.T) {
	testCtx := testlib.NewTestCtx(t)
	mem := &cli.exitMemento{}

	root.NewRootCmd(mem.Exit, testCtx.Config, nil, nil).Execute([]string{"config", "init"})
	require.Equal(t, 0, mem.code)

	require.FileExists(t, filepath.Join(testCtx.ConfigDir, "config.yaml"))

	root.NewRootCmd(mem.Exit, testCtx.Config, nil, nil).Execute([]string{"config", "init"})
	require.Equal(t, 1, mem.code)
}

func TestConfigCmdShowConfig(t *testing.T) {
	testCtx := testlib.NewTestCtx(t)
	mem := &cli.exitMemento{}

	root.NewRootCmd(mem.Exit, testCtx.Config, nil, nil).Execute([]string{"config", "init"})
	require.Equal(t, 0, mem.code)

	require.FileExists(t, filepath.Join(testCtx.ConfigDir, "config.yaml"))

	root.NewRootCmd(mem.Exit, testCtx.Config, nil, nil).Execute([]string{"config", "show"})
	require.Equal(t, 0, mem.code)
}

func TestConfigCmdShowConfigNonExistent(t *testing.T) {
	testCtx := testlib.NewTestCtx(t)
	mem := &cli.exitMemento{}

	require.NoFileExists(t, filepath.Join(testCtx.ConfigDir, "config.yaml"))

	root.NewRootCmd(mem.Exit, testCtx.Config, nil, nil).Execute([]string{"config", "show"})
	require.Equal(t, 1, mem.code)
}
