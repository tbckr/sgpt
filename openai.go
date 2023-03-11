package sgpt

import (
	"context"
	"strings"

	"github.com/sashabaranov/go-openai"
)

func getCompletion(ctx context.Context, client *openai.Client) (string, error) {
	var prompt string
	if config.shell {
		// add directive for shell to prompt
		prompt = config.prompt + ". Provide only shell command as output."
	} else if config.code {
		// add directive for code to prompt
		prompt = config.prompt + ". Provide only code as output."
	}

	req := openai.CompletionRequest{
		Model:       config.model,
		MaxTokens:   config.maxTokens,
		Prompt:      prompt,
		Temperature: config.temperature,
		TopP:        config.topP,
	}

	resp, err := client.CreateCompletion(ctx, req)
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(resp.Choices[0].Text), nil
}

func getChatCompletion(ctx context.Context, client *openai.Client) (string, error) {
	var messages []openai.ChatCompletionMessage
	if config.shell {
		messages = append(messages, openai.ChatCompletionMessage{
			Role:    openai.ChatMessageRoleSystem,
			Content: "Provide only shell command as output.",
		})
	} else if config.code {
		messages = append(messages, openai.ChatCompletionMessage{
			Role:    openai.ChatMessageRoleSystem,
			Content: "Provide only code as output.",
		})
	}
	messages = append(messages, openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleUser,
		Content: config.prompt,
	})

	req := openai.ChatCompletionRequest{
		Model:       config.model,
		MaxTokens:   config.maxTokens,
		Messages:    messages,
		Temperature: config.temperature,
		TopP:        config.topP,
	}

	resp, err := client.CreateChatCompletion(ctx, req)
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(resp.Choices[0].Message.Content), nil
}
