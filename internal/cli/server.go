package cli

import (
	"log"
	"net/http"

	"github.com/spf13/cobra"
	"github.com/yanicksenn/ruthless/internal/auth"
	"github.com/yanicksenn/ruthless/internal/auth/fake"
	"github.com/yanicksenn/ruthless/internal/server"
	"github.com/yanicksenn/ruthless/internal/storage"
	"github.com/yanicksenn/ruthless/internal/storage/memory"
)

var (
	storageFlag string
	authFlag    string
)

var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Start the HTTP API server",
	Run: func(cmd *cobra.Command, args []string) {
		log.Printf("Starting server with storage=%s auth=%s", storageFlag, authFlag)

		// Setup Storage
		var store storage.Storage
		if storageFlag == "memory" {
			store = memory.New()
		} else {
			log.Fatalf("Unsupported storage type: %s", storageFlag)
		}

		// Setup Auth
		var authenticator auth.Authenticator
		if authFlag == "fake" {
			authenticator = fake.New()
		} else {
			log.Fatalf("Unsupported auth type: %s", authFlag)
		}

		srv := server.New(store, authenticator)

		log.Println("Listening on :8080")
		if err := http.ListenAndServe(":8080", srv); err != nil {
			log.Fatalf("Server failed: %v", err)
		}
	},
}

func init() {
	rootCmd.AddCommand(serverCmd)
	serverCmd.Flags().StringVar(&storageFlag, "storage", "memory", "storage engine (memory|postgres)")
	serverCmd.Flags().StringVar(&authFlag, "auth", "fake", "auth mechanism (fake|oauth)")
}
