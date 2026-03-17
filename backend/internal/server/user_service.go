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

func (s *Server) Register(ctx context.Context, req *pb.RegisterRequest) (*pb.User, error) {
	player, ok := getPlayer(ctx)
	if !ok {
		return nil, status.Errorf(codes.Unauthenticated, "unauthorized")
	}

	name := req.Name
	if name == "" {
		name = player.Name
	}

	user := &pb.User{
		Id:   player.Id,
		Name: name,
		// Identifier remains empty -> pending_completion
	}

	if err := s.store.CreateUser(ctx, user); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create user")
	}

	user.PendingCompletion = true
	return user, nil
}

func (s *Server) CompleteRegistration(ctx context.Context, req *pb.CompleteRegistrationRequest) (*pb.User, error) {
	player, ok := getPlayer(ctx)
	if !ok {
		return nil, status.Errorf(codes.Unauthenticated, "unauthorized")
	}

	if req.Name == "" {
		return nil, status.Errorf(codes.InvalidArgument, "name is required")
	}

	maxRetries := int(s.config.GetRegistration().GetMaxUniqueIdentifierRecreations())
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

		err = s.store.CreateUser(ctx, user)
		if err == nil {
			// Success!
			user.Name = fmt.Sprintf("%s#%s", user.Name, user.Identifier)
			user.PendingCompletion = false
			return user, nil
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

	if user.Identifier == "" {
		user.PendingCompletion = true
	} else {
		user.Name = fmt.Sprintf("%s#%s", user.Name, user.Identifier)
		user.PendingCompletion = false
	}

	return user, nil
}

func generateIdentifier() (string, error) {
	var n uint32
	err := binary.Read(rand.Reader, binary.BigEndian, &n)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%08d", n%100000000), nil
}
