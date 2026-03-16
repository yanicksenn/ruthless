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
	"google.golang.org/protobuf/types/known/timestamppb"
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
	_, err := s.db.ExecContext(ctx, "INSERT INTO cards (id, text, color, owner_id) VALUES ($1, $2, $3, $4)", card.Id, card.Text, int32(card.Color), card.OwnerId)
	return err
}

func (s *Storage) GetCard(ctx context.Context, id string) (*pb.Card, error) {
	row := s.db.QueryRowContext(ctx, "SELECT id, text, color, owner_id, created_at, updated_at FROM cards WHERE id = $1", id)
	var c pb.Card
	var color int32
	var createdAt sql.NullTime
	var updatedAt sql.NullTime
	var ownerID sql.NullString
	if err := row.Scan(&c.Id, &c.Text, &color, &ownerID, &createdAt, &updatedAt); err != nil {
		if err == sql.ErrNoRows {
			return nil, storage.ErrNotFound
		}
		return nil, err
	}
	c.Color = pb.CardColor(color)
	if ownerID.Valid {
		c.OwnerId = ownerID.String
	}
	if createdAt.Valid {
		c.CreatedAt = timestamppb.New(createdAt.Time)
	}
	if updatedAt.Valid {
		c.UpdatedAt = timestamppb.New(updatedAt.Time)
	}
	return &c, nil
}

func (s *Storage) ListCards(ctx context.Context, pageSize, pageNumber int32) ([]*pb.Card, int32, error) {
	var totalCount int32
	err := s.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM cards").Scan(&totalCount)
	if err != nil {
		return nil, 0, err
	}

	query := "SELECT id, text, color, owner_id, created_at, updated_at FROM cards ORDER BY id"
	var args []interface{}
	if pageSize > 0 {
		query += " LIMIT $1 OFFSET $2"
		args = append(args, pageSize, (pageNumber-1)*pageSize)
	}

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var cards []*pb.Card
	for rows.Next() {
		var c pb.Card
		var color int32
		var createdAt sql.NullTime
		var updatedAt sql.NullTime
		var ownerID sql.NullString
		if err := rows.Scan(&c.Id, &c.Text, &color, &ownerID, &createdAt, &updatedAt); err != nil {
			return nil, 0, err
		}
		c.Color = pb.CardColor(color)
		if ownerID.Valid {
			c.OwnerId = ownerID.String
		}
		if createdAt.Valid {
			c.CreatedAt = timestamppb.New(createdAt.Time)
		}
		if updatedAt.Valid {
			c.UpdatedAt = timestamppb.New(updatedAt.Time)
		}
		cards = append(cards, &c)
	}
	return cards, totalCount, nil
}

func (s *Storage) DeleteCard(ctx context.Context, id string) error {
	_, err := s.db.ExecContext(ctx, "DELETE FROM cards WHERE id = $1", id)
	return err
}

// User operations
func (s *Storage) CreateUser(ctx context.Context, user *pb.User) error {
	_, err := s.db.ExecContext(ctx, "INSERT INTO users (id, name) VALUES ($1, $2) ON CONFLICT (id) DO UPDATE SET name = EXCLUDED.name", user.Id, user.Name)
	return err
}

func (s *Storage) GetUser(ctx context.Context, id string) (*pb.User, error) {
	row := s.db.QueryRowContext(ctx, "SELECT id, name, created_at FROM users WHERE id = $1", id)
	var u pb.User
	var createdAt sql.NullTime
	if err := row.Scan(&u.Id, &u.Name, &createdAt); err != nil {
		if err == sql.ErrNoRows {
			return nil, storage.ErrNotFound
		}
		return nil, err
	}
	if createdAt.Valid {
		u.CreatedAt = timestamppb.New(createdAt.Time)
	}
	return &u, nil
}

