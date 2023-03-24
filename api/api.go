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
	"os"
	"strings"

	"github.com/sashabaranov/go-openai"
	jww "github.com/spf13/jwalterweatherman"
	"github.com/tbckr/sgpt/chat"
	"github.com/tbckr/sgpt/modifiers"
)

const (
	envKeyOpenAIApi = "OPENAI_API_KEY"

	ImageURL  = "ImageURL"
	ImageData = "ImageData"
)

var (
	DefaultModel     = strings.Clone(openai.GPT3Dot5Turbo)
	DefaultImageSize = strings.Clone(openai.CreateImageSize256x256)

	ErrMissingAPIKey        = fmt.Errorf("%s env variable is not set", envKeyOpenAIApi)
	ErrModelCurieSize       = fmt.Errorf("model %s must not have more than 1024 in total", openai.GPT3TextCurie001)
	ErrChatNotSupported     = errors.New("chat is not supported for this model")
	ErrModifierNotSupported = errors.New("modifier is not supported for this model")
)

var SupportedModels = []string{
	openai.GPT4,
	openai.GPT40314,
	openai.GPT432K,
	openai.GPT432K0314,
	openai.GPT3Dot5Turbo0301,
	openai.GPT3Dot5Turbo,
	openai.GPT3TextDavinci003,
	openai.GPT3TextDavinci002,
	openai.GPT3TextCurie001,
	openai.GPT3TextBabbage001,
	openai.GPT3TextAda001,
	openai.GPT3TextDavinci001,
	openai.GPT3DavinciInstructBeta,
	openai.GPT3Davinci,
	openai.GPT3CurieInstructBeta,
	openai.GPT3Curie,
	openai.GPT3Ada,
	openai.GPT3Babbage,
}

type Client struct {
	api *openai.Client
}

type CompletionOptions struct {
	Model       string
	MaxTokens   int
	Temperature float32
	TopP        float32
	Modifier    string
	ChatSession string
}

type ImageOptions struct {
	Count          int
	Size           string
	ResponseFormat string
}

func CreateClient() (*Client, error) {
	// Check, if api key was set
	apiKey, exists := os.LookupEnv(envKeyOpenAIApi)
	if !exists {
		jww.ERROR.Println(ErrMissingAPIKey)
		return nil, ErrMissingAPIKey
	}
	client := &Client{
		api: openai.NewClient(apiKey),
	}
	jww.DEBUG.Println("OpenAI API client successfully initialized")
	return client, nil
}

func (c *Client) validateCompletionOptions(options CompletionOptions) error {
	// curie has a max limit of 2048 for input and output
	if options.Model == openai.GPT3TextCurie001 && options.MaxTokens > 1024 {
		jww.ERROR.Println(ErrModelCurieSize)
		return ErrModelCurieSize
	}
	// A completion does not support chat
	if options.ChatSession != "" {
		jww.ERROR.Printf("Chat with model %s is not supported\n", options.Model)
		return ErrChatNotSupported
	}
	// A completion does not support modifiers
	if options.Modifier != "" {
		jww.ERROR.Println("Modifiers are not supported for not chat based models")
		return ErrModifierNotSupported
	}
	jww.DEBUG.Println("Completion options are valid")
	return nil
}

func (c *Client) GetCompletion(ctx context.Context, options CompletionOptions, prompt string) (string, error) {
	if err := c.validateCompletionOptions(options); err != nil {
		return "", err
	}
	// Do request
	req := openai.CompletionRequest{
		Prompt:      prompt,
		Model:       options.Model,
		MaxTokens:   options.MaxTokens,
		Temperature: options.Temperature,
		TopP:        options.TopP,
	}
	jww.DEBUG.Printf("Completion request: %+v\n", req)
	resp, err := c.api.CreateCompletion(ctx, req)
	if err != nil {
		return "", err
	}
	jww.DEBUG.Printf("Completion response: %+v\n", resp)
	return strings.TrimSpace(resp.Choices[0].Text), nil
}

