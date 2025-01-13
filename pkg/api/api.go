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

	"github.com/tbckr/sgpt/v2/pkg/fs"

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
	httpClient         *http.Client
	config             *viper.Viper
	api                *openai.Client
	out                io.Writer
	chatSessionManager chat.SessionManager
}

// GetHTTPClient implements the Provider interface
func (c *OpenAIClient) GetHTTPClient() *http.Client {
	return c.httpClient
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

	// Initialize chat session manager
	chatSessionManager, err := chat.NewFilesystemChatSessionManager(config)
	if err != nil {
		return nil, err
	}
	slog.Debug("Chat session manager initialized")

	// Create client
	client := &OpenAIClient{
		HTTPClient:         httpClient,
		config:             config,
		api:                openai.NewClientWithConfig(clientConfig),
		out:                out,
		chatSessionManager: chatSessionManager,
	}
	slog.Debug("OpenAI client created")
	return client, nil
}

// CreateCompletion creates a completion for the given prompt and modifier. If chatID is provided, the chat is reused
// and the completion is added to the chat with this ID. If no chatID is provided, only the modifier and prompt are
// used to create the completion. The completion is printed to the out writer of the client and returned as a string.
func (c *OpenAIClient) CreateCompletion(ctx context.Context, chatID string, prompt []string, modifier string, input []string) (string, error) {
	var messages []openai.ChatCompletionMessage
	var err error

	isChat := false
	if chatID != "" {
		isChat = true
	}

	// Load existing chat messages:
	// If this is a chat, load existing messages from chat session.
	// Optionally, adds a modifier message to the chat as well.
	var loadedMessages []openai.ChatCompletionMessage
	loadedMessages, err = c.loadChatMessages(isChat, chatID, modifier)
	if err != nil {
		return "", err
	}
	messages = append(messages, loadedMessages...)

	// Add prompt to messages
	var promptMessages []openai.ChatCompletionMessage
	promptMessages, err = c.createPromptMessages(prompt, input)
	if err != nil {
		return "", err
	}
	messages = append(messages, promptMessages...)
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
	// Retrieve the completion and print to the out writer. The received message is returned to save it to the chat and
	// to return it as a string (copy to clipboard).
	var receivedMessage openai.ChatCompletionMessage
	if c.config.GetBool("stream") {
		receivedMessage, err = c.retrieveChatCompletionStream(ctx, req)
	} else {
		receivedMessage, err = c.retrieveChatCompletion(ctx, req)
	}
	if err != nil {
		return "", err
	}
	// Set role of received message, if not set
	// This seems to be a bug in the OpenAI API for now
	if receivedMessage.Role == "" {
		receivedMessage.Role = openai.ChatMessageRoleAssistant
	}

	slog.Debug("Received message from OpenAI API")

	// If a session was provided, save received message to this chat
	if isChat {
		messages = append(messages, receivedMessage)
		if err = c.chatSessionManager.SaveSession(chatID, messages); err != nil {
			return "", err
		}
		slog.Debug("Saved chat session")
	}
	// Return received message
	return receivedMessage.Content, nil
}

func (c *OpenAIClient) loadChatMessages(isChat bool, chatID, modifier string) (messages []openai.ChatCompletionMessage, err error) {
	chatExists := false
	// Load existing chat messages
	if isChat {
		chatExists, err = c.chatSessionManager.SessionExists(chatID)
		if err != nil {
			return
		}
		if chatExists {
			var loadedMessages []openai.ChatCompletionMessage
			loadedMessages, err = c.chatSessionManager.GetSession(chatID)
			if err != nil {
				return
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
			return
		}
		if modifierPrompt != "" {
			messages = append(messages, openai.ChatCompletionMessage{
				Role:    openai.ChatMessageRoleSystem,
				Content: modifierPrompt,
			})
			slog.Debug("Added modifier message")
		}
	}
	return
}

func (c *OpenAIClient) createPromptMessages(prompts, input []string) (messages []openai.ChatCompletionMessage, err error) {
	if len(input) > 0 {
		// Request to the gpt-4-vision API
		slog.Warn("The GPT-4 Vision API is in beta and may not work as expected")

		var messageParts []openai.ChatMessagePart
		// Add prompt to message
		// We append the stdin as part of the prompt as a message part
		for _, p := range prompts {
			messageParts = append(messageParts, openai.ChatMessagePart{
				Type: openai.ChatMessagePartTypeText,
				Text: p,
			})
		}

		// Add images to messages
		for _, i := range input {
			// By default, assume that the input is a URL
			imageData := i

			// Check, if input is a file
			if !strings.HasPrefix(i, "http") || !strings.HasPrefix(i, "https") {
				// Input is a file, load image data
				imageData, err = c.buildImageFileData(i)
				if err != nil {
					return []openai.ChatCompletionMessage{}, err
				}
			}

			messageParts = append(messageParts, openai.ChatMessagePart{
				Type: openai.ChatMessagePartTypeImageURL,
				ImageURL: &openai.ChatMessageImageURL{
					URL: imageData,
				},
			})
		}

		messages = append(messages, openai.ChatCompletionMessage{
			Role:         openai.ChatMessageRoleUser,
			MultiContent: messageParts,
		})
	} else {
		// Normal prompt
		// We append the stdin as part of the prompt
		// This means we just add the prompt as a message
		for _, p := range prompts {
			messages = append(messages, openai.ChatCompletionMessage{
				Role:    openai.ChatMessageRoleUser,
				Content: p,
			})
		}
	}
	slog.Debug("Added prompt messages")
	return messages, nil
}

func (c *OpenAIClient) buildImageFileData(inputFile string) (imageData string, err error) {
	// Get image filetype
	var filetype string
	filetype, err = fs.GetImageFileType(inputFile)
	if err != nil {
		return
	}

	// Load image from file
	var b64Image string
	b64Image, err = fs.LoadBase64ImageFromFile(inputFile)
	if err != nil {
		return
	}

	imageData = fmt.Sprintf("data:%s;base64,%s", filetype, b64Image)
	return
}

func (c *OpenAIClient) retrieveChatCompletion(ctx context.Context, req openai.ChatCompletionRequest) (message openai.ChatCompletionMessage, err error) {
	var resp openai.ChatCompletionResponse
	resp, err = c.api.CreateChatCompletion(ctx, req)
	if err != nil {
		return
	}
	slog.Debug("Received response")
	message = resp.Choices[0].Message

	_, err = fmt.Fprintln(c.out, message.Content)
	if err != nil {
		return openai.ChatCompletionMessage{}, err
	}
	slog.Debug("Printed response")
	return
}

func (c *OpenAIClient) retrieveChatCompletionStream(ctx context.Context, req openai.ChatCompletionRequest) (openai.ChatCompletionMessage, error) {
	stream, err := c.api.CreateChatCompletionStream(ctx, req)
	if err != nil {
		return openai.ChatCompletionMessage{}, err
	}
	defer stream.Close()
	slog.Debug("Streaming response")

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
