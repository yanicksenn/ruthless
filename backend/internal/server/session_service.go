package server

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	pb "github.com/yanicksenn/ruthless/api/v1"
	"github.com/yanicksenn/ruthless/backend/internal/domain"
)

func (s *Server) CreateSession(ctx context.Context, req *pb.CreateSessionRequest) (*pb.Session, error) {
	player, ok := getPlayer(ctx)
	if !ok {
		return nil, status.Errorf(codes.Unauthenticated, "unauthorized")
	}

	session := domain.NewSession(player.Id)
	if err := s.store.CreateSession(ctx, session); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create session")
	}
	return session, nil
}

func (s *Server) JoinSession(ctx context.Context, req *pb.JoinSessionRequest) (*pb.Session, error) {
	session, err := s.store.GetSession(ctx, req.SessionId)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "session not found")
	}

	player, ok := getPlayer(ctx)
	if !ok {
		return nil, status.Errorf(codes.Unauthenticated, "unauthorized")
	}

	domain.AddPlayerToSession(session, player.Id)

	if err := s.store.UpdateSession(ctx, session); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to update session")
	}

	return session, nil
}

func (s *Server) GetSession(ctx context.Context, req *pb.GetSessionRequest) (*pb.Session, error) {
	session, err := s.store.GetSession(ctx, req.Id)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "session not found")
	}
	return session, nil
}

func (s *Server) AddDeckToSession(ctx context.Context, req *pb.AddDeckToSessionRequest) (*pb.AddDeckToSessionResponse, error) {
	session, err := s.store.GetSession(ctx, req.SessionId)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "session not found")
	}

	if _, err := s.store.GetDeck(ctx, req.DeckId); err != nil {
		return nil, status.Errorf(codes.NotFound, "deck not found")
	}

	player, ok := getPlayer(ctx)
	if !ok {
		return nil, status.Errorf(codes.Unauthenticated, "unauthorized")
	}

	if !domain.CanModifySession(session, player.Id) {
		return nil, status.Errorf(codes.PermissionDenied, "only the owner can add decks to this session")
	}

	domain.AddDeckToSession(session, req.DeckId)

	if err := s.store.UpdateSession(ctx, session); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to update session")
	}

	return &pb.AddDeckToSessionResponse{}, nil
}

func (s *Server) ListSessions(ctx context.Context, req *pb.ListSessionsRequest) (*pb.ListSessionsResponse, error) {
	sessions, err := s.store.ListSessions(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to list sessions")
	}

	return &pb.ListSessionsResponse{Sessions: sessions}, nil
}
