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

var codeArgs struct {
	model       string
	maxTokens   int
	temperature float64
	topP        float64
	chatSession string
}

func codeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "code <prompt>",
		Short: "Query the openai models for code-specific questions",
		Long: strings.TrimSpace(`
Query a openai model for code specific questions.
The supported completion models can be listed via: "sgpt txt --help"
`),
		RunE: runCode,
	}
	fs := cmd.Flags()
	fs.StringVarP(&codeArgs.model, "model", "m", api.DefaultModel, "model name")
	fs.IntVarP(&codeArgs.maxTokens, "max-tokens", "s", 2048, "strict length of output (tokens)")
	fs.Float64VarP(&codeArgs.temperature, "temperature", "t", 0.8, "randomness of generated output")
	fs.Float64VarP(&codeArgs.topP, "top-p", "p", 0.2, "limits highest probable tokens")
	fs.StringVarP(&codeArgs.chatSession, "chat", "c", "", "use an existing chat session")
	return cmd
}

func runCode(cmd *cobra.Command, args []string) error {
	prompt, err := shell.GetInput(args)
	if err != nil {
		return err
	}

	options := api.CompletionOptions{
		Model:       codeArgs.model,
		MaxTokens:   codeArgs.maxTokens,
		Temperature: float32(codeArgs.temperature),
		TopP:        float32(codeArgs.topP),
		Modifier:    modifiers.Code,
		ChatSession: codeArgs.chatSession,
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
