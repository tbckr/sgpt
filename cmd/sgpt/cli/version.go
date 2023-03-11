package cli

import (
	"context"
	"flag"
	"github.com/peterbourgon/ff/v3/ffcli"
	"strings"
)

// TODO implement command
var versionCmd = &ffcli.Command{
	Name:       "",
	ShortUsage: "",
	ShortHelp:  "",
	LongHelp:   strings.TrimSpace(``),
	Exec:       runVersion,
	FlagSet: (func() *flag.FlagSet {
		fs := newFlagSet("version")
		fs.BoolVar(&versionArgs.json, "json", false, "execute shell command")
		return fs
	})(),
}

var versionArgs struct {
	json bool
}

func runVersion(ctx context.Context, args []string) error {
	return nil
}
