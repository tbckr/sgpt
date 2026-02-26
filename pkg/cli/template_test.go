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
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParseTemplateData_YAML(t *testing.T) {
	data := "name: Dave\ncountry: France\n"
	vars, err := parseTemplateData(data)
	require.NoError(t, err)
	require.Equal(t, "Dave", vars["name"])
	require.Equal(t, "France", vars["country"])
}

func TestParseTemplateData_JSON(t *testing.T) {
	data := `{"name": "Dave", "country": "France"}`
	vars, err := parseTemplateData(data)
	require.NoError(t, err)
	require.Equal(t, "Dave", vars["name"])
	require.Equal(t, "France", vars["country"])
}

func TestParseTemplateData_Empty(t *testing.T) {
	_, err := parseTemplateData("{}")
	require.ErrorIs(t, err, ErrTemplateDataEmpty)
}

func TestParseTemplateData_Invalid(t *testing.T) {
	_, err := parseTemplateData("not: valid: yaml: [")
	require.ErrorIs(t, err, ErrTemplateDataParse)
}

func TestParseTemplateData_Nested(t *testing.T) {
	data := "person:\n  name: Dave\n  age: 30\n"
	vars, err := parseTemplateData(data)
	require.NoError(t, err)
	person, ok := vars["person"].(map[string]interface{})
	require.True(t, ok)
	require.Equal(t, "Dave", person["name"])
}

func TestRenderTemplate_Basic(t *testing.T) {
	vars := map[string]interface{}{"name": "Dave", "country": "France"}
	result, err := renderTemplate("What would {{ .name }} be called in {{ .country }}?", vars)
	require.NoError(t, err)
	require.Equal(t, "What would Dave be called in France?", result)
}

func TestRenderTemplate_MissingKey(t *testing.T) {
	vars := map[string]interface{}{"name": "Dave"}
	_, err := renderTemplate("Hello {{ .name }}, you live in {{ .country }}", vars)
	require.Error(t, err)
}

func TestRenderTemplate_ExtraKeysIgnored(t *testing.T) {
	vars := map[string]interface{}{"name": "Dave", "unused": "extra"}
	result, err := renderTemplate("Hello {{ .name }}", vars)
	require.NoError(t, err)
	require.Equal(t, "Hello Dave", result)
}

func TestRenderTemplate_InvalidSyntax(t *testing.T) {
	vars := map[string]interface{}{"name": "Dave"}
	_, err := renderTemplate("Hello {{ .name", vars)
	require.Error(t, err)
}
