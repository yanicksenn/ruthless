package server_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	pb "github.com/yanicksenn/ruthless/api/v1"
	"github.com/yanicksenn/ruthless/backend/internal/domain"
	"github.com/yanicksenn/ruthless/backend/internal/server"
	"github.com/yanicksenn/ruthless/backend/internal/storage/memory"
)

type mockAuth struct {
	player *pb.Player
}

func (m *mockAuth) Authenticate(ctx context.Context, token string) (*pb.Player, error) {
	return m.player, nil
}

func TestServer_CreateDeck(t *testing.T) {
	store := memory.New()
	player := domain.NewPlayer("Alice")
	auth := &mockAuth{player: player}
	srv := server.New(store, auth)

	ctx := context.WithValue(context.Background(), server.PlayerContextKey, player)
	req := &pb.CreateDeckRequest{Name: "My New Deck"}

	resp, err := srv.CreateDeck(ctx, req)
	require.NoError(t, err)

	assert.NotEmpty(t, resp.Id)
	assert.Equal(t, "My New Deck", resp.Name)
	assert.Equal(t, player.Id, resp.OwnerId)
}

func TestServer_AddCardToDeck(t *testing.T) {
	store := memory.New()
	player := domain.NewPlayer("Alice")
	auth := &mockAuth{player: player}
	srv := server.New(store, auth)

	deck := domain.NewDeck("My Deck", player.Id)
	err := store.CreateDeck(context.Background(), deck)
	require.NoError(t, err)

	card := &pb.Card{Id: "card123", Text: "A ___ card."}
	ctx := context.WithValue(context.Background(), server.PlayerContextKey, player)
	req := &pb.AddCardToDeckRequest{
		DeckId: deck.Id,
		Card:   card,
	}

	resp, err := srv.AddCardToDeck(ctx, req)
	require.NoError(t, err)
	assert.NotNil(t, resp)

	fetchDeck, err := store.GetDeck(context.Background(), deck.Id)
	require.NoError(t, err)
	assert.Len(t, fetchDeck.Cards, 1)
	assert.Equal(t, "A ___ card.", fetchDeck.Cards[0].Text)
}
