package memory

import (
	"context"
	"sort"
	"strings"
	"sync"
	"time"

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
	revokedTokens map[string]time.Time
	invitations   map[string]*pb.FriendInvitation
	friendships   map[string]map[string]time.Time // userID -> friendID -> createdAt
	notifications      map[string]map[pb.NotificationType]int32 // userID -> type -> count
	sessionInvitations map[string]*pb.SessionInvitation
}

func New() *Storage {
	return &Storage{
		users:    make(map[string]*pb.User),
		cards:    make(map[string]*pb.Card),
		sessions: make(map[string]*pb.Session),
		decks:    make(map[string]*pb.Deck),
		games:    make(map[string]*pb.Game),
		revokedTokens: make(map[string]time.Time),
		invitations:   make(map[string]*pb.FriendInvitation),
		friendships:        make(map[string]map[string]time.Time),
		notifications:      make(map[string]map[pb.NotificationType]int32),
		sessionInvitations: make(map[string]*pb.SessionInvitation),
	}
}

func (s *Storage) CreateCard(ctx context.Context, card *pb.Card) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if card.CreatedAt == nil {
		card.CreatedAt = timestamppb.Now()
	}
	// Recalculate color to ensure consistency, matching Postgres generated column behavior
	card.Color = pb.CardColor_CARD_COLOR_WHITE
	if strings.Contains(card.Text, "___") {
		card.Color = pb.CardColor_CARD_COLOR_BLACK
	}
	s.cards[card.Id] = card
	return nil
}

// -- User Methods
func (s *Storage) CreateUser(ctx context.Context, user *pb.User) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.users[user.Id]; ok {
		return storage.ErrAlreadyExists
	}
	if user.CreatedAt == nil {
		user.CreatedAt = timestamppb.Now()
	}
	s.users[user.Id] = user
	return nil
}

func (s *Storage) UpdateUser(ctx context.Context, user *pb.User) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.users[user.Id]; !ok {
		return storage.ErrNotFound
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
	if user.Identifier == "" {
		user.PendingCompletion = true
	} else {
		user.PendingCompletion = false
	}
	return user, nil
}

func (s *Storage) GetUserByNameAndIdentifier(ctx context.Context, name, identifier string) (*pb.User, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, user := range s.users {
		if user.Name == name && user.Identifier == identifier {
			if user.Identifier == "" {
				user.PendingCompletion = true
			} else {
				user.PendingCompletion = false
			}
			return user, nil
		}
	}
	return nil, storage.ErrNotFound
}

func (s *Storage) UpdateUserLastActive(ctx context.Context, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	user, ok := s.users[id]
	if !ok {
		return storage.ErrNotFound
	}
	user.LastActiveAt = timestamppb.Now()
	return nil
}

// -- Token Methods
func (s *Storage) RevokeToken(ctx context.Context, token string, expiresAt time.Time) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.revokedTokens[token] = expiresAt
	return nil
}

