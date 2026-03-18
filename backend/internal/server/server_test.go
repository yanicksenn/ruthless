package server_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	pb "github.com/yanicksenn/ruthless/api/v1"
	"github.com/yanicksenn/ruthless/backend/internal/domain"
	"github.com/yanicksenn/ruthless/backend/internal/server"
	"github.com/yanicksenn/ruthless/backend/internal/storage"
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
	srv := server.New(store, auth, &pb.Config{})

	ctx := context.WithValue(context.Background(), server.PlayerKey{}, player)
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
	srv := server.New(store, auth, &pb.Config{})

	deck := domain.NewDeck("My Deck", player.Id)
	err := store.CreateDeck(context.Background(), deck)
	require.NoError(t, err)

	card := &pb.Card{Id: "card123", Text: "A ___ card.", Color: pb.CardColor_CARD_COLOR_BLACK, OwnerId: player.Id}
	err = store.CreateCard(context.Background(), card)
	require.NoError(t, err)

	ctx := context.WithValue(context.Background(), server.PlayerKey{}, player)
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
	srv := server.New(store, auth, &pb.Config{})

	session := domain.NewSession("owner-1")
	err := store.CreateUser(context.Background(), &pb.User{Id: "owner-1", Name: "Owner"})
	require.NoError(t, err)
	err = store.CreateSession(context.Background(), session)
	require.NoError(t, err)

	player := domain.NewPlayer("Alice")
	err = store.CreateUser(context.Background(), &pb.User{Id: player.Id, Name: player.Name})
	require.NoError(t, err)

	ctx := context.WithValue(context.Background(), server.PlayerKey{}, player)

	req := &pb.JoinSessionRequest{
		SessionId:  session.Id,
		PlayerName: "Alice",
	}

	resp, err := srv.JoinSession(ctx, req)
	require.NoError(t, err)
	assert.Equal(t, session.Id, resp.Id)
	assert.Len(t, resp.PlayerIds, 2)

	// Verify player was created in storage (Alice is the second one now)
	playerID := resp.PlayerIds[1]
	retrievedUser, err := store.GetUser(context.Background(), playerID)
	require.NoError(t, err)
	assert.Equal(t, "Alice", retrievedUser.Name)
}

