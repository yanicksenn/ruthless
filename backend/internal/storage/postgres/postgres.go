package postgres

import (
	"context"
	"database/sql"
	"embed"
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	_ "github.com/lib/pq"
	pb "github.com/yanicksenn/ruthless/api/v1"
	"github.com/yanicksenn/ruthless/backend/internal/storage"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

type Storage struct {
	db *sql.DB
}

func New(connStr string) (*Storage, error) {
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	s := &Storage{db: db}
	if err := s.Init(); err != nil {
		return nil, err
	}

	return s, nil
}

func (s *Storage) Init() error {
	// 1. Ensure migrations table exists
	_, err := s.db.Exec(`CREATE TABLE IF NOT EXISTS migrations (
		id TEXT PRIMARY KEY,
		applied_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
	)`)
	if err != nil {
		return fmt.Errorf("failed to create migrations table: %w", err)
	}

	// 2. Read migration files
	entries, err := migrationsFS.ReadDir("migrations")
	if err != nil {
		return fmt.Errorf("failed to read migrations directory: %w", err)
	}

	var files []string
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".sql") {
			files = append(files, entry.Name())
		}
	}
	sort.Strings(files)

	// 3. Apply migrations in order
	for _, file := range files {
		var exists bool
		err := s.db.QueryRow("SELECT EXISTS(SELECT 1 FROM migrations WHERE id = $1)", file).Scan(&exists)
		if err != nil {
			return fmt.Errorf("failed to check if migration %s exists: %w", file, err)
		}

		if exists {
			continue
		}

		fmt.Printf("Applying migration: %s\n", file)
		content, err := migrationsFS.ReadFile("migrations/" + file)
		if err != nil {
			return fmt.Errorf("failed to read migration file %s: %w", file, err)
		}

		tx, err := s.db.Begin()
		if err != nil {
			return fmt.Errorf("failed to begin transaction for migration %s: %w", file, err)
		}

		if _, err := tx.Exec(string(content)); err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to execute migration %s: %w", file, err)
		}

		if _, err := tx.Exec("INSERT INTO migrations (id) VALUES ($1)", file); err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to record migration %s: %w", file, err)
		}

		if err := tx.Commit(); err != nil {
			return fmt.Errorf("failed to commit migration %s: %w", file, err)
		}
	}

	return nil
}

// Card operations
func (s *Storage) CreateCard(ctx context.Context, card *pb.Card) error {
	_, err := s.db.ExecContext(ctx, "INSERT INTO cards (id, text, blanks) VALUES ($1, $2, $3)", card.Id, card.Text, card.Blanks)
	return err
}

func (s *Storage) GetCard(ctx context.Context, id string) (*pb.Card, error) {
	row := s.db.QueryRowContext(ctx, "SELECT id, text, blanks FROM cards WHERE id = $1", id)
	var c pb.Card
	if err := row.Scan(&c.Id, &c.Text, &c.Blanks); err != nil {
		if err == sql.ErrNoRows {
			return nil, storage.ErrNotFound
		}
		return nil, err
	}
	return &c, nil
}

func (s *Storage) ListCards(ctx context.Context) ([]*pb.Card, error) {
	rows, err := s.db.QueryContext(ctx, "SELECT id, text, blanks FROM cards")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var cards []*pb.Card
	for rows.Next() {
		var c pb.Card
		if err := rows.Scan(&c.Id, &c.Text, &c.Blanks); err != nil {
			return nil, err
		}
		cards = append(cards, &c)
	}
	return cards, nil
}

// User operations
func (s *Storage) CreateUser(ctx context.Context, user *pb.User) error {
	_, err := s.db.ExecContext(ctx, "INSERT INTO users (id, name) VALUES ($1, $2) ON CONFLICT (id) DO UPDATE SET name = EXCLUDED.name", user.Id, user.Name)
	return err
}

// Session operations
func (s *Storage) CreateSession(ctx context.Context, session *pb.Session) error {
	pids, _ := json.Marshal(session.PlayerIds)
	dids, _ := json.Marshal(session.DeckIds)
	_, err := s.db.ExecContext(ctx, "INSERT INTO sessions (id, player_ids, deck_ids) VALUES ($1, $2, $3)", session.Id, pids, dids)
	return err
}

func (s *Storage) GetSession(ctx context.Context, id string) (*pb.Session, error) {
	row := s.db.QueryRowContext(ctx, "SELECT id, player_ids, deck_ids FROM sessions WHERE id = $1", id)
	var s_id string
	var pids, dids []byte
	if err := row.Scan(&s_id, &pids, &dids); err != nil {
		if err == sql.ErrNoRows {
			return nil, storage.ErrNotFound
		}
		return nil, err
	}
	var session pb.Session
	session.Id = s_id
	json.Unmarshal(pids, &session.PlayerIds)
	json.Unmarshal(dids, &session.DeckIds)
	return &session, nil
}

func (s *Storage) UpdateSession(ctx context.Context, session *pb.Session) error {
	pids, _ := json.Marshal(session.PlayerIds)
	dids, _ := json.Marshal(session.DeckIds)
	res, err := s.db.ExecContext(ctx, "UPDATE sessions SET player_ids = $1, deck_ids = $2 WHERE id = $3", pids, dids, session.Id)
	if err != nil {
		return err
	}
	rows, _ := res.RowsAffected()
	if rows == 0 {
		return storage.ErrNotFound
	}
	return nil
}

func (s *Storage) ListSessions(ctx context.Context) ([]*pb.Session, error) {
	rows, err := s.db.QueryContext(ctx, "SELECT id, player_ids, deck_ids FROM sessions")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sessions []*pb.Session
	for rows.Next() {
		var s_id string
		var pids, dids []byte
		if err := rows.Scan(&s_id, &pids, &dids); err != nil {
			return nil, err
		}
		var session pb.Session
		session.Id = s_id
		json.Unmarshal(pids, &session.PlayerIds)
		json.Unmarshal(dids, &session.DeckIds)
		sessions = append(sessions, &session)
	}
	return sessions, nil
}

