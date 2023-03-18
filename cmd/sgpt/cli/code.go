package cli

import (
	"context"
	"flag"
	"fmt"
	"strings"

	"github.com/peterbourgon/ff/v3/ffcli"
	"github.com/sashabaranov/go-openai"
	"github.com/tbckr/sgpt"
	"github.com/tbckr/sgpt/internal/shell"
)

var codeCmd = &ffcli.Command{
	Name:       "code",
	ShortUsage: "sgpt code [command flags] <prompt>",
	ShortHelp:  "Query the openai models for code-specific questions.",
	LongHelp: strings.TrimSpace(`
Query a openai model for code specific questions.
The supported completion models can be listed via: "sgpt txt --help"
`),
	Exec: runCode,
	FlagSet: (func() *flag.FlagSet {
		fs := newFlagSet("code")
		fs.StringVar(&codeArgs.model, "model", openai.GPT3Dot5Turbo, "GPT-3 model name")
		fs.IntVar(&codeArgs.maxTokens, "max-tokens", 2048, "Strict length of output (tokens)")
		fs.Float64Var(&codeArgs.temperature, "temperature", 0.8, "Randomness of generated output")
		fs.Float64Var(&codeArgs.topP, "top-p", 0.2, "Limits highest probable tokens")
		fs.StringVar(&codeArgs.chatSession, "chat", "", "Use an existing chat session")
		return fs
	})(),
}

var codeArgs struct {
	model       string
	maxTokens   int
	temperature float64
	topP        float64
	chatSession string
}

func runCode(ctx context.Context, args []string) error {
	prompt, err := shell.GetPrompt(args)
	if err != nil {
		return err
	}

	options := sgpt.CompletionOptions{
		Model:       codeArgs.model,
		MaxTokens:   codeArgs.maxTokens,
		Temperature: float32(codeArgs.temperature),
		TopP:        float32(codeArgs.topP),
		Modifier:    sgpt.ModifierCode,
		ChatSession: codeArgs.chatSession,
	}
	if err = sgpt.ValidateCompletionOptions(options); err != nil {
		return err
	}

	var client *openai.Client
	client, err = sgpt.CreateClient()
	if err != nil {
		return err
	}

	var response string
	if options.Model == openai.GPT3Dot5Turbo || options.Model == openai.GPT3Dot5Turbo0301 {
		response, err = sgpt.GetChatCompletion(ctx, client, options, prompt)
	} else {
		response, err = sgpt.GetCompletion(ctx, client, options, prompt)
	}
	if err != nil {
		return err
	}
	_, err = fmt.Fprintln(stdout, response)
	return err
}
