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

	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime"
	"github.com/spf13/viper"
)

// BedrockInvoker defines the minimal interface needed for Bedrock API calls
type BedrockInvoker interface {
	InvokeModel(ctx context.Context, params *bedrockruntime.InvokeModelInput, optFns ...func(*bedrockruntime.Options)) (*bedrockruntime.InvokeModelOutput, error)
	InvokeModelWithResponseStream(ctx context.Context, params *bedrockruntime.InvokeModelWithResponseStreamInput, optFns ...func(*bedrockruntime.Options)) (*bedrockruntime.InvokeModelWithResponseStreamOutput, error)
}

// Provider defines the interface that all AI providers must implement
type Provider interface {
	// CreateCompletion creates a completion for the given prompt and modifier
	CreateCompletion(ctx context.Context, chatID string, prompt []string, modifier string, input []string) (string, error)
	// GetHTTPClient returns the HTTP client used by the provider
	GetHTTPClient() *http.Client
}

// CreateProvider creates a new provider based on the configuration
func CreateProvider(config *viper.Viper, out io.Writer) (Provider, error) {
	provider := config.GetString("provider")
	// Log the provider selection
	fmt.Printf("Creating provider: %s\n", provider)

	// Try creating AWS Bedrock provider first if specified
	if provider == "bedrock" {
		fmt.Println("Using AWS Bedrock provider")
		return NewAWSBedrockProvider(config, out)
	}

	// Default to OpenAI for empty or "openai" provider
	if provider == "" || provider == "openai" {
		fmt.Println("Using OpenAI provider")
		return CreateClient(config, out)
	}

	return nil, fmt.Errorf("unknown provider: %s", provider)
}
