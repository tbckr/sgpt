package sgpt

import (
	"bufio"
	"context"
	"errors"
	"flag"
	"fmt"
	openai "github.com/sashabaranov/go-openai"
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
	model       = flag.String("model", "gpt-3.5-turbo", "GPT-3 model name")
	maxTokens   = flag.Int("max-tokens", 2048, "Strict length of output (words)")
	temperature = flag.Float64("temperature", float64(1), "Randomness of generated output")
	topP        = flag.Float64("top-p", float64(1), "Limits highest probable tokens")
	shell       bool
	execute     bool
	code        = flag.Bool("code", false, "Provide code as output")
)

type configStruct struct {
	prompt      string
	model       string
	maxTokens   int
	temperature float32
	topP        float32
	shell       bool
	execute     bool
	code        bool
}

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

func createConfig(prompt string) configStruct {
	// If default values where not changed, make it more accurate
	if *temperature == float64(1) && *temperature == *topP {
		if shell {
			*temperature = 0.2
			*topP = 0.9
		} else if *code {
			*temperature = 0.8
			*topP = 0.2
		}
	}

	// curie has a max limit of 2048 for input and output
	if *model == "text-curie-001" && *maxTokens == 2048 {
		*maxTokens = 1024
	}

	return configStruct{
		prompt:      prompt,
		model:       *model,
		maxTokens:   *maxTokens,
		temperature: float32(*temperature),
		topP:        float32(*topP),
		shell:       shell,
		execute:     execute,
		code:        *code,
	}
}

func getCompletion(ctx context.Context, client *openai.Client, config configStruct) (string, error) {
	var prompt string
	if config.shell {
		// add directive for shell to prompt
		prompt = config.prompt + ". Provide only shell command as output."
	} else if config.code {
		// add directive for code to prompt
		prompt = config.prompt + ". Provide only code as output."
	}

	req := openai.CompletionRequest{
		Model:       config.model,
		MaxTokens:   config.maxTokens,
		Prompt:      prompt,
		Temperature: config.temperature,
		TopP:        config.topP,
	}

	resp, err := client.CreateCompletion(ctx, req)
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(resp.Choices[0].Text), nil
}

func getChatCompletion(ctx context.Context, client *openai.Client, config configStruct) (string, error) {
	var messages []openai.ChatCompletionMessage
	if config.shell {
		messages = append(messages, openai.ChatCompletionMessage{
			Role:    openai.ChatMessageRoleSystem,
			Content: "Provide only shell command as output.",
		})
	} else if config.code {
		messages = append(messages, openai.ChatCompletionMessage{
			Role:    openai.ChatMessageRoleSystem,
			Content: "Provide only code as output.",
		})
	}
	messages = append(messages, openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleUser,
		Content: config.prompt,
	})

	req := openai.ChatCompletionRequest{
		Model:       config.model,
		MaxTokens:   config.maxTokens,
		Messages:    messages,
		Temperature: config.temperature,
		TopP:        config.topP,
	}

	resp, err := client.CreateChatCompletion(ctx, req)
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(resp.Choices[0].Message.Content), nil
}

func Run() error {
	flag.Parse()

	// Check, if prompt was provided via command line
	if flag.NArg() != 1 {
		return errors.New("A prompt must be provided")
	}

	// Check, if api key was set
	apiKey, exists := os.LookupEnv(EnvKey)
	if !exists {
		return fmt.Errorf("%s is not set", EnvKey)
	}

	// Create config with flags
	config := createConfig(flag.Arg(0))

	// Create api client and do request
	client := openai.NewClient(apiKey)
	ctx := context.Background()

	var response string
	var err error
	if config.model == openai.GPT3Dot5Turbo || config.model == openai.GPT3Dot5Turbo0301 {
		response, err = getChatCompletion(ctx, client, config)
	} else {
		response, err = getCompletion(ctx, client, config)
	}
	if err != nil {
		return err
	}

	// Execute response in shell, if flags are provided
	if config.shell && config.execute {
		fmt.Println(ColorRed + response + ColorReset)

		// Require user confirmation
		for {
			fmt.Print("Do you want to execute this command? (Y/n) ")
			reader := bufio.NewReader(os.Stdin)
			char, _, err := reader.ReadRune()
			if err != nil {
				return err
			}
			// 10 = enter
			if char == 10 || char == 'Y' || char == 'y' {
				break
			} else if char == 'N' || char == 'n' {
				return nil
			}
		}

		// Execute cmd from response text
		cmd := exec.CommandContext(ctx, "bash", "-c", response)
		cmd.Stdout = os.Stdout
		if err := cmd.Run(); err != nil {
			return err
		}

	} else {
		fmt.Println(response)
	}

	return nil
}
