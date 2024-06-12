package main

import (
	"log"

	"github.com/knqyf263/boltwiz/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		log.Fatal(err)
	}
}
