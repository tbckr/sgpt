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
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/spf13/cobra"
)

const (
	resetFormat = "\033[0m"
)

var (
	stdout io.Writer = os.Stdout
	stderr io.Writer = os.Stderr
)

var rootArgs struct {
	debug bool
}

// Run runs the CLI. The args do not include the binary name.
func Run(args []string) {
	rootCmd := &cobra.Command{
		Use:   "sgpt",
		Short: "A command-line interface (CLI) tool to access the OpenAI models via the command line.",
		RunE: func(cmd *cobra.Command, _ []string) error {
			return cmd.Help()
		},
	}
	rootCmd.AddCommand(
		textCmd(),
		shellCmd(),
		codeCmd(),
		imageCmd(),
		chatCmd(),
		versionCmd(),
	)
	persistentFs := rootCmd.PersistentFlags()
	persistentFs.BoolVarP(&rootArgs.debug, "verbose", "v", false,
		"enable more verbose output for debugging")

	rootCmd.SetArgs(args)
	if err := rootCmd.ExecuteContext(context.Background()); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return
		}
		if _, err = fmt.Fprintln(stderr, err); err != nil {
			log.Fatal(err)
		}
		os.Exit(1)
	}
}
