package storage

import (
	"context"
	"errors"

	pb "github.com/yanicksenn/ruthless/api/v1"
)

var (
	ErrNotFound = errors.New("not found")
)

type Storage interface {
	// Card operations
	CreateCard(ctx context.Context, card *pb.Card) error
	GetCard(ctx context.Context, id string) (*pb.Card, error)
	ListCards(ctx context.Context) ([]*pb.Card, error)

	// User operations
	CreateUser(ctx context.Context, user *pb.User) error
	GetUser(ctx context.Context, id string) (*pb.User, error)

	// Session operations
	CreateSession(ctx context.Context, session *pb.Session) error
	GetSession(ctx context.Context, id string) (*pb.Session, error)
	UpdateSession(ctx context.Context, session *pb.Session) error
	ListSessions(ctx context.Context) ([]*pb.Session, error)

	// Deck operations
	CreateDeck(ctx context.Context, deck *pb.Deck) error
	GetDeck(ctx context.Context, id string) (*pb.Deck, error)
	UpdateDeck(ctx context.Context, deck *pb.Deck) error
	ListDecks(ctx context.Context) ([]*pb.Deck, error)

	// Game operations
	CreateGame(ctx context.Context, game *pb.Game) error
	GetGame(ctx context.Context, id string) (*pb.Game, error)
	UpdateGame(ctx context.Context, game *pb.Game) error
	GetGameBySession(ctx context.Context, sessionID string) (*pb.Game, error)
}
