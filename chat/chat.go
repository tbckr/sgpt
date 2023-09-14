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

package chat

import (
	"errors"
	"fmt"
	"regexp"
	"unicode/utf8"

	"github.com/sashabaranov/go-openai"
)

const (
	defaultFilePermissions = 0755
	sessionNameMaxLength   = 65
)

var (
	ErrChatSessionDoesNotExist = errors.New("chat session does not exist")
	ErrChatSessionNameInvalid  = fmt.Errorf("chat session name does not match the regex %s", sessionNameRegex)
	ErrChatSessionNameTooLong  = fmt.Errorf("chat session name is greater than %d", sessionNameMaxLength)

	sessionNameRegex   = "^([-a-zA-Z0-9]*[a-zA-Z0-9])?"
	sessionNameMatcher = regexp.MustCompile(sessionNameRegex)
)

type ChatSessionManager interface {
	SessionExists(sessionName string) (bool, error)
	GetSession(sessionName string) ([]openai.ChatCompletionMessage, error)
	SaveSession(sessionName string, messages []openai.ChatCompletionMessage) error
	ListSessions() ([]string, error)
	DeleteSession(sessionName string) error
}

func validateSessionName(sessionName string) error {
	if !sessionNameMatcher.Match([]byte(sessionName)) {
		return ErrChatSessionNameInvalid
	}
	if utf8.RuneCountInString(sessionName) > sessionNameMaxLength {
		return ErrChatSessionNameTooLong
	}
	return nil
}
