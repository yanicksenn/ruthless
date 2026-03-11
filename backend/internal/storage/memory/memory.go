package memory

import (
	"context"
	"sync"

	pb "github.com/yanicksenn/ruthless/api/v1"
	"github.com/yanicksenn/ruthless/backend/internal/storage"
)

type Storage struct {
	mu       sync.RWMutex
	cards    map[string]*pb.Card
	sessions map[string]*pb.Session
	decks    map[string]*pb.Deck
	games    map[string]*pb.Game
}

func New() *Storage {
	return &Storage{
		cards:    make(map[string]*pb.Card),
		sessions: make(map[string]*pb.Session),
		decks:    make(map[string]*pb.Deck),
		games:    make(map[string]*pb.Game),
	}
}

func (s *Storage) CreateCard(ctx context.Context, card *pb.Card) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.cards[card.Id] = card
	return nil
}

func (s *Storage) GetCard(ctx context.Context, id string) (*pb.Card, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	card, ok := s.cards[id]
	if !ok {
		return nil, storage.ErrNotFound
	}
	return card, nil
}

func (s *Storage) ListCards(ctx context.Context) ([]*pb.Card, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	list := make([]*pb.Card, 0, len(s.cards))
	for _, c := range s.cards {
		list = append(list, c)
	}
	return list, nil
}

func (s *Storage) CreateSession(ctx context.Context, session *pb.Session) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.sessions[session.Id] = session
	return nil
}

func (s *Storage) GetSession(ctx context.Context, id string) (*pb.Session, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	session, ok := s.sessions[id]
	if !ok {
		return nil, storage.ErrNotFound
	}
	return session, nil
}

func (s *Storage) UpdateSession(ctx context.Context, session *pb.Session) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.sessions[session.Id]; !ok {
		return storage.ErrNotFound
	}
	s.sessions[session.Id] = session
	return nil
}

func (s *Storage) ListSessions(ctx context.Context) ([]*pb.Session, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	list := make([]*pb.Session, 0, len(s.sessions))
	for _, session := range s.sessions {
		list = append(list, session)
	}
	return list, nil
}

// -- Deck Methods
func (s *Storage) CreateDeck(ctx context.Context, deck *pb.Deck) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.decks[deck.Id] = deck
	return nil
}

func (s *Storage) GetDeck(ctx context.Context, id string) (*pb.Deck, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	deck, ok := s.decks[id]
	if !ok {
		return nil, storage.ErrNotFound
	}
	return deck, nil
}

func (s *Storage) UpdateDeck(ctx context.Context, deck *pb.Deck) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.decks[deck.Id]; !ok {
		return storage.ErrNotFound
	}
	s.decks[deck.Id] = deck
	return nil
}

func (s *Storage) ListDecks(ctx context.Context) ([]*pb.Deck, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	list := make([]*pb.Deck, 0, len(s.decks))
	for _, d := range s.decks {
		list = append(list, d)
	}
	return list, nil
}

// -- Game Methods
func (s *Storage) CreateGame(ctx context.Context, game *pb.Game) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.games[game.Id] = game
	return nil
}

func (s *Storage) GetGame(ctx context.Context, id string) (*pb.Game, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	game, ok := s.games[id]
	if !ok {
		return nil, storage.ErrNotFound
	}
	return game, nil
}

func (s *Storage) UpdateGame(ctx context.Context, game *pb.Game) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.games[game.Id]; !ok {
		return storage.ErrNotFound
	}
	s.games[game.Id] = game
	return nil
}

func (s *Storage) GetGameBySession(ctx context.Context, sessionID string) (*pb.Game, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, game := range s.games {
		if game.SessionId == sessionID {
			return game, nil
		}
	}
	return nil, storage.ErrNotFound
}

// Ensure Storage implements storage.Storage at compile time.
var _ storage.Storage = (*Storage)(nil)
