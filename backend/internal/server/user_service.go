package server

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	pb "github.com/yanicksenn/ruthless/api/v1"
)

func (s *Server) Register(ctx context.Context, req *pb.RegisterRequest) (*pb.User, error) {
	player, ok := getPlayer(ctx)
	if !ok {
		return nil, status.Errorf(codes.Unauthenticated, "unauthorized")
	}

	user := &pb.User{
		Id:   player.Id,
		Name: player.Name,
	}

	if err := s.store.CreateUser(ctx, user); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create user")
	}

	return user, nil
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

	return user, nil
}
