package domain_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yanicksenn/ruthless/backend/internal/domain"
)

func TestDeck_CanModify(t *testing.T) {
	deck := domain.NewDeck("My Deck", "owner-1")
	deck.Contributors = append(deck.Contributors, "contributor-1")

	assert.True(t, domain.CanModifyDeck(deck, "owner-1"))
	assert.True(t, domain.CanModifyDeck(deck, "contributor-1"))
	assert.False(t, domain.CanModifyDeck(deck, "other-user"))
}

func TestDeck_AddRemoveContributor(t *testing.T) {
	deck := domain.NewDeck("My Deck", "owner-1")

	// Unauthorized addition
	err := domain.AddContributorToDeck(deck, "other-user", "contributor-1")
	assert.ErrorIs(t, err, domain.ErrUnauthorized)

	// Authorized addition
	err = domain.AddContributorToDeck(deck, "owner-1", "contributor-1")
	assert.NoError(t, err)
	assert.Contains(t, deck.Contributors, "contributor-1")

	// Authorized removal
	err = domain.RemoveContributorFromDeck(deck, "owner-1", "contributor-1")
	assert.NoError(t, err)
	assert.NotContains(t, deck.Contributors, "contributor-1")
}

func TestDeck_AddRemoveCard(t *testing.T) {
	deck := domain.NewDeck("My Deck", "owner-1")
	card, err := domain.NewCard("A ___ card.")
	require.NoError(t, err)

	// Unauthorized addition
	err = domain.AddCardToDeck(deck, "other-user", card)
	assert.ErrorIs(t, err, domain.ErrUnauthorized)

	// Authorized addition
	err = domain.AddCardToDeck(deck, "owner-1", card)
	assert.NoError(t, err)
	assert.Len(t, deck.Cards, 1)

	// Unauthorized removal
	err = domain.RemoveCardFromDeck(deck, "other-user", card.Id)
	assert.ErrorIs(t, err, domain.ErrUnauthorized)

	// Authorized removal
	err = domain.RemoveCardFromDeck(deck, "owner-1", card.Id)
	assert.NoError(t, err)
	assert.Len(t, deck.Cards, 0)
}
