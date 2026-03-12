package jwt

import (
	"context"
	"fmt"

	"github.com/golang-jwt/jwt/v5"
	pb "github.com/yanicksenn/ruthless/api/v1"
	"github.com/yanicksenn/ruthless/backend/internal/auth"
)

type Authenticator struct {
	secret []byte
}

func New(secret string) *Authenticator {
	return &Authenticator{
		secret: []byte(secret),
	}
}

func (a *Authenticator) Authenticate(ctx context.Context, tokenString string) (*pb.Player, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return a.secret, nil
	})

	if err != nil || !token.Valid {
		return nil, auth.ErrUnauthorized
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, auth.ErrUnauthorized
	}

	// In a real production system (e.g. Firebase), these would be "sub" and "name"
	id, _ := claims["sub"].(string)
	name, _ := claims["name"].(string)

	if id == "" {
		return nil, auth.ErrUnauthorized
	}

	return &pb.Player{
		Id:   id,
		Name: name,
	}, nil
}

// Ensure JWT Authenticator implements auth.Authenticator at compile time.
var _ auth.Authenticator = (*Authenticator)(nil)
