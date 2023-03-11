package cli

import (
	"context"
	"flag"
	"github.com/peterbourgon/ff/v3/ffcli"
	"strings"
)

// TODO implement command
var shellCmd = &ffcli.Command{
	Name:       "",
	ShortUsage: "",
	ShortHelp:  "",
	LongHelp:   strings.TrimSpace(``),
	Exec:       runShell,
	FlagSet: (func() *flag.FlagSet {
		fs := newFlagSet("shell")
		fs.BoolVar(&shellArgs.execute, "execute", false, "execute shell command")
		return fs
	})(),
}

var shellArgs struct {
	execute bool
}

func runShell(ctx context.Context, args []string) error {
	return nil
}
