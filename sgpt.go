package sgpt

import (
	"context"
	"fmt"
	"os"
	"strings"

	openai "github.com/sashabaranov/go-openai"
)

const envKey = "OPENAI_API_KEY"

var (
	ErrMissingAPIKey = fmt.Errorf("%s env variable is not set", envKey)
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
}

func CreateClient() (*openai.Client, error) {
	// Check, if api key was set
	apiKey, exists := os.LookupEnv(envKey)
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
	if modifier != "" {
		prompt = prompt + ". " + modifier
	}
	req := openai.CompletionRequest{
		Prompt:      prompt,
		Model:       options.Model,
		MaxTokens:   options.MaxTokens,
		Temperature: options.Temperature,
		TopP:        options.TopP,
	}
	resp, err := client.CreateCompletion(ctx, req)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(resp.Choices[0].Text), nil
}

func GetChatCompletion(ctx context.Context, client *openai.Client, options CompletionOptions, prompt, modifier string) (string, error) {
	var messages []openai.ChatCompletionMessage
	if modifier != "" {
		messages = append(messages, openai.ChatCompletionMessage{
			Role:    openai.ChatMessageRoleSystem,
			Content: modifier,
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
	resp, err := client.CreateChatCompletion(ctx, req)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(resp.Choices[0].Message.Content), nil
}

func ValidateImageOptions(options ImageOptions) error {
	// curie has a max limit of 2048 for input and output
	//if options.Model == "text-curie-001" && (*options).MaxTokens == 2048 {
	//	(*options).MaxTokens = 1024
	//}
	return nil
}

func GetImage(ctx context.Context, client *openai.Client, options ImageOptions, prompt string) ([]string, error) {
	req := openai.ImageRequest{
		Prompt: prompt,
		N:      options.Count,
	}
	resp, err := client.CreateImage(ctx, req)
	if err != nil {
		return []string{}, err
	}
	var imageUrls []string
	for _, image := range resp.Data {
		imageUrls = append(imageUrls, strings.TrimSpace(image.URL))
	}
	return imageUrls, nil
}
