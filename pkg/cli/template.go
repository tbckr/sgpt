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
	"bytes"
	"errors"
	"fmt"
	"log/slog"
	"text/template"

	"gopkg.in/yaml.v3"
)

var (
	// ErrTemplateRequiresPipe is returned when --template is used without piped input.
	ErrTemplateRequiresPipe = errors.New("--template requires piped input: pipe YAML or JSON variables via stdin")
	// ErrTemplateWithTwoArgs is returned when --template is used with both a persona and a prompt argument.
	ErrTemplateWithTwoArgs = errors.New("--template cannot be combined with a positional prompt argument; the template string replaces the prompt")
	// ErrTemplateWithExecute is returned when --template and --execute are both set.
	// --template drains stdin for variable data, leaving nothing for the --execute confirmation prompt.
	ErrTemplateWithExecute = errors.New("--template cannot be combined with --execute: stdin is consumed by template variables")
	// ErrTemplateDataParse is returned when the piped data cannot be parsed as YAML or JSON.
	ErrTemplateDataParse = errors.New("failed to parse piped data as YAML or JSON")
	// ErrTemplateDataEmpty is returned when the piped data parses to an empty map.
	ErrTemplateDataEmpty = errors.New("piped data parsed to an empty map; provide at least one key-value pair")
)

// parseTemplateData parses YAML or JSON data into a variable map.
// yaml.Unmarshal accepts JSON as valid YAML (JSON is a subset of YAML).
// The original parse error is preserved via wrapping so callers get diagnostic detail.
func parseTemplateData(data string) (map[string]interface{}, error) {
	var vars map[string]interface{}
	if err := yaml.Unmarshal([]byte(data), &vars); err != nil {
		// Wrap with both the sentinel (for errors.Is) and the underlying cause (for diagnostics).
		return nil, fmt.Errorf("%w: %w", ErrTemplateDataParse, err)
	}
	if len(vars) == 0 {
		return nil, ErrTemplateDataEmpty
	}
	return vars, nil
}

// renderTemplate renders a Go text/template string with the provided variables.
// Missing keys in the template cause an error; extra keys in vars are silently ignored.
//
// Security note: templateStr is executed as a fully trusted Go text/template. This function
// must only be called with template strings sourced from the authenticated local user (CLI flag).
// Do not pass template strings from files, environment variables, or network inputs without
// additional validation.
func renderTemplate(templateStr string, vars map[string]interface{}) (string, error) {
	slog.Debug("Rendering prompt template")
	t, err := template.New("prompt").Option("missingkey=error").Parse(templateStr)
	if err != nil {
		return "", fmt.Errorf("template parse error: %w", err)
	}
	var buf bytes.Buffer
	if err = t.Execute(&buf, vars); err != nil {
		return "", fmt.Errorf("template execution error: %w", err)
	}
	slog.Debug("Prompt template rendered")
	return buf.String(), nil
}
