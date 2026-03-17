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
	
	// Add initial decks
	for _, deckID := range req.DeckIds {
		// Verify deck exists
		if _, err := s.store.GetDeck(ctx, deckID); err != nil {
			return nil, status.Errorf(codes.NotFound, "deck %s not found", deckID)
		}
		domain.AddDeckToSession(session, deckID)
	}

	if err := s.store.CreateSession(ctx, session); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create session")
	}

	// Automatically create an associated game
	if _, err := s.createGameInternal(ctx, session.Id); err != nil {
		return nil, err
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

	// Check if player is already in session OR was in the game (re-join)
	game, err := s.store.GetGameBySession(ctx, session.Id)
	alreadyInGame := false
	if err == nil {
		for _, pid := range game.PlayerIds {
			if pid == player.Id {
				alreadyInGame = true
				break
			}
		}
	}

	domain.AddPlayerToSession(session, player.Id)

	if err := s.store.UpdateSession(ctx, session); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to update session")
	}

	// Check game state - can only join if WAITING, unless already in game (re-join)
	// We allow re-join even if ABANDONED/FINISHED to be helpful.
	if err == nil && !alreadyInGame && game.State != pb.GameState_GAME_STATE_WAITING {
		domain.RemovePlayerFromSession(session, player.Id)
		_ = s.store.UpdateSession(ctx, session)
		return nil, status.Errorf(codes.FailedPrecondition, "cannot join session: game already started")
	}

	// Sync session changes to game
	if err := s.syncSessionToGame(ctx, session); err != nil {
		return nil, err
	}

	return session, nil
}


func (s *Server) LeaveSession(ctx context.Context, req *pb.LeaveSessionRequest) (*pb.LeaveSessionResponse, error) {
	session, err := s.store.GetSession(ctx, req.SessionId)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "session not found")
	}

	player, ok := getPlayer(ctx)
	if !ok {
		return nil, status.Errorf(codes.Unauthenticated, "unauthorized")
	}

	domain.RemovePlayerFromSession(session, player.Id)

	if err := s.store.UpdateSession(ctx, session); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to update session")
	}

	game, err := s.store.GetGameBySession(ctx, session.Id)
	if err == nil {
		domain.HandlePlayerLeave(game, player.Id)
		if err := s.store.UpdateGame(ctx, game); err != nil {
			return nil, status.Errorf(codes.Internal, "failed to update game state: %v", err)
		}
	}

	// Sync session changes to game
	if err := s.syncSessionToGame(ctx, session); err != nil {
		return nil, err
	}

	return &pb.LeaveSessionResponse{}, nil
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

	// Sync session changes to game
	if err := s.syncSessionToGame(ctx, session); err != nil {
		return nil, err
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

func (s *Server) syncSessionToGame(ctx context.Context, session *pb.Session) error {
	game, err := s.store.GetGameBySession(ctx, session.Id)
	if err != nil {
		// If game doesn't exist, we don't sync. In the new model it should exist,
		// but let's be robust for existing sessions during transition.
		return nil
	}

	if err := s.syncGamePlayers(ctx, game, session); err != nil {
		return err
	}
	if err := s.syncGameDecks(ctx, game, session); err != nil {
		return err
	}

	if err := s.store.UpdateGame(ctx, game); err != nil {
		return status.Errorf(codes.Internal, "failed to sync game state: %v", err)
	}

	return nil
}

