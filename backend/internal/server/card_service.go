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
	player, ok := getPlayer(ctx)
	if !ok {
		return nil, status.Errorf(codes.Unauthenticated, "unauthorized")
	}
	fetchOwnerID := player.Id
	if len(req.Ids) > 0 {
		fetchOwnerID = ""
	}

	if len(req.IncludeDeckIds) > 0 {
		for _, deckID := range req.IncludeDeckIds {
			deck, err := s.store.GetDeck(ctx, deckID)
			if err != nil {
				if err == storage.ErrNotFound {
					return nil, status.Errorf(codes.NotFound, "deck %s not found", deckID)
				}
				return nil, status.Errorf(codes.Internal, "failed to get deck %s", deckID)
			}
			if !domain.CanViewDeck(deck, player.Id) {
				return nil, status.Errorf(codes.PermissionDenied, "unauthorized to view cards in deck %s", deckID)
			}
		}
		// If authorized to view the decks, allow listing all cards in them
		fetchOwnerID = ""
	}

	cards, totalCount, err := s.store.ListCards(ctx, fetchOwnerID, req.PageSize, req.PageNumber, req.Ids, req.Filter, req.OrderBy, req.IncludeDeckIds, req.Color, req.ExcludeDeckIds)
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

	if s.config != nil && s.config.Public != nil && s.config.Public.Limits != nil {
		// Limit text length
		if s.config.Public.Limits.MaxCardTextLength > 0 && uint32(len(req.Text)) > s.config.Public.Limits.MaxCardTextLength {
			return nil, status.Errorf(codes.InvalidArgument, "card text exceeds limit of %d characters", s.config.Public.Limits.MaxCardTextLength)
		}

		// Limit total cards per user
		if s.config.Public.Limits.MaxCardsPerUser > 0 {
			count, err := s.store.CountCardsByOwner(ctx, ownerID)
			if err != nil {
				return nil, status.Errorf(codes.Internal, "failed to check card count: %v", err)
			}
			if uint32(count) >= s.config.Public.Limits.MaxCardsPerUser {
				return nil, status.Errorf(codes.ResourceExhausted, "maximum number of cards (%d) reached", s.config.Public.Limits.MaxCardsPerUser)
			}
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

func (s *Server) UpdateCard(ctx context.Context, req *pb.UpdateCardRequest) (*pb.Card, error) {
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
		return nil, status.Errorf(codes.PermissionDenied, "unauthorized to update this card")
	}

	if s.config != nil && s.config.Public != nil && s.config.Public.Limits != nil {
		if s.config.Public.Limits.MaxCardTextLength > 0 && uint32(len(req.Text)) > s.config.Public.Limits.MaxCardTextLength {
			return nil, status.Errorf(codes.InvalidArgument, "card text exceeds limit of %d characters", s.config.Public.Limits.MaxCardTextLength)
		}
	}

	card.Text = req.Text
	if err := s.store.UpdateCard(ctx, card); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to update card: %v", err)
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
