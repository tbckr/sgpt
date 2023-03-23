package cli

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"strings"

	"github.com/peterbourgon/ff/v3/ffcli"
	"github.com/tbckr/sgpt/internal/buildinfo"
)

var versionCmd = &ffcli.Command{
	Name:       "version",
	ShortUsage: "sgpt version [command flags]",
	ShortHelp:  "Get the programs version.",
	LongHelp: strings.TrimSpace(`
Get the program version.
`),
	Exec: runVersion,
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
	var output string
	if versionArgs.json {
		versionMap := map[string]string{
			"version":    buildinfo.Version(),
			"commit":     buildinfo.Commit(),
			"commitDate": buildinfo.CommitDate(),
		}
		byteData, err := json.Marshal(versionMap)
		if err != nil {
			return err
		}
		output = string(byteData)
	} else {
		output = fmt.Sprintf("version: %s\ncommit: %s\ncommitDate: %s",
			buildinfo.Version(), buildinfo.Commit(), buildinfo.CommitDate())
	}
	if _, err := fmt.Fprintln(stdout, output); err != nil {
		return err
	}
	return nil
}
