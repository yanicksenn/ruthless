package memory

import (
	"context"
	"sort"
	"sync"

	pb "github.com/yanicksenn/ruthless/api/v1"
	"github.com/yanicksenn/ruthless/backend/internal/storage"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type Storage struct {
	mu       sync.RWMutex
	users    map[string]*pb.User
	cards    map[string]*pb.Card
	sessions map[string]*pb.Session
	decks    map[string]*pb.Deck
	games    map[string]*pb.Game
}

func New() *Storage {
	return &Storage{
		users:    make(map[string]*pb.User),
		cards:    make(map[string]*pb.Card),
		sessions: make(map[string]*pb.Session),
		decks:    make(map[string]*pb.Deck),
		games:    make(map[string]*pb.Game),
	}
}

func (s *Storage) CreateCard(ctx context.Context, card *pb.Card) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if card.CreatedAt == nil {
		card.CreatedAt = timestamppb.Now()
	}
	s.cards[card.Id] = card
	return nil
}

// -- User Methods
func (s *Storage) CreateUser(ctx context.Context, user *pb.User) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if user.CreatedAt == nil {
		user.CreatedAt = timestamppb.Now()
	}
	s.users[user.Id] = user
	return nil
}

func (s *Storage) GetUser(ctx context.Context, id string) (*pb.User, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	user, ok := s.users[id]
	if !ok {
		return nil, storage.ErrNotFound
	}
	return user, nil
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

func (s *Storage) ListCards(ctx context.Context, pageSize, pageNumber int32) ([]*pb.Card, int32, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	totalCount := int32(len(s.cards))
	allCards := make([]*pb.Card, 0, totalCount)
	for _, c := range s.cards {
		allCards = append(allCards, c)
	}

	// Simple sort to ensure stable pagination for memory storage
	sort.Slice(allCards, func(i, j int) bool {
		return allCards[i].Id < allCards[j].Id
	})

	if pageSize <= 0 {
		return allCards, totalCount, nil
	}

	start := (pageNumber - 1) * pageSize
	if start >= totalCount {
		return nil, totalCount, nil
	}

	end := start + pageSize
	if end > totalCount {
		end = totalCount
	}

	return allCards[start:end], totalCount, nil
}

func (s *Storage) DeleteCard(ctx context.Context, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.cards, id)
	return nil
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
	if deck.CreatedAt == nil {
		deck.CreatedAt = timestamppb.Now()
	}
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
	deck.UpdatedAt = timestamppb.Now()
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
