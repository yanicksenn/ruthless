package server

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	pb "github.com/yanicksenn/ruthless/api/v1"
	"github.com/yanicksenn/ruthless/backend/internal/domain"
)

func (s *Server) CreateSession(ctx context.Context, req *pb.CreateSessionRequest) (*pb.Session, error) {
	player, ok := getPlayer(ctx)
	if !ok {
		return nil, status.Errorf(codes.Unauthenticated, "unauthorized")
	}

	if s.config != nil && s.config.Public != nil && s.config.Public.Limits != nil {
		// Limit session name length
		name := req.Name
		if name == "" {
			// Fetch owner's name for default session name if not provided
			owner, err := s.store.GetUser(ctx, player.Id)
			if err != nil {
				return nil, status.Errorf(codes.Internal, "failed to fetch owner information")
			}
			name = owner.Name + "'s Session"
		}

		if s.config.Public.Limits.MaxSessionNameLength > 0 && uint32(len(name)) > s.config.Public.Limits.MaxSessionNameLength {
			return nil, status.Errorf(codes.InvalidArgument, "session name exceeds limit of %d characters", s.config.Public.Limits.MaxSessionNameLength)
		}

		// Limit total sessions per user
		if s.config.Public.Limits.MaxSessionsPerUser > 0 {
			count, err := s.store.CountSessionsByOwner(ctx, player.Id)
			if err != nil {
				return nil, status.Errorf(codes.Internal, "failed to check session count: %v", err)
			}
			if uint32(count) >= s.config.Public.Limits.MaxSessionsPerUser {
				return nil, status.Errorf(codes.ResourceExhausted, "maximum number of sessions (%d) reached", s.config.Public.Limits.MaxSessionsPerUser)
			}
		}

		// Limit decks in session
		if s.config.Public.Limits.MaxDecksPerSession > 0 && uint32(len(req.DeckIds)) > s.config.Public.Limits.MaxDecksPerSession {
			return nil, status.Errorf(codes.InvalidArgument, "maximum number of decks per session (%d) reached", s.config.Public.Limits.MaxDecksPerSession)
		}

		// Limit total cards in session
		if s.config.Public.Limits.MaxCardsPerSession > 0 {
			var totalCards uint32
			for _, deckID := range req.DeckIds {
				deck, err := s.store.GetDeck(ctx, deckID)
				if err != nil {
					return nil, status.Errorf(codes.NotFound, "deck %s not found", deckID)
				}
				totalCards += uint32(len(deck.CardIds))
			}
			if totalCards > s.config.Public.Limits.MaxCardsPerSession {
				return nil, status.Errorf(codes.InvalidArgument, "total cards in session (%d) exceeds limit of %d", totalCards, s.config.Public.Limits.MaxCardsPerSession)
			}
		}
	}

	session := domain.NewSession(player.Id)
	session.CreatedAt = timestamppb.Now()
	
	if req.Name != "" {
		session.Name = req.Name
	} else {
		// Re-fetch or use pre-fetched name if we had to check limit
		owner, _ := s.store.GetUser(ctx, player.Id)
		session.Name = owner.Name + "'s Session"
	}
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

	if s.config != nil && s.config.Public != nil && s.config.Public.Limits != nil {
		if s.config.Public.Limits.MaxPlayersPerSession > 0 && uint32(len(session.PlayerIds)) >= s.config.Public.Limits.MaxPlayersPerSession {
			// Check if player is already in (allow re-join)
			isParticipant := false
			for _, pid := range session.PlayerIds {
				if pid == player.Id {
					isParticipant = true
					break
				}
			}
			if !isParticipant {
				return nil, status.Errorf(codes.ResourceExhausted, "maximum number of players (%d) reached", s.config.Public.Limits.MaxPlayersPerSession)
			}
		}
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

	if s.config != nil && s.config.Public != nil && s.config.Public.Limits != nil {
		if s.config.Public.Limits.MaxDecksPerSession > 0 && uint32(len(session.DeckIds)) >= s.config.Public.Limits.MaxDecksPerSession {
			return nil, status.Errorf(codes.ResourceExhausted, "maximum number of decks per session (%d) reached", s.config.Public.Limits.MaxDecksPerSession)
		}
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
	player, ok := getPlayer(ctx)
	if !ok {
		return nil, status.Errorf(codes.Unauthenticated, "unauthorized")
	}

	sessions, err := s.store.ListSessions(ctx, player.Id)
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

