package cli

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/spf13/cobra"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"

	pb "github.com/yanicksenn/ruthless/api/v1"
)

var playCmd = &cobra.Command{
	Use:   "play",
	Short: "Join a game session and play",
}

var playStartCmd = &cobra.Command{
	Use:   "start",
	Short: "Start a new game session",
	Run: func(cmd *cobra.Command, args []string) {
		conn, err := grpc.NewClient(grpcHost, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			log.Fatalf("Failed to connect: %v", err)
		}
		defer conn.Close()

		client := pb.NewSessionServiceClient(conn)
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		resp, err := client.CreateSession(ctx, &pb.CreateSessionRequest{})
		if err != nil {
			log.Fatalf("Failed to start session: %v", err)
		}

		fmt.Printf("Session started:\nID: %s\n", resp.Id)
	},
}

var playJoinCmd = &cobra.Command{
	Use:   "join [session_id]",
	Short: "Join an existing game session",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		sessionID := args[0]
		playerName, _ := cmd.Flags().GetString("name")
		if playerName == "" {
			log.Fatal("Player name is required")
		}

		token, _ := cmd.Flags().GetString("token")
		if token == "" {
			// For testing with fake auth, we just use the name as token
			token = playerName
		}

		conn, err := grpc.NewClient(grpcHost, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			log.Fatalf("Failed to connect: %v", err)
		}
		defer conn.Close()

		client := pb.NewSessionServiceClient(conn)
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		ctx = metadata.AppendToOutgoingContext(ctx, "authorization", "Bearer "+token)

		resp, err := client.JoinSession(ctx, &pb.JoinSessionRequest{
			SessionId:  sessionID,
			PlayerName: playerName,
		})
		if err != nil {
			log.Fatalf("Failed to join session: %v", err)
		}

		fmt.Printf("Successfully joined session:\nID: %s\nPlayers: %d\n", resp.Id, len(resp.PlayerIds))
	},
}

var playAddDeckCmd = &cobra.Command{
	Use:   "add-deck [session_id] [deck_id]",
	Short: "Add a deck to an existing game session",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		sessionID := args[0]
		deckID := args[1]

		conn, err := grpc.NewClient(grpcHost, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			log.Fatalf("Failed to connect: %v", err)
		}
		defer conn.Close()

		client := pb.NewSessionServiceClient(conn)
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		_, err = client.AddDeckToSession(ctx, &pb.AddDeckToSessionRequest{
			SessionId: sessionID,
			DeckId:    deckID,
		})
		if err != nil {
			log.Fatalf("Failed to add deck to session: %v", err)
		}

		fmt.Println("Successfully added deck to session!")
	},
}

func init() {
	rootCmd.AddCommand(playCmd)
	playCmd.AddCommand(playStartCmd)
	playCmd.AddCommand(playJoinCmd)
	playCmd.AddCommand(playAddDeckCmd)

	playJoinCmd.Flags().String("name", "", "Your player name")
	playJoinCmd.Flags().String("token", "", "Auth token (optional, uses name for fake auth if omitted)")
	playJoinCmd.MarkFlagRequired("name")
}
