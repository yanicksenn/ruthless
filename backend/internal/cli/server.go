package cli

import (
	"log"
	"net"

	"github.com/spf13/cobra"
	"google.golang.org/grpc"
	
	"github.com/yanicksenn/ruthless/backend/internal/auth"
	"github.com/yanicksenn/ruthless/backend/internal/auth/fake"
	"github.com/yanicksenn/ruthless/backend/internal/server"
	"github.com/yanicksenn/ruthless/backend/internal/storage"
	"github.com/yanicksenn/ruthless/backend/internal/storage/memory"
)

var (
	storageFlag string
	authFlag    string
	seedFlag    string
)

var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Start the gRPC API server",
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

		// Handle Seeding
		if seedFlag != "" {
			if storageFlag != "memory" {
				log.Fatalf("Seeding is only supported with 'memory' storage")
			}
			log.Printf("Seeding data from %s", seedFlag)
			if err := server.LoadSeed(cmd.Context(), store, seedFlag); err != nil {
				log.Fatalf("Failed to seed data: %v", err)
			}
		}

		listener, err := net.Listen("tcp", ":8080")
		if err != nil {
			log.Fatalf("Failed to listen: %v", err)
		}

		grpcServer := grpc.NewServer(
			grpc.UnaryInterceptor(srv.UnaryAuthInterceptor()),
		)

		srv.Register(grpcServer)

		log.Println("Listening on gRPC :8080")
		if err := grpcServer.Serve(listener); err != nil {
			log.Fatalf("Server failed: %v", err)
		}
	},
}

func init() {
	rootCmd.AddCommand(serverCmd)
	serverCmd.Flags().StringVar(&storageFlag, "storage", "memory", "storage engine (memory|postgres)")
	serverCmd.Flags().StringVar(&authFlag, "auth", "fake", "auth mechanism (fake|oauth)")
	serverCmd.Flags().StringVar(&seedFlag, "seed", "", "path to a JSON seed file (only works with --storage=memory)")
}
