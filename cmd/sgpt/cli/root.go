package cli

import (
	"context"
	"errors"
	"flag"
	"github.com/peterbourgon/ff/v3/ffcli"
	//flag "github.com/spf13/pflag"
	"io"
	"os"
	"strings"
)

var Stderr io.Writer = os.Stderr

func newFlagSet(name string) *flag.FlagSet {
	fs := flag.NewFlagSet(name, flag.ExitOnError)
	fs.SetOutput(Stderr)
	return fs
}

// Run runs the CLI. The args do not include the binary name.
func Run(args []string) error {
	rootCmd := &ffcli.Command{
		Name:       "sgpt",
		ShortUsage: "sgpt [flags] <subcommand> [command flags]",
		ShortHelp:  "A command-line interface (CLI) tool to access the OpenAI models via the command line.",
		LongHelp: strings.TrimSpace(`
For help on subcommands, add --help after: "sgpt shell --help".
`),
		Subcommands: []*ffcli.Command{
			shellCmd,
			codeCmd,
			imageCmd,
			versionCmd,
		},
		Exec:      runCompletion,
		UsageFunc: usageFunc,
	}

	// parse args
	if err := rootCmd.ParseAndRun(context.Background(), args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return nil
		}
		return err
	}
	return nil
}

func runCompletion(ctx context.Context, args []string) error {
	return nil
}

func usageFunc(c *ffcli.Command) string {
	return ""
}
