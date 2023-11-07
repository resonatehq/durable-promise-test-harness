package main

import (
	"log"

	"github.com/resonatehq/durable-promise-test-harness/cmd"
)

func main() {
	if err := cmd.New().Execute(); err != nil {
		log.Fatal(err)
	}
}
