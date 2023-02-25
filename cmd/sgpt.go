package main

import (
	"github.com/tbckr/sgpt"
	"log"
)

func main() {
	if err := sgpt.Run(); err != nil {
		log.Fatal(err)
	}
}
