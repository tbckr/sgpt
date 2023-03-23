package cli

import (
	"fmt"
	"strings"

	openai "github.com/sashabaranov/go-openai"
	"github.com/spf13/cobra"
	"github.com/tbckr/sgpt/modifier"
	sgpt "github.com/tbckr/sgpt/openai"
	"github.com/tbckr/sgpt/shell"
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
}

var textArgs struct {
	model       string
	maxTokens   int
	temperature float64
	topP        float64
	chatSession string
}

func textCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "txt <prompt>",
		Short: "Query the different openai models for a text completion",
		Long: strings.TrimSpace(fmt.Sprintf(`
Query a openai model for a text completion. The following models are supported: %s.
`, strings.Join(openaiModels, ", "))),
		Args: cobra.ExactArgs(1),
		RunE: runText,
	}
	fs := cmd.Flags()
	fs.StringVarP(&textArgs.model, "model", "m", openai.GPT3Dot5Turbo, "model name to use")
	fs.IntVarP(&textArgs.maxTokens, "max-tokens", "s", 2048, "strict length of output (tokens)")
	fs.Float64VarP(&textArgs.temperature, "temperature", "t", 1.0, "randomness of generated output")
	fs.Float64VarP(&textArgs.topP, "top-p", "p", 1.0, "limits highest probable tokens")
	fs.StringVarP(&textArgs.chatSession, "chat", "c", "", "use an existing chat session")
	return cmd
}

func runText(cmd *cobra.Command, args []string) error {
	prompt, err := shell.GetPrompt(args)
	if err != nil {
		return err
	}

	options := sgpt.CompletionOptions{
		Model:       textArgs.model,
		MaxTokens:   textArgs.maxTokens,
		Temperature: float32(textArgs.temperature),
		TopP:        float32(textArgs.topP),
		Modifier:    modifier.ModifierNil,
		ChatSession: textArgs.chatSession,
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
		response, err = sgpt.GetChatCompletion(cmd.Context(), client, options, prompt)
	} else {
		response, err = sgpt.GetCompletion(cmd.Context(), client, options, prompt)
	}
	if err != nil {
		return err
	}
	_, err = fmt.Fprintln(stdout, response)
	return err
}
