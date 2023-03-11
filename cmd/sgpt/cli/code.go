package cli

import (
	"context"
	"flag"
	"github.com/peterbourgon/ff/v3/ffcli"
	"strings"
)

// TODO implement command
var codeCmd = &ffcli.Command{
	Name:       "",
	ShortUsage: "",
	ShortHelp:  "",
	LongHelp:   strings.TrimSpace(``),
	Exec:       runCode,
	FlagSet: (func() *flag.FlagSet {
		fs := newFlagSet("code")
		fs.BoolVar(&codeArgs.execute, "execute", false, "execute shell command")
		fs.BoolVar(&codeArgs.execute, "e", false, "execute shell command")
		return fs
	})(),
}

var codeArgs struct {
	execute bool
}

func runCode(ctx context.Context, args []string) error {
	return nil
}
