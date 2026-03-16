package domain

import (
	"github.com/google/uuid"
	pb "github.com/yanicksenn/ruthless/api/v1"
)

func NewSession(ownerID string) *pb.Session {
	return &pb.Session{
		Id:        uuid.New().String(),
		PlayerIds: []string{ownerID},
		DeckIds:   []string{},
		OwnerId:   ownerID,
	}
}

func AddPlayerToSession(s *pb.Session, playerID string) {
	for _, pid := range s.PlayerIds {
		if pid == playerID {
			return
		}
	}
	s.PlayerIds = append(s.PlayerIds, playerID)
}

func RemovePlayerFromSession(s *pb.Session, playerID string) {
	newPlayerIds := make([]string, 0)
	for _, pid := range s.PlayerIds {
		if pid != playerID {
			newPlayerIds = append(newPlayerIds, pid)
		}
	}
	s.PlayerIds = newPlayerIds
}

func AddDeckToSession(s *pb.Session, deckID string) {
	s.DeckIds = append(s.DeckIds, deckID)
}

func CanModifySession(s *pb.Session, userID string) bool {
	return s.OwnerId == userID
}
