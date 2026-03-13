package domain_test

import (
	"testing"

	pb "github.com/yanicksenn/ruthless/api/v1"
	"github.com/yanicksenn/ruthless/backend/internal/domain"
)

func TestNewCard(t *testing.T) {
	c1, _ := domain.NewCard("A big black ___", "test-owner")
	if c1.Color != pb.CardColor_CARD_COLOR_BLACK {
		t.Errorf("expected black color, got %v", c1.Color)
	}
	if domain.CountBlanks(c1.Text) != 1 {
		t.Errorf("expected 1 blank, got %d", domain.CountBlanks(c1.Text))
	}

	c2, _ := domain.NewCard("___ is better than ___", "test-owner")
	if c2.Color != pb.CardColor_CARD_COLOR_BLACK {
		t.Errorf("expected black color, got %v", c2.Color)
	}
	if domain.CountBlanks(c2.Text) != 2 {
		t.Errorf("expected 2 blanks, got %d", domain.CountBlanks(c2.Text))
	}

	c3, _ := domain.NewCard("Just some text without blanks", "test-owner")
	if c3.Color != pb.CardColor_CARD_COLOR_WHITE {
		t.Errorf("expected white color, got %v", c3.Color)
	}
	if domain.CountBlanks(c3.Text) != 0 {
		t.Errorf("expected 0 blanks, got %d", domain.CountBlanks(c3.Text))
	}
}
