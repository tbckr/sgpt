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

// BEGIN: 7d8f3c7d7f8d
package cli_test

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/tbckr/sgpt/api"
	"github.com/tbckr/sgpt/cli"
	"github.com/tbckr/sgpt/modifiers"
)

type mockClient struct {
	mock.Mock
}

func (m *mockClient) GetCompletion(ctx context.Context, options api.CompletionOptions, prompt string) (string, error) {
	args := m.Called(ctx, options, prompt)
	return args.String(0), args.Error(1)
}

func (m *mockClient) GetChatCompletion(ctx context.Context, options api.CompletionOptions, prompt string) (string, error) {
	args := m.Called(ctx, options, prompt)
	return args.String(0), args.Error(1)
}

func (m *mockClient) Close() error {
	args := m.Called()
	return args.Error(0)
}

func TestCodeCmd(t *testing.T) {
	tests := []struct {
		name           string
		args           []string
		model          string
		maxTokens      int
		temperature    float64
		topP           float64
		chatSession    string
		expectedOutput string
		expectedError  error
	}{
		{
			name:           "success",
			args:           []string{"prompt"},
			model:          "test-model",
			maxTokens:      2048,
			temperature:    0.8,
			topP:           0.2,
			chatSession:    "",
			expectedOutput: "test-output",
			expectedError:  nil,
		},
		{
			name:           "error",
			args:           []string{"prompt"},
			model:          "test-model",
			maxTokens:      2048,
			temperature:    0.8,
			topP:           0.2,
			chatSession:    "",
			expectedOutput: "",
			expectedError:  errors.New("test error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			var stdout bytes.Buffer
			var stderr bytes.Buffer
			mockClient := new(mockClient)
			mockClient.On("GetCompletion", mock.Anything, api.CompletionOptions{
				Model:       tt.model,
				MaxTokens:   tt.maxTokens,
				Temperature: float32(tt.temperature),
				TopP:        float32(tt.topP),
				Modifier:    modifiers.Code,
				ChatSession: tt.chatSession,
			}, tt.args[0]).Return(tt.expectedOutput, tt.expectedError)

			api.CreateClient = func() (*api.Client, error) {
				return mockClient, nil
			}
			defer func() {
				api.CreateClient = api.NewClient
			}()

			// Execute
			cmd := cli.CodeCmd()
			cmd.SetOut(&stdout)
			cmd.SetErr(&stderr)
			cmd.SetArgs(tt.args)
			err := cmd.Execute()

			// Assert
			assert.Equal(t, tt.expectedError, err)
			if tt.expectedError != nil {
				assert.Empty(t, stdout.String())
			} else {
				assert.Equal(t, tt.expectedOutput+"\n", stdout.String())
			}
			assert.Empty(t, stderr.String())
			mockClient.AssertExpectations(t)
		})
	}
}

func TestCodeCmdWithChatModel(t *testing.T) {
	// Setup
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	mockClient := new(mockClient)
	mockClient.On("GetChatCompletion", mock.Anything, api.CompletionOptions{
		Model:       "test-model",
		MaxTokens:   2048,
		Temperature: 0.8,
		TopP:        0.2,
		Modifier:    modifiers.Code,
		ChatSession: "test-session",
	}, "prompt").Return("test-output", nil)

	api.CreateClient = func() (*api.Client, error) {
		return mockClient, nil
	}
	defer func() {
		api.CreateClient = api.NewClient
	}()

	// Execute
	cmd := cli.CodeCmd()
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)
	cmd.SetArgs([]string{"--model", "test-model", "--chat", "test-session", "prompt"})
	err := cmd.Execute()

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, "test-output\n", stdout.String())
	assert.Empty(t, stderr.String())
	mockClient.AssertExpectations(t)
}

func TestCodeCmdWithError(t *testing.T) {
	// Setup
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	mockClient := new(mockClient)
	mockClient.On("GetCompletion", mock.Anything, api.CompletionOptions{
		Model:       "test-model",
		MaxTokens:   2048,
		Temperature: 0.8,
		TopP:        0.2,
		Modifier:    modifiers.Code,
		ChatSession: "",
	}, "prompt").Return("", errors.New("test error"))

	api.CreateClient = func() (*api.Client, error) {
		return mockClient, nil
	}
	defer func() {
		api.CreateClient = api.NewClient
	}()

	// Execute
	cmd := cli.CodeCmd()
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)
	cmd.SetArgs([]string{"--model", "test-model", "prompt"})
	err := cmd.Execute()

	// Assert
	assert.EqualError(t, err, "test error")
	assert.Empty(t, stdout.String())
	assert.Empty(t, stderr.String())
	mockClient.AssertExpectations(t)
}

