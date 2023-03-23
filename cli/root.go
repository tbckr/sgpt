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
