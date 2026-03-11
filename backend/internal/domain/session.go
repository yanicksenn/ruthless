package domain

import (
	"github.com/google/uuid"
	pb "github.com/yanicksenn/ruthless/api/v1"
)

func NewSession() *pb.Session {
	return &pb.Session{
		Id:      uuid.New().String(),
		Players: []*pb.Player{},
		Decks:   []*pb.Deck{},
	}
}

func AddPlayerToSession(s *pb.Session, player *pb.Player) {
	s.Players = append(s.Players, player)
}

func AddDeckToSession(s *pb.Session, deck *pb.Deck) {
	s.Decks = append(s.Decks, deck)
}
