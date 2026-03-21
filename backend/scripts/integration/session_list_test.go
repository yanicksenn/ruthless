package main

import (
	"context"
	"testing"

	pb "github.com/yanicksenn/ruthless/api/v1"
	"github.com/yanicksenn/ruthless/backend/scripts/testutil"
)

func runSessionListTests(t *testing.T, ctx context.Context, c *testutil.TestClient, runID string) {
	aliceName := "Alice_List_" + runID
	aliceCtx := c.GetAuthContext(ctx, aliceName)
	if c.Store != nil {
		c.Store.CreateUser(ctx, &pb.User{Id: aliceName, Name: aliceName})
	}
	c.UserClient.CompleteRegistration(aliceCtx, &pb.CompleteRegistrationRequest{Name: aliceName})

	bobName := "Bob_List_" + runID
	bobCtx := c.GetAuthContext(ctx, bobName)
	if c.Store != nil {
		c.Store.CreateUser(ctx, &pb.User{Id: bobName, Name: bobName})
	}
	c.UserClient.CompleteRegistration(bobCtx, &pb.CompleteRegistrationRequest{Name: bobName})

	t.Log("\n--- Session List Integration Suite ---")

	// 1. Setup Deck & Cards for testing
	deck, _ := c.DeckClient.CreateDeck(aliceCtx, &pb.CreateDeckRequest{Name: "List Deck"})
	bc, _ := c.CardClient.CreateCard(aliceCtx, &pb.CreateCardRequest{Text: "Test Black ___"})
	c.DeckClient.AddCardToDeck(aliceCtx, &pb.AddCardToDeckRequest{DeckId: deck.Id, CardId: bc.Id})
	for i := 0; i < 20; i++ {
		card, _ := c.CardClient.CreateCard(aliceCtx, &pb.CreateCardRequest{Text: "Test White"})
		c.DeckClient.AddCardToDeck(aliceCtx, &pb.AddCardToDeckRequest{DeckId: deck.Id, CardId: card.Id})
	}

	// 2. Alice creates a session (WAITING status)
	t.Log("  [RUN] Alice creates session (WAITING)...")
	session, err := c.SessionClient.CreateSession(aliceCtx, &pb.CreateSessionRequest{
		Name:    "Alice's Waiting Session",
		DeckIds: []string{deck.Id},
	})
	testutil.AssertSuccess(t, err, "CreateSession")

	// 3. Verify both Alice and Bob see the session
	t.Log("  [RUN] Verifying session visibility for Alice and Bob...")
	aliceList, err := c.SessionClient.ListSessions(aliceCtx, &pb.ListSessionsRequest{})
	testutil.AssertSuccess(t, err, "Alice ListSessions")
	
	foundForAlice := false
	for _, s := range aliceList.Sessions {
		if s.Id == session.Id {
			foundForAlice = true
			break
		}
	}
	if !foundForAlice {
		t.Errorf("Alice should see her own WAITING session")
	}

	bobList, err := c.SessionClient.ListSessions(bobCtx, &pb.ListSessionsRequest{})
	testutil.AssertSuccess(t, err, "Bob ListSessions")
	
	foundForBob := false
	for _, s := range bobList.Sessions {
		if s.Id == session.Id {
			foundForBob = true
			break
		}
	}
	if !foundForBob {
		t.Errorf("Bob should see Alice's WAITING session")
	}

	// 4. Start the game (Alice and Bob join first)
	t.Log("  [RUN] Alice and Bob join and Alice starts the game (PLAYING)...")
	c.SessionClient.JoinSession(aliceCtx, &pb.JoinSessionRequest{SessionId: session.Id})
	c.SessionClient.JoinSession(bobCtx, &pb.JoinSessionRequest{SessionId: session.Id})
	
	game, _ := c.GameClient.GetGameBySession(aliceCtx, &pb.GetGameBySessionRequest{SessionId: session.Id})
	_, err = c.GameClient.StartGame(aliceCtx, &pb.StartGameRequest{Id: game.Id})
	testutil.AssertSuccess(t, err, "StartGame")

	// 5. Create a 3rd user (Charlie) who is NOT in the session
	charlieName := "Charlie_List_" + runID
	charlieCtx := c.GetAuthContext(ctx, charlieName)
	if c.Store != nil {
		c.Store.CreateUser(ctx, &pb.User{Id: charlieName, Name: charlieName})
	}
	c.UserClient.CompleteRegistration(charlieCtx, &pb.CompleteRegistrationRequest{Name: charlieName})

	// 6. Verify Alice and Bob see the PLAYING session, but Charlie does NOT
	t.Log("  [RUN] Verifying session visibility for participants vs non-participants...")
	
	aliceList, _ = c.SessionClient.ListSessions(aliceCtx, &pb.ListSessionsRequest{
		View: pb.SessionView_SESSION_VIEW_ACTIVE,
	})
	foundForAlice = false
	for _, s := range aliceList.Sessions {
		if s.Id == session.Id {
			foundForAlice = true
			break
		}
	}
	if !foundForAlice {
		t.Errorf("Alice should see the PLAYING session she is in")
	}

	charlieList, err := c.SessionClient.ListSessions(charlieCtx, &pb.ListSessionsRequest{
		View: pb.SessionView_SESSION_VIEW_ACTIVE,
	})
	testutil.AssertSuccess(t, err, "Charlie ListSessions")
	
	foundForCharlie := false
	for _, s := range charlieList.Sessions {
		if s.Id == session.Id {
			foundForCharlie = true
			break
		}
	}
	if foundForCharlie {
		t.Errorf("Charlie should NOT see the PLAYING session he is not in")
	}

    t.Log("  [OK] Session filtering logic verified")
}
