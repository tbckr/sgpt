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
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGetChatModifierShell(t *testing.T) {
	modifier, err := GetChatModifier("sh")
	require.NoError(t, err)
	require.NotEmpty(t, modifier)
	require.NotContains(t, modifier, "%s")
}

func TestGetChatModifierNoShell(t *testing.T) {
	shellEnv := os.Getenv("SHELL")
	defer require.NoError(t, os.Setenv("SHELL", shellEnv))

	require.NoError(t, os.Unsetenv("SHELL"))

	modifier, err := GetChatModifier("sh")
	require.Error(t, err)
	require.Empty(t, modifier)
}

func TestGetChatModifierCode(t *testing.T) {
	modifier, err := GetChatModifier("code")
	require.NoError(t, err)
	require.NotEmpty(t, modifier)
}

func TestGetChatModifierTxt(t *testing.T) {
	modifier, err := GetChatModifier("txt")
	require.NoError(t, err)
	require.Empty(t, modifier)
}

func TestGetChatModifierInvalid(t *testing.T) {
	_, err := GetChatModifier("abcd")
	require.Error(t, err)
}