func (c *Client) GetChatCompletion(ctx context.Context, options CompletionOptions, prompt string) (string, error) {
	var err error
	var messages []openai.ChatCompletionMessage

	// Evaluate base decision factors
	isChat := options.ChatSession != ""
	chatExists := false

	// Load existing chat messages
	if isChat {
		jww.DEBUG.Println("Loading existing chat session ", options.ChatSession)
		chatExists, err = chat.SessionExists(options.ChatSession)
		if err != nil {
			return "", err
		}
		if chatExists {
			jww.DEBUG.Println("Chat session exists, loading messages")
			var loadedMessages []openai.ChatCompletionMessage
			loadedMessages, err = chat.GetSession(options.ChatSession)
			if err != nil {
				return "", err
			}
			messages = append(messages, loadedMessages...)
		}
	}

	// If this message is not part of a chat
	// OR
	// if this is the initial message of a chat,
	// then add modifier message
	if !isChat || (isChat && !chatExists) {
		var modifierPrompt string
		modifierPrompt, err = modifiers.GetChatModifier(options.Modifier)
		if err != nil {
			return "", err
		}
		if modifierPrompt != "" {
			jww.DEBUG.Println("Adding modifier message")
			messages = append(messages, openai.ChatCompletionMessage{
				Role:    openai.ChatMessageRoleSystem,
				Content: modifierPrompt,
			})
		}
	}

	// Add prompt to messages
	messages = append(messages, openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleUser,
		Content: prompt,
	})

	// Do request
	req := openai.ChatCompletionRequest{
		Messages:    messages,
		Model:       options.Model,
		MaxTokens:   options.MaxTokens,
		Temperature: options.Temperature,
		TopP:        options.TopP,
	}
	jww.DEBUG.Printf("Chat completion request: %+v\n", req)
	var resp openai.ChatCompletionResponse
	resp, err = c.api.CreateChatCompletion(ctx, req)
	if err != nil {
		jww.DEBUG.Println("Chat completion request failed")
		return "", err
	}
	jww.DEBUG.Printf("Chat completion response: %+v\n", resp)
	receivedMessage := resp.Choices[0].Message
	// Remove surrounding white spaces
	receivedMessage.Content = strings.TrimSpace(receivedMessage.Content)
	// If a session was provided, save received message to this chat
	if isChat {
		jww.DEBUG.Println("Saving received message to chat session")
		messages = append(messages, receivedMessage)
		if err = chat.SaveSession(options.ChatSession, messages); err != nil {
			return "", err
		}
	}
	// Return received message
	return receivedMessage.Content, nil
}

// GetImage creates an image via the DALLE API. It either returns a URL to the image or the image data based on the
// provided ResponseFormat in the options.
func (c *Client) GetImage(ctx context.Context, options ImageOptions, prompt string) ([]string, error) {
	var responseFormat string
	switch options.ResponseFormat {
	case ImageData:
		responseFormat = openai.CreateImageResponseFormatB64JSON
	case ImageURL:
	default: // defaulting to URL
		responseFormat = openai.CreateImageResponseFormatURL
	}
	jww.DEBUG.Printf("Image response format: %s\n", responseFormat)

	req := openai.ImageRequest{
		Prompt:         prompt,
		N:              options.Count,
		Size:           options.Size,
		ResponseFormat: responseFormat,
	}
	jww.DEBUG.Printf("Image request: %+v\n", req)
	resp, err := c.api.CreateImage(ctx, req)
	if err != nil {
		jww.DEBUG.Println("Image request failed")
		return []string{}, err
	}
	jww.DEBUG.Printf("Image response: %+v\n", resp)

	var imageData []string
	for _, image := range resp.Data {
		if responseFormat == openai.CreateImageResponseFormatURL {
			imageData = append(imageData, strings.TrimSpace(image.URL))
		} else if responseFormat == openai.CreateImageResponseFormatB64JSON {
			imageData = append(imageData, strings.TrimSpace(image.B64JSON))
		}
	}
	return imageData, nil
}

func IsChatModel(model string) bool {
	switch model {
	case openai.GPT3Dot5Turbo, openai.GPT3Dot5Turbo0301,
		openai.GPT4, openai.GPT432K,
		openai.GPT40314, openai.GPT432K0314:
		jww.DEBUG.Printf("Model %s is a chat model\n", model)
		return true
	default:
		jww.DEBUG.Printf("Model %s is not a chat model\n", model)
		return false
	}
}
