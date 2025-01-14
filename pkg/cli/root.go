// Copyright (c) 2023 Tim <tbckr>
//
// Permission is hereby granted, free of charge, to any person obtaining a copy of
// this software and associated documentation files (the "Software"), to deal in
// the Software without restriction, including without limitation the rights to
// use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of
// the Software, and to permit persons to whom the Software is furnished to do so,
// subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS
// FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR
// COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER
// IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN
// CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
//
// SPDX-License-Identifier: MIT

package cli

import (
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"strings"

	"path/filepath"

	"github.com/tbckr/sgpt/v2/pkg/api"
	"github.com/tbckr/sgpt/v2/pkg/chat"
	"github.com/tbckr/sgpt/v2/pkg/fs"
	"github.com/tbckr/sgpt/v2/pkg/shell"
	
	"github.com/sashabaranov/go-openai"

	"github.com/atotto/clipboard"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type rootCmd struct {
	cmd  *cobra.Command
	exit func(int)

	chat            string
	execute         bool
	copyToClipboard bool
	input           []string

	verbose bool
}

// We have to create our own viper config, because a singleton does not work in test mode
func createViperConfig() (*viper.Viper, error) {
	config := viper.New()
	appConfigPath, err := fs.GetAppConfigPath()
	if err != nil {
		return nil, err
	}
	config.AddConfigPath(appConfigPath)
	config.SetConfigName("config")
	config.SetConfigType("yaml")
	return config, nil
}

func Execute(args []string) {
	viperConfig, err := createViperConfig()
	if err != nil {
		slog.Error("Failed to create viper config", "error", err)
		os.Exit(1)
	}
	newRootCmd(os.Exit, viperConfig, shell.IsPipedShell, api.CreateProvider).Execute(args)
}

func (r *rootCmd) Execute(args []string) {
	defer func() {
		if err := recover(); err != nil {
			slog.Error("Panic occurred", "error", err)
		}
	}()

	// Set args for root command
	r.cmd.SetArgs(args)

	if err := r.cmd.Execute(); err != nil {
		// Defaults
		code := 1
		msg := "command failed"

		// Override defaults if possible
		exitErr := &exitError{}
		if errors.As(err, &exitErr) {
			code = exitErr.code
			if exitErr.details != "" {
				msg = exitErr.details
			}
		}

		// Log error with details and exit
		slog.Debug(msg, "error", err)
		r.exit(code)
		return
	}
	r.exit(0)
}

func newRootCmd(exit func(int), config *viper.Viper, isPipedShell func() (bool, error), createClientFn func(*viper.Viper, io.Writer) (api.Provider, error)) *rootCmd {
	root := &rootCmd{
		exit: exit,
	}

	cmd := &cobra.Command{
		Use:   "sgpt [persona] [prompt]",
		Short: "A command-line interface (CLI) tool to access the OpenAI models via the command line.",
		Long: `SGPT is a command-line interface (CLI) tool to access the OpenAI models via the command line.

Provide your prompt via stdin or as an argument and the tool will return the generated text. You can also provide a persona as an argument before the prompt to add further customize the generated responses.

By default the personas "code" and "sh" are included and can be used to generate code or shell commands. You can also add your own personas in a "personas"" directory of SGPT's config directory.

The tool can also be used to chat with the OpenAI models. You can start a new chat session or continue an existing one. The chat session is stored in the cache directory and can be deleted at any time.

The OpenAI GPT-4-vision API is also supported. You can provide images via command line args to a file or url. This feature is currently in beta.`,
		Example: `
# Ask questions
$ sgpt "mass of sun"
The mass of the sun is approximately 1.989 x 10^30 kilograms.

# Provide prompt via stdin
$ echo -n "mass of sun" | sgpt
The mass of the sun is approximately 1.989 x 10^30 kilograms.

# Generate code
$ sgpt code "Solve classic fizz buzz problem using Python"
for i in range(1, 101):
    if i % 3 == 0 and i % 5 == 0:
        print("FizzBuzz")
    elif i % 3 == 0:
        print("Fizz")
    elif i % 5 == 0:
        print("Buzz")
    else:
        print(i)

# Generate shell commands
$ sgpt sh "list all files in the current directory"
ls

# Use a chat to further customize the generated text
$ sgpt sh --chat ls-files "list all files directory"
ls
$ sgpt sh --chat ls-files "sort by name"
ls | sort
`,
		DisableFlagsInUseLine: true,
		SilenceUsage:          true,
		Args:                  cobra.RangeArgs(0, 2),
		ValidArgsFunction:     cobra.NoFileCompletions,
		PersistentPreRun: func(_ *cobra.Command, _ []string) {
			if root.verbose {
				opts := &slog.HandlerOptions{
					Level: slog.LevelDebug,
				}
				handler := slog.NewTextHandler(os.Stdout, opts)
				slog.SetDefault(slog.New(handler))
			}
		},
		PreRunE: func(_ *cobra.Command, _ []string) error {
			return loadViperConfig(config)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			isPiped, err := isPipedShell()
			if err != nil {
				return err
			}

			var prompts []string
			mode := "txt"

			if isPiped {
				var stdinInput string
				slog.Debug("Piped shell detected")
				// input is provided via stdin
				stdinInput, err = fs.ReadString(cmd.InOrStdin())
				if err != nil {
					return err
				}
				if len(stdinInput) == 0 {
					slog.Debug("No input via pipe provided")
					return ErrMissingInput
				}
				prompts = append(prompts, stdinInput)
				// mode is provided via command line args
				if len(args) == 1 {
					slog.Debug("Mode provided via command line args")
					mode = args[0]
				} else if len(args) == 2 {
					slog.Debug("Mode and prompt provided via command line args")
					mode = args[0]
					prompts = append(prompts, args[1])
				}

			} else {
				// input is provided via command line args
				if len(args) == 0 {
					return ErrMissingInput
				} else if len(args) == 1 {
					// input is provided via command line args
					slog.Debug("No mode provided via command line args - using default mode")
					prompts = append(prompts, args[0])
				} else {
					// input and mode are provided via command line args
					slog.Debug("Mode and prompt provided via command line args")
					mode = strings.ToLower(args[0])
					prompts = append(prompts, args[1])
				}
			}

			// Create provider
			var provider api.Provider
			provider, err = createClientFn(config, cmd.OutOrStdout())
			if err != nil {
				return err
			}

			var response string
			response, err = provider.CreateCompletion(cmd.Context(), root.chat, prompts, mode, root.input)
			if err != nil {
				return err
			}

			// If using chat, save the response to the chat session
			if root.chat != "" {
				var manager chat.SessionManager
				manager, err = chat.NewFilesystemChatSessionManager(config)
				if err != nil {
					return fmt.Errorf("failed to create chat session manager: %w", err)
				}

				// Get existing messages
				var messages []openai.ChatCompletionMessage
				messages, err = manager.GetSession(root.chat)
				if err != nil && !errors.Is(err, chat.ErrChatSessionDoesNotExist) {
					return fmt.Errorf("failed to get chat session: %w", err)
				}

				// If this is a new chat session and we're using a persona,
				// add the persona as a system message first
				if len(messages) == 0 && len(args) > 0 {
					// Check if the first argument is a persona file
					personaPath := filepath.Join(config.GetString("personas"), args[0])
					if _, err := os.Stat(personaPath); err == nil {
						// Read the persona content
						var persona []byte
						persona, err = os.ReadFile(personaPath)
						if err != nil {
							return fmt.Errorf("failed to read persona file: %w", err)
						}

						// Add persona as system message
						messages = append(messages, openai.ChatCompletionMessage{
							Role:    openai.ChatMessageRoleSystem,
							Content: string(persona),
						})

						// When using a persona in chat mode, we want to:
						// 1. Keep the persona as a system message
						// 2. Use the original prompt for the completion
						// 3. Use the prompt (without persona) for chat messages
						originalPrompt := args[1] // The actual prompt is in args[1]
						prompts = []string{originalPrompt} // Use only the actual prompt
					}
				}

				// Add user prompt
				messages = append(messages, openai.ChatCompletionMessage{
					Role:    openai.ChatMessageRoleUser,
					Content: strings.Join(prompts, "\n"),
				})

				// Add assistant response
				messages = append(messages, openai.ChatCompletionMessage{
					Role:    openai.ChatMessageRoleAssistant,
					Content: strings.TrimSpace(response),
				})

				// Save updated messages
				if err = manager.SaveSession(root.chat, messages); err != nil {
					return fmt.Errorf("failed to save chat session: %w", err)
				}
			}

			if root.copyToClipboard {
				slog.Debug("Sending client response to clipboard")
				err = clipboard.WriteAll(response)
				if err != nil {
					slog.Debug("Failed to send client response to clipboard", "error", err)
					return err
				}
			}

			if root.execute {
				slog.Debug("Trying to execute response in shell")
				return shell.ExecuteCommandWithConfirmation(cmd.Context(), cmd.InOrStdin(), cmd.OutOrStdout(), response)
			}
			return nil
		},
	}

	// flags
	cmd.Flags().BoolVarP(&root.execute, "execute", "e", false, "execute a response in the shell")
	cmd.Flags().BoolVarP(&root.copyToClipboard, "clipboard", "b", false, "send client response to clipboard")
	cmd.Flags().StringVarP(&root.chat, "chat", "c", "", "use an existing chat session or create a new one")
	cmd.Flags().StringSliceVarP(&root.input, "input", "i", nil, "provide images via command line args to a file or url (experimental)")

	// flags with config binding
	createFlagsWithConfigBinding(cmd, config)

	// verbose persistent flag
	cmd.PersistentFlags().BoolVarP(&root.verbose, "verbose", "v", false,
		"enable more verbose output for debugging")

	cmd.AddCommand(
		newChatCmd(config).cmd,
		newCheckCmd(config, createClientFn).cmd,
		newVersionCmd().cmd,
		newLicensesCmd().cmd,
		newManCmd().cmd,
		newConfigCmd(config).cmd,
	)

	root.cmd = cmd
	return root
}

func createFlagsWithConfigBinding(cmd *cobra.Command, config *viper.Viper) {
	var bindErrors []error
	var err error
	// text based commands
	cmd.Flags().StringP("model", "m", api.DefaultModel, "model name")
	err = config.BindPFlag("model", cmd.Flags().Lookup("model"))
	if err != nil {
		bindErrors = append(bindErrors, err)
	}

	cmd.Flags().IntP("max-tokens", "s", 2048, "strict length of output (tokens)")
	err = config.BindPFlag("maxTokens", cmd.Flags().Lookup("max-tokens"))
	if err != nil {
		bindErrors = append(bindErrors, err)
	}

	cmd.Flags().Float64P("temperature", "t", 1, "randomness of generated output")
	err = config.BindPFlag("temperature", cmd.Flags().Lookup("temperature"))
	if err != nil {
		bindErrors = append(bindErrors, err)
	}

	cmd.Flags().Float64P("top-p", "p", 1, "limits highest probable tokens")
	err = config.BindPFlag("topP", cmd.Flags().Lookup("top-p"))
	if err != nil {
		bindErrors = append(bindErrors, err)
	}

	cmd.Flags().Bool("stream", false, "stream output")
	err = config.BindPFlag("stream", cmd.Flags().Lookup("stream"))
	if err != nil {
		bindErrors = append(bindErrors, err)
	}

	// provider flag
	cmd.Flags().String("provider", "", "Name of the AI provider (e.g. 'bedrock' or 'openai')")
	err = config.BindPFlag("provider", cmd.Flags().Lookup("provider"))
	if err != nil {
		bindErrors = append(bindErrors, err)
	}

	if len(bindErrors) > 0 {
		for _, err = range bindErrors {
			slog.Error("Failed to bind flag to viper", "error", err)
		}
		panic("Failed to bind flags to viper")
	}
}

func loadViperConfig(config *viper.Viper) error {
	if !viper.IsSet("TESTING") {
		slog.Debug("Loading config")
		err := setViperDefaults(config)
		if err != nil {
			return err
		}
	}
	
	// Debug log all important config values
	fmt.Printf("Provider: %s\n", config.GetString("provider"))
	fmt.Printf("Model: %s\n", config.GetString("model"))
	fmt.Printf("Stream: %v\n", config.GetBool("stream"))
	if err := config.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			// Config file not found; ignore error
			slog.Debug("Config file not found - using defaults")
			return nil
		}
		// Config file was found but another error was produced
		return err
	}
	slog.Debug("Config file loaded")
	return nil
}

func setViperDefaults(config *viper.Viper) error {
	// cache dir
	appCacheDir, err := fs.GetAppCacheDir()
	if err != nil {
		return err
	}
	config.SetDefault("cacheDir", appCacheDir)
	// personas dir
	var personasDir string
	personasDir, err = fs.GetPersonasPath()
	if err != nil {
		return err
	}
	config.SetDefault("personas", personasDir)

	// model
	config.SetDefault("model", api.DefaultModel)
	// max-tokens
	config.SetDefault("maxTokens", 2048)
	// temperature
	config.SetDefault("temperature", 1)
	// top-p
	config.SetDefault("topP", 1)
	// stream
	config.SetDefault("stream", false)
	// provider
	config.SetDefault("provider", "")

	return nil
}
