package main

import (
	"github.com/kolah/runner/cli"
	"log"
)

func main() {
	if err := cli.RootCommand().Execute(); err != nil {
		log.Fatal(err)
	}
}
