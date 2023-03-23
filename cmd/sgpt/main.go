package main

import (
	"fmt"
	"github.com/tbckr/sgpt/cli"
	"log"
	"os"
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
