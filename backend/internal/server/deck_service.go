package server

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	pb "github.com/yanicksenn/ruthless/api/v1"
	"github.com/yanicksenn/ruthless/backend/internal/domain"
)

func (s *Server) ListDecks(ctx context.Context, req *pb.ListDecksRequest) (*pb.ListDecksResponse, error) {
	decks, err := s.store.ListDecks(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to list decks")
	}

	return &pb.ListDecksResponse{Decks: decks}, nil
}

func (s *Server) GetDeck(ctx context.Context, req *pb.GetDeckRequest) (*pb.Deck, error) {
	deck, err := s.store.GetDeck(ctx, req.Id)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "deck not found")
	}

	return deck, nil
}

func (s *Server) CreateDeck(ctx context.Context, req *pb.CreateDeckRequest) (*pb.Deck, error) {
	player, ok := getPlayer(ctx)
	if !ok {
		return nil, status.Errorf(codes.Unauthenticated, "unauthorized")
	}

	deck := domain.NewDeck(req.Name, player.Id)
	if err := s.store.CreateDeck(ctx, deck); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create deck")
	}

	return deck, nil
}

func (s *Server) AddContributor(ctx context.Context, req *pb.AddContributorRequest) (*pb.AddContributorResponse, error) {
	player, ok := getPlayer(ctx)
	if !ok {
		return nil, status.Errorf(codes.Unauthenticated, "unauthorized")
	}

	deck, err := s.store.GetDeck(ctx, req.DeckId)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "deck not found")
	}

	if err := domain.AddContributorToDeck(deck, player.Id, req.ContributorId); err != nil {
		return nil, status.Errorf(codes.PermissionDenied, err.Error())
	}

	if err := s.store.UpdateDeck(ctx, deck); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to update deck")
	}

	return &pb.AddContributorResponse{}, nil
}

func (s *Server) RemoveContributor(ctx context.Context, req *pb.RemoveContributorRequest) (*pb.RemoveContributorResponse, error) {
	player, ok := getPlayer(ctx)
	if !ok {
		return nil, status.Errorf(codes.Unauthenticated, "unauthorized")
	}

	deck, err := s.store.GetDeck(ctx, req.DeckId)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "deck not found")
	}

	if err := domain.RemoveContributorFromDeck(deck, player.Id, req.ContributorId); err != nil {
		return nil, status.Errorf(codes.PermissionDenied, err.Error())
	}

	if err := s.store.UpdateDeck(ctx, deck); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to update deck")
	}

	return &pb.RemoveContributorResponse{}, nil
}

func (s *Server) AddCardToDeck(ctx context.Context, req *pb.AddCardToDeckRequest) (*pb.AddCardToDeckResponse, error) {
	player, ok := getPlayer(ctx)
	if !ok {
		return nil, status.Errorf(codes.Unauthenticated, "unauthorized")
	}

	deck, err := s.store.GetDeck(ctx, req.DeckId)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "deck not found")
	}

	// Validate card exists and is owned by the player
	card, err := s.store.GetCard(ctx, req.CardId)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "card not found")
	}

	if card.OwnerId != player.Id {
		return nil, status.Errorf(codes.PermissionDenied, "you do not own this card")
	}

	if err := domain.AddCardToDeck(deck, player.Id, req.CardId); err != nil {
		return nil, status.Errorf(codes.PermissionDenied, err.Error())
	}

	if err := s.store.UpdateDeck(ctx, deck); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to update deck")
	}

	return &pb.AddCardToDeckResponse{}, nil
}

func (s *Server) RemoveCardFromDeck(ctx context.Context, req *pb.RemoveCardFromDeckRequest) (*pb.RemoveCardFromDeckResponse, error) {
	player, ok := getPlayer(ctx)
	if !ok {
		return nil, status.Errorf(codes.Unauthenticated, "unauthorized")
	}

	deck, err := s.store.GetDeck(ctx, req.DeckId)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "deck not found")
	}

	if err := domain.RemoveCardFromDeck(deck, player.Id, req.CardId); err != nil {
		return nil, status.Errorf(codes.PermissionDenied, err.Error())
	}

	if err := s.store.UpdateDeck(ctx, deck); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to update deck")
	}

	return &pb.RemoveCardFromDeckResponse{}, nil
}
