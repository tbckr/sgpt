package sgpt

import (
	"bufio"
	"context"
	"errors"
	"flag"
	"fmt"
	gogpt "github.com/sashabaranov/go-gpt3"
	"os"
	"os/exec"
	"runtime"
	"strings"
)

const EnvKey = "OPENAI_API_KEY"

var (
	ColorReset = "\033[0m"
	ColorRed   = "\033[31m"
)

var (
	model       = flag.String("model", "text-davinci-003", "GPT-3 model name")
	maxTokens   = flag.Int("max-tokens", 2048, "Strict length of output (words)")
	temperature = flag.Float64("temperature", float64(1), "Randomness of generated output")
	topP        = flag.Float64("top-p", float64(1), "Limits highest probable tokens")
	shell       bool
	execute     bool
	code        = flag.Bool("code", false, "Provide code as output")
)

func init() {
	if runtime.GOOS == "windows" {
		ColorReset = ""
		ColorRed = ""
	}

	const (
		shellDefault   = false
		shellUsage     = "Provide shell command as output"
		executeDefault = false
		executeUsage   = "Will execute --shell command"
	)

	flag.BoolVar(&shell, "shell", shellDefault, shellUsage)
	flag.BoolVar(&shell, "s", shellDefault, shellUsage)
	flag.BoolVar(&execute, "execute", executeDefault, executeUsage)
	flag.BoolVar(&execute, "e", executeDefault, executeUsage)
}

func Run() error {
	flag.Parse()

	// TODO: retrieve prompt via stdin

	var prompt = ""

	// Check, if prompt was provided via command line
	if prompt == "" && flag.NArg() != 1 {
		return errors.New("A prompt must be provided")
	}
	prompt = flag.Arg(0)

	if shell {
		// If default values where not changed, make it more accurate
		if *temperature == float64(1) && *temperature == *topP {
			*temperature = 0.2
			*topP = 0.9
			// add directive for shell to prompt
			prompt += ". Provide only shell command as output."
		}
	} else if *code {
		// If default values where not changed, make it more accurate
		if *temperature == float64(1) && *temperature == *topP {
			*temperature = 0.8
			*topP = 0.2
			// add directive for code to prompt
			prompt += ". Provide only code as output."
		}
	}

	// curie has a max limit of 2048 for input and output
	if *model == "text-curie-001" && *maxTokens == 2048 {
		*maxTokens = 1024
	}

	// Check if api key was set
	apiKey, exists := os.LookupEnv(EnvKey)
	if !exists {
		return fmt.Errorf("%s is not set", EnvKey)
	}

	// Create api client and do request
	c := gogpt.NewClient(apiKey)
	ctx := context.Background()

	req := gogpt.CompletionRequest{
		Model:       *model,
		MaxTokens:   *maxTokens,
		Prompt:      prompt,
		Temperature: float32(*temperature),
	}
	resp, err := c.CreateCompletion(ctx, req)
	if err != nil {
		return err
	}
	responseText := strings.TrimSpace(resp.Choices[0].Text)

	// Execute response in shell, if flags are provided
	if shell && execute {
		fmt.Println(ColorRed + responseText + ColorReset)
		fmt.Print("Do you want to execute this command? (Y/n) ")

		// Require user confirmation
		for {
			reader := bufio.NewReader(os.Stdin)
			char, size, err := reader.ReadRune()
			if err != nil {
				return err
			}
			if size == 1 || char == 'Y' || char == 'y' {
				break
			} else if char == 'N' || char == 'n' {
				return nil
			}
		}

		// Execute cmd from response text
		cmd := exec.CommandContext(ctx, "bash", "-c", responseText)
		cmd.Stdout = os.Stdout
		if err := cmd.Run(); err != nil {
			return err
		}
	} else {
		fmt.Println(responseText)
	}

	return nil
}
