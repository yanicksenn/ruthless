package cli

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/spf13/cobra"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

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

		resp, err := client.CreateCard(ctx, &pb.CreateCardRequest{Text: text})
		if err != nil {
			log.Fatalf("Failed to create card: %v", err)
		}

		fmt.Printf("Success! Card created:\nID: %s\nText: %s\n", resp.Id, resp.Text)
	},
}

var cardsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all cards",
	Run: func(cmd *cobra.Command, args []string) {
		conn, err := grpc.NewClient(grpcHost, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			log.Fatalf("Failed to connect: %v", err)
		}
		defer conn.Close()

		client := pb.NewCardServiceClient(conn)
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		resp, err := client.ListCards(ctx, &pb.ListCardsRequest{})
		if err != nil {
			log.Fatalf("Failed to list cards: %v", err)
		}

		fmt.Println("Cards:")
		for _, card := range resp.Cards {
			fmt.Printf("- [%s] %s\n", card.Id, card.Text)
		}
	},
}

func init() {
	rootCmd.AddCommand(cardsCmd)
	cardsCmd.AddCommand(cardsCreateCmd)
	cardsCmd.AddCommand(cardsListCmd)

	cardsCreateCmd.Flags().String("text", "", "The text of the card, containing at least one '___'")
	cardsCreateCmd.MarkFlagRequired("text")
}
