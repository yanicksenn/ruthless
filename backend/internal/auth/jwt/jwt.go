package jwt

import (
	"context"
	"fmt"

	"github.com/coreos/go-oidc/v3/oidc"
	pb "github.com/yanicksenn/ruthless/api/v1"
	"github.com/yanicksenn/ruthless/backend/internal/auth"
)

type Authenticator struct {
	verifier *oidc.IDTokenVerifier
}

func NewGoogle(ctx context.Context, audience string) (*Authenticator, error) {
	provider, err := oidc.NewProvider(ctx, "https://accounts.google.com")
	if err != nil {
		return nil, fmt.Errorf("failed to get provider: %v", err)
	}

	config := &oidc.Config{
		ClientID: audience,
	}
	verifier := provider.Verifier(config)

	return &Authenticator{
		verifier: verifier,
	}, nil
}

func (a *Authenticator) Authenticate(ctx context.Context, tokenString string) (*pb.Player, error) {
	idToken, err := a.verifier.Verify(ctx, tokenString)
	if err != nil {
		return nil, auth.ErrUnauthorized
	}

	var claims struct {
		Subject string `json:"sub"`
		Name    string `json:"name"`
		Email   string `json:"email"`
	}
	if err := idToken.Claims(&claims); err != nil {
		return nil, auth.ErrUnauthorized
	}

	if claims.Subject == "" {
		return nil, auth.ErrUnauthorized
	}

	// For Google, we use the display name or email if name is missing
	name := claims.Name
	if name == "" {
		name = claims.Email
	}

	return &pb.Player{
		Id:   claims.Subject,
		Name: name,
	}, nil
}

// Ensure JWT Authenticator implements auth.Authenticator at compile time.
var _ auth.Authenticator = (*Authenticator)(nil)
