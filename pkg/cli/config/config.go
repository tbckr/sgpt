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

package config

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type RootCmd struct {
	Command *cobra.Command
}

type InitCmd struct {
	Command *cobra.Command
}

type ShowCmd struct {
	Command *cobra.Command
}

func NewConfigCmd(config *viper.Viper) *RootCmd {
	configStruct := &RootCmd{}
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
		NewConfigInitCmd(config).Command,
		NewConfigShowCmd(config).Command,
	)
	configStruct.Command = cmd
	return configStruct
}

func NewConfigInitCmd(config *viper.Viper) *InitCmd {
	configInit := &InitCmd{}
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
	configInit.Command = cmd
	return configInit
}

func NewConfigShowCmd(config *viper.Viper) *ShowCmd {
	configShow := &ShowCmd{}
	cmd := &cobra.Command{
		Use:               "show",
		Short:             "Show configuration",
		Long:              "Show configuration",
		Args:              cobra.NoArgs,
		ValidArgsFunction: cobra.NoFileCompletions,
		PreRunE: func(_ *cobra.Command, _ []string) error {
			//return loadViperConfig(config)
			return nil
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
	configShow.Command = cmd
	return configShow
}
