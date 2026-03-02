package domain_test

import (
	"testing"

	"github.com/yanicksenn/ruthless/internal/domain"
)

func TestValidateCardText(t *testing.T) {
	tests := []struct {
		name    string
		text    string
		wantErr error
	}{
		{
			name:    "valid card with one blank",
			text:    "A big black ___",
			wantErr: nil,
		},
		{
			name:    "valid card with multiple blanks",
			text:    "___ is better than ___",
			wantErr: nil,
		},
		{
			name:    "invalid card with no blanks",
			text:    "Just some text without blanks",
			wantErr: domain.ErrInvalidCardBlank,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := domain.ValidateCardText(tt.text)
			if err != tt.wantErr {
				t.Errorf("ValidateCardText() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