func (s *Storage) IsTokenRevoked(ctx context.Context, token string) (bool, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	exp, ok := s.revokedTokens[token]
	if !ok {
		return false, nil
	}
	// Note: in-memory we could lazily clean up expired blocks, but simple bool is fine
	if exp.Before(time.Now()) {
		return false, nil
	}
	return true, nil
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

func (s *Storage) ListCards(ctx context.Context, ownerID string, pageSize, pageNumber int32, ids []string, filter string, orderBy *pb.CardOrder, includeDeckIDs []string, color pb.CardColor, excludeDeckIDs []string) ([]*pb.Card, int32, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	includeCardIds := make(map[string]bool)
	if len(includeDeckIDs) > 0 {
		for _, dID := range includeDeckIDs {
			deck, ok := s.decks[dID]
			if !ok {
				return nil, 0, storage.ErrNotFound
			}
			for _, id := range deck.CardIds {
				includeCardIds[id] = true
			}
		}
	}

	excludeCardIds := make(map[string]bool)
	if len(excludeDeckIDs) > 0 {
		for _, dID := range excludeDeckIDs {
			deck, ok := s.decks[dID]
			if !ok {
				continue // Or return error? Postgres implementation wouldn't error if deck doesn't exist, it just wouldn't exclude anything.
			}
			for _, id := range deck.CardIds {
				excludeCardIds[id] = true
			}
		}
	}

	var filteredCards []*pb.Card
	idMap := make(map[string]bool)
	for _, id := range ids {
		idMap[id] = true
	}

	for _, c := range s.cards {
		// Filter by inclusion decks if provided
		if len(includeDeckIDs) > 0 && !includeCardIds[c.Id] {
			continue
		}
		// Filter by exclusion decks if provided
		if len(excludeDeckIDs) > 0 && excludeCardIds[c.Id] {
			continue
		}
		// Filter by ID list if provided
		if len(ids) > 0 && !idMap[c.Id] {
			continue
		}
		// Filter by owner if provided
		if ownerID != "" && c.OwnerId != ownerID {
			continue
		}
		// Filter by substring if provided
		if filter != "" && !strings.Contains(strings.ToLower(c.Text), strings.ToLower(filter)) {
			continue
		}
		// Filter by color if provided
		if color != pb.CardColor_CARD_COLOR_UNSPECIFIED && c.Color != color {
			continue
		}
		filteredCards = append(filteredCards, c)
	}

	// Dynamic sorting
	if orderBy != nil {
		sort.Slice(filteredCards, func(i, j int) bool {
			var less bool
			switch orderBy.Field {
			case pb.CardOrderField_CARD_ORDER_FIELD_TEXT:
				less = filteredCards[i].Text < filteredCards[j].Text
			case pb.CardOrderField_CARD_ORDER_FIELD_CREATED_AT:
				ti := filteredCards[i].CreatedAt.AsTime()
				tj := filteredCards[j].CreatedAt.AsTime()
				less = ti.Before(tj)
			default:
				less = filteredCards[i].Id < filteredCards[j].Id
			}
			if orderBy.Descending {
				return !less
			}
			return less
		})
	} else {
		// Default sort
		sort.Slice(filteredCards, func(i, j int) bool {
			return filteredCards[i].Id < filteredCards[j].Id
		})
	}

	totalCount := int32(len(filteredCards))

	if pageSize <= 0 {
		return filteredCards, totalCount, nil
	}

	start := (pageNumber - 1) * pageSize
	if start >= totalCount {
		return []*pb.Card{}, totalCount, nil
	}

	end := start + pageSize
	if end > totalCount {
		end = totalCount
	}

	return filteredCards[start:end], totalCount, nil
}

func (s *Storage) DeleteCard(ctx context.Context, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.cards, id)
	return nil
}

func (s *Storage) UpdateCard(ctx context.Context, card *pb.Card) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	stored, ok := s.cards[card.Id]
	if !ok {
		return storage.ErrNotFound
	}
	stored.Text = card.Text
	stored.UpdatedAt = timestamppb.Now()

	// Recalculate color
	stored.Color = pb.CardColor_CARD_COLOR_WHITE
	if strings.Contains(stored.Text, "___") {
		stored.Color = pb.CardColor_CARD_COLOR_BLACK
	}

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