func TestServer_DeleteCard(t *testing.T) {
	store := memory.New()
	alice := domain.NewPlayer("Alice")
	bob := domain.NewPlayer("Bob")
	auth := &mockAuth{player: alice}
	srv := server.New(store, auth, &pb.Config{})

	card, _ := domain.NewCard("My card", alice.Id)
	store.CreateCard(context.Background(), card)

	// Alice deletes her own card
	ctxAlice := context.WithValue(context.Background(), server.PlayerKey{}, alice)
	_, err := srv.DeleteCard(ctxAlice, &pb.DeleteCardRequest{Id: card.Id})
	require.NoError(t, err)

	_, err = store.GetCard(context.Background(), card.Id)
	assert.Equal(t, storage.ErrNotFound, err)

	// Alice tries to delete a card owned by Bob
	cardBob, _ := domain.NewCard("Bob's card", bob.Id)
	store.CreateCard(context.Background(), cardBob)

	_, err = srv.DeleteCard(ctxAlice, &pb.DeleteCardRequest{Id: cardBob.Id})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unauthorized to delete this card")
}
func TestServer_CreateCard_Limit(t *testing.T) {
	store := memory.New()
	alice := domain.NewPlayer("Alice")
	auth := &mockAuth{player: alice}
	cfg := &pb.Config{
		Limits: &pb.Config_Limits{
			MaxCardTextLength: 10,
			MaxCardsPerUser:   2,
		},
	}
	srv := server.New(store, auth, cfg)

	ctx := context.WithValue(context.Background(), server.PlayerKey{}, alice)

	// Under limit
	_, err := srv.CreateCard(ctx, &pb.CreateCardRequest{Text: "Short"})
	assert.NoError(t, err)

	// Over text limit
	_, err = srv.CreateCard(ctx, &pb.CreateCardRequest{Text: "This is way too long"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "card text exceeds limit of 10 characters")

	// Create second card (success)
	_, err = srv.CreateCard(ctx, &pb.CreateCardRequest{Text: "Ok"})
	assert.NoError(t, err)

	// Over count limit
	_, err = srv.CreateCard(ctx, &pb.CreateCardRequest{Text: "Third"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "maximum number of cards (2) reached")
}

func TestServer_CreateDeck_Limit(t *testing.T) {
	store := memory.New()
	alice := domain.NewPlayer("Alice")
	auth := &mockAuth{player: alice}
	cfg := &pb.Config{
		Limits: &pb.Config_Limits{
			MaxDeckNameLength: 10,
			MaxDecksPerUser:   1,
		},
	}
	srv := server.New(store, auth, cfg)

	ctx := context.WithValue(context.Background(), server.PlayerKey{}, alice)

	// Under limit
	_, err := srv.CreateDeck(ctx, &pb.CreateDeckRequest{Name: "Deck 1"})
	assert.NoError(t, err)

	// Over count limit
	_, err = srv.CreateDeck(ctx, &pb.CreateDeckRequest{Name: "Deck 2"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "maximum number of decks (1) reached")

	// Over name limit
	// Reset store and use a new server with higher deck count limit but same name limit
	store2 := memory.New()
	srv2 := server.New(store2, auth, &pb.Config{
		Limits: &pb.Config_Limits{MaxDeckNameLength: 5, MaxDecksPerUser: 10},
	})
	_, err = srv2.CreateDeck(ctx, &pb.CreateDeckRequest{Name: "Too Long"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "deck name exceeds limit of 5 characters")
}

func TestServer_CompleteRegistration_Limit(t *testing.T) {
	store := memory.New()
	alice := &pb.Player{Id: "alice-1", Name: "Alice"}
	auth := &mockAuth{player: alice}
	cfg := &pb.Config{
		Limits: &pb.Config_Limits{
			MinUserNameLength: 3,
			MaxUserNameLength: 8,
		},
	}
	srv := server.New(store, auth, cfg)

	ctx := context.WithValue(context.Background(), server.PlayerKey{}, alice)

	// Pre-create user in store (pending completion)
	err := store.CreateUser(ctx, &pb.User{Id: alice.Id, Name: alice.Name})
	require.NoError(t, err)

	// Too short
	_, err = srv.CompleteRegistration(ctx, &pb.CompleteRegistrationRequest{Name: "Al"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "name is too short")

	// Too long
	_, err = srv.CompleteRegistration(ctx, &pb.CompleteRegistrationRequest{Name: "Alice Wonderland"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "name is too long")

	// OK
	_, err = srv.CompleteRegistration(ctx, &pb.CompleteRegistrationRequest{Name: "Alice"})
	assert.NoError(t, err)
}

func TestServer_CreateSession_Limit(t *testing.T) {
	store := memory.New()
	alice := &pb.Player{Id: "alice-1", Name: "Alice"}
	auth := &mockAuth{player: alice}
	cfg := &pb.Config{
		Limits: &pb.Config_Limits{
			MaxSessionNameLength: 10,
			MaxSessionsPerUser:   1,
			MaxCardsPerSession:   10,
		},
	}
	srv := server.New(store, auth, cfg)

	ctx := context.WithValue(context.Background(), server.PlayerKey{}, alice)
	require.NoError(t, store.CreateUser(ctx, &pb.User{Id: alice.Id, Name: alice.Name}))

	// Create a deck for testing
	deck := &pb.Deck{Id: "deck-1", Name: "My Deck", OwnerId: alice.Id, CardIds: []string{"c1", "c2", "c3", "c4", "c5", "c6"}}
	require.NoError(t, store.CreateDeck(ctx, deck))

	// Under limit
	_, err := srv.CreateSession(ctx, &pb.CreateSessionRequest{Name: "S1", DeckIds: []string{"deck-1"}})
	assert.NoError(t, err)

	// Over count limit
	_, err = srv.CreateSession(ctx, &pb.CreateSessionRequest{Name: "S2"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "maximum number of sessions (1) reached")

	// Over card limit
	// Reset store and use a new server with higher session count limit
	store2 := memory.New()
	require.NoError(t, store2.CreateUser(ctx, &pb.User{Id: alice.Id, Name: alice.Name}))
	require.NoError(t, store2.CreateDeck(ctx, &pb.Deck{Id: "large-deck", Name: "Large", OwnerId: alice.Id, CardIds: make([]string, 11)}))
	
	srv2 := server.New(store2, auth, &pb.Config{
		Limits: &pb.Config_Limits{MaxSessionsPerUser: 10, MaxCardsPerSession: 10},
	})
	_, err = srv2.CreateSession(ctx, &pb.CreateSessionRequest{Name: "Too Big", DeckIds: []string{"large-deck"}})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "total cards in session (11) exceeds limit of 10")
}
