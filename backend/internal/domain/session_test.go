package domain_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yanicksenn/ruthless/backend/internal/domain"
)

func TestSession_AddPlayer(t *testing.T) {
	session := domain.NewSession("owner-1")
	player := domain.NewPlayer("Alice")

	domain.AddPlayerToSession(session, player.Id)

	assert.Len(t, session.PlayerIds, 2)
	assert.Contains(t, session.PlayerIds, "owner-1")
	assert.Contains(t, session.PlayerIds, player.Id)
}

func TestSession_RemovePlayer(t *testing.T) {
	session := domain.NewSession("owner-1")
	player := domain.NewPlayer("Alice")
	domain.AddPlayerToSession(session, player.Id)
	assert.Len(t, session.PlayerIds, 2)

	domain.RemovePlayerFromSession(session, player.Id)

	assert.Len(t, session.PlayerIds, 1)
	assert.Equal(t, "owner-1", session.PlayerIds[0])
}

func TestSession_AddDeck(t *testing.T) {
	session := domain.NewSession("owner-1")
	deck := domain.NewDeck("My Deck", "owner-1")

	domain.AddDeckToSession(session, deck.Id)

	assert.Len(t, session.DeckIds, 1)
	assert.Equal(t, deck.Id, session.DeckIds[0])
}
