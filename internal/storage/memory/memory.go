package memory

import (
	"context"
	"sync"

	"github.com/yanicksenn/ruthless/internal/domain"
	"github.com/yanicksenn/ruthless/internal/storage"
)

type Storage struct {
	mu       sync.RWMutex
	cards    map[string]domain.Card
	sessions map[string]domain.Session
}

func New() *Storage {
	return &Storage{
		cards:    make(map[string]domain.Card),
		sessions: make(map[string]domain.Session),
	}
}

func (s *Storage) CreateCard(ctx context.Context, card domain.Card) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.cards[card.ID] = card
	return nil
}

func (s *Storage) GetCard(ctx context.Context, id string) (domain.Card, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	card, ok := s.cards[id]
	if !ok {
		return domain.Card{}, storage.ErrNotFound
	}
	return card, nil
}

func (s *Storage) ListCards(ctx context.Context) ([]domain.Card, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	list := make([]domain.Card, 0, len(s.cards))
	for _, c := range s.cards {
		list = append(list, c)
	}
	return list, nil
}

func (s *Storage) CreateSession(ctx context.Context, session domain.Session) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.sessions[session.ID] = session
	return nil
}

func (s *Storage) GetSession(ctx context.Context, id string) (domain.Session, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	session, ok := s.sessions[id]
	if !ok {
		return domain.Session{}, storage.ErrNotFound
	}
	return session, nil
}

func (s *Storage) UpdateSession(ctx context.Context, session domain.Session) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.sessions[session.ID]; !ok {
		return storage.ErrNotFound
	}
	s.sessions[session.ID] = session
	return nil
}

// Ensure Storage implements storage.Storage at compile time.
var _ storage.Storage = (*Storage)(nil)
