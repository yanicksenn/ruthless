package domain

import (
	"errors"
	"strings"

	"github.com/google/uuid"
	pb "github.com/yanicksenn/ruthless/api/v1"
)

var (
	ErrInvalidCardBlank = errors.New("a card must contain at least one blank ('___')")
)

func NewCard(text string) (*pb.Card, error) {
	blanks := uint32(strings.Count(text, "___"))

	return &pb.Card{
		Id:     uuid.New().String(),
		Text:   text,
		Blanks: blanks,
	}, nil
}
