package memory_test

import (
	"context"
	"testing"

	"github.com/yanicksenn/ruthless/backend/internal/domain"
	"github.com/yanicksenn/ruthless/backend/internal/storage"
	"github.com/yanicksenn/ruthless/backend/internal/storage/memory"
)

func TestMemoryStorage_Cards(t *testing.T) {
	store := memory.New()
	ctx := context.Background()

	card, err := domain.NewCard("A fast ___", "owner-1")
	if err != nil {
		t.Fatalf("Failed to create card: %v", err)
	}

	err = store.CreateCard(ctx, card)
	if err != nil {
		t.Fatalf("Failed to store card: %v", err)
	}

	retrieved, err := store.GetCard(ctx, card.Id)
	if err != nil {
		t.Fatalf("Failed to get card: %v", err)
	}

	if retrieved.Id != card.Id || retrieved.Text != card.Text {
		t.Errorf("Retrieved card does not match original")
	}

	list, total, err := store.ListCards(ctx, 0, 0, nil)
	if err != nil {
		t.Fatalf("Failed to list cards: %v", err)
	}

	if len(list) != 1 || total != 1 {
		t.Errorf("Expected 1 card and total 1, got %d cards and total %d", len(list), total)
	}

	// Pagination test
	for i := 0; i < 5; i++ {
		c, _ := domain.NewCard("Card", "owner-1")
		_ = store.CreateCard(ctx, c)
	}

	list, total, err = store.ListCards(ctx, 2, 1, nil) // First page
	if err != nil || len(list) != 2 || total != 6 {
		t.Errorf("Page 1 failed: got %d cards, total %d, err %v", len(list), total, err)
	}

	list, total, err = store.ListCards(ctx, 2, 2, nil) // Second page
	if err != nil || len(list) != 2 || total != 6 {
		t.Errorf("Page 2 failed: got %d cards, total %d, err %v", len(list), total, err)
	}

	list, total, err = store.ListCards(ctx, 2, 4, nil) // Out of bounds
	if err != nil || len(list) != 0 || total != 6 {
		t.Errorf("Out of bounds failed: got %d cards, total %d, err %v", len(list), total, err)
	}

	_, err = store.GetCard(ctx, "nonexistent")
	if err != storage.ErrNotFound {
		t.Errorf("Expected ErrNotFound, got %v", err)
	}
}

func TestMemoryStorage_Sessions(t *testing.T) {
	store := memory.New()
	ctx := context.Background()

	sess := domain.NewSession("owner-1")

	err := store.CreateSession(ctx, sess)
	if err != nil {
		t.Fatalf("Failed to store session: %v", err)
	}

	retrieved, err := store.GetSession(ctx, sess.Id)
	if err != nil {
		t.Fatalf("Failed to get session: %v", err)
	}

	if retrieved.Id != sess.Id {
		t.Errorf("Retrieved session does not match original")
	}

	// Update session
	sess.PlayerIds = append(sess.PlayerIds, "player-1")
	err = store.UpdateSession(ctx, sess)
	if err != nil {
		t.Fatalf("Failed to update session: %v", err)
	}

	retrieved, err = store.GetSession(ctx, sess.Id)
	if err != nil {
		t.Fatalf("Failed to get updated session: %v", err)
	}

	if len(retrieved.PlayerIds) != 1 {
		t.Errorf("Expected 1 player, got %d", len(retrieved.PlayerIds))
	}
}
