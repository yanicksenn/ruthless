package auth

import (
	"context"
	"errors"

	"github.com/yanicksenn/ruthless/internal/domain"
)

var (
	ErrUnauthorized = errors.New("unauthorized request")
)

type Authenticator interface {
	// Authenticate validates a request (e.g., via token string) and returns the authenticated Player
	Authenticate(ctx context.Context, token string) (domain.Player, error)
}
