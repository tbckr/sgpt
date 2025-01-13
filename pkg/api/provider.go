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
	"io"
	"net/http"

	"github.com/spf13/viper"
)

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
	switch provider {
	case "bedrock":
		return NewAWSBedrockProvider(config, out)
	default:
		return CreateClient(config, out)
	}
}
