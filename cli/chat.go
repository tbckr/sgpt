package cli

import (
	"errors"
	"fmt"
	"strings"

	"github.com/tbckr/sgpt/chat"

	"github.com/sashabaranov/go-openai"
	"github.com/spf13/cobra"
)

const (
	chatRoleFormat = "\033[1m"
)

var (
	ErrMissingChatSession  = errors.New("no chat session name provided")
	ErrChatSessionNotExist = errors.New("given chat does not exist")
)

var chatRmArgs struct {
	deleteAll bool
}

func chatCmd() *cobra.Command {
	rootChatCmd := &cobra.Command{
		Use:   "chat",
		Short: "Manage chat sessions",
		Long: strings.TrimSpace(`
Manage all open chat sessions - list, show, and delete chat sessions.
`),
		RunE: func(cmd *cobra.Command, _ []string) error {
			return cmd.Help()
		},
		Args: cobra.NoArgs,
	}

	lsCmd := &cobra.Command{
		Use:   "ls",
		Short: "List all chat sessions",
		Long: strings.TrimSpace(`
List all chat sessions.
`),
		RunE: runChatLsCmd,
		Args: cobra.NoArgs,
	}

	showCmd := &cobra.Command{
		Use:     "show <session name>",
		Aliases: []string{"cat"},
		Short:   "Show the conversation for the given chat session",
		Long: strings.TrimSpace(`
Show the conversation for the given chat session.
`),
		RunE: runChatShowCmd,
		Args: cobra.ExactArgs(1),
	}

	rmCmd := &cobra.Command{
		Use:   "rm <session name>",
		Short: "Remove the specified chat session",
		Long: strings.TrimSpace(`
Remove the specified chat session. The --all flag removes all chat sessions.
`),
		DisableFlagsInUseLine: true,
		RunE:                  runChatRmCmd,
		Args:                  cobra.RangeArgs(0, 1),
	}
	rmFs := rmCmd.Flags()
	rmFs.BoolVarP(&chatRmArgs.deleteAll, "all", "a", false, "remove all chat sessions")

	rootChatCmd.AddCommand(
		lsCmd,
		showCmd,
		rmCmd,
	)

	return rootChatCmd
}

func runChatLsCmd(_ *cobra.Command, _ []string) error {
	sessions, err := chat.ListSessions()
	if err != nil {
		return err
	}
	if len(sessions) == 0 {
		return nil
	}
	_, err = fmt.Fprintln(stdout, strings.Join(sessions, "\n"))
	return err
}

func runChatShowCmd(_ *cobra.Command, args []string) error {
	if len(args) != 1 {
		return ErrMissingChatSession
	}
	sessionName := args[0]
	exists, err := chat.SessionExists(sessionName)
	if err != nil {
		return err
	}
	if !exists {
		return ErrChatSessionNotExist
	}
	var messages []openai.ChatCompletionMessage
	messages, err = chat.GetSession(sessionName)
	if err != nil {
		return err
	}
	return showConversation(messages)
}

func showConversation(messages []openai.ChatCompletionMessage) error {
	for _, message := range messages {
		if _, err := fmt.Fprintf(stdout, "%s%s:%s %s\n", chatRoleFormat, message.Role, resetFormat,
			message.Content); err != nil {
			return err
		}
	}
	return nil
}

func runChatRmCmd(_ *cobra.Command, args []string) error {
	// Get session/s
	var chatSessions []string
	if chatRmArgs.deleteAll {
		retrievedSessions, err := chat.ListSessions()
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
	return deleteChatSessions(chatSessions)
}

func deleteChatSessions(chatSessions []string) error {
	for _, chatSession := range chatSessions {
		err := chat.DeleteSession(chatSession)
		if err != nil {
			return err
		}
		_, err = fmt.Fprintln(stdout, chatSession)
		if err != nil {
			return err
		}
	}
	return nil
}
