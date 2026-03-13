package domain

import (
	"errors"
	"strings"

	"github.com/google/uuid"
	pb "github.com/yanicksenn/ruthless/api/v1"
	"google.golang.org/protobuf/types/known/timestamppb"
)

var (
	ErrInvalidCardBlank = errors.New("a card must contain at least one blank ('___')")
)

func CountBlanks(text string) int {
	return strings.Count(text, "___")
}

func NewCard(text string, ownerID string) (*pb.Card, error) {
	color := pb.CardColor_CARD_COLOR_WHITE
	if CountBlanks(text) > 0 {
		color = pb.CardColor_CARD_COLOR_BLACK
	}

	return &pb.Card{
		Id:        uuid.New().String(),
		Text:      text,
		Color:     color,
		OwnerId:   ownerID,
		CreatedAt: timestamppb.Now(),
	}, nil
}
