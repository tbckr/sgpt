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
	"fmt"
	"io"
	"strings"

	"github.com/spf13/viper"

	jww "github.com/spf13/jwalterweatherman"

	"github.com/tbckr/sgpt/chat"

	"github.com/sashabaranov/go-openai"
	"github.com/spf13/cobra"
)

const (
	chatRoleFormat = "\033[1m"
	resetFormat    = "\033[0m"
)

var (
	ErrMissingChatSession  = errors.New("no chat session name provided")
	ErrChatSessionNotExist = errors.New("given chat does not exist")
)

type chatCmd struct {
	cmd *cobra.Command
}

type chatCreateCmd struct {
	cmd *cobra.Command
}

type chatLsCmd struct {
	cmd *cobra.Command
}

type chatShowCmd struct {
	cmd *cobra.Command
}

type chatRmCmd struct {
	cmd       *cobra.Command
	deleteAll bool
}

func newChatCmd(config *viper.Viper) *chatCmd {
	chatStruct := &chatCmd{}
	cmd := &cobra.Command{
		Use:   "chat",
		Short: "Manage chat sessions",
		Long: strings.TrimSpace(`
Manage all open chat sessions - list, show, and delete chat sessions.
`),
		DisableFlagsInUseLine: true,
		Args:                  cobra.NoArgs,
		ValidArgsFunction:     cobra.NoFileCompletions,
		RunE: func(cmd *cobra.Command, _ []string) error {
			// TODO
			return cmd.Help()
		},
	}
	cmd.AddCommand(
		newLsCmd(config).cmd,
		newShowCmd(config).cmd,
		newRmCmd(config).cmd,
	)
	chatStruct.cmd = cmd
	return chatStruct
}

func newLsCmd(config *viper.Viper) *chatLsCmd {
	ls := &chatLsCmd{}
	cmd := &cobra.Command{
		Use:   "ls",
		Short: "List all chat sessions",
		Long: strings.TrimSpace(`
List all chat sessions.
`),
		DisableFlagsInUseLine: true,
		Args:                  cobra.NoArgs,
		ValidArgsFunction:     cobra.NoFileCompletions,
		RunE: func(cmd *cobra.Command, _ []string) error {
			chatSessionManager, err := chat.NewFilesystemChatSessionManager(config)
			if err != nil {
				return err
			}
			var sessions []string
			sessions, err = chatSessionManager.ListSessions()
			if err != nil {
				return err
			}
			if len(sessions) == 0 {
				return nil
			}
			_, err = fmt.Fprintln(cmd.OutOrStdout(), strings.Join(sessions, "\n"))
			return err
		},
	}
	ls.cmd = cmd
	return ls
}

func newShowCmd(config *viper.Viper) *chatShowCmd {
	show := &chatShowCmd{}
	cmd := &cobra.Command{
		Use:     "show <session name>",
		Aliases: []string{"cat"},
		Short:   "Show the conversation for the given chat session",
		Long: strings.TrimSpace(`
Show the conversation for the given chat session.
`),
		DisableFlagsInUseLine: true,
		Args:                  cobra.ExactArgs(1),
		ValidArgsFunction:     cobra.NoFileCompletions,
		RunE: func(cmd *cobra.Command, args []string) error {
			chatSessionManager, err := chat.NewFilesystemChatSessionManager(config)
			if err != nil {
				return err
			}
			sessionName := args[0]
			var exists bool
			exists, err = chatSessionManager.SessionExists(sessionName)
			if err != nil {
				return err
			}
			if !exists {
				return ErrChatSessionNotExist
			}
			var messages []openai.ChatCompletionMessage
			messages, err = chatSessionManager.GetSession(sessionName)
			if err != nil {
				return err
			}
			jww.DEBUG.Println("showing conversation")
			return showConversation(cmd.OutOrStdout(), messages)
		},
	}
	show.cmd = cmd
	return show
}

func newRmCmd(config *viper.Viper) *chatRmCmd {
	rm := &chatRmCmd{}
	cmd := &cobra.Command{
		Use:   "rm [session name]",
		Short: "Remove the specified chat session",
		Long: strings.TrimSpace(`
Remove the specified chat session. The --all flag removes all chat sessions.
`),
		DisableFlagsInUseLine: true,
		Args:                  cobra.RangeArgs(0, 1),
		ValidArgsFunction:     cobra.NoFileCompletions,
		RunE: func(cmd *cobra.Command, args []string) error {
			chatSessionManager, err := chat.NewFilesystemChatSessionManager(config)
			if err != nil {
				return err
			}
			// Get session/s
			var chatSessions []string
			if rm.deleteAll {
				retrievedSessions, err := chatSessionManager.ListSessions()
				if err != nil {
					return err
				}
				chatSessions = append(chatSessions, retrievedSessions...)
			} else {
				if len(args) != 1 {
					return ErrMissingChatSession
				}
				sessionName := args[0]
				chatSessions = append(chatSessions, sessionName)
			}
			return deleteChatSessions(chatSessionManager, cmd.OutOrStdout(), chatSessions)
		},
	}
	cmd.Flags().BoolVarP(&rm.deleteAll, "all", "a", false, "remove all chat sessions")
	rm.cmd = cmd
	return rm
}

func showConversation(out io.Writer, messages []openai.ChatCompletionMessage) error {
	for _, message := range messages {
		if _, err := fmt.Fprintf(out, "%s%s:%s %s\n", chatRoleFormat, message.Role, resetFormat,
			message.Content); err != nil {
			return err
		}
	}
	return nil
}

func deleteChatSessions(manager chat.ChatSessionManager, out io.Writer, chatSessions []string) error {
	for _, chatSession := range chatSessions {
		err := manager.DeleteSession(chatSession)
		if err != nil {
			return err
		}
		_, err = fmt.Fprintln(out, chatSession)
		if err != nil {
			return err
		}
	}
	return nil
}
