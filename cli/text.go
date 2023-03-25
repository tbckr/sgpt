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
	"strings"

	"github.com/spf13/cobra"
	"github.com/tbckr/sgpt/api"
	"github.com/tbckr/sgpt/modifiers"
	"github.com/tbckr/sgpt/shell"
)

var textArgs struct {
	model       string
	maxTokens   int
	temperature float64
	topP        float64
	chatSession string
}

func textCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "txt <prompt>",
		Short: "Query the different openai models for a text completion",
		Long: strings.TrimSpace(fmt.Sprintf(`
Query a openai model for a text completion. The following models are supported: %s.
`, strings.Join(api.SupportedModels, ", "))),
		RunE: runText,
	}
	fs := cmd.Flags()
	fs.StringVarP(&textArgs.model, "model", "m", api.DefaultModel, "model name to use")
	fs.IntVarP(&textArgs.maxTokens, "max-tokens", "s", 2048, "strict length of output (tokens)")
	fs.Float64VarP(&textArgs.temperature, "temperature", "t", 1.0, "randomness of generated output")
	fs.Float64VarP(&textArgs.topP, "top-p", "p", 1.0, "limits highest probable tokens")
	fs.StringVarP(&textArgs.chatSession, "chat", "c", "", "use an existing chat session")
	return cmd
}

func runText(cmd *cobra.Command, args []string) error {
	prompt, err := shell.GetInput(args)
	if err != nil {
		return err
	}

	options := api.CompletionOptions{
		Model:       textArgs.model,
		MaxTokens:   textArgs.maxTokens,
		Temperature: float32(textArgs.temperature),
		TopP:        float32(textArgs.topP),
		Modifier:    modifiers.Nil,
		ChatSession: textArgs.chatSession,
	}

	var cli *api.Client
	cli, err = api.CreateClient()
	if err != nil {
		return err
	}

	var response string
	if api.IsChatModel(options.Model) {
		response, err = cli.GetChatCompletion(cmd.Context(), options, prompt)
	} else {
		response, err = cli.GetCompletion(cmd.Context(), options, prompt)
	}
	if err != nil {
		return err
	}
	_, err = fmt.Fprintln(stdout, response)
	return err
}
