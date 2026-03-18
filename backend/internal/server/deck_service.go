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

	if s.config != nil && s.config.Public != nil && s.config.Public.Limits != nil {
		// Limit deck name length
		if s.config.Public.Limits.MaxDeckNameLength > 0 && uint32(len(req.Name)) > s.config.Public.Limits.MaxDeckNameLength {
			return nil, status.Errorf(codes.InvalidArgument, "deck name exceeds limit of %d characters", s.config.Public.Limits.MaxDeckNameLength)
		}

		// Limit total decks per user
		if s.config.Public.Limits.MaxDecksPerUser > 0 {
			count, err := s.store.CountDecksByOwner(ctx, player.Id)
			if err != nil {
				return nil, status.Errorf(codes.Internal, "failed to check deck count: %v", err)
			}
			if uint32(count) >= s.config.Public.Limits.MaxDecksPerUser {
				return nil, status.Errorf(codes.ResourceExhausted, "maximum number of decks (%d) reached", s.config.Public.Limits.MaxDecksPerUser)
			}
		}
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

	if s.config != nil && s.config.Public != nil && s.config.Public.Limits != nil {
		if s.config.Public.Limits.MaxContributorsPerDeck > 0 && uint32(len(deck.Contributors)) >= s.config.Public.Limits.MaxContributorsPerDeck {
			return nil, status.Errorf(codes.ResourceExhausted, "maximum number of contributors (%d) reached", s.config.Public.Limits.MaxContributorsPerDeck)
		}
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

	// Validate card exists
	_, err = s.store.GetCard(ctx, req.CardId)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "card not found")
	}

	if s.config != nil && s.config.Public != nil && s.config.Public.Limits != nil {
		if s.config.Public.Limits.MaxCardsPerDeck > 0 && uint32(len(deck.CardIds)) >= s.config.Public.Limits.MaxCardsPerDeck {
			return nil, status.Errorf(codes.ResourceExhausted, "maximum number of cards in deck (%d) reached", s.config.Public.Limits.MaxCardsPerDeck)
		}
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
