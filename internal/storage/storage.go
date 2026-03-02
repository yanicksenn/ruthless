package storage

import (
	"context"
	"errors"

	"github.com/yanicksenn/ruthless/internal/domain"
)

var (
	ErrNotFound = errors.New("not found")
)

type Storage interface {
	// Card operations
	CreateCard(ctx context.Context, card domain.Card) error
	GetCard(ctx context.Context, id string) (domain.Card, error)
	ListCards(ctx context.Context) ([]domain.Card, error)

	// Session operations
	CreateSession(ctx context.Context, session domain.Session) error
	GetSession(ctx context.Context, id string) (domain.Session, error)
	UpdateSession(ctx context.Context, session domain.Session) error
}
