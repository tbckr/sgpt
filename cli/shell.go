package cli

import (
	"strings"

	"github.com/spf13/cobra"
	"github.com/tbckr/sgpt/client"
	"github.com/tbckr/sgpt/modifier"
	"github.com/tbckr/sgpt/shell"
)

var shellArgs struct {
	model       string
	maxTokens   int
	temperature float64
	topP        float64
	execute     bool
	chatSession string
}

func shellCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "sh <prompt>",
		Short: "Query the openai models for a shell command",
		Long: strings.TrimSpace(`
Query a openai model for a shell command. The retrieved command can be executed at the same time.
The supported completion models can be listed via: "sgpt txt --help"
`),
		RunE: runShell,
		Args: cobra.ExactArgs(1),
	}
	fs := cmd.Flags()
	fs.StringVar(&shellArgs.model, "model", client.DefaultModel, "model name")
	fs.IntVar(&shellArgs.maxTokens, "max-tokens", 2048, "strict length of output (tokens)")
	fs.Float64Var(&shellArgs.temperature, "temperature", 0.2, "randomness of generated output")
	fs.Float64Var(&shellArgs.topP, "top-p", 0.9, "limits highest probable tokens")
	fs.BoolVar(&shellArgs.execute, "execute", false, "execute shell command")
	fs.StringVar(&shellArgs.chatSession, "chat", "", "use an existing chat session")
	return cmd
}

func runShell(cmd *cobra.Command, args []string) error {
	prompt, err := shell.GetInput(args)
	if err != nil {
		return err
	}
	options := client.CompletionOptions{
		Model:       shellArgs.model,
		MaxTokens:   shellArgs.maxTokens,
		Temperature: float32(shellArgs.temperature),
		TopP:        float32(shellArgs.topP),
		Modifier:    modifier.Shell,
		ChatSession: shellArgs.chatSession,
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
	response = strings.TrimSpace(response)
	return shell.ExecuteCommandWithConfirmation(cmd.Context(), response)
}
