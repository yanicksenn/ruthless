package main

import (
	"context"
	"testing"

	pb "github.com/yanicksenn/ruthless/api/v1"
	"github.com/yanicksenn/ruthless/backend/scripts/testutil"
)

func runVisibilityTests(t *testing.T, ctx context.Context, c *testutil.TestClient, runID string) {
	aliceName := "VisibilityAlice_" + runID
	bobName := "VisibilityBob_" + runID

	aliceCtx := c.GetAuthContext(ctx, aliceName)
	bobCtx := c.GetAuthContext(ctx, bobName)

	if c.Store != nil {
		c.Store.CreateUser(ctx, &pb.User{Id: aliceName, Name: aliceName})
		c.Store.CreateUser(ctx, &pb.User{Id: bobName, Name: bobName})
	}

	// Complete registration for both
	_, _ = c.UserClient.CompleteRegistration(aliceCtx, &pb.CompleteRegistrationRequest{Name: aliceName})
	bobUser, _ := c.UserClient.CompleteRegistration(bobCtx, &pb.CompleteRegistrationRequest{Name: bobName})

	t.Log("\n--- Visibility Isolation Suite ---")

	// 1. Alice creates a card and a deck
	t.Log("  [RUN] Alice creates card and deck...")
	aliceCard, err := c.CardClient.CreateCard(aliceCtx, &pb.CreateCardRequest{Text: "Alice's Secret Card"})
	testutil.AssertSuccess(t, err, "Alice CreateCard")

	aliceDeck, err := c.DeckClient.CreateDeck(aliceCtx, &pb.CreateDeckRequest{Name: "Alice's Secret Deck"})
	testutil.AssertSuccess(t, err, "Alice CreateDeck")

	_, err = c.DeckClient.AddCardToDeck(aliceCtx, &pb.AddCardToDeckRequest{DeckId: aliceDeck.Id, CardId: aliceCard.Id})
	testutil.AssertSuccess(t, err, "Alice AddCardToDeck")

	// 2. Bob lists cards and decks - should be empty
	t.Log("  [RUN] Bob lists cards...")
	bobCards, err := c.CardClient.ListCards(bobCtx, &pb.ListCardsRequest{})
	testutil.AssertSuccess(t, err, "Bob ListCards")
	for _, card := range bobCards.Cards {
		if card.Id == aliceCard.Id {
			t.Errorf("Bob should not see Alice's card %s", card.Id)
		}
	}

	t.Log("  [RUN] Bob lists decks...")
	bobDecks, err := c.DeckClient.ListDecks(bobCtx, &pb.ListDecksRequest{})
	testutil.AssertSuccess(t, err, "Bob ListDecks")
	for _, deck := range bobDecks.Decks {
		if deck.Id == aliceDeck.Id {
			t.Errorf("Bob should not see Alice's deck %s", deck.Id)
		}
	}

	// 3. Alice lists cards and decks - should see them
	t.Log("  [RUN] Alice lists cards...")
	aliceCards, err := c.CardClient.ListCards(aliceCtx, &pb.ListCardsRequest{})
	testutil.AssertSuccess(t, err, "Alice ListCards")
	foundCard := false
	for _, card := range aliceCards.Cards {
		if card.Id == aliceCard.Id {
			foundCard = true
			break
		}
	}
	if !foundCard {
		t.Errorf("Alice should see her own card %s", aliceCard.Id)
	}

	t.Log("  [RUN] Alice lists decks...")
	aliceDecks, err := c.DeckClient.ListDecks(aliceCtx, &pb.ListDecksRequest{})
	testutil.AssertSuccess(t, err, "Alice ListDecks")
	foundDeck := false
	for _, deck := range aliceDecks.Decks {
		if deck.Id == aliceDeck.Id {
			foundDeck = true
			break
		}
	}
	if !foundDeck {
		t.Errorf("Alice should see her own deck %s", aliceDeck.Id)
	}

	// 4. Alice adds Bob as contributor to her deck
	t.Log("  [RUN] Alice adds Bob as contributor...")
	_, err = c.DeckClient.AddContributor(aliceCtx, &pb.AddContributorRequest{DeckId: aliceDeck.Id, Identifier: bobUser.Identifier})
	testutil.AssertSuccess(t, err, "Alice AddContributor Bob")

	// 5. Bob lists decks - should now see Alice's deck
	t.Log("  [RUN] Bob lists decks again...")
	bobDecks, err = c.DeckClient.ListDecks(bobCtx, &pb.ListDecksRequest{})
	testutil.AssertSuccess(t, err, "Bob ListDecks again")
	foundDeck = false
	for _, deck := range bobDecks.Decks {
		if deck.Id == aliceDeck.Id {
			foundDeck = true
			break
		}
	}
	if !foundDeck {
		t.Errorf("Bob should now see Alice's deck %s as a contributor", aliceDeck.Id)
	}

	// 6. Bob fetches Alice's card explicitly by ID - should succeed since cards are public by ID
	t.Log("  [RUN] Bob fetches Alice's card by ID...")
	bobCardsById, err := c.CardClient.ListCards(bobCtx, &pb.ListCardsRequest{Ids: []string{aliceCard.Id}})
	testutil.AssertSuccess(t, err, "Bob ListCards by ID")
	foundCardById := false
	for _, card := range bobCardsById.Cards {
		if card.Id == aliceCard.Id {
			foundCardById = true
			break
		}
	}
	if !foundCardById {
		t.Errorf("Bob should be able to fetch Alice's card %s explicitly by ID", aliceCard.Id)
	}

	// 7. Charlie subscribes to Alice's deck - should be able to list cards in it
	charlieName := "VisibilityCharlie_" + runID
	charlieCtx := c.GetAuthContext(ctx, charlieName)
	if c.Store != nil {
		c.Store.CreateUser(ctx, &pb.User{Id: charlieName, Name: charlieName})
	}
	_, _ = c.UserClient.CompleteRegistration(charlieCtx, &pb.CompleteRegistrationRequest{Name: charlieName})

	t.Log("  [RUN] Charlie subscribes to Alice's deck...")
	_, err = c.DeckClient.SubscribeToDeck(charlieCtx, &pb.SubscribeToDeckRequest{DeckId: aliceDeck.Id})
	testutil.AssertSuccess(t, err, "Charlie SubscribeToDeck")

	t.Log("  [RUN] Charlie lists cards in Alice's deck...")
	charlieCards, err := c.CardClient.ListCards(charlieCtx, &pb.ListCardsRequest{DeckId: aliceDeck.Id})
	testutil.AssertSuccess(t, err, "Charlie ListCards in Alice's deck")
	
	if charlieCards.TotalCount == 0 {
		t.Errorf("Charlie should see cards in Alice's deck, but got 0")
	}
	
	foundAliceCard := false
	for _, card := range charlieCards.Cards {
		if card.Id == aliceCard.Id {
			foundAliceCard = true
			break
		}
	}
	if !foundAliceCard {
		t.Errorf("Charlie should see Alice's card %s in her deck", aliceCard.Id)
	}
}
