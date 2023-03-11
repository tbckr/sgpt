package cli

import (
	"context"
	"flag"
	"github.com/peterbourgon/ff/v3/ffcli"
	"strings"
)

// TODO implement command
var imageCmd = &ffcli.Command{
	Name:       "",
	ShortUsage: "",
	ShortHelp:  "",
	LongHelp:   strings.TrimSpace(``),
	Exec:       runImage,
	FlagSet: (func() *flag.FlagSet {
		fs := newFlagSet("image")
		fs.BoolVar(&imageArgs.execute, "execute", false, "execute shell command")
		return fs
	})(),
}

var imageArgs struct {
	execute bool
}

func runImage(ctx context.Context, args []string) error {
	return nil
}
