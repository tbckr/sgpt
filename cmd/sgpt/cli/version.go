package cli

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"strings"

	"github.com/tbckr/sgpt"

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
	var output string
	if versionArgs.json {
		versionMap := map[string]string{
			"version": sgpt.Version,
		}
		byteData, err := json.Marshal(versionMap)
		if err != nil {
			return err
		}
		output = string(byteData)
	} else {
		output = sgpt.Version
	}
	if _, err := fmt.Fprintln(stdout, output); err != nil {
		return err
	}
	return nil
}
