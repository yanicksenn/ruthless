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

	card := &pb.Card{Id: "card123", Text: "A ___ card.", Color: pb.CardColor_CARD_COLOR_BLACK}
	err = store.CreateCard(context.Background(), card)
	require.NoError(t, err)

	ctx := context.WithValue(context.Background(), server.PlayerContextKey, player)
	req := &pb.AddCardToDeckRequest{
		DeckId: deck.Id,
		CardId: card.Id,
	}

	resp, err := srv.AddCardToDeck(ctx, req)
	require.NoError(t, err)
	assert.NotNil(t, resp)

	fetchDeck, err := store.GetDeck(context.Background(), deck.Id)
	require.NoError(t, err)
	assert.Len(t, fetchDeck.CardIds, 1)
	assert.Equal(t, card.Id, fetchDeck.CardIds[0])
}

func TestServer_JoinSession(t *testing.T) {
	store := memory.New()
	auth := &mockAuth{} // auth not used for JoinSession directly here
	srv := server.New(store, auth)

	session := domain.NewSession("owner-1")
	err := store.CreateSession(context.Background(), session)
	require.NoError(t, err)

	player := domain.NewPlayer("Alice")
	err = store.CreateUser(context.Background(), &pb.User{Id: player.Id, Name: player.Name})
	require.NoError(t, err)

	ctx := context.WithValue(context.Background(), server.PlayerContextKey, player)

	req := &pb.JoinSessionRequest{
		SessionId:  session.Id,
		PlayerName: "Alice",
	}

	resp, err := srv.JoinSession(ctx, req)
	require.NoError(t, err)
	assert.Equal(t, session.Id, resp.Id)
	assert.Len(t, resp.PlayerIds, 1)

	// Verify player was created in storage
	playerID := resp.PlayerIds[0]
	retrievedUser, err := store.GetUser(context.Background(), playerID)
	require.NoError(t, err)
	assert.Equal(t, "Alice", retrievedUser.Name)
}
