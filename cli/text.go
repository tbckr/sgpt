package cli

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/tbckr/sgpt/client"
	"github.com/tbckr/sgpt/modifier"
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
`, strings.Join(client.SupportedModels, ", "))),
		Args: cobra.ExactArgs(1),
		RunE: runText,
	}
	fs := cmd.Flags()
	fs.StringVarP(&textArgs.model, "model", "m", client.DefaultModel, "model name to use")
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

	options := client.CompletionOptions{
		Model:       textArgs.model,
		MaxTokens:   textArgs.maxTokens,
		Temperature: float32(textArgs.temperature),
		TopP:        float32(textArgs.topP),
		Modifier:    modifier.Nil,
		ChatSession: textArgs.chatSession,
	}

	var cli *client.Client
	cli, err = client.CreateClient()
	if err != nil {
		return err
	}

	var response string
	if client.IsChatModel(options.Model) {
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
