package domain_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yanicksenn/ruthless/backend/internal/domain"
)

func TestGame_NewGame(t *testing.T) {
	sessionID := "session-123"
	game := domain.NewGame(sessionID)

	assert.NotEmpty(t, game.Id)
	assert.Equal(t, sessionID, game.SessionId)
}

func TestPlayer_NewPlayer(t *testing.T) {
	player := domain.NewPlayer("Alice")

	assert.NotEmpty(t, player.Id)
	assert.Equal(t, "Alice", player.Name)
}
