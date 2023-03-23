package cli

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/tbckr/sgpt/client"
	"github.com/tbckr/sgpt/modifier"
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
	fs.StringVarP(&codeArgs.model, "model", "m", client.DefaultModel, "model name")
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

	options := client.CompletionOptions{
		Model:       codeArgs.model,
		MaxTokens:   codeArgs.maxTokens,
		Temperature: float32(codeArgs.temperature),
		TopP:        float32(codeArgs.topP),
		Modifier:    modifier.Code,
		ChatSession: codeArgs.chatSession,
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
