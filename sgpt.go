package sgpt

import (
	"context"
	"fmt"
	"os"
	"runtime"
	"strings"

	openai "github.com/sashabaranov/go-openai"
)

const (
	envKeyOpenAIApi = "OPENAI_API_KEY"
	envKeyShell     = "SHELL"
)

var (
	ErrMissingAPIKey = fmt.Errorf("%s env variable is not set", envKeyOpenAIApi)
	Version          = "dev"
)

type CompletionOptions struct {
	Model       string
	MaxTokens   int
	Temperature float32
	TopP        float32
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

func GetCompletion(ctx context.Context, client *openai.Client, options CompletionOptions, prompt, modifier string) (string, error) {
	var err error
	var modifierPrompt string
	switch modifier {
	case ModifierShell:
		modifierPrompt, err = completeShellModifier(completionModifierTemplate[ModifierShell])
	case ModifierCode:
		modifierPrompt, err = completionModifierTemplate[ModifierCode], nil
	case ModifierNil:
		modifierPrompt, err = "", nil
	default:
		return "", fmt.Errorf("unsupported modifier %s", modifier)
	}
	if err != nil {
		return "", err
	}
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

func GetChatCompletion(ctx context.Context, client *openai.Client, options CompletionOptions, prompt, modifier string) (string, error) {
	var err error
	var modifierPrompt string
	switch modifier {
	case ModifierShell:
		modifierPrompt, err = completeShellModifier(chatCompletionModifierTemplate[ModifierShell])
	case ModifierCode:
		modifierPrompt, err = chatCompletionModifierTemplate[ModifierCode], nil
	case ModifierNil:
		modifierPrompt, err = "", nil
	default:
		return "", fmt.Errorf("unsupported modifier %s", modifier)
	}
	if err != nil {
		return "", err
	}
	var messages []openai.ChatCompletionMessage
	if modifierPrompt != "" {
		messages = append(messages, openai.ChatCompletionMessage{
			Role:    openai.ChatMessageRoleSystem,
			Content: modifierPrompt,
		})
	}
	messages = append(messages, openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleUser,
		Content: prompt,
	})
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
	return strings.TrimSpace(resp.Choices[0].Message.Content), nil
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
