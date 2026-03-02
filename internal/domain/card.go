package domain

import (
	"errors"
	"strings"

	"github.com/google/uuid"
)

var (
	ErrInvalidCardBlank = errors.New("a card must contain at least one blank ('___')")
)

type Card struct {
	ID   string `json:"id"`
	Text string `json:"text"`
}

func NewCard(text string) (Card, error) {
	if err := ValidateCardText(text); err != nil {
		return Card{}, err
	}

	return Card{
		ID:   uuid.New().String(),
		Text: text,
	}, nil
}

func ValidateCardText(text string) error {
	if !strings.Contains(text, "___") {
		return ErrInvalidCardBlank
	}
	return nil
}
