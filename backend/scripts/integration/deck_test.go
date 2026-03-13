package main

import (
	"context"
	"testing"

	"google.golang.org/grpc/codes"
	pb "github.com/yanicksenn/ruthless/api/v1"
	"github.com/yanicksenn/ruthless/backend/scripts/testutil"
)

func runDeckTests(t *testing.T, ctx context.Context, c *testutil.TestClient, runID string) {
	aliceCtx, err := c.GetAuthContextForUser(ctx, "Alice")
	if err != nil {
		t.Fatalf("Failed to get Alice context: %v", err)
	}
	tok, _ := c.UserTokenSources["Alice"].Token()
	idToken, _ := testutil.GetIDToken(tok)
	aliceSub := testutil.GetSub(idToken)

	bobName := "DeckBob_" + runID
	bobCtx := c.GetAuthContext(ctx, bobName)
	_, err = c.UserClient.Register(bobCtx, &pb.RegisterRequest{})
	testutil.AssertSuccess(t, err, "Register Bob (Fake)")

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
	testutil.AssertError(t, err, codes.PermissionDenied, "user not registered")

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
	testutil.AssertError(t, err, codes.PermissionDenied, "you do not own this card")

	// 5. SUCCESS: Alice adds Bob as contributor
	t.Log("  [RUN] Alice adds Bob as contributor...")
	_, err = c.DeckClient.AddContributor(aliceCtx, &pb.AddContributorRequest{DeckId: deck.Id, ContributorId: bobName})
	testutil.AssertSuccess(t, err, "AddContributor")
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
}
