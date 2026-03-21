package server

import (
	"context"
	"strings"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	pb "github.com/yanicksenn/ruthless/api/v1"
	"github.com/yanicksenn/ruthless/backend/internal/domain"
	"github.com/yanicksenn/ruthless/backend/internal/storage"
)

func (s *Server) ListDecks(ctx context.Context, req *pb.ListDecksRequest) (*pb.ListDecksResponse, error) {
	player, ok := getPlayer(ctx)
	if !ok {
		return nil, status.Errorf(codes.Unauthenticated, "unauthorized")
	}
	decks, err := s.store.ListDecks(ctx, player.Id)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to list decks")
	}

	for _, deck := range decks {
		s.populatePlayers(ctx, deck)
	}

	return &pb.ListDecksResponse{Decks: decks}, nil
}

func (s *Server) GetDeck(ctx context.Context, req *pb.GetDeckRequest) (*pb.Deck, error) {
	deck, err := s.store.GetDeck(ctx, req.Id)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "deck not found")
	}

	s.populatePlayers(ctx, deck)
	return deck, nil
}

func (s *Server) populatePlayers(ctx context.Context, deck *pb.Deck) {
	// Populate owner
	owner, err := s.store.GetUser(ctx, deck.OwnerId)
	if err == nil {
		deck.OwnerPlayer = &pb.Player{
			Id:         owner.Id,
			Name:       owner.Name,
			Identifier: owner.Identifier,
		}
	}

	// Populate contributors
	for _, cID := range deck.Contributors {
		user, err := s.store.GetUser(ctx, cID)
		if err == nil {
			deck.ContributorPlayers = append(deck.ContributorPlayers, &pb.Player{
				Id:         user.Id,
				Name:       user.Name,
				Identifier: user.Identifier,
			})
		}
	}
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

	s.LogUsageEvent(EventDeckCreated, player.Id, map[string]interface{}{
		"deck_id": deck.Id,
		"name":    deck.Name,
	})

	s.populatePlayers(ctx, deck)
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

	identifier := strings.TrimSpace(req.Identifier)
	name, ident, err := parseIdentifier(identifier)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid identifier format: %v", err)
	}

	user, err := s.store.GetUserByNameAndIdentifier(ctx, name, ident)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "user not found")
	}

	if err := domain.AddContributorToDeck(deck, player.Id, user.Id); err != nil {
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

	identifier := strings.TrimSpace(req.Identifier)
	name, ident, err := parseIdentifier(identifier)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid identifier format: %v", err)
	}

	user, err := s.store.GetUserByNameAndIdentifier(ctx, name, ident)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "user not found")
	}

	if err := domain.RemoveContributorFromDeck(deck, player.Id, user.Id); err != nil {
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

func (s *Server) SubscribeToDeck(ctx context.Context, req *pb.SubscribeToDeckRequest) (*pb.SubscribeToDeckResponse, error) {
	player, ok := getPlayer(ctx)
	if !ok {
		return nil, status.Errorf(codes.Unauthenticated, "unauthorized")
	}

	deck, err := s.store.GetDeck(ctx, req.DeckId)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "deck not found")
	}

	if deck.OwnerId == player.Id {
		return nil, status.Errorf(codes.InvalidArgument, "cannot subscribe to your own deck")
	}

	if err := s.store.SubscribeToDeck(ctx, req.DeckId, player.Id); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to subscribe to deck")
	}

	return &pb.SubscribeToDeckResponse{}, nil
}

func (s *Server) UnsubscribeFromDeck(ctx context.Context, req *pb.UnsubscribeFromDeckRequest) (*pb.UnsubscribeFromDeckResponse, error) {
	player, ok := getPlayer(ctx)
	if !ok {
		return nil, status.Errorf(codes.Unauthenticated, "unauthorized")
	}

	if err := s.store.UnsubscribeFromDeck(ctx, req.DeckId, player.Id); err != nil {
		if err == storage.ErrNotFound {
			return nil, status.Errorf(codes.NotFound, "subscription not found")
		}
		return nil, status.Errorf(codes.Internal, "failed to unsubscribe from deck")
	}

	return &pb.UnsubscribeFromDeckResponse{}, nil
}
