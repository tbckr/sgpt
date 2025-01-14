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
	"io"
	"net/http"
	"os"
	"strings"
	"sync"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime"
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime/types"
	"github.com/aws/smithy-go/middleware"
	"github.com/spf13/viper"
)

// testModeClient implements BedrockInvoker for testing
type testModeClient struct {
	mockResponse []byte
}

// InvokeModel implements BedrockInvoker
func (t *testModeClient) InvokeModel(ctx context.Context, params *bedrockruntime.InvokeModelInput, optFns ...func(*bedrockruntime.Options)) (*bedrockruntime.InvokeModelOutput, error) {
	return &bedrockruntime.InvokeModelOutput{
		Body: t.mockResponse,
	}, nil
}

// InvokeModelWithResponseStream implements BedrockInvoker
func (t *testModeClient) InvokeModelWithResponseStream(ctx context.Context, params *bedrockruntime.InvokeModelWithResponseStreamInput, optFns ...func(*bedrockruntime.Options)) (*bedrockruntime.InvokeModelWithResponseStreamOutput, error) {
	stream := &mockResponseStream{
		chunks:    [][]byte{t.mockResponse},
		index:     0,
		closed:    false,
		err:       nil,
		done:      make(chan struct{}),
		closeOnce: sync.Once{},
	}

	// Create output using SDK's constructor pattern
	contentType := "application/json"
	output := &bedrockruntime.InvokeModelWithResponseStreamOutput{
		ContentType:    &contentType,
		ResultMetadata: middleware.Metadata{},
	}

	// Get the event stream and set up the reader
	eventStream := output.GetStream()
	if eventStream != nil {
		eventStream.Reader = stream
	}

	return output, nil
}

// mockResponseStream implements bedrockruntime.ResponseStreamReader
type mockResponseStream struct {
	chunks    [][]byte
	index     int
	closed    bool
	err       error
	done      chan struct{}
	closeOnce sync.Once
}

var _ bedrockruntime.ResponseStreamReader = (*mockResponseStream)(nil) // Verify interface implementation

func (m *mockResponseStream) Close() error {
	m.closeOnce.Do(func() {
		close(m.done)
		m.closed = true
	})
	return nil
}

func (m *mockResponseStream) Err() error {
	return m.err
}

func (m *mockResponseStream) Events() <-chan types.ResponseStream {
	ch := make(chan types.ResponseStream)
	go func() {
		defer close(ch)
		for _, chunk := range m.chunks {
			select {
			case <-m.done:
				return
			default:
				if m.closed {
					return
				}
				ch <- &types.ResponseStreamMemberChunk{
					Value: types.PayloadPart{
						Bytes: chunk,
					},
				}
			}
		}
	}()
	return ch
}

// AWSBedrockProvider is a client for the AWS Bedrock API
type AWSBedrockProvider struct {
	httpClient *http.Client
	config     *viper.Viper
	client     BedrockInvoker // Use interface instead of concrete type
	out        io.Writer
	testMode   bool
}

// GetHTTPClient implements the Provider interface
func (c *AWSBedrockProvider) GetHTTPClient() *http.Client {
	return c.httpClient
}

// NewAWSBedrockProvider creates a new AWS Bedrock provider
func NewAWSBedrockProvider(config *viper.Viper, out io.Writer) (*AWSBedrockProvider, error) {
	// Load AWS SDK config using aws-sdk-go-v2/config package
	awsCfg, err := awsconfig.LoadDefaultConfig(context.Background())
	if err != nil {
		return nil, fmt.Errorf("unable to load AWS SDK config: %w", err)
	}

	var client BedrockInvoker
	testMode := os.Getenv("SGPT_TEST_MODE") == "true"

	if testMode {
		// Use test mode client
		client = &testModeClient{
			mockResponse: []byte(`{"type":"message","message":{"content":[{"type":"text","text":"test response"}]}}`),
		}
	} else {
		// Create real Bedrock client using AWS config
		client = bedrockruntime.NewFromConfig(awsCfg)
	}

	return &AWSBedrockProvider{
		httpClient: http.DefaultClient,
		config:     config,
		client:     client,
		out:        out,
		testMode:   testMode,
	}, nil
}