func TestCodeCmdWithHelp(t *testing.T) {
	// Setup
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	// Execute
	cmd := cli.CodeCmd()
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)
	cmd.SetArgs([]string{"--help"})
	err := cmd.Execute()

	// Assert
	assert.NoError(t, err)
	assert.Contains(t, stdout.String(), "Query the openai models for code-specific questions")
	assert.Empty(t, stderr.String())
}

func TestCodeCmdWithInvalidModel(t *testing.T) {
	// Setup
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	// Execute
	cmd := cli.CodeCmd()
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)
	cmd.SetArgs([]string{"--model", "invalid-model", "prompt"})
	err := cmd.Execute()

	// Assert
	assert.EqualError(t, err, "invalid model name: invalid-model")
	assert.Empty(t, stdout.String())
	assert.Empty(t, stderr.String())
}

func TestCodeCmdWithInvalidMaxTokens(t *testing.T) {
	// Setup
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	// Execute
	cmd := cli.CodeCmd()
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)
	cmd.SetArgs([]string{"--max-tokens", "invalid-tokens", "prompt"})
	err := cmd.Execute()

	// Assert
	assert.EqualError(t, err, "strconv.Atoi: parsing \"invalid-tokens\": invalid syntax")
	assert.Empty(t, stdout.String())
	assert.Empty(t, stderr.String())
}

func TestCodeCmdWithInvalidTemperature(t *testing.T) {
	// Setup
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	// Execute
	cmd := cli.CodeCmd()
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)
	cmd.SetArgs([]string{"--temperature", "invalid-temperature", "prompt"})
	err := cmd.Execute()

	// Assert
	assert.EqualError(t, err, "strconv.ParseFloat: parsing \"invalid-temperature\": invalid syntax")
	assert.Empty(t, stdout.String())
	assert.Empty(t, stderr.String())
}

func TestCodeCmdWithInvalidTopP(t *testing.T) {
	// Setup
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	// Execute
	cmd := cli.CodeCmd()
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)
	cmd.SetArgs([]string{"--top-p", "invalid-top-p", "prompt"})
	err := cmd.Execute()

	// Assert
	assert.EqualError(t, err, "strconv.ParseFloat: parsing \"invalid-top-p\": invalid syntax")
	assert.Empty(t, stdout.String())
	assert.Empty(t, stderr.String())
}

func TestCodeCmdWithInvalidFlag(t *testing.T) {
	// Setup
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	// Execute
	cmd := cli.CodeCmd()
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)
	cmd.SetArgs([]string{"--invalid-flag", "prompt"})
	err := cmd.Execute()

	// Assert
	assert.EqualError(t, err, "unknown flag: --invalid-flag")
	assert.Empty(t, stdout.String())
	assert.Empty(t, stderr.String())
}

func TestCodeCmdWithNoPrompt(t *testing.T) {
	// Setup
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	// Execute
	cmd := cli.CodeCmd()
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)
	cmd.SetArgs([]string{})
	err := cmd.Execute()

	// Assert
	assert.EqualError(t, err, "requires a prompt argument")
	assert.Empty(t, stdout.String())
	assert.Empty(t, stderr.String())
}

func TestCodeCmdWithInput(t *testing.T) {
	// Setup
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	mockClient := new(mockClient)
	mockClient.On("GetCompletion", mock.Anything, api.CompletionOptions{
		Model:       "test-model",
		MaxTokens:   2048,
		Temperature: 0.8,
		TopP:        0.2,
		Modifier:    modifiers.Code,
		ChatSession: "",
	}, "test-prompt").Return("test-output", nil)

	api.CreateClient = func() (*api.Client, error) {
		return mockClient, nil
	}
	defer func() {
		api.CreateClient = api.NewClient
	}()

	// Execute
	cmd := cli.CodeCmd()
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)
	cmd.SetIn(strings.NewReader("test-prompt\n"))
	cmd.SetArgs([]string{"--model", "test-model"})
	err := cmd.Execute()

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, "test-output\n", stdout.String())
	assert.Empty(t, stderr.String())
	mockClient.AssertExpectations(t)
}

// END: 7d8f3c7d7f8d
