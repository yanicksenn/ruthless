package domain

import (
	"github.com/google/uuid"
	pb "github.com/yanicksenn/ruthless/api/v1"
)

func NewPlayer(name string) *pb.Player {
	return &pb.Player{
		Id:   uuid.New().String(),
		Name: name,
	}
}

func NewGame(sessionID string) *pb.Game {
	return &pb.Game{
		Id:        uuid.New().String(),
		SessionId: sessionID,
		State:     pb.GameState_GAME_STATE_WAITING,
	}
}
