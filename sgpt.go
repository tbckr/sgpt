package sgpt

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"

	openai "github.com/sashabaranov/go-openai"
	"github.com/spf13/pflag"
)

const EnvKey = "OPENAI_API_KEY"

var (
	Stdout     io.Writer = os.Stdout
	Stderr     io.Writer = os.Stderr
	ColorReset           = "\033[0m"
	ColorRed             = "\033[31m"
)

var config struct {
	apiKey      string
	prompt      string
	model       string
	maxTokens   int
	temperature float32
	topP        float32
	shell       bool
	execute     bool
	code        bool
}

func validateConfig() error {
	// Check, if prompt was provided via command line
	if pflag.NArg() != 1 {
		return errors.New("a prompt must be provided")
	}
	config.prompt = pflag.Arg(0)

	// Check, if api key was set
	apiKey, exists := os.LookupEnv(EnvKey)
	if !exists {
		return fmt.Errorf("%s is not set", EnvKey)
	}
	config.apiKey = apiKey

	// If default values where not changed, make it more accurate
	if config.temperature == float32(1) && config.topP == float32(1) {
		if config.shell {
			config.temperature = 0.2
			config.topP = 0.9
		} else if config.code {
			config.temperature = 0.8
			config.topP = 0.2
		}
	}

	// curie has a max limit of 2048 for input and output
	if config.model == "text-curie-001" && config.maxTokens == 2048 {
		config.maxTokens = 1024
	}
	return nil
}

func usage() error {
	if _, err := fmt.Fprintf(Stderr, "Usage of %s:\n", os.Args[0]); err != nil {
		return err
	}
	pflag.PrintDefaults()
	return nil
}

func print(a ...any) error {
	if _, err := fmt.Fprint(Stdout, a...); err != nil {
		return err
	}
	return nil
}

func requireUserConfirmation() (bool, error) {
	// Require user confirmation
	for {
		if err := print("Do you want to execute this command? (Y/n) "); err != nil {
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

func handleShell(ctx context.Context, response string) error {
	if err := print(ColorRed, response, ColorReset, "\n"); err != nil {
		return err
	}
	choiceConfirmed, err := requireUserConfirmation()
	if err != nil {
		return err
	}
	if choiceConfirmed {
		// Execute cmd from response text
		cmd := exec.CommandContext(ctx, "bash", "-c", response)
		cmd.Stdout = Stdout
		cmd.Stderr = Stderr
		return cmd.Run()
	}
	return nil
}

func Run() (err error) {
	pflag.StringVarP(&config.model, "model", "m", "gpt-3.5-turbo", "GPT-3 model name")
	pflag.IntVar(&config.maxTokens, "max-tokens", 2048, "Strict length of output (words)")
	pflag.Float32Var(&config.temperature, "temperature", float32(1), "Randomness of generated output")
	pflag.Float32Var(&config.topP, "top-p", float32(1), "Limits highest probable tokens")
	pflag.BoolVar(&config.code, "code", false, "Provide code as output")
	pflag.BoolVarP(&config.shell, "shell", "s", false, "Provide shell command as output")
	pflag.BoolVarP(&config.execute, "execute", "e", false, "Will execute --shell command")
	printUsage := pflag.BoolP("help", "h", false, "Usage overview")
	pflag.Parse()

	if *printUsage {
		return usage()
	}

	if err = validateConfig(); err != nil {
		return
	}

	// Create api client and do request
	client := openai.NewClient(config.apiKey)
	ctx := context.Background()

	var response string
	if config.model == openai.GPT3Dot5Turbo || config.model == openai.GPT3Dot5Turbo0301 {
		response, err = getChatCompletion(ctx, client)
	} else {
		response, err = getCompletion(ctx, client)
	}
	if err != nil {
		return
	}

	if config.shell {
		err = handleShell(ctx, response)
	} else {
		err = print(response, "\n")
	}
	return err
}
