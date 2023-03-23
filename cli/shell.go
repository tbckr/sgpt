package cli

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"

	openai "github.com/sashabaranov/go-openai"
	"github.com/spf13/cobra"
	"github.com/tbckr/sgpt/modifier"
	sgpt "github.com/tbckr/sgpt/openai"
	"github.com/tbckr/sgpt/shell"
)

const (
	shellFormat = "\033[31m" // color red
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
	fs.StringVar(&shellArgs.model, "model", openai.GPT3Dot5Turbo, "model name")
	fs.IntVar(&shellArgs.maxTokens, "max-tokens", 2048, "strict length of output (tokens)")
	fs.Float64Var(&shellArgs.temperature, "temperature", 0.2, "randomness of generated output")
	fs.Float64Var(&shellArgs.topP, "top-p", 0.9, "limits highest probable tokens")
	fs.BoolVar(&shellArgs.execute, "execute", false, "execute shell command")
	fs.StringVar(&shellArgs.chatSession, "chat", "", "use an existing chat session")
	return cmd
}

func runShell(cmd *cobra.Command, args []string) error {
	prompt, err := shell.GetPrompt(args)
	if err != nil {
		return err
	}

	options := sgpt.CompletionOptions{
		Model:       shellArgs.model,
		MaxTokens:   shellArgs.maxTokens,
		Temperature: float32(shellArgs.temperature),
		TopP:        float32(shellArgs.topP),
		Modifier:    modifier.Shell,
		ChatSession: shellArgs.chatSession,
	}
	if err = sgpt.ValidateCompletionOptions(options); err != nil {
		return err
	}

	var client *openai.Client
	client, err = sgpt.CreateAPIClient()
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

	if _, err = fmt.Fprintln(stdout, shellFormat, response, resetFormat); err != nil {
		return err
	}

	if shellArgs.execute {
		var ok bool
		ok, err = getUserConfirmation()
		if err != nil {
			return err
		}
		if ok {
			return executeShellCommand(cmd.Context(), response)
		}
	}
	return nil
}

func getUserConfirmation() (bool, error) {
	// Require user confirmation
	for {
		if _, err := fmt.Fprint(stdout, "Do you want to execute this command? (Y/n) "); err != nil {
			return false, err
		}
		reader := bufio.NewReader(os.Stdin)
		char, _, err := reader.ReadRune()
		if err != nil {
			return false, err
		}
		// 10 = enter
		if char == 10 || char == 'Y' || char == 'y' {
			return true, nil
		} else if char == 'N' || char == 'n' {
			return false, nil
		}
	}
}

func executeShellCommand(ctx context.Context, response string) error {
	// Execute cmd from response text
	cmd := exec.CommandContext(ctx, "bash", "-c", response)
	cmd.Stdout = stdout
	return cmd.Run()
}
