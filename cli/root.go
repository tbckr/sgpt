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
	"log/slog"
	"os"
	"strings"

	"github.com/atotto/clipboard"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/tbckr/sgpt/v2/api"
	"github.com/tbckr/sgpt/v2/fs"
	"github.com/tbckr/sgpt/v2/shell"
)

type rootCmd struct {
	cmd     *cobra.Command
	exit    func(int)
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
	newRootCmd(os.Exit, viperConfig, shell.IsPipedShell, api.CreateClient).Execute(args)
}

func (r *rootCmd) Execute(args []string) {
	defer func() {
		if err := recover(); err != nil {
			slog.Error("Panic occured", "error", err)
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

func newRootCmd(exit func(int), config *viper.Viper, isPipedShell func() (bool, error), createClientFn func() (*api.OpenAIClient, error)) *rootCmd {
	root := &rootCmd{
		exit: exit,
	}

	cmd := &cobra.Command{
		Use:   "sgpt [persona] [prompt]",
		Short: "A command-line interface (CLI) tool to access the OpenAI models via the command line.",
		Long: `SGPT is a command-line interface (CLI) tool to access the OpenAI models via the command line.

Provide your prompt via stdin or as an argument and the tool will return the generated text. You can also provide a persona as an argument before the prompt to add further customize the generated responses.

By default the personas "code" and "sh" are included and can be used to generate code or shell commands. You can also add your own personas in a "personas"" directory of SGPT's config directory.

The tool can also be used to chat with the OpenAI models. You can start a new chat session or continue an existing one. The chat session is stored in the cache directory and can be deleted at any time.`,
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

			var promptBuilder strings.Builder
			var prompt, input string
			mode := "txt"

			if isPiped {
				slog.Debug("Piped shell detected")
				// input is provided via stdin
				input, err = fs.ReadString(cmd.InOrStdin())
				if err != nil {
					return err
				}
				if len(input) == 0 {
					slog.Debug("No input via pipe provided")
					return ErrMissingInput
				}
				_, err = promptBuilder.WriteString(input)
				if err != nil {
					return err
				}

				// check if args are provided
				if len(args) == 0 {
					// no mode or prompt provided via command line args
					slog.Debug("No mode or additional prompt provided via command line args")

				} else if len(args) == 1 {
					// additional prompt was provided via command line args
					slog.Debug("Additional prompt provided via command line args")
					_, err = promptBuilder.WriteString("\n\n" + args[0])
					if err != nil {
						return err
					}

				} else {
					// mode and additional prompt were provided via command line args
					mode = strings.ToLower(args[0])
					_, err = promptBuilder.WriteString("\n\n" + args[1])
					if err != nil {
						return err
					}
				}
				prompt = promptBuilder.String()

			} else {
				// input is provided via command line args
				if len(args) == 0 {
					return ErrMissingInput
				} else if len(args) == 1 {
					// input is provided via command line args
					slog.Debug("No mode provided via command line args - using default mode")
					prompt = args[0]
				} else {
					// input and mode are provided via command line args
					slog.Debug("Mode and prompt provided via command line args")
					mode = strings.ToLower(args[0])
					prompt = args[1]
				}
			}

			// Create client
			var client *api.OpenAIClient
			client, err = createClientFn()
			if err != nil {
				return err
			}

			var response string
			response, err = client.GetChatCompletion(cmd.Context(), config, prompt, mode)
			if err != nil {
				return err
			}
			_, err = fmt.Fprintln(cmd.OutOrStdout(), response)
			if err != nil {
				return err
			}

			if config.GetBool("clipboard") {
				slog.Debug("Sending client response to clipboard")
				err = clipboard.WriteAll(response)
				if err != nil {
					slog.Debug("Failed to send client response to clipboard", "error", err)
					return err
				}
			}

			if config.GetBool("execute") {
				slog.Debug("Trying to execute response in shell")
				return shell.ExecuteCommandWithConfirmation(cmd.Context(), cmd.InOrStdin(), cmd.OutOrStdout(), response)
			}
			return nil
		},
	}

	cmd.PersistentFlags().BoolVarP(&root.verbose, "verbose", "v", false,
		"enable more verbose output for debugging")

	createFlags(cmd, config)

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

func createFlags(cmd *cobra.Command, config *viper.Viper) {
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

	// shell command
	cmd.Flags().BoolP("execute", "e", false, "execute a response in the shell")
	err = config.BindPFlag("execute", cmd.Flags().Lookup("execute"))
	if err != nil {
		bindErrors = append(bindErrors, err)
	}

	// clipboard flags
	cmd.Flags().BoolP("clipboard", "b", false, "send client response to clipboard")
	err = config.BindPFlag("clipboard", cmd.Flags().Lookup("clipboard"))
	if err != nil {
		bindErrors = append(bindErrors, err)
	}

	// chat flags
	cmd.Flags().StringP("chat", "c", "", "use an existing chat session or create a new one")
	err = config.BindPFlag("chat", cmd.Flags().Lookup("chat"))
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

	// model
	config.SetDefault("model", api.DefaultModel)
	// max-tokens
	config.SetDefault("maxTokens", 2048)
	// temperature
	config.SetDefault("temperature", 1)
	// top-p
	config.SetDefault("topP", 1)
	// execute
	config.SetDefault("execute", false)

	return nil
}
