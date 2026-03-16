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
	aliceCtx := c.GetAuthContext(ctx, "Alice")
	aliceSub := "Alice"

	bobName := "GameBob_" + runID
	bobCtx := c.GetAuthContext(ctx, bobName)
	_, err := c.UserClient.Register(bobCtx, &pb.RegisterRequest{})
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
	t.Log("  [RUN] Alice creates session with deck...")
	session, err := c.SessionClient.CreateSession(aliceCtx, &pb.CreateSessionRequest{
		DeckIds: []string{deck.Id},
	})
	testutil.AssertSuccess(t, err, "CreateSession with deck")
	
	// Verify deck is in session
	if len(session.DeckIds) != 1 || session.DeckIds[0] != deck.Id {
		t.Errorf("Expected 1 deck %s in session, got %v", deck.Id, session.DeckIds)
	}

	// 3. Join players
	t.Log("  [RUN] Alice and Bob join...")
	_, err = c.SessionClient.JoinSession(aliceCtx, &pb.JoinSessionRequest{SessionId: session.Id})
	testutil.AssertSuccess(t, err, "Alice Join")
	_, err = c.SessionClient.JoinSession(bobCtx, &pb.JoinSessionRequest{SessionId: session.Id})
	testutil.AssertSuccess(t, err, "Bob Join")

	// 3.5 Bob leaves and re-joins
	t.Log("  [RUN] Bob leaves session...")
	_, err = c.SessionClient.LeaveSession(bobCtx, &pb.LeaveSessionRequest{SessionId: session.Id})
	testutil.AssertSuccess(t, err, "Bob Leave")

	// Verify Bob is gone
	session, _ = c.SessionClient.GetSession(aliceCtx, &pb.GetSessionRequest{Id: session.Id})
	if len(session.PlayerIds) != 1 {
		t.Errorf("Expected 1 player after Bob left, got %d", len(session.PlayerIds))
	}

	t.Log("  [RUN] Bob re-joins session...")
	_, err = c.SessionClient.JoinSession(bobCtx, &pb.JoinSessionRequest{SessionId: session.Id})
	testutil.AssertSuccess(t, err, "Bob Join again")

	// 4. Create Game
	t.Log("  [RUN] Alice creates game...")
	game, err := c.GameClient.CreateGame(aliceCtx, &pb.CreateGameRequest{SessionId: session.Id})
	testutil.AssertSuccess(t, err, "CreateGame")

	// 5. Late Join: Bob joins AFTER CreateGame
	t.Log("  [RUN] Bob late joins after CreateGame...")
	_, err = c.SessionClient.JoinSession(bobCtx, &pb.JoinSessionRequest{SessionId: session.Id})
	testutil.AssertSuccess(t, err, "Bob Late Join")

	// 5.5 Late Deck: Alice adds another deck AFTER CreateGame
	t.Log("  [RUN] Alice adds another deck after CreateGame...")
	deck2, err := c.DeckClient.CreateDeck(aliceCtx, &pb.CreateDeckRequest{Name: "Late Deck"})
	testutil.AssertSuccess(t, err, "CreateDeck 2")
	for i := 0; i < 20; i++ {
		card, _ := c.CardClient.CreateCard(aliceCtx, &pb.CreateCardRequest{Text: fmt.Sprintf("Late White %d", i)})
		c.DeckClient.AddCardToDeck(aliceCtx, &pb.AddCardToDeckRequest{DeckId: deck2.Id, CardId: card.Id})
	}
	_, err = c.SessionClient.AddDeckToSession(aliceCtx, &pb.AddDeckToSessionRequest{SessionId: session.Id, DeckId: deck2.Id})
	testutil.AssertSuccess(t, err, "AddDeckToSession 2")

	// 6. FAILURE: Bob (Non-owner) tries to start game
	t.Log("  [RUN] Bob (Non-owner) tries to start game...")
	_, err = c.GameClient.StartGame(bobCtx, &pb.StartGameRequest{Id: game.Id})
	testutil.AssertError(t, err, codes.PermissionDenied, "owner")

	// 7. SUCCESS: Alice starts game (should sync Bob)
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
