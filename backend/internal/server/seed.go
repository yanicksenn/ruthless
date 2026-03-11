package server

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	pb "github.com/yanicksenn/ruthless/api/v1"
	"github.com/yanicksenn/ruthless/backend/internal/domain"
	"github.com/yanicksenn/ruthless/backend/internal/storage"
)

type SeedData struct {
	Cards    []SeedCard    `json:"cards"`
	Decks    []SeedDeck    `json:"decks"`
	Sessions []SeedSession `json:"sessions"`
}

type SeedCard struct {
	Text string `json:"text"`
}

type SeedDeck struct {
	Name      string   `json:"name"`
	OwnerId   string   `json:"owner_id"`
	CardTexts []string `json:"card_texts"`
}

type SeedSession struct {
	Id        string   `json:"id"`
	DeckNames []string `json:"deck_names"`
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

	// 1. Create Cards
	textToCard := make(map[string]*pb.Card)
	for _, sc := range seed.Cards {
		card, _ := domain.NewCard(sc.Text)
		if err := store.CreateCard(ctx, card); err != nil {
			return fmt.Errorf("failed to seed card %q: %w", sc.Text, err)
		}
		textToCard[sc.Text] = card
	}

	// 2. Create Decks
	nameToDeck := make(map[string]*pb.Deck)
	for _, sd := range seed.Decks {
		deck := domain.NewDeck(sd.Name, sd.OwnerId)
		for _, ct := range sd.CardTexts {
			card, ok := textToCard[ct]
			if !ok {
				return fmt.Errorf("card text %q not found for deck %q", ct, sd.Name)
			}
			domain.AddCardToDeck(deck, sd.OwnerId, card)
		}
		if err := store.CreateDeck(ctx, deck); err != nil {
			return fmt.Errorf("failed to seed deck %q: %w", sd.Name, err)
		}
		nameToDeck[sd.Name] = deck
	}

	// 3. Create Sessions
	for _, ss := range seed.Sessions {
		session := domain.NewSession()
		if ss.Id != "" {
			session.Id = ss.Id
		}
		for _, dn := range ss.DeckNames {
			deck, ok := nameToDeck[dn]
			if !ok {
				return fmt.Errorf("deck name %q not found for session %q", dn, session.Id)
			}
			domain.AddDeckToSession(session, deck)
		}
		if err := store.CreateSession(ctx, session); err != nil {
			return fmt.Errorf("failed to seed session %q: %w", session.Id, err)
		}
	}

	return nil
}
