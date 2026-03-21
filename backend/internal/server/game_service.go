package server

import (
	"context"
	"errors"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	pb "github.com/yanicksenn/ruthless/api/v1"
	"github.com/yanicksenn/ruthless/backend/internal/domain"
)

func (s *Server) createGameInternal(ctx context.Context, sessionID string) (*pb.Game, error) {
	session, err := s.store.GetSession(ctx, sessionID)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "session not found")
	}

	minPlayers := 2
	if s.config != nil && s.config.Public != nil && s.config.Public.Game != nil && s.config.Public.Game.MinRequiredPlayers > 0 {
		minPlayers = int(s.config.Public.Game.MinRequiredPlayers)
	}

	game := domain.NewGame(sessionID, uint32(minPlayers))
	if err := s.syncGamePlayers(ctx, game, session); err != nil {
		return nil, err
	}

	if err := s.syncGameDecks(ctx, game, session); err != nil {
		return nil, err
	}

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
		return nil, status.Errorf(codes.NotFound, "session not found")
	}

	player, ok := getPlayer(ctx)
	if !ok {
		return nil, status.Errorf(codes.Unauthenticated, "unauthorized")
	}

	if !domain.CanModifySession(session, player.Id) {
		return nil, status.Errorf(codes.PermissionDenied, "only the session owner can start the game")
	}

	minPlayers := 2
	if s.config != nil && s.config.Public != nil && s.config.Public.Game != nil && s.config.Public.Game.MinRequiredPlayers > 0 {
		minPlayers = int(s.config.Public.Game.MinRequiredPlayers)
	}

	if err := s.syncGamePlayers(ctx, game, session); err != nil {
		return nil, err
	}

	if err := s.syncGameDecks(ctx, game, session); err != nil {
		return nil, err
	}

	if err := domain.StartGame(game, minPlayers); err != nil {
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
		code := codes.FailedPrecondition
		if err == domain.ErrCzarCannotPlay || errors.Is(err, domain.ErrCzarCannotPlay) {
			code = codes.PermissionDenied
		}
		return nil, status.Errorf(code, "failed to play cards: %v", err)
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

	if err := domain.SelectWinner(game, player.Id, req.PlayId); err != nil {
		code := codes.FailedPrecondition
		if err == domain.ErrNotCzar || errors.Is(err, domain.ErrNotCzar) {
			code = codes.PermissionDenied
		}
		return nil, status.Errorf(code, "failed to select winner: %v", err)
	}

	if err := s.store.UpdateGame(ctx, game); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to save game state")
	}

	s.LogUsageEvent(EventRoundCompleted, player.Id, map[string]interface{}{
		"game_id":     game.Id,
		"round_index": len(game.Rounds),
	})

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
func (s *Server) syncGameDecks(ctx context.Context, game *pb.Game, session *pb.Session) error {
	game.HiddenBlackDeck = nil
	game.HiddenWhiteDeck = nil

	cards, _, err := s.store.ListCards(ctx, "", 0, 0, nil, "", nil, nil, pb.CardColor_CARD_COLOR_UNSPECIFIED, nil)
	if err != nil {
		return status.Errorf(codes.Internal, "failed to fetch cards")
	}
	cardMap := make(map[string]*pb.Card)
	for _, c := range cards {
		cardMap[c.Id] = c
	}

	decks, err := s.store.ListDecks(ctx, "")
	if err != nil {
		return status.Errorf(codes.Internal, "failed to fetch decks")
	}
	deckMap := make(map[string]*pb.Deck)
	for _, d := range decks {
		deckMap[d.Id] = d
	}

	domain.ConsolidateDecks(game, session.DeckIds, deckMap, cardMap)
	return nil
}

func (s *Server) syncGamePlayers(ctx context.Context, game *pb.Game, session *pb.Session) error {
	game.PlayerIds = session.PlayerIds
	game.Players = nil // Clear existing players before sync

	for _, pid := range session.PlayerIds {
		user, err := s.store.GetUser(ctx, pid)
		if err != nil {
			return status.Errorf(codes.Internal, "failed to fetch player %s: %v", pid, err)
		}
		game.Players = append(game.Players, &pb.Player{
			Id:         user.Id,
			Name:       user.Name,
			Identifier: user.Identifier,
		})
	}
	return nil
}
