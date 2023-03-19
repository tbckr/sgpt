package cli

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"strings"

	"github.com/sashabaranov/go-openai"

	"github.com/peterbourgon/ff/v3/ffcli"
	"github.com/tbckr/sgpt/internal/chat"
)

var (
	ErrMissingChatSession  = errors.New("no chat session name provided")
	ErrChatSessionNotExist = errors.New("given chat does not exist")
)

var chatCmd = &ffcli.Command{
	Name:       "chat",
	ShortUsage: "sgpt chat <subcommand> [subcommand flags]",
	ShortHelp:  "Manage chat sessions",
	// TODO
	LongHelp: strings.TrimSpace(`
`),
	Subcommands: []*ffcli.Command{
		lsCmd,
		showCmd,
		rmCmd,
	},
	Exec: func(_ context.Context, _ []string) error {
		return flag.ErrHelp
	},
	UsageFunc: usageFunc,
}

var lsCmd = &ffcli.Command{
	Name:       "ls",
	ShortUsage: "sgpt chat ls",
	ShortHelp:  "List all chat sessions",
	// TODO
	LongHelp: strings.TrimSpace(`
`),
	Exec:      runChatLsCmd,
	UsageFunc: usageFunc,
}

var showCmd = &ffcli.Command{
	Name:       "show",
	ShortUsage: "sgpt chat show <chat session>",
	ShortHelp:  "Show the conversation for the given chat session",
	// TODO
	LongHelp: strings.TrimSpace(`
`),
	Exec:      runChatShowCmd,
	UsageFunc: usageFunc,
}

var rmCmd = &ffcli.Command{
	Name:       "rm",
	ShortUsage: "sgpt chat rm [command flags] [chat session]",
	ShortHelp:  "Remove the specified chat session",
	// TODO
	LongHelp: strings.TrimSpace(`
`),
	Exec: runChatRmCmd,
	FlagSet: (func() *flag.FlagSet {
		fs := newFlagSet("chat-rm")
		fs.BoolVar(&chatRmArgs.deleteAll, "all", false, "Remove all chat sessions")
		return fs
	})(),
	UsageFunc: usageFunc,
}

var chatRmArgs struct {
	deleteAll bool
}

func runChatLsCmd(_ context.Context, _ []string) error {
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

func runChatShowCmd(_ context.Context, args []string) error {
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
		if _, err := fmt.Fprintf(stdout, "%s: %s\n", message.Role, message.Content); err != nil {
			return err
		}
	}
	return nil
}

func runChatRmCmd(_ context.Context, args []string) error {
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
