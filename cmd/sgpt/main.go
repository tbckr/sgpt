package main

import (
	"fmt"
	"os"

	"github.com/tbckr/sgpt/cmd/sgpt/cli"
)

func main() {
	args := os.Args[1:]
	if err := cli.Run(args); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
