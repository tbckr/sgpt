package main

import (
	"log"
	"os"

	"github.com/tbckr/sgpt/cmd/sgpt/cli"
)

func main() {
	args := os.Args[1:]
	if err := cli.Run(args); err != nil {
		log.Fatal(err)
	}
}
