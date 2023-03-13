package cli

import (
	"context"
	"flag"
	"fmt"
	"strings"

	"github.com/peterbourgon/ff/v3/ffcli"
	"github.com/sashabaranov/go-openai"
	"github.com/tbckr/sgpt"
)

const codeModifier = "Provide only code as output."

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
		fs.StringVar(&textArgs.model, "model", "gpt-3.5-turbo", "GPT-3 model name")
		fs.IntVar(&textArgs.maxTokens, "max-tokens", 2048, "Strict length of output (tokens)")
		fs.Float64Var(&textArgs.temperature, "temperature", 0.8, "Randomness of generated output")
		fs.Float64Var(&textArgs.topP, "top-p", 0.2, "Limits highest probable tokens")
		return fs
	})(),
}

var codeArgs struct {
	model       string
	maxTokens   int
	temperature float64
	topP        float64
}

func runCode(ctx context.Context, args []string) error {
	// Check, if prompt was provided via command line
	if len(args) != 1 {
		return ErrMissingPrompt
	}
	prompt := args[0]

	options := sgpt.CompletionOptions{
		Model:       codeArgs.model,
		MaxTokens:   codeArgs.maxTokens,
		Temperature: float32(codeArgs.temperature),
		TopP:        float32(codeArgs.topP),
	}
	sgpt.ValidateCompletionOptions(&options)

	client, err := sgpt.CreateClient()
	if err != nil {
		return err
	}

	var response string
	if options.Model == openai.GPT3Dot5Turbo || options.Model == openai.GPT3Dot5Turbo0301 {
		response, err = sgpt.GetChatCompletion(ctx, client, options, prompt, codeModifier)
	} else {
		response, err = sgpt.GetCompletion(ctx, client, options, prompt, codeModifier)
	}
	if err != nil {
		return err
	}
	_, err = fmt.Fprint(stdout, response)
	return err
}