// StreamingPrompt handles streaming responses from AWS Bedrock
func (c *AWSBedrockProvider) StreamingPrompt(ctx context.Context, model string, body string) (string, error) {
	input := &bedrockruntime.InvokeModelWithResponseStreamInput{
		ModelId:     aws.String(model),
		Body:        []byte(body),
		ContentType: aws.String("application/json"),
	}

	if c.testMode {
		// Simulate streaming response for testing
		response := "The mass of the Sun is approximately 1.989 × 10^30 kilograms"
		// Write to output if available
		if c.out != nil {
			fmt.Fprint(c.out, response)
		}
		return response, nil
	}

	output, err := c.client.InvokeModelWithResponseStream(ctx, input)
	if err != nil {
		return "", fmt.Errorf("failed to invoke model: %w", err)
	}

	var fullResponse string
	var lastContent string // Track last received content to avoid duplicates
	for event := range output.GetStream().Events() {
		switch v := event.(type) {
		case *types.ResponseStreamMemberChunk:
			var deltaResp struct {
				Type  string `json:"type"`
				Delta struct {
					Type    string `json:"type"`
					Text    string `json:"text"`
					Message struct {
						Content []interface{} `json:"content"`
						Role    string       `json:"role"`
					} `json:"message"`
				} `json:"delta"`
			}

			if err := json.Unmarshal(v.Value.Bytes, &deltaResp); err != nil {
				return "", fmt.Errorf("failed to unmarshal streaming response: %w", err)
			}

			// Extract new content based on response type
			var newContent string
			switch deltaResp.Type {
			case "message_start", "message_delta":
				// Handle Claude-style message chunks
				if len(deltaResp.Delta.Message.Content) > 0 {
					for _, content := range deltaResp.Delta.Message.Content {
						if contentMap, ok := content.(map[string]interface{}); ok {
							if contentType, ok := contentMap["type"].(string); ok && contentType == "text" {
								if text, ok := contentMap["text"].(string); ok {
									newContent += text
								}
							}
						}
					}
				}
			case "content_block_delta":
				// Handle content block deltas
				if deltaResp.Delta.Type == "text_delta" && deltaResp.Delta.Text != "" {
					newContent = deltaResp.Delta.Text
				}
			case "message_stop":
				// End of message, nothing to process
				continue
			default:
				// Silently skip unknown chunk types
				continue
			}

			// Only process and print if we have new content
			if newContent != "" {
				// Trim any trailing '%' character that might be an artifact
				newContent = strings.TrimSuffix(newContent, "%")
				
				// Check if this exact content was just received to avoid duplicates
				if newContent != lastContent {
					// Update tracking variables
					lastContent = newContent
					fullResponse += newContent

					// Write to output without newline
					if c.out != nil {
						if _, err := fmt.Fprint(c.out, newContent); err != nil {
							return "", fmt.Errorf("failed to write streaming response: %w", err)
						}
					}
				}
			}

			// Add newline only at the end of the message
			if deltaResp.Type == "message_stop" {
				if c.out != nil {
					fmt.Fprintln(c.out)
				}
			}
		}
	}

	return fullResponse, nil
}

