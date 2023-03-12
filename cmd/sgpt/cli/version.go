package cli

import (
	"context"
	"flag"
	"strings"

	"github.com/peterbourgon/ff/v3/ffcli"
)

var versionCmd = &ffcli.Command{
	Name:       "version",
	ShortUsage: "",
	ShortHelp:  "",
	LongHelp:   strings.TrimSpace(``),
	Exec:       runVersion,
	FlagSet: (func() *flag.FlagSet {
		fs := newFlagSet("version")
		fs.BoolVar(&versionArgs.json, "json", false, "Output in json format")
		return fs
	})(),
}

var versionArgs struct {
	json bool
}

func runVersion(_ context.Context, _ []string) error {
	return nil
}
