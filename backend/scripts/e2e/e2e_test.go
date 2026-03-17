package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"testing"
	"time"

	pb "github.com/yanicksenn/ruthless/api/v1"
	"github.com/yanicksenn/ruthless/backend/internal/storage"
	"github.com/yanicksenn/ruthless/backend/scripts/testutil"
)

var (
	addr           = flag.String("addr", "", "the address to connect to (if empty, starts in-process server)")
	clientSecret   = flag.String("client-secret", "", "path to Google Client Secret JSON")
	aliceRefresh   = flag.String("alice-refresh", "", "path to Alice's refresh token")
	bobRefresh     = flag.String("bob-refresh", "", "path to Bob's refresh token")
	googleAudience = flag.String("google-audience", "964146605436-cn068f6livloacebbi0itbhgvh9t5uns.apps.googleusercontent.com", "Google OAuth Client ID")
)

func TestE2E(t *testing.T) {
	if !flag.Parsed() {
		flag.Parse()
	}

	// Fallback for local dev if flags are empty
	if *clientSecret == "" {
		*clientSecret = "secrets/client_secret_dev.json"
	}
	if *aliceRefresh == "" {
		*aliceRefresh = "secrets/ruthless.alice.sec"
	}
	if *bobRefresh == "" {
		*bobRefresh = "secrets/ruthless.bob.sec"
	}

	// Resolve paths relative to execution root if needed (for Bazel)
	for _, p := range []*string{clientSecret, aliceRefresh, bobRefresh} {
		if _, err := os.Stat(*p); os.IsNotExist(err) {
			// Try going up 3 levels (backend/scripts/e2e -> root)
			newPath := "../../../" + *p
			if _, err := os.Stat(newPath); err == nil {
				*p = newPath
			}
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	var connectAddr string
	var cleanup func()
	var store storage.Storage

	if *addr != "" {
		connectAddr = *addr
	} else {
		var err error
		connectAddr, store, cleanup, err = testutil.StartTestServer(ctx, *googleAudience)
		if err != nil {
			t.Fatalf("Failed to start test server: %v", err)
		}
		defer cleanup()
	}

	client := testutil.NewTestClient(connectAddr, "")
	defer client.Close()

	// 0. Setup Alice
	if *aliceRefresh != "" {
		ts, err := testutil.TokenSourceFromRefresh(ctx, *clientSecret, *aliceRefresh)
		if err != nil {
			t.Fatalf("Failed to create Alice token source: %v", err)
		}
		client.UserTokenSources["Alice"] = ts
	} else {
		t.Fatalf("Alice refresh token is required for E2E tests")
	}

	aliceSub, err := client.EnsureUserRegistered(ctx, "Alice", store)
	if err != nil {
		t.Fatalf("Failed to ensure Alice is registered: %v", err)
	}

	aliceCtx, err := client.GetAuthContextForUser(ctx, "Alice")
	if err != nil {
		t.Fatalf("Failed to get Alice Auth Context: %v", err)
	}

	// Setup Bob (using real refresh token if available, else fail for E2E)
	var bobSub string
	var bobCtx context.Context
	if *bobRefresh != "" {
		ts, err := testutil.TokenSourceFromRefresh(ctx, *clientSecret, *bobRefresh)
		if err != nil {
			t.Fatalf("Failed to create Bob token source: %v", err)
		}
		client.UserTokenSources["Bob"] = ts
		bobSub, err = client.EnsureUserRegistered(ctx, "Bob", store)
		if err != nil {
			t.Fatalf("Failed to ensure Bob is registered: %v", err)
		}
		bobCtx, err = client.GetAuthContextForUser(ctx, "Bob")
		if err != nil {
			t.Fatalf("Failed to get Bob Auth Context: %v", err)
		}
	} else {
		t.Fatalf("Bob refresh token is required for E2E tests")
	}

	t.Log("--- Starting E2E Validation ---")

	// 1. Alice creates cards
	t.Log("1. Creating cards...")
	deck, _ := client.DeckClient.CreateDeck(aliceCtx, &pb.CreateDeckRequest{Name: "E2E Deck"})
	bc, _ := client.CardClient.CreateCard(aliceCtx, &pb.CreateCardRequest{Text: "E2E Black Card ___"})
	client.DeckClient.AddCardToDeck(aliceCtx, &pb.AddCardToDeckRequest{DeckId: deck.Id, CardId: bc.Id})
	
	for i := 0; i < 20; i++ {
		wc, _ := client.CardClient.CreateCard(aliceCtx, &pb.CreateCardRequest{Text: fmt.Sprintf("E2E White Card %d", i)})
		client.DeckClient.AddCardToDeck(aliceCtx, &pb.AddCardToDeckRequest{DeckId: deck.Id, CardId: wc.Id})
	}

	// 2. Play Flow
	t.Log("2. Setting up session...")
	session, _ := client.SessionClient.CreateSession(aliceCtx, &pb.CreateSessionRequest{})
	client.SessionClient.AddDeckToSession(aliceCtx, &pb.AddDeckToSessionRequest{SessionId: session.Id, DeckId: deck.Id})

	client.SessionClient.JoinSession(aliceCtx, &pb.JoinSessionRequest{SessionId: session.Id})
	client.SessionClient.JoinSession(bobCtx, &pb.JoinSessionRequest{SessionId: session.Id})

	game, _ := client.GameClient.GetGameBySession(aliceCtx, &pb.GetGameBySessionRequest{SessionId: session.Id})
	client.GameClient.StartGame(aliceCtx, &pb.StartGameRequest{Id: game.Id})


	// Czar roles
	game, _ = client.GameClient.GetGame(aliceCtx, &pb.GetGameRequest{Id: game.Id})
	czarID := game.Rounds[len(game.Rounds)-1].CzarId

	var playerCtx context.Context
	var playerSub string
	if czarID == aliceSub {
		playerCtx = bobCtx
		playerSub = bobSub
	} else {
		playerCtx = aliceCtx
		playerSub = aliceSub
	}

	t.Logf("3. Player %s plays card...", playerSub)
	handResp, err := client.GameClient.GetHand(playerCtx, &pb.GetHandRequest{GameId: game.Id})
	if err != nil {
		t.Fatalf("GetHand failed: %v", err)
	}
	if len(handResp.Cards) == 0 {
		t.Fatalf("Player has no cards in hand")
	}

	playResp, err := client.GameClient.PlayCards(playerCtx, &pb.PlayCardsRequest{
		GameId:  game.Id,
		CardIds: []string{handResp.Cards[0].Id},
	})
	if err != nil {
		t.Fatalf("PlayCards failed: %v", err)
	}

	t.Log("4. Czar selects winner...")
	var czarCtx context.Context
	if czarID == bobSub {
		czarCtx = bobCtx
	} else {
		czarCtx = aliceCtx
	}
	_, err = client.GameClient.SelectWinner(czarCtx, &pb.SelectWinnerRequest{
		GameId: game.Id,
		PlayId: playResp.PlayId,
	})
	if err != nil {
		t.Fatalf("SelectWinner failed: %v", err)
	}

	t.Log("✅ E2E Validation Successful!")
}
