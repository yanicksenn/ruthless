package auth

import (
	"context"
	"errors"

	pb "github.com/yanicksenn/ruthless/api/v1"
)

var (
	ErrUnauthorized = errors.New("unauthorized request")
)

type Authenticator interface {
	// Authenticate validates a request (e.g., via token string) and returns the authenticated Player
	Authenticate(ctx context.Context, token string) (*pb.Player, error)
}
