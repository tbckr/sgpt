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
	"encoding/json"
	"fmt"
	"strings"
)

// CreateCompletion implements the Provider interface for AWS Bedrock
func (c *AWSBedrockProvider) CreateCompletion(ctx context.Context, chatID string, prompt []string, modifier string, input []string) (string, error) {
	// Build messages array
	var messages []map[string]interface{}
	var systemMessage map[string]interface{}

	// Extract system prompt if present
	if len(prompt) > 0 && strings.HasPrefix(prompt[0], "System:") {
		systemMessage = map[string]interface{}{
			"role": "system",
			"content": []map[string]string{
				{
					"type": "text",
					"text": strings.TrimPrefix(prompt[0], "System:"),
				},
			},
		}
		// Remove system prompt from the array
		prompt = prompt[1:]
	}

	// Add user messages first (required for Claude models)
	for _, p := range prompt {
		messages = append(messages, map[string]interface{}{
			"role": "user",
			"content": []map[string]string{
				{
					"type": "text",
					"text": p,
				},
			},
		})
	}

	// Append system message after user messages if present
	if systemMessage != nil {
		messages = append(messages, systemMessage)
	}

	// Create the request body
	requestBody := map[string]interface{}{
		"anthropic_version": "bedrock-2023-05-31",
		"temperature":       c.config.GetFloat64("temperature"),
		"messages":         messages,
		"max_tokens":       c.config.GetInt("maxtokens"),
	}

	// Marshal the request body
	jsonBytes, err := json.Marshal(requestBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request body: %w", err)
	}

	// Use streaming if configured
	if c.config.GetBool("stream") {
		return c.StreamingPrompt(ctx, c.config.GetString("model"), string(jsonBytes))
	}
	return c.SimplePrompt(ctx, c.config.GetString("model"), string(jsonBytes))
}
