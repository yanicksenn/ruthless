package memory_test

import (
	"context"
	"testing"

	"github.com/yanicksenn/ruthless/internal/domain"
	"github.com/yanicksenn/ruthless/internal/storage"
	"github.com/yanicksenn/ruthless/internal/storage/memory"
)

func TestMemoryStorage_Cards(t *testing.T) {
	store := memory.New()
	ctx := context.Background()

	card, err := domain.NewCard("A fast ___")
	if err != nil {
		t.Fatalf("Failed to create card: %v", err)
	}

	err = store.CreateCard(ctx, card)
	if err != nil {
		t.Fatalf("Failed to store card: %v", err)
	}

	retrieved, err := store.GetCard(ctx, card.ID)
	if err != nil {
		t.Fatalf("Failed to get card: %v", err)
	}

	if retrieved.ID != card.ID || retrieved.Text != card.Text {
		t.Errorf("Retrieved card does not match original")
	}

	list, err := store.ListCards(ctx)
	if err != nil {
		t.Fatalf("Failed to list cards: %v", err)
	}

	if len(list) != 1 {
		t.Errorf("Expected 1 card, got %d", len(list))
	}

	_, err = store.GetCard(ctx, "nonexistent")
	if err != storage.ErrNotFound {
		t.Errorf("Expected ErrNotFound, got %v", err)
	}
}

func TestMemoryStorage_Sessions(t *testing.T) {
	store := memory.New()
	ctx := context.Background()

	sess := domain.NewSession()

	err := store.CreateSession(ctx, sess)
	if err != nil {
		t.Fatalf("Failed to store session: %v", err)
	}

	retrieved, err := store.GetSession(ctx, sess.ID)
	if err != nil {
		t.Fatalf("Failed to get session: %v", err)
	}

	if retrieved.ID != sess.ID {
		t.Errorf("Retrieved session does not match original")
	}

	// Update session
	sess.Players = append(sess.Players, domain.NewPlayer("Alice"))
	err = store.UpdateSession(ctx, sess)
	if err != nil {
		t.Fatalf("Failed to update session: %v", err)
	}

	retrieved, err = store.GetSession(ctx, sess.ID)
	if err != nil {
		t.Fatalf("Failed to get updated session: %v", err)
	}

	if len(retrieved.Players) != 1 {
		t.Errorf("Expected 1 player, got %d", len(retrieved.Players))
	}
}
