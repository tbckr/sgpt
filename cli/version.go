package cli

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/tbckr/sgpt/internal/buildinfo"
)

var versionArgs struct {
	json bool
}

func versionCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "version",
		Short: "Get the programs version",
		Long: strings.TrimSpace(`
Get the program version.
`),
		RunE: runVersion,
		Args: cobra.NoArgs,
	}
	fs := cmd.Flags()
	fs.BoolVar(&versionArgs.json, "json", false, "Output in json format")
	return cmd
}

func runVersion(_ *cobra.Command, _ []string) error {
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
