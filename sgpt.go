package sgpt

import (
	"context"
	"errors"
	"fmt"
	"os"
	"runtime"
	"strings"

	"github.com/tbckr/sgpt/internal/chat"

	openai "github.com/sashabaranov/go-openai"
)

const (
	envKeyOpenAIApi = "OPENAI_API_KEY"
	envKeyShell     = "SHELL"
)

var (
	ErrMissingAPIKey       = fmt.Errorf("%s env variable is not set", envKeyOpenAIApi)
	ErrUnsupportedModifier = errors.New("unsupported modifier")
	ErrChatNotSupported    = errors.New("chat is not supported with this model")
	Version                = "dev"
)

type CompletionOptions struct {
	Model       string
	MaxTokens   int
	Temperature float32
	TopP        float32
	Modifier    string
	ChatSession string
}

type ImageOptions struct {
	Count int
	Size  string
}

func CreateClient() (*openai.Client, error) {
	// Check, if api key was set
	apiKey, exists := os.LookupEnv(envKeyOpenAIApi)
	if !exists {
		return nil, ErrMissingAPIKey
	}
	return openai.NewClient(apiKey), nil
}

func ValidateCompletionOptions(options CompletionOptions) error {
	// curie has a max limit of 2048 for input and output
	if options.Model == openai.GPT3TextCurie001 && options.MaxTokens > 1024 {
		options.MaxTokens = 1024
		return fmt.Errorf("model %s must not have more than 1024 in total", openai.GPT3TextCurie001)
	}
	return nil
}

func GetCompletion(ctx context.Context, client *openai.Client, options CompletionOptions, prompt string) (string, error) {
	var err error

	// A completion does not support chat
	if options.ChatSession != "" {
		return "", ErrChatNotSupported
	}

	// Add modifier
	var modifierPrompt string
	switch options.Modifier {
	case ModifierShell:
		modifierPrompt, err = completeShellModifier(completionModifierTemplate[ModifierShell])
	case ModifierCode:
		modifierPrompt, err = completionModifierTemplate[ModifierCode], nil
	case ModifierNil:
		modifierPrompt, err = "", nil
	default:
		return "", ErrUnsupportedModifier
	}
	if err != nil {
		return "", err
	}

	// Do request
	req := openai.CompletionRequest{
		Prompt:      modifierPrompt + prompt,
		Model:       options.Model,
		MaxTokens:   options.MaxTokens,
		Temperature: options.Temperature,
		TopP:        options.TopP,
	}
	var resp openai.CompletionResponse
	resp, err = client.CreateCompletion(ctx, req)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(resp.Choices[0].Text), nil
}

func GetChatCompletion(ctx context.Context, client *openai.Client, options CompletionOptions, prompt string) (string, error) {
	var err error
	var messages []openai.ChatCompletionMessage

	// Load existing chat messages
	isChat := options.ChatSession != ""
	chatExists := false
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
		var modifierPrompt string
		switch options.Modifier {
		case ModifierShell:
			modifierPrompt, err = completeShellModifier(chatCompletionModifierTemplate[ModifierShell])
		case ModifierCode:
			modifierPrompt, err = chatCompletionModifierTemplate[ModifierCode], nil
		case ModifierNil:
			modifierPrompt, err = "", nil
		default:
			return "", ErrUnsupportedModifier
		}
		if err != nil {
			return "", err
		}

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
	resp, err = client.CreateChatCompletion(ctx, req)
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

func GetImage(ctx context.Context, client *openai.Client, options ImageOptions, prompt, responseFormat string) ([]string, error) {
	req := openai.ImageRequest{
		Prompt:         prompt,
		N:              options.Count,
		Size:           options.Size,
		ResponseFormat: responseFormat,
	}
	resp, err := client.CreateImage(ctx, req)
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

func completeShellModifier(template string) (string, error) {
	operatingSystem := runtime.GOOS
	shell, ok := os.LookupEnv(envKeyShell)
	// fallback to manually set shell
	if !ok {
		if operatingSystem == "windows" {
			shell = "powershell"
		} else if operatingSystem == "linux" {
			shell = "bash"
		} else if operatingSystem == "darwin" {
			shell = "zsh"
		} else {
			return "", fmt.Errorf("unsupported os %s", operatingSystem)
		}
	}
	return fmt.Sprintf(template, shell, operatingSystem, shell, operatingSystem, shell), nil
}
