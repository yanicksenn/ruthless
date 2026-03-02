package domain

import "github.com/google/uuid"

type Player struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	// Additional auth/oauth references could go here or in a separate User model
}

func NewPlayer(name string) Player {
	return Player{
		ID:   uuid.New().String(),
		Name: name,
	}
}

type Session struct {
	ID      string            `json:"id"`
	Players []Player          `json:"players"`
	Cards   []Card            `json:"cards"`
	// This is a simplified state for now, we can expand on the Game rules later
}

func NewSession() Session {
	return Session{
		ID:      uuid.New().String(),
		Players: []Player{},
		Cards:   []Card{},
	}
}
