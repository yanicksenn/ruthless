package server

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	pb "github.com/yanicksenn/ruthless/api/v1"
	"github.com/yanicksenn/ruthless/backend/internal/domain"
	"github.com/yanicksenn/ruthless/backend/internal/storage"
	"google.golang.org/protobuf/types/known/emptypb"
)

func (s *Server) ListCards(ctx context.Context, req *pb.ListCardsRequest) (*pb.ListCardsResponse, error) {
	cards, totalCount, err := s.store.ListCards(ctx, req.PageSize, req.PageNumber)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to list cards")
	}

	return &pb.ListCardsResponse{Cards: cards, TotalCount: totalCount}, nil
}

func (s *Server) CreateCard(ctx context.Context, req *pb.CreateCardRequest) (*pb.Card, error) {
	player, ok := getPlayer(ctx)
	if !ok {
		return nil, status.Errorf(codes.Unauthenticated, "unauthorized")
	}
	ownerID := player.Id

	if s.config != nil && s.config.Limits != nil && s.config.Limits.CardTextLimit > 0 {
		if uint32(len(req.Text)) > s.config.Limits.CardTextLimit {
			return nil, status.Errorf(codes.InvalidArgument, "card text exceeds limit of %d characters", s.config.Limits.CardTextLimit)
		}
	}

	card, err := domain.NewCard(req.Text, ownerID)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, err.Error())
	}

	if err := s.store.CreateCard(ctx, card); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to save card: %v", err)
	}

	return card, nil
}

func (s *Server) DeleteCard(ctx context.Context, req *pb.DeleteCardRequest) (*emptypb.Empty, error) {
	player, ok := getPlayer(ctx)
	if !ok {
		return nil, status.Errorf(codes.Unauthenticated, "unauthorized")
	}

	card, err := s.store.GetCard(ctx, req.Id)
	if err != nil {
		if err == storage.ErrNotFound {
			return nil, status.Errorf(codes.NotFound, "card not found")
		}
		return nil, status.Errorf(codes.Internal, "failed to get card")
	}

	if card.OwnerId != player.Id {
		return nil, status.Errorf(codes.PermissionDenied, "unauthorized to delete this card")
	}

	if err := s.store.DeleteCard(ctx, req.Id); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to delete card")
	}

	return &emptypb.Empty{}, nil
}
