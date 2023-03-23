// Copyright (c) 2023 Tim <tbckr>
//
// Permission is hereby granted, free of charge, to any person obtaining a copy of
// this software and associated documentation files (the "Software"), to deal in
// the Software without restriction, including without limitation the rights to
// use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of
// the Software, and to permit persons to whom the Software is furnished to do so,
// subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS
// FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR
// COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER
// IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN
// CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
//
// SPDX-License-Identifier: MIT

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