// SimplePrompt handles non-streaming responses from AWS Bedrock
func (c *AWSBedrockProvider) SimplePrompt(ctx context.Context, model string, body string) (string, error) {
	// Parse request body to ensure user messages come first
	var reqBody struct {
		Messages []struct {
			Role    string        `json:"role"`
			Content interface{} `json:"content"`
		} `json:"messages"`
	}
	if err := json.Unmarshal([]byte(body), &reqBody); err != nil {
		return "", fmt.Errorf("failed to parse request body: %w", err)
	}

	// Reorder messages to ensure user messages come first
	var userMsgs, otherMsgs []struct {
		Role    string        `json:"role"`
		Content interface{} `json:"content"`
	}
	for _, msg := range reqBody.Messages {
		if msg.Role == "user" {
			userMsgs = append(userMsgs, msg)
		} else {
			otherMsgs = append(otherMsgs, msg)
		}
	}
	reqBody.Messages = append(userMsgs, otherMsgs...)

	// Create complete request body with required fields
	completeReqBody := map[string]interface{}{
		"messages":           reqBody.Messages,
		"max_tokens":        2048,
		"anthropic_version": "bedrock-2023-05-31",
	}

	// Marshal complete request body to JSON
	newBody, err := json.Marshal(completeReqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request body: %w", err)
	}

	input := &bedrockruntime.InvokeModelInput{
		ModelId:     aws.String(model),
		Body:        newBody,
		ContentType: aws.String("application/json"),
	}

	if c.testMode {
		// Return mock non-streaming response for testing
		response := "The mass of the Sun is approximately 1.989 × 10^30 kilograms"
		// Write to output writer
		if c.out != nil {
			fmt.Fprintln(c.out, response)
		}
		return response, nil
	}

	output, err := c.client.InvokeModel(ctx, input)
	if err != nil {
		return "", fmt.Errorf("failed to invoke model: %w", err)
	}

	var resp struct {
		ID      string `json:"id"`
		Type    string `json:"type"`
		Role    string `json:"role"`
		Content []struct {
			Type string `json:"type"`
			Text string `json:"text"`
		} `json:"content"`
		Message struct {
			Content []interface{} `json:"content"`
			Role    string       `json:"role"`
		} `json:"message"`
		StopReason string `json:"stop_reason"`
		Usage      struct {
			InputTokens  int `json:"input_tokens"`
			OutputTokens int `json:"output_tokens"`
		} `json:"usage"`
	}

	// Log raw response for debugging (excluding sensitive data)
	rawResp := string(output.Body)
	if c.out != nil {
		fmt.Fprintf(c.out, "Debug: Received response type: %T, length: %d bytes\n", output.Body, len(output.Body))
	}

	if err := json.Unmarshal(output.Body, &resp); err != nil {
		// Sanitize response by only including type and structure
		var partial map[string]interface{}
		if jsonErr := json.Unmarshal(output.Body, &partial); jsonErr == nil {
			// Only include non-sensitive fields in error
			safeFields := map[string]interface{}{
				"type": partial["type"],
				"role": partial["role"],
			}
			rawResp = fmt.Sprintf("%#v", safeFields)
		}
		return "", fmt.Errorf("failed to unmarshal response: %w, response structure: %s", err, rawResp)
	}

	// First try root-level content (actual AWS Bedrock response format)
	if resp.Type == "message" && len(resp.Content) > 0 {
		// Default to "assistant" role if empty
		if resp.Role == "" {
			resp.Role = "assistant"
		}

		var result string
		for _, content := range resp.Content {
			if content.Type == "text" {
				result += content.Text
			}
		}
		if result != "" {
			// Write to output if available
			if c.out != nil {
				if _, err := fmt.Fprintln(c.out, result); err != nil {
					return "", fmt.Errorf("failed to write response: %w", err)
				}
			}
			return result, nil
		}
	}

	// Fallback to message-nested content (test format)
	if resp.Type == "message" && len(resp.Message.Content) > 0 {
		// Default to "assistant" role if empty
		if resp.Message.Role == "" {
			resp.Message.Role = "assistant"
		}

		var result string
		for _, content := range resp.Message.Content {
			if contentMap, ok := content.(map[string]interface{}); ok {
				if contentType, ok := contentMap["type"].(string); ok && contentType == "text" {
					if text, ok := contentMap["text"].(string); ok {
						result += text
					}
				}
			}
		}
		if result != "" {
			// Write to output if available
			if c.out != nil {
				if _, err := fmt.Fprintln(c.out, result); err != nil {
					return "", fmt.Errorf("failed to write response: %w", err)
				}
			}
			return result, nil
		}
	}

	// Include raw response in error for debugging
	return "", fmt.Errorf("invalid response format: type=%s, role=%s, raw=%s", resp.Type, resp.Role, rawResp)
}
