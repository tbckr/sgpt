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
	"net/http"
	"os"
	"testing"

	"github.com/tbckr/sgpt/v2/pkg/api"
	"github.com/tbckr/sgpt/v2/internal/testlib"
	"github.com/stretchr/testify/require"
)

func TestCheckCmd(t *testing.T) {
	testCtx := testlib.NewTestCtx(t)
	testlib.SetAPIKey(t)
	mem := &exitMemento{}

	testlib.SetAPIKey(t)

	client := &api.MockProvider{
		HTTPClient: &http.Client{},
		Out:        os.Stdout,
	}
	newRootCmd(mem.Exit, testCtx.Config, mockIsPipedShell(false, nil), useMockClient(client)).Execute([]string{"check"})
	require.Equal(t, 0, mem.code)
}

func TestCheckCmdUnsetEnvAPIKey(t *testing.T) {
	testCtx := testlib.NewTestCtx(t)
	mem := &exitMemento{}

	// Save current API key
	apiKey := os.Getenv("OPENAI_API_KEY")
	
	// Unset OpenAI API key for this test
	if err := os.Unsetenv("OPENAI_API_KEY"); err != nil {
		t.Fatal(err)
	}

	// Test with OpenAI provider (default)
	client := &api.MockProvider{HTTPClient: &http.Client{}}
	newRootCmd(mem.Exit, testCtx.Config, mockIsPipedShell(false, nil), useMockClient(client)).Execute([]string{"check"})
	require.Equal(t, 1, mem.code)

	// Test with Bedrock provider (should pass without OpenAI API key)
	testCtx.Config.Set("provider", "bedrock")
	mem = &exitMemento{}
	newRootCmd(mem.Exit, testCtx.Config, mockIsPipedShell(false, nil), useMockClient(client)).Execute([]string{"check"})
	require.Equal(t, 0, mem.code)

	// Reset API key
	if apiKey != "" {
		if err := os.Setenv("OPENAI_API_KEY", apiKey); err != nil {
			t.Fatal(err)
		}
	}
}
