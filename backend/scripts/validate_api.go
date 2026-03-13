package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"

	"github.com/golang-jwt/jwt/v5"
	pb "github.com/yanicksenn/ruthless/api/v1"
)

var (
	addr       = flag.String("addr", "localhost:8080", "the address to connect to")
	authSecret = flag.String("auth-secret", "", "JWT secret (if provided, generates JWT tokens)")
)

func main() {
	flag.Parse()

	// Set up a connection to the server.
	conn, err := grpc.NewClient(*addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()

	cardClient := pb.NewCardServiceClient(conn)
	deckClient := pb.NewDeckServiceClient(conn)
	sessionClient := pb.NewSessionServiceClient(conn)
	gameClient := pb.NewGameServiceClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	fmt.Print("--- Starting E2E Validation ---\r\n")

	signToken := func(name string) string {
		if *authSecret == "" {
			return name
		}
		t := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
			"sub":  name,
			"name": name,
		})
		signed, err := t.SignedString([]byte(*authSecret))
		if err != nil {
			log.Fatalf("could not sign token for %s: %v", name, err)
		}
		return signed
	}

	// 1. Create some cards
	fmt.Print("1. Creating cards...\r\n")
	blackCards := make([]*pb.Card, 10)
	for i := 0; i < 10; i++ {
		bc, err := cardClient.CreateCard(ctx, &pb.CreateCardRequest{Text: fmt.Sprintf("Black Card %d with blank ___", i)})
		if err != nil {
			log.Fatalf("could not create black card %d: %v", i, err)
		}
		blackCards[i] = bc
	}
	fmt.Print("   Created 10 black cards.\r\n")

	whiteCards := make([]*pb.Card, 100)
	for i := 0; i < 100; i++ {
		wc, err := cardClient.CreateCard(ctx, &pb.CreateCardRequest{Text: fmt.Sprintf("White Card %d", i)})
		if err != nil {
			log.Fatalf("could not create white card %d: %v", err)
		}
		whiteCards[i] = wc
	}
	fmt.Print("   Created 100 white cards.\r\n")

	// 2. Create a deck
	fmt.Print("2. Creating deck...\r\n")
	aliceCtx := metadata.AppendToOutgoingContext(ctx, "authorization", "Bearer "+signToken("Alice"))
	deck, err := deckClient.CreateDeck(aliceCtx, &pb.CreateDeckRequest{Name: "E2E Test Deck"})
	if err != nil {
		log.Fatalf("could not create deck: %v", err)
	}
	fmt.Printf("   Created deck: %s\r\n", deck.Id)

	// 3. Add cards to deck
	fmt.Print("3. Adding cards to deck...\r\n")
	for _, bc := range blackCards {
		_, err = deckClient.AddCardToDeck(aliceCtx, &pb.AddCardToDeckRequest{DeckId: deck.Id, CardId: bc.Id})
		if err != nil {
			log.Fatalf("could not add black card %s to deck: %v", bc.Id, err)
		}
	}
	for _, wc := range whiteCards {
		_, err = deckClient.AddCardToDeck(aliceCtx, &pb.AddCardToDeckRequest{DeckId: deck.Id, CardId: wc.Id})
		if err != nil {
			log.Fatalf("could not add white card %s to deck: %v", wc.Id, err)
		}
	}

	// 4. Create session and join
	fmt.Print("4. Setting up session...\r\n")
	session, err := sessionClient.CreateSession(ctx, &pb.CreateSessionRequest{})
	if err != nil {
		log.Fatalf("could not create session: %v", err)
	}
	fmt.Printf("   Created session: %s\r\n", session.Id)

	_, err = sessionClient.AddDeckToSession(ctx, &pb.AddDeckToSessionRequest{SessionId: session.Id, DeckId: deck.Id})
	if err != nil {
		log.Fatalf("could not add deck to session: %v", err)
	}

	playerTokens := []string{"Bob", "Alice", "Charlie"}
	playerContexts := make(map[string]context.Context)
	for _, token := range playerTokens {
		pCtx := metadata.AppendToOutgoingContext(ctx, "authorization", "Bearer "+signToken(token))
		_, err = sessionClient.JoinSession(pCtx, &pb.JoinSessionRequest{SessionId: session.Id, PlayerName: token})
		if err != nil {
			log.Fatalf("%s could not join session: %v", token, err)
		}
		playerContexts[token] = pCtx
	}
	fmt.Print("   Three players joined.\r\n")

	// 5. Create and start game
	fmt.Print("5. Starting game...\r\n")
	session, err = sessionClient.GetSession(ctx, &pb.GetSessionRequest{Id: session.Id})
	if err != nil {
		log.Fatalf("could not get session: %v", err)
	}
	fmt.Printf("   Session players: %v\r\n", session.PlayerIds)

	game, err := gameClient.CreateGame(ctx, &pb.CreateGameRequest{SessionId: session.Id})
	if err != nil {
		log.Fatalf("could not create game: %v", err)
	}
	fmt.Printf("   Created game: %s\r\n", game.Id)

	_, err = gameClient.StartGame(aliceCtx, &pb.StartGameRequest{Id: game.Id})
	if err != nil {
		log.Fatalf("could not start game: %v", err)
	}

	// Fetch game state to find czar
	game, err = gameClient.GetGame(ctx, &pb.GetGameRequest{Id: game.Id})
	if err != nil {
		log.Fatalf("could not get game state: %v", err)
	}
	currentRound := game.Rounds[len(game.Rounds)-1]
	czarID := currentRound.CzarId
	fmt.Printf("   Czar for this round: %s\r\n", czarID)

	// 6. Non-czars play cards
	fmt.Print("6. Non-czars playing cards...\r\n")
	var lastPlayID string
	for _, token := range playerTokens {
		// In fake auth, ID == Token
		if token == czarID {
			continue
		}
		pCtx := playerContexts[token]
		hand, err := gameClient.GetHand(pCtx, &pb.GetHandRequest{GameId: game.Id})
		if err != nil {
			log.Fatalf("could not get %s's hand: %v", token, err)
		}
		if len(hand.Cards) == 0 {
			log.Fatalf("%s has no cards in hand", token)
		}

		playResp, err := gameClient.PlayCards(pCtx, &pb.PlayCardsRequest{
			GameId:  game.Id,
			CardIds: []string{hand.Cards[0].Id},
		})
		if err != nil {
			log.Fatalf("%s could not play card: %v", token, err)
		}
		fmt.Printf("   %s played a card: %s\r\n", token, playResp.PlayId)
		lastPlayID = playResp.PlayId
	}

	// 7. Judge
	fmt.Print("7. Judging...\r\n")
	czarCtx := playerContexts[czarID]
	_, err = gameClient.SelectWinner(czarCtx, &pb.SelectWinnerRequest{
		GameId: game.Id,
		PlayId: lastPlayID,
	})
	if err != nil {
		log.Fatalf("Judging failed: %v", err)
	}
	fmt.Print("   Judging successful!\r\n")

	// 8. Verify persistence / Round transition
	fmt.Print("8. Verifying round transition...\r\n")
	game, err = gameClient.GetGame(ctx, &pb.GetGameRequest{Id: game.Id})
	if err != nil {
		log.Fatalf("could not get game state: %v", err)
	}
	if len(game.Rounds) != 2 {
		log.Fatalf("expected 2 rounds, got %d", len(game.Rounds))
	}
	fmt.Print("   Successfully transitioned to Round 2.\r\n")

	fmt.Print("--- E2E Validation Successful! ---\r\n")
}
