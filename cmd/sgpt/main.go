package main

import (
	"fmt"
	"log"
	"os"

	"github.com/tbckr/sgpt/cmd/sgpt/cli"
)

func main() {
	args := os.Args[1:]
	if err := cli.Run(args); err != nil {
		if _, err := fmt.Fprintln(os.Stderr, err); err != nil {
			log.Fatal(err)
		}
		os.Exit(1)
	}
}
