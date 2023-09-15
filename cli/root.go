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

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/tbckr/sgpt/api"
	"github.com/tbckr/sgpt/fs"
	"github.com/tbckr/sgpt/shell"
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
	config.SetConfigFile("config")
	config.SetConfigType("yaml")
	return config, nil
}

func Execute(args []string) {
	viperConfig, err := createViperConfig()
	if err != nil {
		slog.Error("Failed to create viper config", "error", err)
		os.Exit(1)
	}
	newRootCmd(os.Exit, viperConfig, api.CreateClient).Execute(args)
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
		slog.Error(msg, "error", err)
		r.exit(code)
		return
	}
	r.exit(0)
}

func newRootCmd(exit func(int), config *viper.Viper, createClientFn func() (*api.OpenAIClient, error)) *rootCmd {
	root := &rootCmd{
		exit: exit,
	}

	cmd := &cobra.Command{
		Use:                   "sgpt",
		Short:                 "A command-line interface (CLI) tool to access the OpenAI models via the command line.",
		DisableFlagsInUseLine: true,
		Args:                  cobra.RangeArgs(1, 2),
		ValidArgsFunction:     cobra.NoFileCompletions,
		PersistentPreRunE: func(cmd *cobra.Command, _ []string) error {
			if root.verbose {
				opts := &slog.HandlerOptions{
					Level: slog.LevelDebug,
				}
				handler := slog.NewTextHandler(os.Stdout, opts)
				slog.SetDefault(slog.New(handler))
			}
			err := setViperDefaults(config)
			if err != nil {
				return err
			}
			if err = config.ReadInConfig(); err != nil {
				if errors.Is(err, os.ErrNotExist) {
					// Config file not found; skip this then
					return nil
				}
				// Config file was found but another error was produced
				return err
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			mode := strings.ToLower(args[0])

			var prompt string
			if len(args) == 1 {
				// input is provided via stdin
				input, err := fs.ReadString(cmd.InOrStdin())
				if err != nil {
					return err
				}
				if len(input) == 0 {
					return ErrMissingInput
				}
				prompt = input

			} else if len(args) == 2 {
				// input is provided via command line args
				prompt = args[1]
			} else {
				// input is missing
				return ErrMissingInput
			}

			// TODO validate flag combination
			// TODO override config with mode config

			// Create client
			client, err := createClientFn()
			if err != nil {
				return err
			}

			switch mode {
			case "txt":
			case "sh":
			case "code":
			default:
				var response string
				response, err = client.GetChatCompletion(cmd.Context(), config, prompt, mode)
				if err != nil {
					return err
				}
				_, err = fmt.Fprintln(cmd.OutOrStdout(), response)
				if err != nil {
					return err
				}
				if config.GetBool("execute") {
					return shell.ExecuteCommandWithConfirmation(cmd.Context(), cmd.InOrStdin(), cmd.OutOrStdout(), response)
				}
				break
			}
			return nil
		},
	}

	cmd.PersistentFlags().BoolVarP(&root.verbose, "verbose", "v", false,
		"enable more verbose output for debugging")

	// text based commands
	cmd.Flags().StringP("model", "m", api.DefaultModel, "model name")
	cmd.Flags().IntP("max-tokens", "s", 2048, "strict length of output (tokens)")
	cmd.Flags().Float64P("temperature", "t", 0.8, "randomness of generated output")
	cmd.Flags().Float64P("top-p", "p", 0.2, "limits highest probable tokens")
	// additionally: shell command
	cmd.Flags().BoolP("execute", "e", false, "execute shell command")
	// chat flags
	cmd.Flags().StringP("chat", "c", "", "use an existing chat session")

	// Bind flags to viper
	err := config.BindPFlags(cmd.Flags())
	if err != nil {
		panic(err)
	}

	cmd.AddCommand(
		newChatCmd(config).cmd,
		newCheckCmd().cmd,
		newVersionCmd().cmd,
		newLicensesCmd().cmd,
		newManCmd().cmd,
	)

	root.cmd = cmd
	return root
}

func setViperDefaults(config *viper.Viper) error {
	// cache dir
	appCacheDir, err := fs.GetAppCacheDir()
	if err != nil {
		return err
	}
	config.SetDefault("cacheDir", appCacheDir)
	return nil
}
