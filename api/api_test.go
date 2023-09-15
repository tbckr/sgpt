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

package api

import (
	"context"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/spf13/viper"
)

type MockClient struct {
	fn func(ctx context.Context, config *viper.Viper, prompt string, mode string) (string, error)
}

func NewMockClient(fn func(ctx context.Context, config *viper.Viper, prompt string, mode string) (string, error)) (Client, error) {
	return MockClient{
		fn: fn,
	}, nil
}

func (c MockClient) GetChatCompletion(ctx context.Context, config *viper.Viper, prompt string, mode string) (string, error) {
	return c.fn(ctx, config, prompt, mode)
}

func createTestConfig(t *testing.T) *viper.Viper {
	cacheDir, err := createTempDir("cache")
	require.NoError(t, err)

	var configDir string
	configDir, err = createTempDir("config")
	require.NoError(t, err)

	config := viper.New()
	config.AddConfigPath(configDir)
	config.SetConfigFile("config")
	config.SetConfigType("yaml")
	config.Set("cacheDir", cacheDir)

	return config
}

func teardownTestDirs(t *testing.T, config *viper.Viper) {
	cacheDir := config.GetString("cacheDir")
	require.NoError(t, removeTempDir(cacheDir))

	configDir := config.ConfigFileUsed()
	require.NoError(t, removeTempDir(configDir))
}

func createTempDir(suffix string) (string, error) {
	if suffix != "" {
		suffix = "_" + suffix
	}
	return os.MkdirTemp("", strings.Join([]string{"sgpt_temp_*", suffix}, ""))
}

func removeTempDir(fullConfigPath string) error {
	return os.RemoveAll(fullConfigPath)
}

func TestSimplePrompt(t *testing.T) {
	prompt := "Say: Hello World!"
	expected := "Hello World!"

	client, err := NewMockClient(func(ctx context.Context, config *viper.Viper, prompt string, mode string) (string, error) {
		return strings.Clone(expected), nil
	})
	require.NoError(t, err)
	config := createTestConfig(t)
	defer teardownTestDirs(t, config)

	var result string
	result, err = client.GetChatCompletion(context.Background(), config, prompt, "txt")
	require.NoError(t, err)
	require.Equal(t, expected, result)
}
