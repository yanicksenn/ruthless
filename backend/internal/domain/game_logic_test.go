package domain_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	pb "github.com/yanicksenn/ruthless/api/v1"
	"github.com/yanicksenn/ruthless/backend/internal/domain"
)

func TestPlayCards_Validation(t *testing.T) {
	// Setup a game in PLAYING state with a black card having 2 blanks
	game := &pb.Game{
		Id:    "game-1",
		State: pb.GameState_GAME_STATE_PLAYING,
		Rounds: []*pb.Round{
			{
				Id:     "round-1",
				CzarId: "Czar",
				BlackCard: &pb.Card{
					Id:     "black-1",
					Text:   "___ and ___",
					Blanks: 2,
				},
			},
		},
		HiddenHands: map[string]*pb.PlayerHand{
			"Player1": {
				Cards: []*pb.Card{
					{Id: "white-1", Text: "W1"},
					{Id: "white-2", Text: "W2"},
					{Id: "white-3", Text: "W3"},
				},
			},
		},
		Scores: map[string]uint32{
			"Czar":    0,
			"Player1": 0,
			"Player2": 0,
		},
	}

	// 1. Try to play 1 card (should fail)
	_, err := domain.PlayCards(game, "Player1", []string{"white-1"})
	assert.Equal(t, domain.ErrInvalidNumberOfCards, err)

	// 2. Try to play 3 cards (should fail)
	_, err = domain.PlayCards(game, "Player1", []string{"white-1", "white-2", "white-3"})
	assert.Equal(t, domain.ErrInvalidNumberOfCards, err)

	// 3. Try to play 2 cards (should succeed)
	play, err := domain.PlayCards(game, "Player1", []string{"white-1", "white-2"})
	require.NoError(t, err)
	assert.NotNil(t, play)
	assert.Len(t, play.Cards, 2)
	assert.Equal(t, "white-1", play.Cards[0].Id)
	assert.Equal(t, "white-2", play.Cards[1].Id)
}
