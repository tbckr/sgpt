package cli

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"strings"

	"github.com/peterbourgon/ff/v3/ffcli"
	"github.com/sashabaranov/go-openai"
	"github.com/tbckr/sgpt"
)

const nilModifier = ""

var ErrMissingPrompt = errors.New("a prompt must be provided")

var textCmd = &ffcli.Command{
	Name:       "txt",
	ShortUsage: "",
	ShortHelp:  "",
	LongHelp:   strings.TrimSpace(``),
	Exec:       runText,
	FlagSet: (func() *flag.FlagSet {
		fs := newFlagSet("text")
		fs.StringVar(&textArgs.model, "model", "gpt-3.5-turbo", "GPT-3 model name")
		fs.IntVar(&textArgs.maxTokens, "max-tokens", 2048, "Strict length of output (tokens)")
		fs.Float64Var(&textArgs.temperature, "temperature", 1.0, "Randomness of generated output")
		fs.Float64Var(&textArgs.topP, "top-p", 1.0, "Limits highest probable tokens")
		return fs
	})(),
}

var textArgs struct {
	model       string
	maxTokens   int
	temperature float64
	topP        float64
}

func runText(ctx context.Context, args []string) error {
	// Check, if prompt was provided via command line
	if len(args) != 1 {
		return ErrMissingPrompt
	}
	prompt := args[0]

	options := sgpt.CompletionOptions{
		Model:       textArgs.model,
		MaxTokens:   textArgs.maxTokens,
		Temperature: float32(textArgs.temperature),
		TopP:        float32(textArgs.topP),
	}
	sgpt.ValidateOptions(&options)

	client, err := sgpt.CreateClient()
	if err != nil {
		return err
	}

	var response string
	if options.Model == openai.GPT3Dot5Turbo || options.Model == openai.GPT3Dot5Turbo0301 {
		response, err = sgpt.GetChatCompletion(ctx, client, options, prompt, nilModifier)
	} else {
		response, err = sgpt.GetCompletion(ctx, client, options, prompt, nilModifier)
	}
	if err != nil {
		return err
	}

	_, err = fmt.Fprint(stdout, response)
	return err
}
