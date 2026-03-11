package server

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	pb "github.com/yanicksenn/ruthless/api/v1"
	"github.com/yanicksenn/ruthless/backend/internal/storage"
)

type SeedData struct {
	Users    map[string]*pb.User    `json:"users"`
	Cards    map[string]*pb.Card    `json:"cards"`
	Decks    map[string]*pb.Deck    `json:"decks"`
	Sessions map[string]*pb.Session `json:"sessions"`
	Games    map[string]*pb.Game    `json:"games"`
}

func LoadSeed(ctx context.Context, store storage.Storage, filePath string) error {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read seed file: %w", err)
	}

	var seed SeedData
	if err := json.Unmarshal(data, &seed); err != nil {
		return fmt.Errorf("failed to parse seed JSON: %w", err)
	}

	for _, u := range seed.Users {
		if err := store.CreateUser(ctx, u); err != nil {
			return fmt.Errorf("failed to seed user %s: %w", u.Id, err)
		}
	}

	for _, c := range seed.Cards {
		if c.Blanks == 0 && strings.Contains(c.Text, "___") {
			c.Blanks = uint32(strings.Count(c.Text, "___"))
		}
		if err := store.CreateCard(ctx, c); err != nil {
			return fmt.Errorf("failed to seed card %s: %w", c.Id, err)
		}
	}

	for _, d := range seed.Decks {
		if err := store.CreateDeck(ctx, d); err != nil {
			return fmt.Errorf("failed to seed deck %s: %w", d.Id, err)
		}
	}

	for _, s := range seed.Sessions {
		if err := store.CreateSession(ctx, s); err != nil {
			return fmt.Errorf("failed to seed session %s: %w", s.Id, err)
		}
	}

	for _, g := range seed.Games {
		if err := store.CreateGame(ctx, g); err != nil {
			return fmt.Errorf("failed to seed game %s: %w", g.Id, err)
		}
	}

	return nil
}
