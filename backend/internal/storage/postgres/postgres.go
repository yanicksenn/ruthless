package postgres

import (
	"context"
	"database/sql"
	"embed"
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

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
	_, err := s.db.ExecContext(ctx, "INSERT INTO cards (id, text, owner_id) VALUES ($1, $2, $3)", card.Id, card.Text, card.OwnerId)
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

func (s *Storage) ListCards(ctx context.Context, ownerID string, pageSize, pageNumber int32, ids []string, filter string, orderBy *pb.CardOrder, includeDeckIDs []string, color pb.CardColor, excludeDeckIDs []string) ([]*pb.Card, int32, error) {
	var totalCount int32
	var whereClauses []string
	var args []interface{}
	argId := 1

	if len(includeDeckIDs) > 0 {
		placeholders := make([]string, len(includeDeckIDs))
		for i, id := range includeDeckIDs {
			placeholders[i] = "$" + strconv.Itoa(argId)
			args = append(args, id)
			argId++
		}
		whereClauses = append(whereClauses, "EXISTS (SELECT 1 FROM deck_cards dc WHERE dc.card_id = cards.id AND dc.deck_id IN ("+strings.Join(placeholders, ", ")+"))")
	}

	if len(excludeDeckIDs) > 0 {
		placeholders := make([]string, len(excludeDeckIDs))
		for i, id := range excludeDeckIDs {
			placeholders[i] = "$" + strconv.Itoa(argId)
			args = append(args, id)
			argId++
		}
		whereClauses = append(whereClauses, "NOT EXISTS (SELECT 1 FROM deck_cards dc WHERE dc.card_id = cards.id AND dc.deck_id IN ("+strings.Join(placeholders, ", ")+"))")
	}

	if ownerID != "" {
		whereClauses = append(whereClauses, "owner_id = $"+strconv.Itoa(argId))
		args = append(args, ownerID)
		argId++
	}

	if len(ids) > 0 {
		placeholders := make([]string, len(ids))
		for i, id := range ids {
			placeholders[i] = "$" + strconv.Itoa(argId)
			args = append(args, id)
			argId++
		}
		whereClauses = append(whereClauses, "cards.id IN ("+strings.Join(placeholders, ", ")+")")
	}

	if filter != "" {
		whereClauses = append(whereClauses, "text ILIKE $"+strconv.Itoa(argId))
		args = append(args, "%"+filter+"%")
		argId++
	}

	if color != pb.CardColor_CARD_COLOR_UNSPECIFIED {
		whereClauses = append(whereClauses, "cards.color = $"+strconv.Itoa(argId))
		args = append(args, int32(color))
		argId++
	}

	whereClause := ""
	if len(whereClauses) > 0 {
		whereClause = " WHERE " + strings.Join(whereClauses, " AND ")
	}

	countQuery := "SELECT COUNT(*) FROM cards" + whereClause
	dataQuery := "SELECT cards.id, cards.text, cards.color, cards.owner_id, cards.created_at, cards.updated_at FROM cards" + whereClause

	err := s.db.QueryRowContext(ctx, countQuery, args...).Scan(&totalCount)
	if err != nil {
		return nil, 0, err
	}

	orderClause := " ORDER BY cards.id"
	if orderBy != nil {
		column := "cards.id"
		switch orderBy.Field {
		case pb.CardOrderField_CARD_ORDER_FIELD_TEXT:
			column = "cards.text"
		case pb.CardOrderField_CARD_ORDER_FIELD_CREATED_AT:
			column = "cards.created_at"
		}
		dir := "ASC"
		if orderBy.Descending {
			dir = "DESC"
		}
		orderClause = fmt.Sprintf(" ORDER BY %s %s, cards.id ASC", column, dir)
	}
	dataQuery += orderClause

	if pageSize > 0 {
		dataQuery += " LIMIT $" + strconv.Itoa(argId) + " OFFSET $" + strconv.Itoa(argId+1)
		args = append(args, pageSize, (pageNumber-1)*pageSize)
	}

	rows, err := s.db.QueryContext(ctx, dataQuery, args...)
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

func (s *Storage) UpdateCard(ctx context.Context, card *pb.Card) error {
	res, err := s.db.ExecContext(ctx, "UPDATE cards SET text = $1, updated_at = CURRENT_TIMESTAMP WHERE id = $2", card.Text, card.Id)
	if err != nil {
		return err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return storage.ErrNotFound
	}
	return nil
}

// User operations
func (s *Storage) CreateUser(ctx context.Context, user *pb.User) error {
	var identifier sql.NullString
	if user.Identifier != "" {
		identifier = sql.NullString{String: user.Identifier, Valid: true}
	}
	_, err := s.db.ExecContext(ctx, "INSERT INTO users (id, name, identifier) VALUES ($1, $2, $3)", user.Id, user.Name, identifier)
	if err != nil {
		if strings.Contains(err.Error(), "unique_name_identifier") || strings.Contains(err.Error(), "users_pkey") {
			return storage.ErrAlreadyExists
		}
		return err
	}
	return nil
}

func (s *Storage) UpdateUser(ctx context.Context, user *pb.User) error {
	var identifier sql.NullString
	if user.Identifier != "" {
		identifier = sql.NullString{String: user.Identifier, Valid: true}
	}
	res, err := s.db.ExecContext(ctx, "UPDATE users SET name = $2, identifier = $3 WHERE id = $1", user.Id, user.Name, identifier)
	if err != nil {
		if strings.Contains(err.Error(), "unique_name_identifier") {
			return storage.ErrAlreadyExists
		}
		return err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return storage.ErrNotFound
	}
	return nil
}

func (s *Storage) GetUser(ctx context.Context, id string) (*pb.User, error) {
	row := s.db.QueryRowContext(ctx, "SELECT id, name, identifier, created_at FROM users WHERE id = $1", id)
	var u pb.User
	var identifier sql.NullString
	var createdAt sql.NullTime
	if err := row.Scan(&u.Id, &u.Name, &identifier, &createdAt); err != nil {
		if err == sql.ErrNoRows {
			return nil, storage.ErrNotFound
		}
		return nil, err
	}
	if identifier.Valid {
		u.Identifier = identifier.String
		u.PendingCompletion = false
	} else {
		u.PendingCompletion = true
	}
	if createdAt.Valid {
		u.CreatedAt = timestamppb.New(createdAt.Time)
	}
	return &u, nil
}

func (s *Storage) GetUserByIdentifier(ctx context.Context, identifierStr string) (*pb.User, error) {
	row := s.db.QueryRowContext(ctx, "SELECT id, name, identifier, created_at FROM users WHERE identifier = $1", identifierStr)
	var u pb.User
	var identifier sql.NullString
	var createdAt sql.NullTime
	if err := row.Scan(&u.Id, &u.Name, &identifier, &createdAt); err != nil {
		if err == sql.ErrNoRows {
			return nil, storage.ErrNotFound
		}
		return nil, err
	}
	if identifier.Valid {
		u.Identifier = identifier.String
		u.PendingCompletion = false
	} else {
		u.PendingCompletion = true
	}
	if createdAt.Valid {
		u.CreatedAt = timestamppb.New(createdAt.Time)
	}
	return &u, nil
}

// Token operations
func (s *Storage) RevokeToken(ctx context.Context, token string, expiresAt time.Time) error {
	_, err := s.db.ExecContext(ctx, "INSERT INTO revoked_tokens (token, expires_at) VALUES ($1, $2) ON CONFLICT (token) DO UPDATE SET expires_at = EXCLUDED.expires_at", token, expiresAt)
	return err
}

func (s *Storage) IsTokenRevoked(ctx context.Context, token string) (bool, error) {
	var expiresAt time.Time
	err := s.db.QueryRowContext(ctx, "SELECT expires_at FROM revoked_tokens WHERE token = $1", token).Scan(&expiresAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		return false, err
	}
	// Let's delete it if it's already expired to clean up the DB
	if expiresAt.Before(time.Now()) {
		go s.db.ExecContext(context.Background(), "DELETE FROM revoked_tokens WHERE token = $1", token)
		return false, nil // Expired token doesn't matter, but technically the JWT validate will catch it anyway
	}
	return true, nil
}

// Session operations
func (s *Storage) CreateSession(ctx context.Context, session *pb.Session) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	_, err = tx.ExecContext(ctx, "INSERT INTO sessions (id, owner_id, name, created_at) VALUES ($1, $2, $3, $4)", session.Id, session.OwnerId, session.Name, session.CreatedAt.AsTime())
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
	row := s.db.QueryRowContext(ctx, "SELECT id, owner_id, name, created_at FROM sessions WHERE id = $1", id)
	var session pb.Session
	var createdAt sql.NullTime
	if err := row.Scan(&session.Id, &session.OwnerId, &session.Name, &createdAt); err != nil {
		if err == sql.ErrNoRows {
			return nil, storage.ErrNotFound
		}
		return nil, err
	}
	if createdAt.Valid {
		session.CreatedAt = timestamppb.New(createdAt.Time)
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

func (s *Storage) ListSessions(ctx context.Context, playerID string) ([]*pb.Session, error) {
	// Filter for sessions that:
	// 1. Are WAITING (state = 1)
	// 2. OR Are PLAYING/JUDGING (state = 2, 3) AND the player is already a participant
	query := `
		SELECT s.id, s.owner_id, s.name, s.created_at
		FROM sessions s
		JOIN games g ON s.id = g.session_id
		WHERE g.state = 1
		   OR (g.state IN (2, 3) AND EXISTS (
			   SELECT 1 FROM session_players sp 
			   WHERE sp.session_id = s.id AND sp.player_id = $1
		   ))
	`
	rows, err := s.db.QueryContext(ctx, query, playerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
 
	var sessions []*pb.Session
	for rows.Next() {
		var sID, oID, name string
		var createdAt sql.NullTime
		if err := rows.Scan(&sID, &oID, &name, &createdAt); err != nil {
			return nil, err
		}
		sess := &pb.Session{Id: sID, OwnerId: oID, Name: name}
		if createdAt.Valid {
			sess.CreatedAt = timestamppb.New(createdAt.Time)
		}
		sessions = append(sessions, sess)
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
		contributorID := deck.CardContributorIds[cardID]
		if contributorID == "" {
			contributorID = deck.OwnerId
		}
		_, err = tx.ExecContext(ctx, "INSERT INTO deck_cards (deck_id, card_id, contributor_id) VALUES ($1, $2, $3)", deck.Id, cardID, contributorID)
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
	rows, err = s.db.QueryContext(ctx, "SELECT card_id, contributor_id FROM deck_cards WHERE deck_id = $1", id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	if d.CardContributorIds == nil {
		d.CardContributorIds = make(map[string]string)
	}
	for rows.Next() {
		var cardID string
		var contributorID sql.NullString
		if err := rows.Scan(&cardID, &contributorID); err != nil {
			return nil, err
		}
		d.CardIds = append(d.CardIds, cardID)
		if contributorID.Valid {
			d.CardContributorIds[cardID] = contributorID.String
		}
	}

	// Fetch subscribers
	rows, err = s.db.QueryContext(ctx, "SELECT user_id FROM deck_subscriptions WHERE deck_id = $1", id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var userID string
		if err := rows.Scan(&userID); err != nil {
			return nil, err
		}
		d.Subscribers = append(d.Subscribers, userID)
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
		contributorID := deck.CardContributorIds[cardID]
		if contributorID == "" {
			contributorID = deck.OwnerId
		}
		_, err = tx.ExecContext(ctx, "INSERT INTO deck_cards (deck_id, card_id, contributor_id) VALUES ($1, $2, $3)", deck.Id, cardID, contributorID)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (s *Storage) ListDecks(ctx context.Context, ownerID string) ([]*pb.Deck, error) {
	var rows *sql.Rows
	var err error
	if ownerID == "" {
		rows, err = s.db.QueryContext(ctx, "SELECT id, name, owner_id, created_at, updated_at FROM decks")
	} else {
		query := `
			SELECT id, name, owner_id, created_at, updated_at 
			FROM decks 
			WHERE owner_id = $1 
			   OR EXISTS (SELECT 1 FROM deck_contributors dc WHERE dc.deck_id = decks.id AND dc.user_id = $1)
			   OR EXISTS (SELECT 1 FROM deck_subscriptions ds WHERE ds.deck_id = decks.id AND ds.user_id = $1)
		`
		rows, err = s.db.QueryContext(ctx, query, ownerID)
	}
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
		cardRows, err := s.db.QueryContext(ctx, "SELECT card_id, contributor_id FROM deck_cards WHERE deck_id = $1", d.Id)
		if err != nil {
			return nil, err
		}
		if d.CardContributorIds == nil {
			d.CardContributorIds = make(map[string]string)
		}
		for cardRows.Next() {
			var cardID string
			var contributorID sql.NullString
			if err := cardRows.Scan(&cardID, &contributorID); err != nil {
				cardRows.Close()
				return nil, err
			}
			d.CardIds = append(d.CardIds, cardID)
			if contributorID.Valid {
				d.CardContributorIds[cardID] = contributorID.String
			}
		}
		cardRows.Close()

		// Subscribers
		subRows, err := s.db.QueryContext(ctx, "SELECT user_id FROM deck_subscriptions WHERE deck_id = $1", d.Id)
		if err != nil {
			return nil, err
		}
		for subRows.Next() {
			var userID string
			if err := subRows.Scan(&userID); err != nil {
				subRows.Close()
				return nil, err
			}
			d.Subscribers = append(d.Subscribers, userID)
		}
		subRows.Close()
	}

	return decks, nil
}

func (s *Storage) SubscribeToDeck(ctx context.Context, deckID, userID string) error {
	_, err := s.db.ExecContext(ctx, "INSERT INTO deck_subscriptions (deck_id, user_id) VALUES ($1, $2) ON CONFLICT DO NOTHING", deckID, userID)
	return err
}

func (s *Storage) UnsubscribeFromDeck(ctx context.Context, deckID, userID string) error {
	res, err := s.db.ExecContext(ctx, "DELETE FROM deck_subscriptions WHERE deck_id = $1 AND user_id = $2", deckID, userID)
	if err != nil {
		return err
	}
	rows, _ := res.RowsAffected()
	if rows == 0 {
		return storage.ErrNotFound
	}
	return nil
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

func (s *Storage) CountCardsByOwner(ctx context.Context, ownerID string) (int32, error) {
	var count int32
	err := s.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM cards WHERE owner_id = $1", ownerID).Scan(&count)
	return count, err
}

func (s *Storage) CountDecksByOwner(ctx context.Context, ownerID string) (int32, error) {
	var count int32
	err := s.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM decks WHERE owner_id = $1", ownerID).Scan(&count)
	return count, err
}

func (s *Storage) CountSessionsByOwner(ctx context.Context, ownerID string) (int32, error) {
	var count int32
	err := s.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM sessions WHERE owner_id = $1", ownerID).Scan(&count)
	return count, err
}

// Ensure Storage implements storage.Storage at compile time.
var _ storage.Storage = (*Storage)(nil)
