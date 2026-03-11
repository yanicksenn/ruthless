package domain_test

import (
	"testing"

	"github.com/yanicksenn/ruthless/backend/internal/domain"
)

func TestNewCard(t *testing.T) {
	c1, _ := domain.NewCard("A big black ___")
	if c1.Blanks != 1 {
		t.Errorf("expected 1 blank, got %d", c1.Blanks)
	}

	c2, _ := domain.NewCard("___ is better than ___")
	if c2.Blanks != 2 {
		t.Errorf("expected 2 blanks, got %d", c2.Blanks)
	}

	c3, _ := domain.NewCard("Just some text without blanks")
	if c3.Blanks != 0 {
		t.Errorf("expected 0 blanks, got %d", c3.Blanks)
	}
}
