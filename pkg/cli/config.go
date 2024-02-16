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
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type configCmd struct {
	cmd *cobra.Command
}

type configInitCmd struct {
	cmd *cobra.Command
}

type configShowCmd struct {
	cmd *cobra.Command
}

func newConfigCmd(config *viper.Viper) *configCmd {
	configStruct := &configCmd{}
	cmd := &cobra.Command{
		Use:               "config",
		Short:             "Manage configuration",
		Long:              "Manage configuration",
		Args:              cobra.NoArgs,
		ValidArgsFunction: cobra.NoFileCompletions,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return cmd.Help()
		},
	}
	cmd.AddCommand(
		newConfigInitCmd(config).cmd,
		newConfigShowCmd(config).cmd,
	)
	configStruct.cmd = cmd
	return configStruct
}

func newConfigInitCmd(config *viper.Viper) *configInitCmd {
	configInit := &configInitCmd{}
	cmd := &cobra.Command{
		Use:               "init",
		Short:             "Initialize configuration",
		Long:              "Initialize configuration",
		Args:              cobra.NoArgs,
		ValidArgsFunction: cobra.NoFileCompletions,
		RunE: func(_ *cobra.Command, _ []string) error {
			return config.SafeWriteConfig()
		},
	}
	configInit.cmd = cmd
	return configInit
}

func newConfigShowCmd(config *viper.Viper) *configShowCmd {
	configShow := &configShowCmd{}
	cmd := &cobra.Command{
		Use:               "show",
		Short:             "Show configuration",
		Long:              "Show configuration",
		Args:              cobra.NoArgs,
		ValidArgsFunction: cobra.NoFileCompletions,
		PreRunE: func(_ *cobra.Command, _ []string) error {
			return loadViperConfig(config)
		},
		RunE: func(cmd *cobra.Command, _ []string) error {
			configFilepath := config.ConfigFileUsed()
			data, err := os.ReadFile(configFilepath)
			if err != nil {
				return err
			}
			_, err = fmt.Fprint(cmd.OutOrStdout(), string(data))
			return err
		},
	}
	configShow.cmd = cmd
	return configShow
}
