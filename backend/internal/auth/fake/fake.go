package fake

import (
	"context"

	pb "github.com/yanicksenn/ruthless/api/v1"
	"github.com/yanicksenn/ruthless/backend/internal/auth"
)

type Authenticator struct {
	// For simplicity in fake auth, any non-empty string is a valid player name which becomes the "token".
}

func New() *Authenticator {
	return &Authenticator{}
}

func (a *Authenticator) Authenticate(ctx context.Context, token string) (*pb.Player, error) {
	if token == "" {
		return nil, auth.ErrUnauthorized
	}

	// Issue a deterministic or random player. For simplicity, we just use the token as the player name and salt it as ID.
	return &pb.Player{
		Id:   "fake-" + token,
		Name: token,
	}, nil
}

// Ensure Fake Authenticator implements auth.Authenticator at compile time.
var _ auth.Authenticator = (*Authenticator)(nil)
