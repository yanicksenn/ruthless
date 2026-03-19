package main

import (
	"context"
	"testing"
	"time"

	pb "github.com/yanicksenn/ruthless/api/v1"
	"github.com/yanicksenn/ruthless/backend/scripts/testutil"
)

func runCardTests(t *testing.T, ctx context.Context, c *testutil.TestClient, runID string) {
	aliceCtx := c.GetAuthContext(ctx, "Alice")

	t.Log("\n--- Card Filtering & Sorting Suite ---")

	// Pre-requisite: Create some cards with known content and timing
	t.Log("  [RUN] Creating test cards...")
	cardTexts := []string{
		"Alpha card",
		"Beta card",
		"Gamma card",
		"Apple card",
		"Banana card",
	}

	var cardIDs []string
	for _, text := range cardTexts {
		card, err := c.CardClient.CreateCard(aliceCtx, &pb.CreateCardRequest{Text: text})
		testutil.AssertSuccess(t, err, "CreateCard "+text)
		cardIDs = append(cardIDs, card.Id)
		// Small sleep to ensure different timestamps if needed, though they should be different naturally
		time.Sleep(10 * time.Millisecond)
	}

	// 1. Test Filtering
	t.Run("Filtering", func(t *testing.T) {
		t.Log("  [RUN] Testing text filtering...")
		resp, err := c.CardClient.ListCards(aliceCtx, &pb.ListCardsRequest{
			Filter: "Alpha",
		})
		testutil.AssertSuccess(t, err, "ListCards with Filter 'Alpha'")
		if len(resp.Cards) != 1 || resp.Cards[0].Text != "Alpha card" {
			t.Errorf("Expected 1 card 'Alpha card', got %d cards", len(resp.Cards))
		}

		resp, err = c.CardClient.ListCards(aliceCtx, &pb.ListCardsRequest{
			Filter: "card",
		})
		testutil.AssertSuccess(t, err, "ListCards with Filter 'card'")
		if len(resp.Cards) < 5 {
			t.Errorf("Expected at least 5 cards for filter 'card', got %d", len(resp.Cards))
		}
	})

	// 2. Test Sorting by Text
	t.Run("SortingByText", func(t *testing.T) {
		t.Log("  [RUN] Testing sorting by text ASC...")
		resp, err := c.CardClient.ListCards(aliceCtx, &pb.ListCardsRequest{
			Ids: cardIDs,
			OrderBy: &pb.CardOrder{
				Field:      pb.CardOrderField_CARD_ORDER_FIELD_TEXT,
				Descending: false,
			},
		})
		testutil.AssertSuccess(t, err, "ListCards sorting by Text ASC")
		if len(resp.Cards) != 5 {
			t.Fatalf("Expected 5 cards, got %d", len(resp.Cards))
		}
		// Expected order: Alpha, Apple, Banana, Beta, Gamma
		expected := []string{"Alpha card", "Apple card", "Banana card", "Beta card", "Gamma card"}
		for i, card := range resp.Cards {
			if card.Text != expected[i] {
				t.Errorf("At index %d: expected %q, got %q", i, expected[i], card.Text)
			}
		}

		t.Log("  [RUN] Testing sorting by text DESC...")
		resp, err = c.CardClient.ListCards(aliceCtx, &pb.ListCardsRequest{
			Ids: cardIDs,
			OrderBy: &pb.CardOrder{
				Field:      pb.CardOrderField_CARD_ORDER_FIELD_TEXT,
				Descending: true,
			},
		})
		testutil.AssertSuccess(t, err, "ListCards sorting by Text DESC")
		// Expected order: Gamma, Beta, Banana, Apple, Alpha
		expectedDesc := []string{"Gamma card", "Beta card", "Banana card", "Apple card", "Alpha card"}
		for i, card := range resp.Cards {
			if card.Text != expectedDesc[i] {
				t.Errorf("At index %d (DESC): expected %q, got %q", i, expectedDesc[i], card.Text)
			}
		}
	})

	// 3. Test Sorting by CreatedAt
	t.Run("SortingByDate", func(t *testing.T) {
		t.Log("  [RUN] Testing sorting by date ASC...")
		resp, err := c.CardClient.ListCards(aliceCtx, &pb.ListCardsRequest{
			Ids: cardIDs,
			OrderBy: &pb.CardOrder{
				Field:      pb.CardOrderField_CARD_ORDER_FIELD_CREATED_AT,
				Descending: false,
			},
		})
		testutil.AssertSuccess(t, err, "ListCards sorting by Date ASC")
		if len(resp.Cards) != 5 {
			t.Fatalf("Expected 5 cards, got %d", len(resp.Cards))
		}
		// Expected order: Alpha, Beta, Gamma, Apple, Banana (order of creation)
		for i, card := range resp.Cards {
			if card.Text != cardTexts[i] {
				t.Errorf("At index %d (Date ASC): expected %q, got %q", i, cardTexts[i], card.Text)
			}
		}

		t.Log("  [RUN] Testing sorting by date DESC...")
		resp, err = c.CardClient.ListCards(aliceCtx, &pb.ListCardsRequest{
			Ids: cardIDs,
			OrderBy: &pb.CardOrder{
				Field:      pb.CardOrderField_CARD_ORDER_FIELD_CREATED_AT,
				Descending: true,
			},
		})
		testutil.AssertSuccess(t, err, "ListCards sorting by Date DESC")
		// Expected order: Banana, Apple, Gamma, Beta, Alpha
		for i, card := range resp.Cards {
			expectedText := cardTexts[4-i]
			if card.Text != expectedText {
				t.Errorf("At index %d (Date DESC): expected %q, got %q", i, expectedText, card.Text)
			}
		}
	})
	// 4. Test Updating Cards
	t.Run("Updating", func(t *testing.T) {
		t.Log("  [RUN] Testing card update...")
		cardID := cardIDs[0]
		newText := "Updated Alpha card"
		updated, err := c.CardClient.UpdateCard(aliceCtx, &pb.UpdateCardRequest{
			Id:   cardID,
			Text: newText,
		})
		testutil.AssertSuccess(t, err, "UpdateCard")
		if updated.Text != newText {
			t.Errorf("Expected updated text %q, got %q", newText, updated.Text)
		}

		// Verify it's actually updated in the list
		resp, err := c.CardClient.ListCards(aliceCtx, &pb.ListCardsRequest{Ids: []string{cardID}})
		testutil.AssertSuccess(t, err, "ListCards after update")
		if len(resp.Cards) != 1 || resp.Cards[0].Text != newText {
			t.Errorf("ListCards returned wrong text after update: %q", resp.Cards[0].Text)
		}

		// 5. Test Update Authorization
		t.Log("  [RUN] Testing update authorization (Bob attempts to update Alice's card)...")
		bobCtx := c.GetAuthContext(ctx, "Bob")
		_, err = c.CardClient.UpdateCard(bobCtx, &pb.UpdateCardRequest{
			Id:   cardID,
			Text: "Bob's malicious update",
		})
		if err == nil {
			t.Error("Expected error when Bob tries to update Alice's card, but got none")
		} else {
			t.Logf("  [OK] Bob update failed as expected: %v", err)
		}
	})

	// 6. Test Deck Filtering
	t.Run("DeckFiltering", func(t *testing.T) {
		t.Log("  [RUN] Testing deck filtering...")

		// Alice creates a deck
		deck, err := c.DeckClient.CreateDeck(aliceCtx, &pb.CreateDeckRequest{Name: "Filter Deck"})
		testutil.AssertSuccess(t, err, "CreateDeck for filtering")

		// Alice adds 2 cards to the deck
		card1 := cardIDs[1] // "Beta card"
		card2 := cardIDs[2] // "Gamma card"
		_, err = c.DeckClient.AddCardToDeck(aliceCtx, &pb.AddCardToDeckRequest{DeckId: deck.Id, CardId: card1})
		testutil.AssertSuccess(t, err, "Add card1 to deck")
		_, err = c.DeckClient.AddCardToDeck(aliceCtx, &pb.AddCardToDeckRequest{DeckId: deck.Id, CardId: card2})
		testutil.AssertSuccess(t, err, "Add card2 to deck")

		// Alice lists cards with deck filter
		resp, err := c.CardClient.ListCards(aliceCtx, &pb.ListCardsRequest{
			DeckId: deck.Id,
		})
		testutil.AssertSuccess(t, err, "ListCards with DeckId")

		if len(resp.Cards) != 2 {
			t.Errorf("Expected 2 cards in deck, got %d", len(resp.Cards))
		}

		found1, found2 := false, false
		for _, card := range resp.Cards {
			if card.Id == card1 {
				found1 = true
			}
			if card.Id == card2 {
				found2 = true
			}
		}

		if !found1 || !found2 {
			t.Errorf("Did not find both cards in deck filter results")
		}

		// Test deck filter combined with text filter
		resp, err = c.CardClient.ListCards(aliceCtx, &pb.ListCardsRequest{
			DeckId: deck.Id,
			Filter: "Beta",
		})
		testutil.AssertSuccess(t, err, "ListCards with DeckId and Filter 'Beta'")
		if len(resp.Cards) != 1 || resp.Cards[0].Id != card1 {
			t.Errorf("Expected 1 card (Beta) in deck filter with text filter 'Beta', got %d", len(resp.Cards))
		}
	})

	// 7. Test Color Filtering
	t.Run("FilteringByColor", func(t *testing.T) {
		t.Log("  [RUN] Testing color filtering...")

		// Create a white card
		white, err := c.CardClient.CreateCard(aliceCtx, &pb.CreateCardRequest{Text: "A simple answer"})
		testutil.AssertSuccess(t, err, "Create white card")
		if white.Color != pb.CardColor_CARD_COLOR_WHITE {
			t.Errorf("Expected White card, got %v", white.Color)
		}

		// Create a black card
		black, err := c.CardClient.CreateCard(aliceCtx, &pb.CreateCardRequest{Text: "What is ___?"})
		testutil.AssertSuccess(t, err, "Create black card")
		if black.Color != pb.CardColor_CARD_COLOR_BLACK {
			t.Errorf("Expected Black card, got %v", black.Color)
		}

		// Filter by White
		resp, err := c.CardClient.ListCards(aliceCtx, &pb.ListCardsRequest{
			Color: pb.CardColor_CARD_COLOR_WHITE,
			Ids:   []string{white.Id, black.Id},
		})
		testutil.AssertSuccess(t, err, "ListCards White")
		if len(resp.Cards) != 1 || resp.Cards[0].Id != white.Id {
			t.Errorf("Expected only white card, got %d cards", len(resp.Cards))
		}

		// Filter by Black
		resp, err = c.CardClient.ListCards(aliceCtx, &pb.ListCardsRequest{
			Color: pb.CardColor_CARD_COLOR_BLACK,
			Ids:   []string{white.Id, black.Id},
		})
		testutil.AssertSuccess(t, err, "ListCards Black")
		if len(resp.Cards) != 1 || resp.Cards[0].Id != black.Id {
			t.Errorf("Expected only black card, got %d cards", len(resp.Cards))
		}

		// Update white card to black
		updated, err := c.CardClient.UpdateCard(aliceCtx, &pb.UpdateCardRequest{
			Id:   white.Id,
			Text: "Now I have ___",
		})
		testutil.AssertSuccess(t, err, "Update to black")
		if updated.Color != pb.CardColor_CARD_COLOR_BLACK {
			t.Errorf("Expected updated card to be Black, got %v", updated.Color)
		}

		// Verify filtering again
		resp, err = c.CardClient.ListCards(aliceCtx, &pb.ListCardsRequest{
			Color: pb.CardColor_CARD_COLOR_BLACK,
			Ids:   []string{white.Id, black.Id},
		})
		testutil.AssertSuccess(t, err, "ListCards Black after update")
		if len(resp.Cards) != 2 {
			t.Errorf("Expected 2 black cards after update, got %d", len(resp.Cards))
		}
	})
}
