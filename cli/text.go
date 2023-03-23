package cli

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/tbckr/sgpt/api"
	"github.com/tbckr/sgpt/modifiers"
	"github.com/tbckr/sgpt/shell"
)

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
`, strings.Join(api.SupportedModels, ", "))),
		Args: cobra.ExactArgs(1),
		RunE: runText,
	}
	fs := cmd.Flags()
	fs.StringVarP(&textArgs.model, "model", "m", api.DefaultModel, "model name to use")
	fs.IntVarP(&textArgs.maxTokens, "max-tokens", "s", 2048, "strict length of output (tokens)")
	fs.Float64VarP(&textArgs.temperature, "temperature", "t", 1.0, "randomness of generated output")
	fs.Float64VarP(&textArgs.topP, "top-p", "p", 1.0, "limits highest probable tokens")
	fs.StringVarP(&textArgs.chatSession, "chat", "c", "", "use an existing chat session")
	return cmd
}

func runText(cmd *cobra.Command, args []string) error {
	prompt, err := shell.GetInput(args)
	if err != nil {
		return err
	}

	options := api.CompletionOptions{
		Model:       textArgs.model,
		MaxTokens:   textArgs.maxTokens,
		Temperature: float32(textArgs.temperature),
		TopP:        float32(textArgs.topP),
		Modifier:    modifiers.Nil,
		ChatSession: textArgs.chatSession,
	}

	var cli *api.Client
	cli, err = api.CreateClient()
	if err != nil {
		return err
	}

	var response string
	if api.IsChatModel(options.Model) {
		response, err = cli.GetChatCompletion(cmd.Context(), options, prompt)
	} else {
		response, err = cli.GetCompletion(cmd.Context(), options, prompt)
	}
	if err != nil {
		return err
	}
	_, err = fmt.Fprintln(stdout, response)
	return err
}
