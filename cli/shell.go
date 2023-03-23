package cli

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/tbckr/sgpt/api"
	"github.com/tbckr/sgpt/modifiers"
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
	fs.StringVarP(&shellArgs.model, "model", "m", api.DefaultModel, "model name")
	fs.IntVarP(&shellArgs.maxTokens, "max-tokens", "s", 2048, "strict length of output (tokens)")
	fs.Float64VarP(&shellArgs.temperature, "temperature", "t", 0.2, "randomness of generated output")
	fs.Float64VarP(&shellArgs.topP, "top-p", "p", 0.9, "limits highest probable tokens")
	fs.BoolVarP(&shellArgs.execute, "execute", "e", false, "execute shell command")
	fs.StringVarP(&shellArgs.chatSession, "chat", "c", "", "use an existing chat session")
	return cmd
}

func runShell(cmd *cobra.Command, args []string) error {
	prompt, err := shell.GetInput(args)
	if err != nil {
		return err
	}
	options := api.CompletionOptions{
		Model:       shellArgs.model,
		MaxTokens:   shellArgs.maxTokens,
		Temperature: float32(shellArgs.temperature),
		TopP:        float32(shellArgs.topP),
		Modifier:    modifiers.Shell,
		ChatSession: shellArgs.chatSession,
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
	response = strings.TrimSpace(response)

	if shellArgs.execute {
		return shell.ExecuteCommandWithConfirmation(cmd.Context(), response)
	}

	_, err = fmt.Fprintln(stdout, response)
	return err
}
