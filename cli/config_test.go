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
	"path/filepath"
	"testing"

	"github.com/spf13/viper"

	"github.com/stretchr/testify/require"
	"github.com/tbckr/sgpt/v2/api"
)

func TestConfigCmd(t *testing.T) {
	mem := &exitMemento{}

	config := createTestConfig(t)

	root := newRootCmd(mem.Exit, config, api.MockClient("", nil))

	root.Execute([]string{"config"})
	require.Equal(t, 0, mem.code)
}

func TestConfigCmdInit(t *testing.T) {
	mem := &exitMemento{}

	configDir := createTempDir(t, "config")

	config := viper.New()
	config.AddConfigPath(configDir)
	config.SetConfigName("config")
	config.SetConfigType("yaml")
	config.Set("TESTING", 1)

	newRootCmd(mem.Exit, config, api.MockClient("", nil)).Execute([]string{"config", "init"})
	require.Equal(t, 0, mem.code)

	require.FileExists(t, filepath.Join(configDir, "config.yaml"))
}

func TestConfigCmdInitAlreadyExists(t *testing.T) {
	mem := &exitMemento{}

	configDir := createTempDir(t, "config")

	config := viper.New()
	config.AddConfigPath(configDir)
	config.SetConfigName("config")
	config.SetConfigType("yaml")
	config.Set("TESTING", 1)

	newRootCmd(mem.Exit, config, api.MockClient("", nil)).Execute([]string{"config", "init"})
	require.Equal(t, 0, mem.code)

	require.FileExists(t, filepath.Join(configDir, "config.yaml"))

	newRootCmd(mem.Exit, config, api.MockClient("", nil)).Execute([]string{"config", "init"})
	require.Equal(t, 1, mem.code)
}

func TestConfigCmdShowConfig(t *testing.T) {
	mem := &exitMemento{}

	configDir := createTempDir(t, "config")

	config := viper.New()
	config.AddConfigPath(configDir)
	config.SetConfigName("config")
	config.SetConfigType("yaml")
	config.Set("TESTING", 1)
	require.NoError(t, setViperDefaults(config))

	newRootCmd(mem.Exit, config, api.MockClient("", nil)).Execute([]string{"config", "init"})
	require.Equal(t, 0, mem.code)

	require.FileExists(t, filepath.Join(configDir, "config.yaml"))

	newRootCmd(mem.Exit, config, api.MockClient("", nil)).Execute([]string{"config", "show"})
	require.Equal(t, 0, mem.code)
}

func TestConfigCmdShowConfigNonExistent(t *testing.T) {
	mem := &exitMemento{}

	configDir := createTempDir(t, "config")

	config := viper.New()
	config.AddConfigPath(configDir)
	config.SetConfigName("config")
	config.SetConfigType("yaml")
	config.Set("TESTING", 1)

	require.NoFileExists(t, filepath.Join(configDir, "config.yaml"))

	newRootCmd(mem.Exit, config, api.MockClient("", nil)).Execute([]string{"config", "show"})
	require.Equal(t, 1, mem.code)
}
