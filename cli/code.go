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

var codeArgs struct {
	model       string
	maxTokens   int
	temperature float64
	topP        float64
	chatSession string
}

func codeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "code <prompt>",
		Short: "Query the openai models for code-specific questions",
		Long: strings.TrimSpace(`
Query a openai model for code specific questions.
The supported completion models can be listed via: "sgpt txt --help"
`),
		RunE: runCode,
		Args: cobra.ExactArgs(1),
	}
	fs := cmd.Flags()
	fs.StringVarP(&codeArgs.model, "model", "m", openai.GPT3Dot5Turbo, "model name")
	fs.IntVarP(&codeArgs.maxTokens, "max-tokens", "s", 2048, "strict length of output (tokens)")
	fs.Float64VarP(&codeArgs.temperature, "temperature", "t", 0.8, "randomness of generated output")
	fs.Float64VarP(&codeArgs.topP, "top-p", "p", 0.2, "limits highest probable tokens")
	fs.StringVarP(&codeArgs.chatSession, "chat", "c", "", "use an existing chat session")
	return cmd
}

func runCode(cmd *cobra.Command, args []string) error {
	prompt, err := shell.GetPrompt(args)
	if err != nil {
		return err
	}

	options := sgpt.CompletionOptions{
		Model:       codeArgs.model,
		MaxTokens:   codeArgs.maxTokens,
		Temperature: float32(codeArgs.temperature),
		TopP:        float32(codeArgs.topP),
		Modifier:    modifier.ModifierCode,
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
