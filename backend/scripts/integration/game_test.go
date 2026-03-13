package main

import (
	"context"
	"fmt"
	"testing"

	"google.golang.org/grpc/codes"
	pb "github.com/yanicksenn/ruthless/api/v1"
	"github.com/yanicksenn/ruthless/backend/scripts/testutil"
)

func runGameTests(t *testing.T, ctx context.Context, c *testutil.TestClient, runID string) {
	aliceCtx, err := c.GetAuthContextForUser(ctx, "Alice")
	if err != nil {
		t.Fatalf("Failed to get Alice context: %v", err)
	}
	
	ts, ok := c.UserTokenSources["Alice"]
	if !ok {
		t.Fatalf("Alice token source not initialized")
	}
	tok, _ := ts.Token()
	idToken, _ := testutil.GetIDToken(tok)
	aliceSub := testutil.GetSub(idToken)

	bobName := "GameBob_" + runID
	bobCtx := c.GetAuthContext(ctx, bobName)
	_, err = c.UserClient.Register(bobCtx, &pb.RegisterRequest{})
	testutil.AssertSuccess(t, err, "Register Bob (Fake)")

	t.Log("\n--- Session & Game Flow Suite ---")

	// 1. Setup Deck & Cards
	t.Log("  [RUN] Alice creates deck and cards...")
	deck, err := c.DeckClient.CreateDeck(aliceCtx, &pb.CreateDeckRequest{Name: "Game Deck"})
	testutil.AssertSuccess(t, err, "CreateDeck")
	
	bc, _ := c.CardClient.CreateCard(aliceCtx, &pb.CreateCardRequest{Text: "The best thing in life is ___."})
	c.DeckClient.AddCardToDeck(aliceCtx, &pb.AddCardToDeckRequest{DeckId: deck.Id, CardId: bc.Id})
	for i := 0; i < 20; i++ {
		card, _ := c.CardClient.CreateCard(aliceCtx, &pb.CreateCardRequest{Text: fmt.Sprintf("White Card %d", i)})
		c.DeckClient.AddCardToDeck(aliceCtx, &pb.AddCardToDeckRequest{DeckId: deck.Id, CardId: card.Id})
	}

	// 2. Setup Session
	t.Log("  [RUN] Alice creates session...")
	session, err := c.SessionClient.CreateSession(aliceCtx, &pb.CreateSessionRequest{})
	testutil.AssertSuccess(t, err, "CreateSession")
	c.SessionClient.AddDeckToSession(aliceCtx, &pb.AddDeckToSessionRequest{SessionId: session.Id, DeckId: deck.Id})

	// 3. Join players
	t.Log("  [RUN] Alice and Bob join...")
	_, err = c.SessionClient.JoinSession(aliceCtx, &pb.JoinSessionRequest{SessionId: session.Id})
	testutil.AssertSuccess(t, err, "Alice Join")
	_, err = c.SessionClient.JoinSession(bobCtx, &pb.JoinSessionRequest{SessionId: session.Id})
	testutil.AssertSuccess(t, err, "Bob Join")

	// 4. FAILURE: Bob (Non-owner) tries to start game
	t.Log("  [RUN] Bob (Non-owner) tries to start game...")
	game, err := c.GameClient.CreateGame(aliceCtx, &pb.CreateGameRequest{SessionId: session.Id})
	testutil.AssertSuccess(t, err, "CreateGame")
	_, err = c.GameClient.StartGame(bobCtx, &pb.StartGameRequest{Id: game.Id})
	testutil.AssertError(t, err, codes.PermissionDenied, "owner")

	// 5. SUCCESS: Alice starts game
	t.Log("  [RUN] Alice starts game...")
	_, err = c.GameClient.StartGame(aliceCtx, &pb.StartGameRequest{Id: game.Id})
	testutil.AssertSuccess(t, err, "Alice StartGame")

	// 6. Czar Roles
	game, _ = c.GameClient.GetGame(aliceCtx, &pb.GetGameRequest{Id: game.Id})
	currentRound := game.Rounds[len(game.Rounds)-1]
	czarID := currentRound.CzarId
	
	var playerCtx, czarCtx context.Context
	if czarID == aliceSub {
		czarCtx = aliceCtx
		playerCtx = bobCtx
	} else {
		czarCtx = bobCtx
		playerCtx = aliceCtx
	}

	// 7. FAILURE: Czar plays card
	t.Logf("  [RUN] Czar (%s) attempts to play card...", czarID)
	_, err = c.GameClient.PlayCards(czarCtx, &pb.PlayCardsRequest{GameId: game.Id, CardIds: []string{"some-card"}})
	testutil.AssertError(t, err, codes.PermissionDenied, "czar")

	// 8. SUCCESS: Player plays card
	t.Log("  [RUN] Player plays card...")
	handResp, err := c.GameClient.GetHand(playerCtx, &pb.GetHandRequest{GameId: game.Id})
	testutil.AssertSuccess(t, err, "GetHand")
	if len(handResp.Cards) == 0 {
		t.Fatalf("Player has no cards in hand")
	}

	// Submit correct number of cards
	blanks := testutil.CountBlanks(currentRound.BlackCard.Text)
	if blanks == 0 {
		blanks = 1
	}

	var cardIDs []string
	for i := 0; i < blanks; i++ {
		cardIDs = append(cardIDs, handResp.Cards[i].Id)
	}

	_, err = c.GameClient.PlayCards(playerCtx, &pb.PlayCardsRequest{
		GameId:  game.Id,
		CardIds: cardIDs,
	})
	testutil.AssertSuccess(t, err, "Player PlayCards")

	// Verify state transition
	game, _ = c.GameClient.GetGame(aliceCtx, &pb.GetGameRequest{Id: game.Id})
	if game.State != pb.GameState_GAME_STATE_JUDGING {
		t.Errorf("Expected game state JUDGING, got %v", game.State)
	}

	// 9. FAILURE: Player judges
	t.Log("  [RUN] Regular Player attempts to judge...")
	_, err = c.GameClient.SelectWinner(playerCtx, &pb.SelectWinnerRequest{GameId: game.Id, PlayId: "someone"})
	testutil.AssertError(t, err, codes.PermissionDenied, "czar")
}
