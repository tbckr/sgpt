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
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestExitError_Error(t *testing.T) {
	baseErr := errors.New("test error")
	exitErr := &exitError{
		err:     baseErr,
		code:    42,
		details: "test details",
	}

	require.Equal(t, "test error", exitErr.Error())
}

func TestExitError_Unwrap(t *testing.T) {
	baseErr := errors.New("test error")
	exitErr := &exitError{
		err:     baseErr,
		code:    42,
		details: "test details",
	}

	require.Equal(t, baseErr, exitErr.Unwrap())
}

func TestExitError_ErrorsIs(t *testing.T) {
	baseErr := errors.New("test error")
	exitErr := &exitError{
		err:     baseErr,
		code:    42,
		details: "test details",
	}

	require.True(t, errors.Is(exitErr, baseErr))
	require.False(t, errors.Is(exitErr, errors.New("different error")))
}

func TestExitError_ErrorsAs(t *testing.T) {
	baseErr := errors.New("test error")
	exitErr := &exitError{
		err:     baseErr,
		code:    42,
		details: "test details",
	}

	var target *exitError
	require.True(t, errors.As(exitErr, &target))
	require.Equal(t, 42, target.code)
	require.Equal(t, "test details", target.details)
}
