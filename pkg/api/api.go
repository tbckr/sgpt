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
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"strings"

	"github.com/sashabaranov/go-openai"
	"github.com/spf13/viper"
	"github.com/tbckr/sgpt/v2/pkg/chat"
	"github.com/tbckr/sgpt/v2/pkg/modifiers"
)

const (
	// envKeyOpenAIApi is the environment variable key for the OpenAI API key.
	envKeyOpenAIApi = "OPENAI_API_KEY"
)

var (
	// DefaultModel is the default model used for chat completions.
	DefaultModel = strings.Clone(openai.GPT3Dot5Turbo)
	// ErrMissingAPIKey is returned, if the OPENAI_API_KEY environment variable is not set.
	ErrMissingAPIKey = fmt.Errorf("%s env variable is not set", envKeyOpenAIApi)
)

// OpenAIClient is a client for the OpenAI API.
type OpenAIClient struct {
	HTTPClient *http.Client
	config     *viper.Viper
	api        *openai.Client
	out        io.Writer
}

// CreateClient creates a new OpenAI client with the given config and output writer.
func CreateClient(config *viper.Viper, out io.Writer) (*OpenAIClient, error) {
	// Check, if api key was set
	apiKey, exists := os.LookupEnv(envKeyOpenAIApi)
	if !exists {
		return nil, ErrMissingAPIKey
	}
	clientConfig := openai.DefaultConfig(apiKey)

	// Set HTTP Proxy
	transport := &http.Transport{
		Proxy: http.ProxyFromEnvironment,
	}
	httpClient := &http.Client{
		Transport: transport,
	}
	clientConfig.HTTPClient = httpClient

	// Check, if API base url was set
	baseURL, isSet := os.LookupEnv("OPENAI_API_BASE")
	if isSet {
		// Set base url
		clientConfig.BaseURL = baseURL
		slog.Debug("Setting API base url to " + baseURL)
	}

	// Create client
	client := &OpenAIClient{
		HTTPClient: httpClient,
		config:     config,
		api:        openai.NewClientWithConfig(clientConfig),
		out:        out,
	}
	slog.Debug("OpenAI client created")
	return client, nil
}

// CreateCompletion creates a completion for the given prompt and modifier. If chatID is provided, the chat is reused
// and the completion is added to the chat with this ID. If no chatID is provided, only the modifier and prompt are
// used to create the completion. The completion is printed to the out writer of the client and returned as a string.
func (c *OpenAIClient) CreateCompletion(ctx context.Context, chatID, prompt, modifier string) (string, error) {
	var err error
	var chatSessionManager chat.SessionManager
	var messages []openai.ChatCompletionMessage

	chatSessionManager, err = chat.NewFilesystemChatSessionManager(c.config)
	if err != nil {
		return "", err
	}

	isChat := false
	if chatID != "" {
		isChat = true
	}
	chatExists := false

	// Load existing chat messages
	if isChat {
		chatExists, err = chatSessionManager.SessionExists(chatID)
		if err != nil {
			return "", err
		}
		if chatExists {
			var loadedMessages []openai.ChatCompletionMessage
			loadedMessages, err = chatSessionManager.GetSession(chatID)
			if err != nil {
				return "", err
			}
			messages = append(messages, loadedMessages...)
			slog.Debug("Loaded chat session")
		}
	}

	// If this message is not part of a chat
	// OR
	// if this is the initial message of a chat,
	// then add modifier message
	if !isChat || (isChat && !chatExists) {
		var modifierPrompt string
		modifierPrompt, err = modifiers.GetChatModifier(c.config, modifier)
		if err != nil {
			return "", err
		}
		if modifierPrompt != "" {
			messages = append(messages, openai.ChatCompletionMessage{
				Role:    openai.ChatMessageRoleSystem,
				Content: modifierPrompt,
			})
			slog.Debug("Added modifier message")
		}
	}

	// Add prompt to messages
	messages = append(messages, openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleUser,
		Content: prompt,
	})
	slog.Debug("Added prompt message")

	// Create request
	req := openai.ChatCompletionRequest{
		Messages:    messages,
		Model:       c.config.GetString("model"),
		MaxTokens:   c.config.GetInt("max-tokens"),
		Temperature: float32(c.config.GetFloat64("temperature")),
		TopP:        float32(c.config.GetFloat64("top-p")),
		Stream:      c.config.GetBool("stream"),
	}

	// Retrieve response
	var receivedMessage openai.ChatCompletionMessage
	if c.config.GetBool("stream") {
		receivedMessage, err = c.retrieveChatCompletionStream(ctx, req)
	} else {
		receivedMessage, err = c.retrieveChatCompletion(ctx, req)
	}
	if err != nil {
		return "", err
	}

	slog.Debug("Received message from OpenAI API")

	// If a session was provided, save received message to this chat
	if isChat {
		messages = append(messages, receivedMessage)
		if err = chatSessionManager.SaveSession(chatID, messages); err != nil {
			return "", err
		}
		slog.Debug("Saved chat session")
	}
	// Return received message
	return receivedMessage.Content, nil
}

func (c *OpenAIClient) retrieveChatCompletion(ctx context.Context, req openai.ChatCompletionRequest) (openai.ChatCompletionMessage, error) {
	resp, err := c.api.CreateChatCompletion(ctx, req)
	if err != nil {
		return openai.ChatCompletionMessage{}, err
	}
	receivedMessage := resp.Choices[0].Message

	_, err = fmt.Fprintln(c.out, receivedMessage.Content)
	if err != nil {
		return openai.ChatCompletionMessage{}, err
	}

	return receivedMessage, nil
}

func (c *OpenAIClient) retrieveChatCompletionStream(ctx context.Context, req openai.ChatCompletionRequest) (openai.ChatCompletionMessage, error) {
	stream, err := c.api.CreateChatCompletionStream(ctx, req)
	if err != nil {
		return openai.ChatCompletionMessage{}, err
	}
	defer stream.Close()
	slog.Debug("Streaming response")

	// Print initial linebreak
	_, err = fmt.Fprintf(c.out, "\n")
	if err != nil {
		return openai.ChatCompletionMessage{}, err
	}

	var receivedMessage openai.ChatCompletionMessage
	for {
		response, streamErr := stream.Recv()
		if errors.Is(streamErr, io.EOF) {
			slog.Debug("Stream finished")
			break
		}
		if streamErr != nil {
			slog.Debug("Stream error encountered")
			return openai.ChatCompletionMessage{}, streamErr
		}

		receivedContent := response.Choices[0].Delta.Content
		// 1. Append received content to message
		receivedMessage.Content += receivedContent
		// 2. Print received content
		_, err = fmt.Fprint(c.out, receivedContent)
		if err != nil {
			return openai.ChatCompletionMessage{}, err
		}
	}
	// Print final linebreak
	_, err = fmt.Fprintf(c.out, "\n")
	if err != nil {
		slog.Warn("Could not print final linebreak")
	}
	// Return received message to save it to the chat session
	return receivedMessage, nil
}
