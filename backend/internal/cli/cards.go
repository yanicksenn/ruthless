package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/spf13/cobra"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"

	pb "github.com/yanicksenn/ruthless/api/v1"
)

var cardsCmd = &cobra.Command{
	Use:   "cards",
	Short: "Manage cards",
}

var cardsCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new card",
	Run: func(cmd *cobra.Command, args []string) {
		text, _ := cmd.Flags().GetString("text")
		if text == "" {
			log.Fatal("Card text is required")
		}

		conn, err := grpc.NewClient(grpcHost, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			log.Fatalf("Failed to connect: %v", err)
		}
		defer conn.Close()

		client := pb.NewCardServiceClient(conn)
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		token, err := ResolveToken(cmd)
		if err != nil {
			log.Fatalf("Token error: %v", err)
		}
		if token != "" {
			ctx = metadata.AppendToOutgoingContext(ctx, "authorization", "Bearer "+token)
		}

		resp, err := client.CreateCard(ctx, &pb.CreateCardRequest{Text: text})
		if err != nil {
			log.Fatalf("Failed to create card: %v", err)
		}

		fmt.Printf("Success! Card created:\nID: %s\nText: %s\n", resp.Id, resp.Text)
	},
}

var cardsBulkCreateCmd = &cobra.Command{
	Use:   "bulk-create",
	Short: "Bulk create cards from a JSON file",
	Run: func(cmd *cobra.Command, args []string) {
		filePath, _ := cmd.Flags().GetString("file")
		if filePath == "" {
			log.Fatal("File path is required")
		}

		data, err := os.ReadFile(filePath)
		if err != nil {
			log.Fatalf("Failed to read file: %v", err)
		}

		var cards []struct {
			Text string `json:"text"`
		}
		if err := json.Unmarshal(data, &cards); err != nil {
			log.Fatalf("Failed to parse JSON: %v", err)
		}

		conn, err := grpc.NewClient(grpcHost, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			log.Fatalf("Failed to connect: %v", err)
		}
		defer conn.Close()

		client := pb.NewCardServiceClient(conn)
		token, err := ResolveToken(cmd)
		if err != nil {
			log.Fatalf("Token error: %v", err)
		}

		fmt.Printf("Creating %d cards...\n", len(cards))
		for i, card := range cards {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			if token != "" {
				ctx = metadata.AppendToOutgoingContext(ctx, "authorization", "Bearer "+token)
			}

			resp, err := client.CreateCard(ctx, &pb.CreateCardRequest{Text: card.Text})
			cancel()
			if err != nil {
				log.Printf("[%d/%d] Failed to create card %q: %v", i+1, len(cards), card.Text, err)
				continue
			}
			fmt.Printf("[%d/%d] Created: %s\n", i+1, len(cards), resp.Id)
		}
		fmt.Println("Done!")
	},
}

var cardsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all cards",
	Run: func(cmd *cobra.Command, args []string) {
		deckID, _ := cmd.Flags().GetString("deck-id")
		conn, err := grpc.NewClient(grpcHost, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			log.Fatalf("Failed to connect: %v", err)
		}
		defer conn.Close()

		client := pb.NewCardServiceClient(conn)
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		token, err := ResolveToken(cmd)
		if err != nil {
			log.Fatalf("Token error: %v", err)
		}
		if token != "" {
			ctx = metadata.AppendToOutgoingContext(ctx, "authorization", "Bearer "+token)
		}

		resp, err := client.ListCards(ctx, &pb.ListCardsRequest{
			DeckId: deckID,
		})
		if err != nil {
			log.Fatalf("Failed to list cards: %v", err)
		}

		for _, card := range resp.Cards {
			fmt.Printf("- [%s] %s\n", card.Id, card.Text)
		}
	},
}

var cardsDeleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete a card",
	Run: func(cmd *cobra.Command, args []string) {
		id, _ := cmd.Flags().GetString("id")
		if id == "" {
			log.Fatal("Card ID is required")
		}

		conn, err := grpc.NewClient(grpcHost, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			log.Fatalf("Failed to connect: %v", err)
		}
		defer conn.Close()

		client := pb.NewCardServiceClient(conn)
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		token, err := ResolveToken(cmd)
		if err != nil {
			log.Fatalf("Token error: %v", err)
		}
		if token != "" {
			ctx = metadata.AppendToOutgoingContext(ctx, "authorization", "Bearer "+token)
		}

		_, err = client.DeleteCard(ctx, &pb.DeleteCardRequest{Id: id})
		if err != nil {
			log.Fatalf("Failed to delete card: %v", err)
		}

		fmt.Printf("Success! Card [%s] deleted.\n", id)
	},
}

func init() {
	rootCmd.AddCommand(cardsCmd)
	cardsCmd.AddCommand(cardsCreateCmd)
	cardsCmd.AddCommand(cardsBulkCreateCmd)
	cardsCmd.AddCommand(cardsListCmd)
	cardsCmd.AddCommand(cardsDeleteCmd)

	cardsCreateCmd.Flags().String("text", "", "The text of the card, containing at least one '___'")
	cardsCreateCmd.MarkFlagRequired("text")

	cardsBulkCreateCmd.Flags().String("file", "", "Path to the JSON file containing cards")
	cardsBulkCreateCmd.MarkFlagRequired("file")

	cardsDeleteCmd.Flags().String("id", "", "The ID of the card to delete")
	cardsDeleteCmd.MarkFlagRequired("id")

	cardsListCmd.Flags().String("deck-id", "", "Filter cards by deck ID")
	AddTokenFlags(cardsCmd)
}
