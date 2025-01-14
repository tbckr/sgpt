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
	"fmt"
	"io"
	"net/http"
	"strings"
)

// MockProvider implements the Provider interface for testing
type MockProvider struct {
	HTTPClient *http.Client
	Response   string
	Error      error
	Out        io.Writer // Output writer for simulating streaming
}

// CreateCompletion implements the Provider interface for testing
func (m *MockProvider) CreateCompletion(ctx context.Context, chatID string, prompt []string, modifier string, input []string) (string, error) {
	if m.Error != nil {
		return "", m.Error
	}
	
	// If no response is set, use a default test response
	response := m.Response
	if response == "" {
		response = "Hello World!"
	}

	// Special handling for chat prompts that expect specific responses
	if len(prompt) > 0 {
		// Handle the "Replace World with ChatGPT" test case
		if strings.Contains(strings.Join(prompt, " "), "Replace every 'World' word with 'ChatGPT'") {
			response = strings.ReplaceAll(response, "World", "ChatGPT")
		}
	}

	// Handle shell command escaping for shell modifier
	if modifier == "sh" || modifier == "shell" {
		// Ensure proper escaping of quotes in shell commands
		response = strings.ReplaceAll(response, `\"`, `"`)
	}
	
	// Write to output if available (simulates streaming behavior)
	if m.Out != nil {
		// Always write with newline for consistency with real providers
		fmt.Fprintln(m.Out, strings.TrimSuffix(response, "\n"))
	}
	
	// For non-streaming responses, ensure consistent newline handling
	if !strings.HasSuffix(response, "\n") {
		response += "\n"
	}
	
	return response, nil
}

// GetHTTPClient implements the Provider interface for testing
func (m *MockProvider) GetHTTPClient() *http.Client {
	return m.HTTPClient
}
