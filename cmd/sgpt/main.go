package main

import (
	"os"

	"github.com/tbckr/sgpt/cli"
)

func main() {
	args := os.Args[1:]
	cli.Run(args)
}
