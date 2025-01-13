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

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime"
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime/types"
	"github.com/spf13/viper"
)

// AWSBedrockProvider is a client for the AWS Bedrock API
type AWSBedrockProvider struct {
	httpClient *http.Client
	config     *viper.Viper
	client     *bedrockruntime.Client
	out        io.Writer
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

	// Create Bedrock client using AWS config
	client := bedrockruntime.NewFromConfig(awsCfg)

	return &AWSBedrockProvider{
		httpClient: http.DefaultClient,
		config:     config,
		client:     client,
		out:        out,
	}, nil
}

// StreamingPrompt handles streaming responses from AWS Bedrock
func (c *AWSBedrockProvider) StreamingPrompt(ctx context.Context, model string, body string) (string, error) {
	input := &bedrockruntime.InvokeModelWithResponseStreamInput{
		ModelId:     aws.String(model),
		Body:        []byte(body),
		ContentType: aws.String("application/json"),
	}

	output, err := c.client.InvokeModelWithResponseStream(ctx, input)
	if err != nil {
		return "", fmt.Errorf("failed to invoke model: %w", err)
	}

	var fullResponse string
	for event := range output.GetStream().Events() {
		switch v := event.(type) {
		case *types.ResponseStreamMemberChunk:
			var deltaResp struct {
				Type  string `json:"type"`
				Delta struct {
					Type    string `json:"type"`
					Text    string `json:"text"`
					Message struct {
						Content []struct {
							Type string `json:"type"`
							Text string `json:"text"`
						} `json:"content"`
					} `json:"message"`
				} `json:"delta"`
			}

			if err := json.Unmarshal(v.Value.Bytes, &deltaResp); err != nil {
				return "", fmt.Errorf("failed to unmarshal response: %w", err)
			}

			switch {
			case deltaResp.Type == "message_start" || deltaResp.Type == "message_delta":
				if len(deltaResp.Delta.Message.Content) > 0 {
					newContent := deltaResp.Delta.Message.Content[0].Text
					fullResponse += newContent
					// Print only the new content
					if _, err := fmt.Fprint(c.out, newContent); err != nil {
						return "", fmt.Errorf("failed to write streaming response: %w", err)
					}
				}
			case deltaResp.Delta.Type == "text_delta":
				newContent := deltaResp.Delta.Text
				fullResponse += newContent
				// Print only the new content
				if _, err := fmt.Fprint(c.out, newContent); err != nil {
					return "", fmt.Errorf("failed to write streaming response: %w", err)
				}
			default:
				// Skip other message types (like message_stop)
				continue
			}
		}
	}

	return fullResponse, nil
}

// SimplePrompt handles non-streaming responses from AWS Bedrock
func (c *AWSBedrockProvider) SimplePrompt(ctx context.Context, model string, body string) (string, error) {
	input := &bedrockruntime.InvokeModelInput{
		ModelId:     aws.String(model),
		Body:        []byte(body),
		ContentType: aws.String("application/json"),
	}

	output, err := c.client.InvokeModel(ctx, input)
	if err != nil {
		return "", fmt.Errorf("failed to invoke model: %w", err)
	}

	var resp struct {
		Completion string `json:"completion"`
		Message    struct {
			Content []struct {
				Type string `json:"type"`
				Text string `json:"text"`
			} `json:"content"`
		} `json:"message"`
	}

	if err := json.Unmarshal(output.Body, &resp); err != nil {
		return "", fmt.Errorf("failed to unmarshal response: %w", err)
	}

	// Try to get completion field first
	if resp.Completion != "" {
		return resp.Completion, nil
	}

	// Try to get text from message content
	if len(resp.Message.Content) > 0 {
		var result string
		for _, content := range resp.Message.Content {
			if content.Type == "text" {
				result += content.Text
			}
		}
		if result != "" {
			return result, nil
		}
	}

	return "", fmt.Errorf("no valid response content found")
}