func (s *Storage) ListSessions(ctx context.Context, playerID string) ([]*pb.Session, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	list := make([]*pb.Session, 0)
	for _, session := range s.sessions {
		// Find associated game
		var game *pb.Game
		for _, g := range s.games {
			if g.SessionId == session.Id {
				game = g
				break
			}
		}

		if game == nil {
			continue
		}

		isParticipant := false
		for _, pid := range session.PlayerIds {
			if pid == playerID {
				isParticipant = true
				break
			}
		}

		// Filter: WAITING or (PLAYING/JUDGING and participant)
		if game.State == pb.GameState_GAME_STATE_WAITING ||
			((game.State == pb.GameState_GAME_STATE_PLAYING || game.State == pb.GameState_GAME_STATE_JUDGING) && isParticipant) {
			list = append(list, session)
		}
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
	if deck.CardContributorIds == nil {
		deck.CardContributorIds = make(map[string]string)
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

func (s *Storage) ListDecks(ctx context.Context, ownerID string) ([]*pb.Deck, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	list := make([]*pb.Deck, 0, len(s.decks))
	for _, d := range s.decks {
		isContributor := false
		for _, c := range d.Contributors {
			if c == ownerID {
				isContributor = true
				break
			}
		}

		isSubscriber := false
		for _, sub := range d.Subscribers {
			if sub == ownerID {
				isSubscriber = true
				break
			}
		}

		if ownerID == "" || d.OwnerId == ownerID || isContributor || isSubscriber {
			list = append(list, d)
		}
	}
	return list, nil
}

func (s *Storage) SubscribeToDeck(ctx context.Context, deckID, userID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	deck, ok := s.decks[deckID]
	if !ok {
		return storage.ErrNotFound
	}
	for _, id := range deck.Subscribers {
		if id == userID {
			return nil
		}
	}
	deck.Subscribers = append(deck.Subscribers, userID)
	return nil
}

func (s *Storage) UnsubscribeFromDeck(ctx context.Context, deckID, userID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	deck, ok := s.decks[deckID]
	if !ok {
		return storage.ErrNotFound
	}
	for i, id := range deck.Subscribers {
		if id == userID {
			deck.Subscribers = append(deck.Subscribers[:i], deck.Subscribers[i+1:]...)
			return nil
		}
	}
	return storage.ErrNotFound
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

func (s *Storage) CountCardsByOwner(ctx context.Context, ownerID string) (int32, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var count int32
	for _, c := range s.cards {
		if c.OwnerId == ownerID {
			count++
		}
	}
	return count, nil
}

func (s *Storage) CountDecksByOwner(ctx context.Context, ownerID string) (int32, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var count int32
	for _, d := range s.decks {
		if d.OwnerId == ownerID {
			count++
		}
	}
	return count, nil
}

func (s *Storage) CountSessionsByOwner(ctx context.Context, ownerID string) (int32, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var count int32
	for _, sess := range s.sessions {
		if sess.OwnerId == ownerID {
			count++
		}
	}
	return count, nil
}

func (s *Storage) CreateInvitation(ctx context.Context, senderID, receiverID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, inv := range s.invitations {
		if inv.Sender.Id == senderID && inv.Receiver.Id == receiverID {
			return storage.ErrAlreadyExists
		}
	}

	sender, ok := s.users[senderID]
	if !ok {
		return storage.ErrNotFound
	}
	receiver, ok := s.users[receiverID]
	if !ok {
		return storage.ErrNotFound
	}

	id := time.Now().Format(time.RFC3339Nano) // Simple ID for memory storage
	invitation := &pb.FriendInvitation{
		Id: id,
		Sender: &pb.Player{
			Id:         sender.Id,
			Name:       sender.Name,
			Identifier: sender.Identifier,
		},
		Receiver: &pb.Player{
			Id:         receiver.Id,
			Name:       receiver.Name,
			Identifier: receiver.Identifier,
		},
		CreatedAt: timestamppb.Now(),
	}
	s.invitations[id] = invitation
	return nil
}

func (s *Storage) GetInvitation(ctx context.Context, id string) (*pb.FriendInvitation, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	inv, ok := s.invitations[id]
	if !ok {
		return nil, storage.ErrNotFound
	}
	return inv, nil
}

func (s *Storage) DeleteInvitation(ctx context.Context, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.invitations, id)
	return nil
}

func (s *Storage) ListInvitations(ctx context.Context, receiverID string, pageSize, pageNumber int32) ([]*pb.FriendInvitation, int32, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var invitations []*pb.FriendInvitation
	for _, inv := range s.invitations {
		if inv.Receiver.Id == receiverID {
			invitations = append(invitations, inv)
		}
	}
	sort.Slice(invitations, func(i, j int) bool {
		return invitations[i].CreatedAt.AsTime().After(invitations[j].CreatedAt.AsTime())
	})

	totalCount := int32(len(invitations))
	if pageSize <= 0 {
		return invitations, totalCount, nil
	}

	start := (pageNumber - 1) * pageSize
	if start >= totalCount {
		return []*pb.FriendInvitation{}, totalCount, nil
	}

	end := start + pageSize
	if end > totalCount {
		end = totalCount
	}

	return invitations[start:end], totalCount, nil
}

func (s *Storage) CreateFriendship(ctx context.Context, userID, friendID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.friendships[userID]; !ok {
		s.friendships[userID] = make(map[string]time.Time)
	}
	if _, ok := s.friendships[friendID]; !ok {
		s.friendships[friendID] = make(map[string]time.Time)
	}

	if _, ok := s.friendships[userID][friendID]; ok {
		return storage.ErrAlreadyExists
	}

	now := time.Now()
	s.friendships[userID][friendID] = now
	s.friendships[friendID][userID] = now
	return nil
}

func (s *Storage) DeleteFriendship(ctx context.Context, userID, friendID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if friends, ok := s.friendships[userID]; ok {
		delete(friends, friendID)
	}
	if friends, ok := s.friendships[friendID]; ok {
		delete(friends, userID)
	}
	return nil
}

func (s *Storage) ListFriends(ctx context.Context, userID string, excludeSessionID string, filter string, pageSize, pageNumber int32) ([]*pb.Player, int32, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	friendsMap, ok := s.friendships[userID]
	if !ok {
		return []*pb.Player{}, 0, nil
	}

	var friends []*pb.Player
	for fID := range friendsMap {
		user, ok := s.users[fID]
		if !ok {
			continue
		}

		if excludeSessionID != "" {
			inSession := false
			if sess, ok := s.sessions[excludeSessionID]; ok {
				for _, pid := range sess.PlayerIds {
					if pid == fID {
						inSession = true
						break
					}
				}
			}
			for _, inv := range s.sessionInvitations {
				if inv.Session.Id == excludeSessionID && inv.Receiver.Id == fID {
					inSession = true
					break
				}
			}
			if inSession {
				continue
			}
		}

		if filter != "" {
			filterLower := strings.ToLower(filter)
			if !strings.Contains(strings.ToLower(user.Name), filterLower) {
				continue
			}
		}

		friends = append(friends, &pb.Player{
			Id:         user.Id,
			Name:       user.Name,
			Identifier: user.Identifier,
		})
	}

	sort.Slice(friends, func(i, j int) bool {
		return friends[i].Name < friends[j].Name
	})

	totalCount := int32(len(friends))
	if pageSize <= 0 {
		return friends, totalCount, nil
	}

	start := (pageNumber - 1) * pageSize
	if start >= totalCount {
		return []*pb.Player{}, totalCount, nil
	}

	end := start + pageSize
	if end > totalCount {
		end = totalCount
	}

	return friends[start:end], totalCount, nil
}

// -- Notification Methods
func (s *Storage) IncrementNotificationCounter(ctx context.Context, userID string, notificationType pb.NotificationType) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	userNotifs, ok := s.notifications[userID]
	if !ok {
		userNotifs = make(map[pb.NotificationType]int32)
		s.notifications[userID] = userNotifs
	}
	userNotifs[notificationType]++
	return nil
}

func (s *Storage) ResetNotificationCounter(ctx context.Context, userID string, notificationType pb.NotificationType) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if userNotifs, ok := s.notifications[userID]; ok {
		delete(userNotifs, notificationType)
	}
	return nil
}

func (s *Storage) GetNotifications(ctx context.Context, userID string) ([]*pb.Notification, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var notifs []*pb.Notification
	if userNotifs, ok := s.notifications[userID]; ok {
		for nType, count := range userNotifs {
			notifs = append(notifs, &pb.Notification{
				Type:  nType,
				Count: count,
			})
		}
	}
	return notifs, nil
}

func (s *Storage) CreateSessionInvitation(ctx context.Context, sessionID, senderID, receiverID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, inv := range s.sessionInvitations {
		if inv.Session.Id == sessionID && inv.Receiver.Id == receiverID {
			return storage.ErrAlreadyExists
		}
	}

	sender, ok := s.users[senderID]
	if !ok {
		return storage.ErrNotFound
	}
	receiver, ok := s.users[receiverID]
	if !ok {
		return storage.ErrNotFound
	}
	session, ok := s.sessions[sessionID]
	if !ok {
		return storage.ErrNotFound
	}

	id := time.Now().Format(time.RFC3339Nano)
	invitation := &pb.SessionInvitation{
		Id:      id,
		Session: session,
		Sender: &pb.Player{
			Id:         sender.Id,
			Name:       sender.Name,
			Identifier: sender.Identifier,
		},
		Receiver: &pb.Player{
			Id:         receiver.Id,
			Name:       receiver.Name,
			Identifier: receiver.Identifier,
		},
		CreatedAt: timestamppb.Now(),
	}
	s.sessionInvitations[id] = invitation
	return nil
}

func (s *Storage) GetSessionInvitation(ctx context.Context, id string) (*pb.SessionInvitation, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	inv, ok := s.sessionInvitations[id]
	if !ok {
		return nil, storage.ErrNotFound
	}
	return inv, nil
}

func (s *Storage) DeleteSessionInvitation(ctx context.Context, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.sessionInvitations, id)
	return nil
}

