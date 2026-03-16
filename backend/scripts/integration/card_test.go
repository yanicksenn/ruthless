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
}
