package server

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	pb "github.com/yanicksenn/ruthless/api/v1"
	"github.com/yanicksenn/ruthless/backend/internal/domain"
)

func (s *Server) ListCards(ctx context.Context, req *pb.ListCardsRequest) (*pb.ListCardsResponse, error) {
	cards, err := s.store.ListCards(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to list cards")
	}

	return &pb.ListCardsResponse{Cards: cards}, nil
}

func (s *Server) CreateCard(ctx context.Context, req *pb.CreateCardRequest) (*pb.Card, error) {
	card, err := domain.NewCard(req.Text)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, err.Error())
	}

	if err := s.store.CreateCard(ctx, card); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to save card")
	}

	return card, nil
}
