package fake

import (
	"context"

	"github.com/yanicksenn/ruthless/internal/auth"
	"github.com/yanicksenn/ruthless/internal/domain"
)

type Authenticator struct {
	// For simplicity in fake auth, any non-empty string is a valid player name which becomes the "token".
}

func New() *Authenticator {
	return &Authenticator{}
}

func (a *Authenticator) Authenticate(ctx context.Context, token string) (domain.Player, error) {
	if token == "" {
		return domain.Player{}, auth.ErrUnauthorized
	}

	// Issue a deterministic or random player. For simplicity, we just use the token as the player name and salt it as ID.
	return domain.Player{
		ID:   "fake-" + token,
		Name: token,
	}, nil
}

// Ensure Fake Authenticator implements auth.Authenticator at compile time.
var _ auth.Authenticator = (*Authenticator)(nil)
