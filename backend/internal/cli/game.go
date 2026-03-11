package cli

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"

	pb "github.com/yanicksenn/ruthless/api/v1"
)

var gameCmd = &cobra.Command{
	Use:   "game",
	Short: "Manage and play games",
}

// Internal helper
func getGameClientAndCtx(cmd *cobra.Command) (pb.GameServiceClient, context.Context, context.CancelFunc, *grpc.ClientConn) {
	token, _ := cmd.Flags().GetString("token")
	if token == "" {
		log.Fatal("Token is required to interact with games. Usually this is your player name when using fake auth.")
	}

	conn, err := grpc.NewClient(grpcHost, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}

	client := pb.NewGameServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	
	ctx = metadata.AppendToOutgoingContext(ctx, "authorization", "Bearer "+token)

	return client, ctx, cancel, conn
}

var gameCreateCmd = &cobra.Command{
	Use:   "create [session_id]",
	Short: "Create a game from a session",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		sessionID := args[0]
		client, ctx, cancel, conn := getGameClientAndCtx(cmd)
		defer cancel()
		defer conn.Close()

		resp, err := client.CreateGame(ctx, &pb.CreateGameRequest{SessionId: sessionID})
		if err != nil {
			log.Fatalf("Failed to create game: %v", err)
		}
		fmt.Printf("Game Created! ID: %s\n", resp.Id)
	},
}

var gameStartCmd = &cobra.Command{
	Use:   "begin [game_id]",
	Short: "Start a created game",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		gameID := args[0]
		client, ctx, cancel, conn := getGameClientAndCtx(cmd)
		defer cancel()
		defer conn.Close()

		_, err := client.StartGame(ctx, &pb.StartGameRequest{Id: gameID})
		if err != nil {
			log.Fatalf("Failed to start game: %v", err)
		}
		fmt.Println("Game has started!")
	},
}

var gameStatusCmd = &cobra.Command{
	Use:   "status [game_id]",
	Short: "Get the current game status",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		gameID := args[0]
		client, ctx, cancel, conn := getGameClientAndCtx(cmd)
		defer cancel()
		defer conn.Close()

		game, err := client.GetGame(ctx, &pb.GetGameRequest{Id: gameID})
		if err != nil {
			log.Fatalf("Failed to get game status: %v", err)
		}

		fmt.Printf("Game ID: %s (%s)\n", game.Id, game.State.String())
		fmt.Println("Scores:")
		for playerID, score := range game.Scores {
			fmt.Printf("  %s: %d\n", playerID, score)
		}

		if len(game.Rounds) > 0 {
			currentRound := game.Rounds[len(game.Rounds)-1]
			fmt.Println("\n--- Current Round ---")
			fmt.Printf("Czar: %s\n", currentRound.CzarId)
			fmt.Printf("Black Card: %s (Blanks: %d) [%s]\n", currentRound.BlackCard.Text, currentRound.BlackCard.Blanks, currentRound.BlackCard.Id)
			fmt.Println("Plays so far:")
			for _, play := range currentRound.Plays {
				cardTexts := []string{}
				for _, c := range play.Cards {
					cardTexts = append(cardTexts, fmt.Sprintf("%s [%s]", c.Text, c.Id))
				}
				fmt.Printf("  - Play %s (by %s): %s\n", play.Id, play.PlayerId, strings.Join(cardTexts, ", "))
			}
		}
	},
}

var gameHandCmd = &cobra.Command{
	Use:   "hand [game_id]",
	Short: "View your current hand",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		gameID := args[0]
		client, ctx, cancel, conn := getGameClientAndCtx(cmd)
		defer cancel()
		defer conn.Close()

		resp, err := client.GetHand(ctx, &pb.GetHandRequest{GameId: gameID})
		if err != nil {
			log.Fatalf("Failed to get hand: %v", err)
		}
		fmt.Println("Your Hand:")
		for i, card := range resp.Cards {
			fmt.Printf("%d: %s [%s]\n", i+1, card.Text, card.Id)
		}
	},
}

var gamePlayCardsCmd = &cobra.Command{
	Use:   "play-cards [game_id] [card_id...]",
	Short: "Submit your white cards for the round",
	Args:  cobra.MinimumNArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		gameID := args[0]
		cardIDs := args[1:]
		client, ctx, cancel, conn := getGameClientAndCtx(cmd)
		defer cancel()
		defer conn.Close()

		resp, err := client.PlayCards(ctx, &pb.PlayCardsRequest{
			GameId:  gameID,
			CardIds: cardIDs,
		})
		if err != nil {
			log.Fatalf("Failed to play cards: %v", err)
		}
		fmt.Printf("Successfully submitted play! Play ID: %s\n", resp.PlayId)
	},
}

var gameJudgeCmd = &cobra.Command{
	Use:   "judge [game_id] [play_id]",
	Short: "Select a winning play (if you are czar)",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		gameID := args[0]
		playID := args[1]
		client, ctx, cancel, conn := getGameClientAndCtx(cmd)
		defer cancel()
		defer conn.Close()

		_, err := client.SelectWinner(ctx, &pb.SelectWinnerRequest{
			GameId: gameID,
			PlayId: playID,
		})
		if err != nil {
			log.Fatalf("Failed to select winner: %v", err)
		}
		fmt.Println("Winner selected! Next round starting.")
	},
}

func init() {
	rootCmd.AddCommand(gameCmd)
	gameCmd.AddCommand(gameCreateCmd)
	gameCmd.AddCommand(gameStartCmd)
	gameCmd.AddCommand(gameStatusCmd)
	gameCmd.AddCommand(gameHandCmd)
	gameCmd.AddCommand(gamePlayCardsCmd)
	gameCmd.AddCommand(gameJudgeCmd)

	// Add persistent flag for token (since it's required to play)
	gameCmd.PersistentFlags().String("token", "", "Your auth token (fake auth uses player name)")
	gameCmd.MarkPersistentFlagRequired("token")
}
