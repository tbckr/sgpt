package cli

import (
	"context"
	"flag"
	"strings"

	"github.com/peterbourgon/ff/v3/ffcli"
)

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
