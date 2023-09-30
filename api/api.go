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
	"log/slog"
	"os"
	"strings"

	"github.com/sashabaranov/go-openai"
	"github.com/spf13/viper"
	"github.com/tbckr/sgpt/v2/chat"
	"github.com/tbckr/sgpt/v2/modifiers"
)

const (
	envKeyOpenAIApi = "OPENAI_API_KEY"
)

var (
	DefaultModel     = strings.Clone(openai.GPT3Dot5Turbo)
	ErrMissingAPIKey = fmt.Errorf("%s env variable is not set", envKeyOpenAIApi)
)

type OpenAIClient struct {
	api                *openai.Client
	retrieveResponseFn func(*openai.Client, context.Context, openai.ChatCompletionRequest) (openai.ChatCompletionResponse, error)
}

func MockClient(response string, err error) func() (*OpenAIClient, error) {
	return func() (*OpenAIClient, error) {
		return &OpenAIClient{
			api: nil,
			retrieveResponseFn: func(_ *openai.Client, _ context.Context, _ openai.ChatCompletionRequest) (openai.ChatCompletionResponse, error) {
				return openai.ChatCompletionResponse{
					Choices: []openai.ChatCompletionChoice{
						{
							Message: openai.ChatCompletionMessage{
								Role:    openai.ChatMessageRoleAssistant,
								Content: response,
							},
						},
					},
				}, nil
			},
		}, err
	}
}

func CreateClient() (*OpenAIClient, error) {
	// Check, if api key was set
	apiKey, exists := os.LookupEnv(envKeyOpenAIApi)
	if !exists {
		return nil, ErrMissingAPIKey
	}
	client := &OpenAIClient{
		api: openai.NewClient(apiKey),
		// This is necessary to be able to mock the api in tests
		retrieveResponseFn: func(api *openai.Client, ctx context.Context, req openai.ChatCompletionRequest) (openai.ChatCompletionResponse, error) {
			return api.CreateChatCompletion(ctx, req)
		},
	}
	slog.Debug("OpenAI client created")
	return client, nil
}

func (c *OpenAIClient) GetChatCompletion(ctx context.Context, config *viper.Viper, prompt, modifier string) (string, error) {
	var err error
	var chatSessionManager chat.SessionManager
	var messages []openai.ChatCompletionMessage

	chatSessionManager, err = chat.NewFilesystemChatSessionManager(config)
	if err != nil {
		return "", err
	}

	var chatID string
	var isChat bool
	if config.IsSet("chat") {
		chatID = config.GetString("chat")
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
		modifierPrompt, err = modifiers.GetChatModifier(config, modifier)
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

	// Do request
	req := openai.ChatCompletionRequest{
		Messages:    messages,
		Model:       config.GetString("model"),
		MaxTokens:   config.GetInt("max-tokens"),
		Temperature: float32(config.GetFloat64("temperature")),
		TopP:        float32(config.GetFloat64("top-p")),
	}
	var resp openai.ChatCompletionResponse
	resp, err = c.retrieveResponseFn(c.api, ctx, req)
	if err != nil {
		return "", err
	}
	receivedMessage := resp.Choices[0].Message
	slog.Debug("Received message from OpenAI API")

	// Remove surrounding white spaces
	receivedMessage.Content = strings.TrimSpace(receivedMessage.Content)

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
