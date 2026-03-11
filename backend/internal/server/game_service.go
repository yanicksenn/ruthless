package server

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	pb "github.com/yanicksenn/ruthless/api/v1"
	"github.com/yanicksenn/ruthless/backend/internal/domain"
)

func (s *Server) CreateGame(ctx context.Context, req *pb.CreateGameRequest) (*pb.Game, error) {
	game := domain.NewGame(req.SessionId)
	if err := s.store.CreateGame(ctx, game); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create game")
	}
	return game, nil
}

func (s *Server) GetGame(ctx context.Context, req *pb.GetGameRequest) (*pb.Game, error) {
	game, err := s.store.GetGame(ctx, req.Id)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "game not found")
	}
	return domain.StripHidden(game), nil
}

func (s *Server) StartGame(ctx context.Context, req *pb.StartGameRequest) (*pb.Game, error) {
	game, err := s.store.GetGame(ctx, req.Id)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "game not found")
	}

	session, err := s.store.GetSession(ctx, game.SessionId)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to fetch session")
	}

	if err := domain.StartGame(game, session); err != nil {
		return nil, status.Errorf(codes.FailedPrecondition, "failed to start game: %v", err)
	}

	if err := s.store.UpdateGame(ctx, game); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to save game state")
	}

	return domain.StripHidden(game), nil
}

func (s *Server) PlayCards(ctx context.Context, req *pb.PlayCardsRequest) (*pb.PlayCardsResponse, error) {
	player, ok := getPlayer(ctx)
	if !ok {
		return nil, status.Errorf(codes.Unauthenticated, "player not in context")
	}

	game, err := s.store.GetGame(ctx, req.GameId)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "game not found")
	}

	play, err := domain.PlayCards(game, player.Id, req.CardIds)
	if err != nil {
		return nil, status.Errorf(codes.FailedPrecondition, "failed to play cards: %v", err)
	}

	if err := s.store.UpdateGame(ctx, game); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to save game state")
	}

	return &pb.PlayCardsResponse{PlayId: play.Id}, nil
}

func (s *Server) SelectWinner(ctx context.Context, req *pb.SelectWinnerRequest) (*pb.SelectWinnerResponse, error) {
	player, ok := getPlayer(ctx)
	if !ok {
		return nil, status.Errorf(codes.Unauthenticated, "player not in context")
	}

	game, err := s.store.GetGame(ctx, req.GameId)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "game not found")
	}

	session, err := s.store.GetSession(ctx, game.SessionId)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to fetch session")
	}

	if err := domain.SelectWinner(game, session, player.Id, req.PlayId); err != nil {
		return nil, status.Errorf(codes.FailedPrecondition, "failed to select winner: %v", err)
	}

	if err := s.store.UpdateGame(ctx, game); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to save game state")
	}

	return &pb.SelectWinnerResponse{}, nil
}

func (s *Server) GetHand(ctx context.Context, req *pb.GetHandRequest) (*pb.GetHandResponse, error) {
	player, ok := getPlayer(ctx)
	if !ok {
		return nil, status.Errorf(codes.Unauthenticated, "player not in context")
	}

	game, err := s.store.GetGame(ctx, req.GameId)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "game not found")
	}

	hand, ok := game.HiddenHands[player.Id]
	if !ok {
		return nil, status.Errorf(codes.NotFound, "player has no hand in this game")
	}

	return &pb.GetHandResponse{Cards: hand.Cards}, nil
}

func (s *Server) GetGameBySession(ctx context.Context, req *pb.GetGameBySessionRequest) (*pb.Game, error) {
	game, err := s.store.GetGameBySession(ctx, req.SessionId)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "game not found for session")
	}
	return domain.StripHidden(game), nil
}
