package cli

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/peterbourgon/ff/v3/ffcli"
	"github.com/sashabaranov/go-openai"
	"github.com/tbckr/sgpt"
)

const shellModifier = "Provide only shell command as output."

var (
	ColorReset = "\033[0m"
	ColorRed   = "\033[31m"
)

var shellCmd = &ffcli.Command{
	Name:       "sh",
	ShortUsage: "",
	ShortHelp:  "",
	LongHelp:   strings.TrimSpace(``),
	Exec:       runShell,
	FlagSet: (func() *flag.FlagSet {
		fs := newFlagSet("shell")
		fs.StringVar(&textArgs.model, "model", "gpt-3.5-turbo", "GPT-3 model name")
		fs.IntVar(&textArgs.maxTokens, "max-tokens", 2048, "Strict length of output (tokens)")
		fs.Float64Var(&textArgs.temperature, "temperature", float64(1), "Randomness of generated output")
		fs.Float64Var(&textArgs.topP, "top-p", float64(1), "Limits highest probable tokens")
		fs.BoolVar(&shellArgs.execute, "execute", false, "Execute shell command")
		return fs
	})(),
}

var shellArgs struct {
	model       string
	maxTokens   int
	temperature float64
	topP        float64
	execute     bool
}

func runShell(ctx context.Context, args []string) error {
	// Check, if prompt was provided via command line
	if len(args) != 1 {
		return ErrMissingPrompt
	}
	prompt := args[0]

	// If default values where not changed, make it more accurate
	if shellArgs.temperature == float64(1) && shellArgs.topP == float64(1) {
		shellArgs.temperature = 0.2
		shellArgs.topP = 0.9
	}

	options := sgpt.CompletionOptions{
		Model:       shellArgs.model,
		MaxTokens:   shellArgs.maxTokens,
		Temperature: float32(shellArgs.temperature),
		TopP:        float32(shellArgs.topP),
	}
	sgpt.ValidateOptions(&options)

	client, err := sgpt.CreateClient()
	if err != nil {
		return err
	}

	var response string
	if options.Model == openai.GPT3Dot5Turbo || options.Model == openai.GPT3Dot5Turbo0301 {
		response, err = sgpt.GetChatCompletion(ctx, client, options, prompt, shellModifier)
	} else {
		response, err = sgpt.GetCompletion(ctx, client, options, prompt, shellModifier)
	}
	if err != nil {
		return err
	}

	fmt.Fprint(Stdout, ColorRed, response, ColorReset)

	if shellArgs.execute {
		var continueConfirmed bool
		continueConfirmed, err = getUserConfirmation()
		if err != nil {
			return err
		}
		if continueConfirmed {
			return executeShellCommand(ctx, response)
		}
	}
	return nil
}

func getUserConfirmation() (bool, error) {
	// Require user confirmation
	for {
		fmt.Fprint(Stdout, "Do you want to execute this command? (Y/n) ")
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
	cmd.Stdout = Stdout
	return cmd.Run()
}
