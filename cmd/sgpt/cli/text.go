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

var openaiModels = []string{
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

	openai.CodexCodeDavinci002,
	openai.CodexCodeCushman001,
	openai.CodexCodeDavinci001,
}

var textCmd = &ffcli.Command{
	Name:       "txt",
	ShortUsage: "sgpt txt [command flags] <prompt>",
	ShortHelp:  "Query the different openai models for a text completion.",
	LongHelp: strings.TrimSpace(fmt.Sprintf(`
Query a openai model for a text completion. The following models are supported:
- %s
`, strings.Join(openaiModels, "\n- "))),
	Exec: runText,
	FlagSet: (func() *flag.FlagSet {
		fs := newFlagSet("text")
		fs.StringVar(&textArgs.model, "model", openai.GPT3Dot5Turbo, "GPT-3 model name")
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
	prompt, err := shell.GetPrompt(args)
	if err != nil {
		return err
	}

	options := sgpt.CompletionOptions{
		Model:       textArgs.model,
		MaxTokens:   textArgs.maxTokens,
		Temperature: float32(textArgs.temperature),
		TopP:        float32(textArgs.topP),
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
		response, err = sgpt.GetChatCompletion(ctx, client, options, prompt, sgpt.ModifierNil)
	} else {
		response, err = sgpt.GetCompletion(ctx, client, options, prompt, sgpt.ModifierNil)
	}
	if err != nil {
		return err
	}
	_, err = fmt.Fprintln(stdout, response)
	return err
}