// Deck operations
func (s *Storage) CreateDeck(ctx context.Context, deck *pb.Deck) error {
	con, _ := json.Marshal(deck.Contributors)
	cids, _ := json.Marshal(deck.CardIds)
	_, err := s.db.ExecContext(ctx, "INSERT INTO decks (id, name, owner_id, contributors, card_ids) VALUES ($1, $2, $3, $4, $5)", deck.Id, deck.Name, deck.OwnerId, con, cids)
	return err
}

func (s *Storage) GetDeck(ctx context.Context, id string) (*pb.Deck, error) {
	row := s.db.QueryRowContext(ctx, "SELECT id, name, owner_id, contributors, card_ids FROM decks WHERE id = $1", id)
	var d pb.Deck
	var con, cids []byte
	if err := row.Scan(&d.Id, &d.Name, &d.OwnerId, &con, &cids); err != nil {
		if err == sql.ErrNoRows {
			return nil, storage.ErrNotFound
		}
		return nil, err
	}
	json.Unmarshal(con, &d.Contributors)
	json.Unmarshal(cids, &d.CardIds)
	return &d, nil
}

func (s *Storage) UpdateDeck(ctx context.Context, deck *pb.Deck) error {
	con, _ := json.Marshal(deck.Contributors)
	cids, _ := json.Marshal(deck.CardIds)
	res, err := s.db.ExecContext(ctx, "UPDATE decks SET name = $1, owner_id = $2, contributors = $3, card_ids = $4 WHERE id = $5", deck.Name, deck.OwnerId, con, cids, deck.Id)
	if err != nil {
		return err
	}
	rows, _ := res.RowsAffected()
	if rows == 0 {
		return storage.ErrNotFound
	}
	return nil
}

func (s *Storage) ListDecks(ctx context.Context) ([]*pb.Deck, error) {
	rows, err := s.db.QueryContext(ctx, "SELECT id, name, owner_id, contributors, card_ids FROM decks")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var decks []*pb.Deck
	for rows.Next() {
		var d pb.Deck
		var con, cids []byte
		if err := rows.Scan(&d.Id, &d.Name, &d.OwnerId, &con, &cids); err != nil {
			return nil, err
		}
		json.Unmarshal(con, &d.Contributors)
		json.Unmarshal(cids, &d.CardIds)
		decks = append(decks, &d)
	}
	return decks, nil
}

// Game operations
func (s *Storage) CreateGame(ctx context.Context, game *pb.Game) error {
	scores, _ := json.Marshal(game.Scores)
	hands, _ := json.Marshal(game.HiddenHands)
	black, _ := json.Marshal(game.HiddenBlackDeck)
	white, _ := json.Marshal(game.HiddenWhiteDeck)
	rounds, _ := json.Marshal(game.Rounds)

	_, err := s.db.ExecContext(ctx, `INSERT INTO games 
		(id, session_id, state, scores, hidden_hands, hidden_black_deck, hidden_white_deck, rounds) 
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
		game.Id, game.SessionId, int32(game.State), scores, hands, black, white, rounds)
	return err
}

func (s *Storage) GetGame(ctx context.Context, id string) (*pb.Game, error) {
	row := s.db.QueryRowContext(ctx, "SELECT id, session_id, state, scores, hidden_hands, hidden_black_deck, hidden_white_deck, rounds FROM games WHERE id = $1", id)
	var g pb.Game
	var state int32
	var scores, hands, black, white, rounds []byte
	if err := row.Scan(&g.Id, &g.SessionId, &state, &scores, &hands, &black, &white, &rounds); err != nil {
		if err == sql.ErrNoRows {
			return nil, storage.ErrNotFound
		}
		return nil, err
	}
	g.State = pb.GameState(state)
	json.Unmarshal(scores, &g.Scores)
	json.Unmarshal(hands, &g.HiddenHands)
	json.Unmarshal(black, &g.HiddenBlackDeck)
	json.Unmarshal(white, &g.HiddenWhiteDeck)
	json.Unmarshal(rounds, &g.Rounds)
	return &g, nil
}

func (s *Storage) UpdateGame(ctx context.Context, game *pb.Game) error {
	scores, _ := json.Marshal(game.Scores)
	hands, _ := json.Marshal(game.HiddenHands)
	black, _ := json.Marshal(game.HiddenBlackDeck)
	white, _ := json.Marshal(game.HiddenWhiteDeck)
	rounds, _ := json.Marshal(game.Rounds)

	res, err := s.db.ExecContext(ctx, `UPDATE games 
		SET state = $1, scores = $2, hidden_hands = $3, hidden_black_deck = $4, hidden_white_deck = $5, rounds = $6 
		WHERE id = $7`,
		int32(game.State), scores, hands, black, white, rounds, game.Id)
	if err != nil {
		return err
	}
	rows, _ := res.RowsAffected()
	if rows == 0 {
		return storage.ErrNotFound
	}
	return nil
}

func (s *Storage) GetGameBySession(ctx context.Context, sessionID string) (*pb.Game, error) {
	row := s.db.QueryRowContext(ctx, "SELECT id FROM games WHERE session_id = $1 LIMIT 1", sessionID)
	var id string
	if err := row.Scan(&id); err != nil {
		if err == sql.ErrNoRows {
			return nil, storage.ErrNotFound
		}
		return nil, err
	}
	return s.GetGame(ctx, id)
}

// Ensure Storage implements storage.Storage at compile time.
var _ storage.Storage = (*Storage)(nil)
