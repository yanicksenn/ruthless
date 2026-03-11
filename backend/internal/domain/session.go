package domain

import (
	"github.com/google/uuid"
	pb "github.com/yanicksenn/ruthless/api/v1"
)

func NewSession() *pb.Session {
	return &pb.Session{
		Id:        uuid.New().String(),
		PlayerIds: []string{},
		DeckIds:   []string{},
	}
}

func AddPlayerToSession(s *pb.Session, playerID string) {
	s.PlayerIds = append(s.PlayerIds, playerID)
}

func AddDeckToSession(s *pb.Session, deckID string) {
	s.DeckIds = append(s.DeckIds, deckID)
}
