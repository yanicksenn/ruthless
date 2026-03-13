package server

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	pb "github.com/yanicksenn/ruthless/api/v1"
	"github.com/yanicksenn/ruthless/backend/internal/storage"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type SeedData struct {
	Users    []*pb.User    `json:"users"`
	Cards    []*pb.Card    `json:"cards"`
	Decks    []*pb.Deck    `json:"decks"`
	Sessions []*pb.Session `json:"sessions"`
	Games    []*pb.Game    `json:"games"`
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
		if u.CreatedAt == nil {
			u.CreatedAt = timestamppb.Now()
		}
		if err := store.CreateUser(ctx, u); err != nil {
			return fmt.Errorf("failed to seed user %s: %w", u.Id, err)
		}
	}

	for _, c := range seed.Cards {
		if c.Color == pb.CardColor_CARD_COLOR_UNSPECIFIED {
			if strings.Contains(c.Text, "___") {
				c.Color = pb.CardColor_CARD_COLOR_BLACK
			} else {
				c.Color = pb.CardColor_CARD_COLOR_WHITE
			}
		}
		if c.CreatedAt == nil {
			c.CreatedAt = timestamppb.Now()
		}
		if err := store.CreateCard(ctx, c); err != nil {
			return fmt.Errorf("failed to seed card %s: %w", c.Id, err)
		}
	}

	for _, d := range seed.Decks {
		if d.CreatedAt == nil {
			d.CreatedAt = timestamppb.Now()
		}
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
