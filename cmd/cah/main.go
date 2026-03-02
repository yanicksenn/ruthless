package main

import (
	"log"

	"github.com/yanicksenn/ruthless/internal/cli"
)

func main() {
	if err := cli.Execute(); err != nil {
		log.Fatalf("Error: %v", err)
	}
}
