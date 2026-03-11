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

var decksCmd = &cobra.Command{
	Use:   "decks",
	Short: "Manage decks of cards",
}

func getDeckClientAndCtx(cmd *cobra.Command) (pb.DeckServiceClient, context.Context, context.CancelFunc, *grpc.ClientConn) {
	token, _ := cmd.Flags().GetString("token")
	conn, err := grpc.NewClient(grpcHost, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}

	client := pb.NewDeckServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	
	if token != "" {
		ctx = metadata.AppendToOutgoingContext(ctx, "authorization", "Bearer "+token)
	}

	return client, ctx, cancel, conn
}

var decksCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new deck",
	Run: func(cmd *cobra.Command, args []string) {
		name, _ := cmd.Flags().GetString("name")
		if name == "" {
			log.Fatal("Deck name is required")
		}

		client, ctx, cancel, conn := getDeckClientAndCtx(cmd)
		defer cancel()
		defer conn.Close()

		resp, err := client.CreateDeck(ctx, &pb.CreateDeckRequest{Name: name})
		if err != nil {
			log.Fatalf("Failed to create deck: %v", err)
		}

		fmt.Printf("Success! Deck created:\nID: %s\nName: %s\n", resp.Id, resp.Name)
	},
}

var decksAddCardCmd = &cobra.Command{
	Use:   "add-card [deck_id] [card_id]",
	Short: "Add an existing card to a deck",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		deckID := args[0]
		cardID := args[1]

		// First, get the card to pass to AddCardToDeck
		conn, err := grpc.NewClient(grpcHost, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			log.Fatalf("Failed to connect: %v", err)
		}
		defer conn.Close()

		clientCards := pb.NewCardServiceClient(conn)

		ctxCards, cancelCards := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancelCards()
		cardsResp, err := clientCards.ListCards(ctxCards, &pb.ListCardsRequest{})
		if err != nil {
			log.Fatalf("Failed to list cards: %v", err)
		}

		var targetCard *pb.Card
		for _, c := range cardsResp.Cards {
			if c.Id == cardID {
				targetCard = c
				break
			}
		}
		if targetCard == nil {
			log.Fatalf("Card with ID %s not found", cardID)
		}

		client := pb.NewDeckServiceClient(conn)
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		token, _ := cmd.Flags().GetString("token")
		if token != "" {
			ctx = metadata.AppendToOutgoingContext(ctx, "authorization", "Bearer "+token)
		}

		_, err = client.AddCardToDeck(ctx, &pb.AddCardToDeckRequest{
			DeckId: deckID,
			Card:   targetCard,
		})
		if err != nil {
			log.Fatalf("Failed to add card to deck: %v", err)
		}

		fmt.Println("Card successfully added to deck.")
	},
}

func init() {
	rootCmd.AddCommand(decksCmd)
	decksCmd.AddCommand(decksCreateCmd)
	decksCmd.AddCommand(decksAddCardCmd)

	decksCreateCmd.Flags().String("name", "", "The name of the deck")
	decksCreateCmd.MarkFlagRequired("name")
	
	decksCmd.PersistentFlags().String("token", "", "Your auth token (fake auth uses player name)")
}
