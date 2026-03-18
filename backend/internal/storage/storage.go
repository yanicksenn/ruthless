package storage

import (
	"context"
	"errors"
	"time"

	pb "github.com/yanicksenn/ruthless/api/v1"
)

var (
	ErrNotFound      = errors.New("not found")
	ErrAlreadyExists = errors.New("already exists")
)

type Storage interface {
	// Card operations
	CreateCard(ctx context.Context, card *pb.Card) error
	GetCard(ctx context.Context, id string) (*pb.Card, error)
	ListCards(ctx context.Context, pageSize, pageNumber int32, ids []string, filter string, orderBy *pb.CardOrder) ([]*pb.Card, int32, error)
	DeleteCard(ctx context.Context, id string) error

	// User operations
	CreateUser(ctx context.Context, user *pb.User) error
	UpdateUser(ctx context.Context, user *pb.User) error
	GetUser(ctx context.Context, id string) (*pb.User, error)

	// Auth Token operations
	RevokeToken(ctx context.Context, token string, expiresAt time.Time) error
	IsTokenRevoked(ctx context.Context, token string) (bool, error)

	// Session operations
	CreateSession(ctx context.Context, session *pb.Session) error
	GetSession(ctx context.Context, id string) (*pb.Session, error)
	UpdateSession(ctx context.Context, session *pb.Session) error
	ListSessions(ctx context.Context, playerID string) ([]*pb.Session, error)

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

	// Admin/Limit operations
	CountCardsByOwner(ctx context.Context, ownerID string) (int32, error)
	CountDecksByOwner(ctx context.Context, ownerID string) (int32, error)
	CountSessionsByOwner(ctx context.Context, ownerID string) (int32, error)
}
