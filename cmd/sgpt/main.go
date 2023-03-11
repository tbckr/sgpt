package main

import (
	"log"

	"github.com/tbckr/sgpt"
)

func main() {
	if err := sgpt.Run(); err != nil {
		log.Fatal(err)
	}
}
