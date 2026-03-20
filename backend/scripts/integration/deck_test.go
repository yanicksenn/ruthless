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
	if c.Store != nil {
		c.Store.CreateUser(ctx, &pb.User{Id: aliceSub, Name: aliceSub})
	}
	aliceUser, err := c.UserClient.CompleteRegistration(aliceCtx, &pb.CompleteRegistrationRequest{Name: "Alice"})
	testutil.AssertSuccess(t, err, "CompleteRegistration Alice (Fake)")

	bobName := "DeckBob_" + runID
	bobCtx := c.GetAuthContext(ctx, bobName)
	if c.Store != nil {
		c.Store.CreateUser(ctx, &pb.User{Id: bobName, Name: bobName})
	}
	bobUser, err := c.UserClient.CompleteRegistration(bobCtx, &pb.CompleteRegistrationRequest{Name: bobName})
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
	_, err = c.DeckClient.AddContributor(aliceCtx, &pb.AddContributorRequest{DeckId: deck.Id, Identifier: bobUser.Identifier})
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
	_, err = c.DeckClient.RemoveContributor(bobCtx, &pb.RemoveContributorRequest{DeckId: deck.Id, Identifier: aliceUser.Identifier})
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

	// 12. SUCCESS: Charlie subscribes to Alice's deck
	t.Log("  [RUN] Charlie subscribes to Alice's deck...")
	_, err = c.DeckClient.SubscribeToDeck(charlieCtx, &pb.SubscribeToDeckRequest{DeckId: deck.Id})
	testutil.AssertSuccess(t, err, "SubscribeToDeck Charlie")

	// 13. Verify Charlie sees the deck in ListDecks
	t.Log("  [RUN] Charlie lists decks and finds Alice's deck...")
	listRes, err := c.DeckClient.ListDecks(charlieCtx, &pb.ListDecksRequest{})
	testutil.AssertSuccess(t, err, "ListDecks Charlie")
	found := false
	for _, d := range listRes.Decks {
		if d.Id == deck.Id {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("Charlie should see deck %s in ListDecks after subscribing", deck.Id)
	}

	// 14. FAILURE: Alice tries to subscribe to her own deck
	t.Log("  [RUN] Alice tries to subscribe to her own deck...")
	_, err = c.DeckClient.SubscribeToDeck(aliceCtx, &pb.SubscribeToDeckRequest{DeckId: deck.Id})
	testutil.AssertError(t, err, codes.InvalidArgument, "cannot subscribe to your own deck")

	// 15. SUCCESS: Charlie unsubscribes from Alice's deck
	t.Log("  [RUN] Charlie unsubscribes from Alice's deck...")
	_, err = c.DeckClient.UnsubscribeFromDeck(charlieCtx, &pb.UnsubscribeFromDeckRequest{DeckId: deck.Id})
	testutil.AssertSuccess(t, err, "UnsubscribeFromDeck Charlie")

	// 16. Verify Charlie no longer sees the deck
	listRes, err = c.DeckClient.ListDecks(charlieCtx, &pb.ListDecksRequest{})
	testutil.AssertSuccess(t, err, "ListDecks Charlie post-unsubscribe")
	for _, d := range listRes.Decks {
		if d.Id == deck.Id {
			t.Fatalf("Charlie should not see deck %s in ListDecks after unsubscribing", deck.Id)
		}
	}

	// 17. Verify card contributor tracking
	t.Log("  [RUN] Alice adds card and Bob adds card, verify contributors...")
	aliceCard2, err := c.CardClient.CreateCard(aliceCtx, &pb.CreateCardRequest{Text: "Alice's Second Secret"})
	testutil.AssertSuccess(t, err, "Alice CreateCard 2")
	_, err = c.DeckClient.AddCardToDeck(aliceCtx, &pb.AddCardToDeckRequest{DeckId: deck.Id, CardId: aliceCard2.Id})
	testutil.AssertSuccess(t, err, "Alice AddCardToDeck 2")

	bobCard2, err := c.CardClient.CreateCard(bobCtx, &pb.CreateCardRequest{Text: "Bob's Second Secret"})
	testutil.AssertSuccess(t, err, "Bob CreateCard 2")
	_, err = c.DeckClient.AddCardToDeck(bobCtx, &pb.AddCardToDeckRequest{DeckId: deck.Id, CardId: bobCard2.Id})
	testutil.AssertSuccess(t, err, "Bob AddCardToDeck 2")

	deckFinal, err := c.DeckClient.GetDeck(aliceCtx, &pb.GetDeckRequest{Id: deck.Id})
	testutil.AssertSuccess(t, err, "GetDeck final")

	if deckFinal.CardContributorIds[aliceCard2.Id] != aliceUser.Id {
		t.Fatalf("Expected Alice (%s) to be contributor for card %s, but got %s", aliceUser.Id, aliceCard2.Id, deckFinal.CardContributorIds[aliceCard2.Id])
	}
	if deckFinal.CardContributorIds[bobCard2.Id] != bobUser.Id {
		t.Fatalf("Expected Bob (%s) to be contributor for card %s, but got %s", bobUser.Id, bobCard2.Id, deckFinal.CardContributorIds[bobCard2.Id])
	}
}
