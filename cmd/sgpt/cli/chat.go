package cli

import (
	"context"
	"flag"
	"strings"

	"github.com/peterbourgon/ff/v3/ffcli"
)

var chatCmd = &ffcli.Command{
	Name:       "chat",
	ShortUsage: "sgpt chat <subcommand> [subcommand flags]",
	ShortHelp:  "Manage chat sessions",
	LongHelp: strings.TrimSpace(`
`),
	Subcommands: []*ffcli.Command{
		lsCmd,
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
	LongHelp: strings.TrimSpace(`
`),
	Exec:      runChatLsCmd,
	UsageFunc: usageFunc,
}

var rmCmd = &ffcli.Command{
	Name:       "rm",
	ShortUsage: "sgpt chat rm [command flags] [chat session]",
	ShortHelp:  "Remove the specified chat session",
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

func runChatLsCmd(ctx context.Context, args []string) error {
	return nil
}

func runChatRmCmd(ctx context.Context, args []string) error {
	return nil
}
