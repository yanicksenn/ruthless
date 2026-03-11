package domain_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yanicksenn/ruthless/backend/internal/domain"
)

func TestSession_AddPlayer(t *testing.T) {
	session := domain.NewSession()
	player := domain.NewPlayer("Alice")

	domain.AddPlayerToSession(session, player)

	assert.Len(t, session.Players, 1)
	assert.Equal(t, "Alice", session.Players[0].Name)
}

func TestSession_AddDeck(t *testing.T) {
	session := domain.NewSession()
	deck := domain.NewDeck("My Deck", "owner-1")

	domain.AddDeckToSession(session, deck)

	assert.Len(t, session.Decks, 1)
	assert.Equal(t, "My Deck", session.Decks[0].Name)
}
