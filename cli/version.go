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
	"fmt"

	"github.com/spf13/cobra"
	"github.com/tbckr/sgpt/internal/buildinfo"
)

type versionCmd struct {
	cmd  *cobra.Command
	full bool
}

func newVersionCmd() *versionCmd {
	version := &versionCmd{}
	cmd := &cobra.Command{
		Use:               "version",
		Short:             "Get the programs version",
		Args:              cobra.NoArgs,
		ValidArgsFunction: cobra.NoFileCompletions,
		RunE: func(cmd *cobra.Command, _ []string) error {
			var err error
			if version.full {
				_, err = fmt.Fprintf(cmd.OutOrStdout(), "version: %s\ncommit: %s\ncommitDate: %s\n", buildinfo.Version(), buildinfo.Commit(), buildinfo.CommitDate())
			} else {
				_, err = fmt.Fprintf(cmd.OutOrStdout(), "version: %s\n", buildinfo.Version())
			}
			return err
		},
	}
	cmd.Flags().BoolVarP(&version.full, "full", "f", false, "Print full version information")
	version.cmd = cmd
	return version
}
