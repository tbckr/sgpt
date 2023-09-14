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
	"runtime"
	"strings"

	"github.com/spf13/cobra"
)

type licensesCmd struct {
	cmd *cobra.Command
}

func newLicensesCmd() *licensesCmd {
	licenses := &licensesCmd{}
	cmd := &cobra.Command{
		Use:   "licenses",
		Short: "List the open source licenses used in this software",
		Long: strings.TrimSpace(`
List the open source licenses used in this software.
`),
		DisableFlagsInUseLine: true,
		Args:                  cobra.NoArgs,
		ValidArgsFunction:     cobra.NoFileCompletions,
		RunE: func(cmd *cobra.Command, _ []string) error {
			licenses := licensesURL()
			_, err := fmt.Fprintln(cmd.OutOrStdout(), `To see the open source packages included in SGPT and
their respective license information, visit:

`+licenses)
			return err
		},
	}
	licenses.cmd = cmd
	return licenses
}

// licensesURL returns the absolute URL containing open source license information for the current platform.
func licensesURL() string {
	switch runtime.GOOS {
	default:
		return "https://github.com/tbckr/sgpt/blob/main/licenses/oss-licenses.md"
	}
}
