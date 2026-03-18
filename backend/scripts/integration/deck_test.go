package main

import (
	"context"
	"testing"

	"google.golang.org/grpc/codes"
	pb "github.com/yanicksenn/ruthless/api/v1"
	"github.com/yanicksenn/ruthless/backend/scripts/testutil"
)

func runDeckTests(t *testing.T, ctx context.Context, c *testutil.TestClient, runID string) {
	aliceCtx := c.GetAuthContext(ctx, "Alice")
	aliceSub := "Alice"

	bobName := "DeckBob_" + runID
	bobCtx := c.GetAuthContext(ctx, bobName)
	if c.Store != nil {
		c.Store.CreateUser(ctx, &pb.User{Id: bobName, Name: bobName})
	}
	_, err := c.UserClient.CompleteRegistration(bobCtx, &pb.CompleteRegistrationRequest{Name: bobName})
	testutil.AssertSuccess(t, err, "CompleteRegistration Bob (Fake)")

	t.Log("\n--- Deck & Card Suite ---")

	// 1. Alice creates a card
	t.Logf("  [RUN] Alice creates card...")
	aliceCard, err := c.CardClient.CreateCard(aliceCtx, &pb.CreateCardRequest{Text: "Alice's Secret"})
	testutil.AssertSuccess(t, err, "Alice CreateCard")

	// 2. Unregistered User tries to create card
	charlieName := "DeckCharlie_" + runID
	t.Logf("  [RUN] Unregistered User (%s) creates card...", charlieName)
	charlieCtx := c.GetAuthContext(ctx, charlieName)
	_, err = c.CardClient.CreateCard(charlieCtx, &pb.CreateCardRequest{Text: "Charlie's Ghost"})
	testutil.AssertError(t, err, codes.Unauthenticated, "user profile not found")

	// 3. Alice creates a deck
	t.Log("  [RUN] Alice creates deck...")
	deck, err := c.DeckClient.CreateDeck(aliceCtx, &pb.CreateDeckRequest{Name: "Test Deck"})
	testutil.AssertSuccess(t, err, "CreateDeck")

	// 4. FAILURE: Bob (Non-contributor) tries to add Alice's card
	t.Logf("  [DEBUG] bobName=%s, aliceCardOwner=%s", bobName, aliceCard.OwnerId)
	_, err = c.DeckClient.AddCardToDeck(bobCtx, &pb.AddCardToDeckRequest{DeckId: deck.Id, CardId: aliceCard.Id})
	if err == nil {
		t.Fatalf("Expected error for non-contributor/non-owner, but got success. bobName=%s, aliceCardOwner=%s", bobName, aliceCard.OwnerId)
	}
	testutil.AssertError(t, err, codes.PermissionDenied, "user is not authorized to modify this deck")

	// 5. SUCCESS: Alice adds Bob as contributor
	t.Log("  [RUN] Alice adds Bob as contributor...")
	_, err = c.DeckClient.AddContributor(aliceCtx, &pb.AddContributorRequest{DeckId: deck.Id, ContributorId: bobName})
	testutil.AssertSuccess(t, err, "AddContributor")

	// 6. SUCCESS: Bob (Now contributor) adds card
	t.Log("  [RUN] Bob (contributor) adds card...")
	bobCard, err := c.CardClient.CreateCard(bobCtx, &pb.CreateCardRequest{Text: "Bob's Secret"})
	testutil.AssertSuccess(t, err, "CreateCard Bob")
	_, err = c.DeckClient.AddCardToDeck(bobCtx, &pb.AddCardToDeckRequest{DeckId: deck.Id, CardId: bobCard.Id})
	testutil.AssertSuccess(t, err, "AddCardToDeck Bob")

	// 7. FAILURE: Unauthenticated creation
	t.Log("  [RUN] Unauthenticated creation...")
	_, err = c.DeckClient.CreateDeck(ctx, &pb.CreateDeckRequest{Name: "Fail Deck"})
	testutil.AssertError(t, err, codes.Unauthenticated, "Authorization header")

	// 8. FAILURE: Contributor removes contributor (only owner)
	t.Log("  [RUN] Bob (contributor) tries to remove Alice (owner)...")
	_, err = c.DeckClient.RemoveContributor(bobCtx, &pb.RemoveContributorRequest{DeckId: deck.Id, ContributorId: aliceSub})
	testutil.AssertError(t, err, codes.PermissionDenied, "authorized")

	// 9. FAILURE: Non-contributor Charlie tries to remove card
	t.Log("  [RUN] Charlie tries to remove Bob's card...")
	if c.Store != nil {
		c.Store.CreateUser(ctx, &pb.User{Id: charlieName, Name: charlieName})
	}
	_, _ = c.UserClient.CompleteRegistration(charlieCtx, &pb.CompleteRegistrationRequest{Name: charlieName})
	_, err = c.DeckClient.RemoveCardFromDeck(charlieCtx, &pb.RemoveCardFromDeckRequest{DeckId: deck.Id, CardId: bobCard.Id})
	testutil.AssertError(t, err, codes.PermissionDenied, "not authorized")

	// 10. SUCCESS: Bob removes card
	t.Log("  [RUN] Bob removes his card...")
	_, err = c.DeckClient.RemoveCardFromDeck(bobCtx, &pb.RemoveCardFromDeckRequest{DeckId: deck.Id, CardId: bobCard.Id})
	testutil.AssertSuccess(t, err, "RemoveCardFromDeck Bob")

	// 11. Verify card is gone
	deckPostRemove, err := c.DeckClient.GetDeck(aliceCtx, &pb.GetDeckRequest{Id: deck.Id})
	testutil.AssertSuccess(t, err, "GetDeck after remove")
	for _, id := range deckPostRemove.CardIds {
		if id == bobCard.Id {
			t.Fatalf("Card %s should have been removed from deck", bobCard.Id)
		}
	}
}