// Session operations
func (s *Storage) CreateSession(ctx context.Context, session *pb.Session) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	_, err = tx.ExecContext(ctx, "INSERT INTO sessions (id, owner_id) VALUES ($1, $2)", session.Id, session.OwnerId)
	if err != nil {
		return err
	}

	for _, playerID := range session.PlayerIds {
		_, err = tx.ExecContext(ctx, "INSERT INTO session_players (session_id, player_id) VALUES ($1, $2)", session.Id, playerID)
		if err != nil {
			return err
		}
	}

	for _, deckID := range session.DeckIds {
		_, err = tx.ExecContext(ctx, "INSERT INTO session_decks (session_id, deck_id) VALUES ($1, $2)", session.Id, deckID)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (s *Storage) GetSession(ctx context.Context, id string) (*pb.Session, error) {
	row := s.db.QueryRowContext(ctx, "SELECT id, owner_id FROM sessions WHERE id = $1", id)
	var session pb.Session
	if err := row.Scan(&session.Id, &session.OwnerId); err != nil {
		if err == sql.ErrNoRows {
			return nil, storage.ErrNotFound
		}
		return nil, err
	}

	// Fetch players
	pRows, err := s.db.QueryContext(ctx, "SELECT player_id FROM session_players WHERE session_id = $1", id)
	if err != nil {
		return nil, err
	}
	defer pRows.Close()
	for pRows.Next() {
		var pID string
		if err := pRows.Scan(&pID); err != nil {
			return nil, err
		}
		session.PlayerIds = append(session.PlayerIds, pID)
	}

	// Fetch decks
	dRows, err := s.db.QueryContext(ctx, "SELECT deck_id FROM session_decks WHERE session_id = $1", id)
	if err != nil {
		return nil, err
	}
	defer dRows.Close()
	for dRows.Next() {
		var dID string
		if err := dRows.Scan(&dID); err != nil {
			return nil, err
		}
		session.DeckIds = append(session.DeckIds, dID)
	}

	return &session, nil
}

func (s *Storage) UpdateSession(ctx context.Context, session *pb.Session) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Ensure session exists
	res, err := tx.ExecContext(ctx, "UPDATE sessions SET owner_id = $1 WHERE id = $2", session.OwnerId, session.Id)
	if err != nil {
		return err
	}
	rowsAffected, _ := res.RowsAffected()
	if rowsAffected == 0 {
		return storage.ErrNotFound
	}
	_, err = tx.ExecContext(ctx, "DELETE FROM session_players WHERE session_id = $1", session.Id)
	if err != nil {
		return err
	}
	for _, pID := range session.PlayerIds {
		_, err = tx.ExecContext(ctx, "INSERT INTO session_players (session_id, player_id) VALUES ($1, $2)", session.Id, pID)
		if err != nil {
			return err
		}
	}

	// Update decks
	_, err = tx.ExecContext(ctx, "DELETE FROM session_decks WHERE session_id = $1", session.Id)
	if err != nil {
		return err
	}
	for _, dID := range session.DeckIds {
		_, err = tx.ExecContext(ctx, "INSERT INTO session_decks (session_id, deck_id) VALUES ($1, $2)", session.Id, dID)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (s *Storage) ListSessions(ctx context.Context) ([]*pb.Session, error) {
	rows, err := s.db.QueryContext(ctx, "SELECT id, owner_id FROM sessions")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
 
	var sessions []*pb.Session
	for rows.Next() {
		var sID string
		var oID string
		if err := rows.Scan(&sID, &oID); err != nil {
			return nil, err
		}
		sessions = append(sessions, &pb.Session{Id: sID, OwnerId: oID})
	}

	for _, sess := range sessions {
		// Players
		pRows, err := s.db.QueryContext(ctx, "SELECT player_id FROM session_players WHERE session_id = $1", sess.Id)
		if err != nil {
			return nil, err
		}
		for pRows.Next() {
			var pID string
			if err := pRows.Scan(&pID); err != nil {
				pRows.Close()
				return nil, err
			}
			sess.PlayerIds = append(sess.PlayerIds, pID)
		}
		pRows.Close()

		// Decks
		dRows, err := s.db.QueryContext(ctx, "SELECT deck_id FROM session_decks WHERE session_id = $1", sess.Id)
		if err != nil {
			return nil, err
		}
		for dRows.Next() {
			var dID string
			if err := dRows.Scan(&dID); err != nil {
				dRows.Close()
				return nil, err
			}
			sess.DeckIds = append(sess.DeckIds, dID)
		}
		dRows.Close()
	}

	return sessions, nil
}

// Deck operations
func (s *Storage) CreateDeck(ctx context.Context, deck *pb.Deck) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	_, err = tx.ExecContext(ctx, "INSERT INTO decks (id, name, owner_id) VALUES ($1, $2, $3)", deck.Id, deck.Name, deck.OwnerId)
	if err != nil {
		return err
	}

	for _, contributorID := range deck.Contributors {
		_, err = tx.ExecContext(ctx, "INSERT INTO deck_contributors (deck_id, user_id) VALUES ($1, $2)", deck.Id, contributorID)
		if err != nil {
			return err
		}
	}

	for _, cardID := range deck.CardIds {
		_, err = tx.ExecContext(ctx, "INSERT INTO deck_cards (deck_id, card_id) VALUES ($1, $2)", deck.Id, cardID)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (s *Storage) GetDeck(ctx context.Context, id string) (*pb.Deck, error) {
	row := s.db.QueryRowContext(ctx, "SELECT id, name, owner_id, created_at, updated_at FROM decks WHERE id = $1", id)
	var d pb.Deck
	var createdAt sql.NullTime
	var updatedAt sql.NullTime
	if err := row.Scan(&d.Id, &d.Name, &d.OwnerId, &createdAt, &updatedAt); err != nil {
		if err == sql.ErrNoRows {
			return nil, storage.ErrNotFound
		}
		return nil, err
	}
	if createdAt.Valid {
		d.CreatedAt = timestamppb.New(createdAt.Time)
	}
	if updatedAt.Valid {
		d.UpdatedAt = timestamppb.New(updatedAt.Time)
	}

	// Fetch contributors
	rows, err := s.db.QueryContext(ctx, "SELECT user_id FROM deck_contributors WHERE deck_id = $1", id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var userID string
		if err := rows.Scan(&userID); err != nil {
			return nil, err
		}
		d.Contributors = append(d.Contributors, userID)
	}

	// Fetch cards
	rows, err = s.db.QueryContext(ctx, "SELECT card_id FROM deck_cards WHERE deck_id = $1", id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var cardID string
		if err := rows.Scan(&cardID); err != nil {
			return nil, err
		}
		d.CardIds = append(d.CardIds, cardID)
	}

	return &d, nil
}

func (s *Storage) UpdateDeck(ctx context.Context, deck *pb.Deck) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	res, err := tx.ExecContext(ctx, "UPDATE decks SET name = $1, owner_id = $2, updated_at = CURRENT_TIMESTAMP WHERE id = $3", deck.Name, deck.OwnerId, deck.Id)
	if err != nil {
		return err
	}
	rowsAffected, _ := res.RowsAffected()
	if rowsAffected == 0 {
		return storage.ErrNotFound
	}

	// Update contributors
	_, err = tx.ExecContext(ctx, "DELETE FROM deck_contributors WHERE deck_id = $1", deck.Id)
	if err != nil {
		return err
	}
	for _, contributorID := range deck.Contributors {
		_, err = tx.ExecContext(ctx, "INSERT INTO deck_contributors (deck_id, user_id) VALUES ($1, $2)", deck.Id, contributorID)
		if err != nil {
			return err
		}
	}

	// Update cards
	_, err = tx.ExecContext(ctx, "DELETE FROM deck_cards WHERE deck_id = $1", deck.Id)
	if err != nil {
		return err
	}
	for _, cardID := range deck.CardIds {
		_, err = tx.ExecContext(ctx, "INSERT INTO deck_cards (deck_id, card_id) VALUES ($1, $2)", deck.Id, cardID)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (s *Storage) ListDecks(ctx context.Context) ([]*pb.Deck, error) {
	rows, err := s.db.QueryContext(ctx, "SELECT id, name, owner_id, created_at, updated_at FROM decks")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var decks []*pb.Deck
	for rows.Next() {
		var d pb.Deck
		var createdAt sql.NullTime
		var updatedAt sql.NullTime
		if err := rows.Scan(&d.Id, &d.Name, &d.OwnerId, &createdAt, &updatedAt); err != nil {
			return nil, err
		}
		if createdAt.Valid {
			d.CreatedAt = timestamppb.New(createdAt.Time)
		}
		if updatedAt.Valid {
			d.UpdatedAt = timestamppb.New(updatedAt.Time)
		}
		decks = append(decks, &d)
	}

	// Fetch relations for each deck
	for _, d := range decks {
		// Contributors
		cRows, err := s.db.QueryContext(ctx, "SELECT user_id FROM deck_contributors WHERE deck_id = $1", d.Id)
		if err != nil {
			return nil, err
		}
		for cRows.Next() {
			var userID string
			if err := cRows.Scan(&userID); err != nil {
				cRows.Close()
				return nil, err
			}
			d.Contributors = append(d.Contributors, userID)
		}
		cRows.Close()

		// Cards
		cardRows, err := s.db.QueryContext(ctx, "SELECT card_id FROM deck_cards WHERE deck_id = $1", d.Id)
		if err != nil {
			return nil, err
		}
		for cardRows.Next() {
			var cardID string
			if err := cardRows.Scan(&cardID); err != nil {
				cardRows.Close()
				return nil, err
			}
			d.CardIds = append(d.CardIds, cardID)
		}
		cardRows.Close()
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
	players, _ := json.Marshal(game.Players)
	playerIDs, _ := json.Marshal(game.PlayerIds)

	_, err := s.db.ExecContext(ctx, `INSERT INTO games 
		(id, session_id, state, scores, hidden_hands, hidden_black_deck, hidden_white_deck, rounds, players, player_ids) 
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`,
		game.Id, game.SessionId, int32(game.State), scores, hands, black, white, rounds, players, playerIDs)
	return err
}

func (s *Storage) GetGame(ctx context.Context, id string) (*pb.Game, error) {
	row := s.db.QueryRowContext(ctx, "SELECT id, session_id, state, scores, hidden_hands, hidden_black_deck, hidden_white_deck, rounds, players, player_ids, created_at FROM games WHERE id = $1", id)
	var g pb.Game
	var state int32
	var scores, hands, black, white, rounds, players, playerIDs []byte
	var createdAt sql.NullTime
	if err := row.Scan(&g.Id, &g.SessionId, &state, &scores, &hands, &black, &white, &rounds, &players, &playerIDs, &createdAt); err != nil {
		if err == sql.ErrNoRows {
			return nil, storage.ErrNotFound
		}
		return nil, err
	}
	g.State = pb.GameState(state)
	if createdAt.Valid {
		g.CreatedAt = timestamppb.New(createdAt.Time)
	}
	json.Unmarshal(scores, &g.Scores)
	json.Unmarshal(hands, &g.HiddenHands)
	json.Unmarshal(black, &g.HiddenBlackDeck)
	json.Unmarshal(white, &g.HiddenWhiteDeck)
	json.Unmarshal(rounds, &g.Rounds)
	json.Unmarshal(players, &g.Players)
	json.Unmarshal(playerIDs, &g.PlayerIds)
	return &g, nil
}

func (s *Storage) UpdateGame(ctx context.Context, game *pb.Game) error {
	scores, _ := json.Marshal(game.Scores)
	hands, _ := json.Marshal(game.HiddenHands)
	black, _ := json.Marshal(game.HiddenBlackDeck)
	white, _ := json.Marshal(game.HiddenWhiteDeck)
	rounds, _ := json.Marshal(game.Rounds)
	players, _ := json.Marshal(game.Players)
	playerIDs, _ := json.Marshal(game.PlayerIds)

	res, err := s.db.ExecContext(ctx, `UPDATE games 
		SET state = $1, scores = $2, hidden_hands = $3, hidden_black_deck = $4, hidden_white_deck = $5, rounds = $6, players = $7, player_ids = $8 
		WHERE id = $9`,
		int32(game.State), scores, hands, black, white, rounds, players, playerIDs, game.Id)
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
