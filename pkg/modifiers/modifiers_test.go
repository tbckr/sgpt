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

package modifiers

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/spf13/viper"

	"github.com/stretchr/testify/require"
)

func TestGetChatModifierShell(t *testing.T) {
	config := createTestConfig(t)

	err := os.Setenv("SHELL", "/bin/bash")
	require.NoError(t, err)
	t.Cleanup(func() {
		require.NoError(t, os.Unsetenv("SHELL"))
	})

	modifier, err := GetChatModifier(config, "sh")
	require.NoError(t, err)
	require.NotEmpty(t, modifier)
	require.NotContains(t, modifier, "{{")
	require.NotContains(t, modifier, "}}")
}

func TestGetChatModifierNoShell(t *testing.T) {
	expectedPattern := `Act as a natural language to command translation engine on %s.
You are an expert in %s and translate the question at the end to valid syntax.`
	expected := fmt.Sprintf(expectedPattern, runtime.GOOS, runtime.GOOS)

	config := createTestConfig(t)

	shellEnv := os.Getenv("SHELL")
	require.NoError(t, os.Unsetenv("SHELL"))
	t.Cleanup(func() {
		require.NoError(t, os.Setenv("SHELL", shellEnv))
	})

	modifier, err := GetChatModifier(config, "sh")
	require.NoError(t, err)
	require.True(t, strings.HasPrefix(modifier, expected))
}

func TestGetChatModifierCode(t *testing.T) {
	config := createTestConfig(t)

	modifier, err := GetChatModifier(config, "code")
	require.NoError(t, err)
	require.NotEmpty(t, modifier)
}

func TestGetChatModifierCodeOverride(t *testing.T) {
	codeOverride := `Act as a natural language to code translation engine.`

	config := createTestConfig(t)

	fileHandler, err := os.Create(filepath.Join(config.GetString("personas"), "code"))
	require.NoError(t, err)

	_, err = fileHandler.WriteString(codeOverride)
	require.NoError(t, err)

	var modifier string
	modifier, err = GetChatModifier(config, "code")
	require.NoError(t, err)
	require.Equal(t, codeOverride, modifier)
	require.NoError(t, fileHandler.Close())
}

func TestGetChatModifierCodeOverrideWithComments(t *testing.T) {
	codeOverride := `Act as a natural language to code translation engine.`
	personaContent := "# This is a comment\n" + codeOverride

	config := createTestConfig(t)

	fileHandler, err := os.Create(filepath.Join(config.GetString("personas"), "code"))
	require.NoError(t, err)

	_, err = fileHandler.WriteString(personaContent)
	require.NoError(t, err)

	var modifier string
	modifier, err = GetChatModifier(config, "code")
	require.NoError(t, err)
	require.Equal(t, codeOverride, modifier)
	require.NoError(t, fileHandler.Close())
}

func TestGetChatModifierCustomPersona(t *testing.T) {
	codeOverride := `This is custom persona.`

	config := createTestConfig(t)

	fileHandler, err := os.Create(filepath.Join(config.GetString("personas"), "my-persona"))
	require.NoError(t, err)

	_, err = fileHandler.WriteString(codeOverride)
	require.NoError(t, err)

	var modifier string
	modifier, err = GetChatModifier(config, "my-persona")
	require.NoError(t, err)
	require.Equal(t, codeOverride, modifier)
	require.NoError(t, fileHandler.Close())
}

func TestGetChatModifierInvalidPersonaInSubdir(t *testing.T) {
	personaText := `This is custom persona.`

	config := createTestConfig(t)

	fileHandler, err := os.Create(filepath.Join(config.GetString("personas"), "my-persona"))
	require.NoError(t, err)
	_, err = fileHandler.WriteString(personaText)
	require.NoError(t, err)
	require.NoError(t, fileHandler.Close())

	err = os.MkdirAll(filepath.Join(config.GetString("personas"), "subdir"), 0755)
	require.NoError(t, err)
	fileHandler, err = os.Create(filepath.Join(config.GetString("personas"), "subdir", "my-persona2"))
	require.NoError(t, err)
	_, err = fileHandler.WriteString(personaText)
	require.NoError(t, err)
	require.NoError(t, fileHandler.Close())

	var modifier string
	modifier, err = GetChatModifier(config, "my-persona2")
	require.Error(t, err)
	require.Empty(t, modifier)
}

func TestGetChatModifierInvalidWhitespacesInName(t *testing.T) {
	personaText := `This is custom persona.`

	config := createTestConfig(t)

	fileHandler, err := os.Create(filepath.Join(config.GetString("personas"), "my persona"))
	require.NoError(t, err)
	_, err = fileHandler.WriteString(personaText)
	require.NoError(t, err)
	require.NoError(t, fileHandler.Close())

	var modifier string
	modifier, err = GetChatModifier(config, "my persona2")
	require.Error(t, err)
	require.Empty(t, modifier)
}

func TestGetChatModifierTxt(t *testing.T) {
	config := createTestConfig(t)

	modifier, err := GetChatModifier(config, "txt")
	require.NoError(t, err)
	require.Empty(t, modifier)
}

func TestGetChatModifierInvalid(t *testing.T) {
	config := createTestConfig(t)

	_, err := GetChatModifier(config, "abcd")
	require.Error(t, err)
}

func createTestConfig(t *testing.T) *viper.Viper {
	cacheDir := t.TempDir()
	configDir := t.TempDir()
	personasDir := t.TempDir()

	config := viper.New()
	config.AddConfigPath(configDir)
	config.SetConfigName("config")
	config.SetConfigType("yaml")
	config.Set("cacheDir", cacheDir)
	config.Set("personas", personasDir)
	config.Set("TESTING", 1)

	return config
}
