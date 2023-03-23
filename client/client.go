package client

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"

	openai "github.com/sashabaranov/go-openai"
	"github.com/tbckr/sgpt/chat"
	"github.com/tbckr/sgpt/modifier"
)

const (
	envKeyOpenAIApi = "OPENAI_API_KEY"

	ImageURL  = "ImageURL"
	ImageData = "ImageData"
)

var (
	DefaultModel     = strings.Clone(openai.GPT3Dot5Turbo)
	DefaultImageSize = strings.Clone(openai.CreateImageSize256x256)

	ErrMissingAPIKey    = fmt.Errorf("%s env variable is not set", envKeyOpenAIApi)
	ErrChatNotSupported = errors.New("chat is not supported with this model")
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
		return nil, ErrMissingAPIKey
	}
	client := &Client{
		api: openai.NewClient(apiKey),
	}
	return client, nil
}

func (c *Client) validateCompletionOptions(options CompletionOptions) error {
	// curie has a max limit of 2048 for input and output
	if options.Model == openai.GPT3TextCurie001 && options.MaxTokens > 1024 {
		options.MaxTokens = 1024
		return fmt.Errorf("model %s must not have more than 1024 in total", openai.GPT3TextCurie001)
	}
	// A completion does not support chat
	if options.ChatSession != "" {
		return ErrChatNotSupported
	}
	return nil
}

func (c *Client) GetCompletion(ctx context.Context, options CompletionOptions, prompt string) (string, error) {
	if err := c.validateCompletionOptions(options); err != nil {
		return "", err
	}
	// TODO: handle this error
	modifierPrompt, _ := modifier.GetModifier(options.Modifier)
	// Do request
	req := openai.CompletionRequest{
		Prompt:      modifierPrompt + prompt,
		Model:       options.Model,
		MaxTokens:   options.MaxTokens,
		Temperature: options.Temperature,
		TopP:        options.TopP,
	}
	resp, err := c.api.CreateCompletion(ctx, req)
	if err != nil {
		return "", err
	}
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
		chatExists, err = chat.SessionExists(options.ChatSession)
		if err != nil {
			return "", err
		}
		if chatExists {
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
		// TODO: handle this error
		modifierPrompt, _ := modifier.GetChatModifier(options.Modifier)
		if modifierPrompt != "" {
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
	var resp openai.ChatCompletionResponse
	resp, err = c.api.CreateChatCompletion(ctx, req)
	if err != nil {
		return "", err
	}
	receivedMessage := resp.Choices[0].Message
	// Remove surrounding white spaces
	receivedMessage.Content = strings.TrimSpace(receivedMessage.Content)
	// If a session was provided, save received message to this chat
	if isChat {
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

	req := openai.ImageRequest{
		Prompt:         prompt,
		N:              options.Count,
		Size:           options.Size,
		ResponseFormat: responseFormat,
	}
	resp, err := c.api.CreateImage(ctx, req)
	if err != nil {
		return []string{}, err
	}

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
	if model == openai.GPT3Dot5Turbo || model == openai.GPT3Dot5Turbo0301 {
		return true
	}
	return false
}
