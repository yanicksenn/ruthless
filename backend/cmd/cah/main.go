package main

import (
	"log"

	"github.com/yanicksenn/ruthless/backend/internal/cli"
)

func main() {
	if err := cli.Execute(); err != nil {
		log.Fatalf("Error: %v", err)
	}
}
