package cli

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/peterbourgon/ff/v3/ffcli"
)

var (
	stdout io.Writer = os.Stdout
	stderr io.Writer = os.Stderr

	ErrMissingPrompt = errors.New("exactly one prompt must be provided")
)

func newFlagSet(name string) *flag.FlagSet {
	fs := flag.NewFlagSet(name, flag.ExitOnError)
	fs.SetOutput(stderr)
	return fs
}

// Run runs the CLI. The args do not include the binary name.
func Run(args []string) error {
	rootCmd := &ffcli.Command{
		Name:       "sgpt",
		ShortUsage: "sgpt <subcommand> [command flags] [prompt]",
		ShortHelp:  "A command-line interface (CLI) tool to access the OpenAI models via the command line.",
		LongHelp: strings.TrimSpace(`
For help on subcommands, add --help after: "sgpt sh --help".
`),
		Subcommands: []*ffcli.Command{
			textCmd,
			shellCmd,
			codeCmd,
			imageCmd,
			versionCmd,
		},
		Exec: func(ctx context.Context, args []string) error {
			return flag.ErrHelp
		},
		UsageFunc: usageFunc,
	}
	// add usage to subcommands
	for _, c := range rootCmd.Subcommands {
		if c.UsageFunc == nil {
			c.UsageFunc = usageFunc
		}
	}
	// parse flags
	if err := rootCmd.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return nil
		}
		return err
	}
	err := rootCmd.Run(context.Background())
	if errors.Is(err, flag.ErrHelp) {
		return nil
	}
	return err
}

func usageFunc(c *ffcli.Command) string {
	var b strings.Builder

	fmt.Fprintf(&b, "USAGE\n")
	if c.ShortUsage != "" {
		fmt.Fprintf(&b, "  %s\n", c.ShortUsage)
	} else {
		fmt.Fprintf(&b, "  %s\n", c.Name)
	}
	fmt.Fprintf(&b, "\n")

	if c.LongHelp != "" {
		fmt.Fprintf(&b, "%s\n\n", c.LongHelp)
	}

	if len(c.Subcommands) > 0 {
		fmt.Fprintf(&b, "SUBCOMMANDS\n")
		tw := tabwriter.NewWriter(&b, 0, 2, 2, ' ', 0)
		for _, subcommand := range c.Subcommands {
			fmt.Fprintf(tw, "  %s\t%s\n", subcommand.Name, subcommand.ShortHelp)
		}
		tw.Flush()
		fmt.Fprintf(&b, "\n")
	}

	if countFlags(c.FlagSet) > 0 {
		fmt.Fprintf(&b, "FLAGS\n")
		tw := tabwriter.NewWriter(&b, 0, 2, 2, ' ', 0)
		c.FlagSet.VisitAll(func(f *flag.Flag) {
			var s string
			name, usage := flag.UnquoteUsage(f)
			if strings.HasPrefix(usage, "HIDDEN: ") {
				return
			}
			if isBoolFlag(f) {
				s = fmt.Sprintf("  --%s, --%s=false", f.Name, f.Name)
			} else {
				s = fmt.Sprintf("  --%s", f.Name) // Two spaces before --; see next two comments.
				if len(name) > 0 {
					s += " " + name
				}
			}
			// Four spaces before the tab triggers good alignment
			// for both 4- and 8-space tab stops.
			s += "\n    \t"
			s += strings.ReplaceAll(usage, "\n", "\n    \t")

			s += fmt.Sprintf(" (default %s)", f.DefValue)

			fmt.Fprintln(&b, s)
		})
		tw.Flush()
		fmt.Fprintf(&b, "\n")
	}

	return strings.TrimSpace(b.String())
}

func isBoolFlag(f *flag.Flag) bool {
	bf, ok := f.Value.(interface {
		IsBoolFlag() bool
	})
	return ok && bf.IsBoolFlag()
}

func countFlags(fs *flag.FlagSet) (n int) {
	fs.VisitAll(func(*flag.Flag) { n++ })
	return n
}
