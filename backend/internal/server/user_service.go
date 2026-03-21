package server

import (
	"context"
	"crypto/rand"
	"encoding/binary"
	"fmt"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	pb "github.com/yanicksenn/ruthless/api/v1"
	"github.com/yanicksenn/ruthless/backend/internal/storage"
)

func (s *Server) CompleteRegistration(ctx context.Context, req *pb.CompleteRegistrationRequest) (*pb.User, error) {
	player, ok := getPlayer(ctx)
	if !ok {
		return nil, status.Errorf(codes.Unauthenticated, "unauthorized")
	}

	if req.Name == "" {
		return nil, status.Errorf(codes.InvalidArgument, "name is required")
	}

	if s.config != nil && s.config.Public != nil && s.config.Public.Limits != nil {
		if s.config.Public.Limits.MinUserNameLength > 0 && uint32(len(req.Name)) < s.config.Public.Limits.MinUserNameLength {
			return nil, status.Errorf(codes.InvalidArgument, "name is too short (minimum %d characters)", s.config.Public.Limits.MinUserNameLength)
		}
		if s.config.Public.Limits.MaxUserNameLength > 0 && uint32(len(req.Name)) > s.config.Public.Limits.MaxUserNameLength {
			return nil, status.Errorf(codes.InvalidArgument, "name is too long (maximum %d characters)", s.config.Public.Limits.MaxUserNameLength)
		}
	}

	maxRetries := int(s.config.GetPublic().GetRegistration().GetMaxUniqueIdentifierRecreations())
	if maxRetries == 0 {
		maxRetries = 10
	}

	var user *pb.User
	var err error

	for i := 0; i <= maxRetries; i++ {
		identifier, genErr := generateIdentifier()
		if genErr != nil {
			return nil, status.Errorf(codes.Internal, "failed to generate identifier")
		}

		user = &pb.User{
			Id:         player.Id,
			Name:       req.Name,
			Identifier: identifier,
		}

		err = s.store.UpdateUser(ctx, user)
		if err == nil {
			// Success!
			s.LogUsageEvent(EventAccountCreated, player.Id, map[string]interface{}{
				"name": req.Name,
			})
			// Return a copy to avoid in-place modification of the object in the store (e.g. for memory store)
			return &pb.User{
				Id:                user.Id,
				Name:              user.Name,
				Identifier:        user.Identifier,
				CreatedAt:         user.CreatedAt,
				PendingCompletion: false,
			}, nil
		}

		if err != storage.ErrAlreadyExists {
			return nil, status.Errorf(codes.Internal, "failed to complete registration: %v", err)
		}
	}

	return nil, status.Errorf(codes.AlreadyExists, "failed to find a unique name identifier after %d retries", maxRetries)
}

func (s *Server) GetMe(ctx context.Context, req *pb.GetMeRequest) (*pb.User, error) {
	player, ok := getPlayer(ctx)
	if !ok {
		return nil, status.Errorf(codes.Unauthenticated, "unauthorized")
	}

	user, err := s.store.GetUser(ctx, player.Id)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "user not found")
	}

	if user.PendingCompletion {
		return user, nil
	}

	// Return a copy with concatenated name (to avoid in-place modification for memory store)
	return &pb.User{
		Id:                user.Id,
		Name:              user.Name,
		Identifier:        user.Identifier,
		CreatedAt:         user.CreatedAt,
		PendingCompletion: false,
	}, nil
}

func generateIdentifier() (string, error) {
	var n uint32
	err := binary.Read(rand.Reader, binary.BigEndian, &n)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%08d", n%100000000), nil
}