func (s *Storage) ListSessionInvitations(ctx context.Context, receiverID string, pageSize, pageNumber int32) ([]*pb.SessionInvitation, int32, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var invitations []*pb.SessionInvitation
	for _, inv := range s.sessionInvitations {
		if inv.Receiver.Id == receiverID {
			invitations = append(invitations, inv)
		}
	}
	sort.Slice(invitations, func(i, j int) bool {
		return invitations[i].CreatedAt.AsTime().After(invitations[j].CreatedAt.AsTime())
	})

	totalCount := int32(len(invitations))
	if pageSize <= 0 {
		return invitations, totalCount, nil
	}
	start := (pageNumber - 1) * pageSize
	if start >= totalCount {
		return []*pb.SessionInvitation{}, totalCount, nil
	}
	end := start + pageSize
	if end > totalCount {
		end = totalCount
	}
	return invitations[start:end], totalCount, nil
}

func (s *Storage) DeleteUnansweredSessionInvitations(ctx context.Context, sessionID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	for id, inv := range s.sessionInvitations {
		if inv.Session.Id == sessionID {
			// Instead of modifying map while iterating, delete is safe in Go over maps, but let's be explicitly safe.
			delete(s.sessionInvitations, id)
		}
	}
	return nil
}

// Ensure Storage implements storage.Storage at compile time.
var _ storage.Storage = (*Storage)(nil)
