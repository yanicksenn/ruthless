package postgres

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
	pb "github.com/yanicksenn/ruthless/api/v1"
	"github.com/yanicksenn/ruthless/backend/internal/storage"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (s *Storage) CreateSessionInvitation(ctx context.Context, sessionID, senderID, receiverID string) error {
	id := uuid.New().String()
	_, err := s.db.ExecContext(ctx, "INSERT INTO session_invitations (id, session_id, sender_id, receiver_id) VALUES ($1, $2, $3, $4)", id, sessionID, senderID, receiverID)
	if err != nil {
		if isUniqueViolation(err) {
			return storage.ErrAlreadyExists
		}
		return err
	}
	return nil
}

func (s *Storage) GetSessionInvitation(ctx context.Context, id string) (*pb.SessionInvitation, error) {
	var invitation pb.SessionInvitation
	var session pb.Session
	var sender, receiver pb.Player
	var createdAt time.Time
	var sessionCreatedAt sql.NullTime

	query := `
		SELECT i.id, i.created_at,
		       s.id, s.owner_id, s.name, s.created_at,
		       u1.id, u1.name, u1.identifier,
		       u2.id, u2.name, u2.identifier
		FROM session_invitations i
		JOIN sessions s ON i.session_id = s.id
		JOIN users u1 ON i.sender_id = u1.id
		JOIN users u2 ON i.receiver_id = u2.id
		WHERE i.id = $1`

	err := s.db.QueryRowContext(ctx, query, id).Scan(
		&invitation.Id, &createdAt,
		&session.Id, &session.OwnerId, &session.Name, &sessionCreatedAt,
		&sender.Id, &sender.Name, &sender.Identifier,
		&receiver.Id, &receiver.Name, &receiver.Identifier)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, storage.ErrNotFound
		}
		return nil, err
	}

	if sessionCreatedAt.Valid {
		session.CreatedAt = timestamppb.New(sessionCreatedAt.Time)
	}

	// Fetch players
	pRows, err := s.db.QueryContext(ctx, "SELECT player_id FROM session_players WHERE session_id = $1", session.Id)
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
	dRows, err := s.db.QueryContext(ctx, "SELECT deck_id FROM session_decks WHERE session_id = $1", session.Id)
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

	invitation.Session = &session
	invitation.Sender = &sender
	invitation.Receiver = &receiver
	invitation.CreatedAt = timestamppb.New(createdAt)
	return &invitation, nil
}

func (s *Storage) DeleteSessionInvitation(ctx context.Context, id string) error {
	_, err := s.db.ExecContext(ctx, "DELETE FROM session_invitations WHERE id = $1", id)
	return err
}

func (s *Storage) ListSessionInvitations(ctx context.Context, receiverID string, pageSize, pageNumber int32) ([]*pb.SessionInvitation, int32, error) {
	var totalCount int32
	err := s.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM session_invitations WHERE receiver_id = $1", receiverID).Scan(&totalCount)
	if err != nil {
		return nil, 0, err
	}

	query := `
		SELECT i.id, i.created_at,
		       s.id, s.owner_id, s.name, s.created_at,
		       u1.id, u1.name, u1.identifier,
		       u2.id, u2.name, u2.identifier
		FROM session_invitations i
		JOIN sessions s ON i.session_id = s.id
		JOIN users u1 ON i.sender_id = u1.id
		JOIN users u2 ON i.receiver_id = u2.id
		WHERE i.receiver_id = $1
		ORDER BY i.created_at DESC`

	var args []interface{}
	args = append(args, receiverID)

	if pageSize > 0 {
		query += " LIMIT $2 OFFSET $3"
		args = append(args, pageSize, (pageNumber-1)*pageSize)
	}

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var invitations []*pb.SessionInvitation
	for rows.Next() {
		var invitation pb.SessionInvitation
		var session pb.Session
		var sender, receiver pb.Player
		var createdAt time.Time
		var sessionCreatedAt sql.NullTime

		err := rows.Scan(
			&invitation.Id, &createdAt,
			&session.Id, &session.OwnerId, &session.Name, &sessionCreatedAt,
			&sender.Id, &sender.Name, &sender.Identifier,
			&receiver.Id, &receiver.Name, &receiver.Identifier)
		if err != nil {
			return nil, 0, err
		}

		if sessionCreatedAt.Valid {
			session.CreatedAt = timestamppb.New(sessionCreatedAt.Time)
		}

		// We could fetch players/decks per session if needed, but it might be heavy. For list views, basic session info + players count is often enough, but let's fetch for completeness.
		pRows, err := s.db.QueryContext(ctx, "SELECT player_id FROM session_players WHERE session_id = $1", session.Id)
		if err == nil {
			for pRows.Next() {
				var pID string
				if pRows.Scan(&pID) == nil {
					session.PlayerIds = append(session.PlayerIds, pID)
				}
			}
			pRows.Close()
		}

		invitation.Session = &session
		invitation.Sender = &sender
		invitation.Receiver = &receiver
		invitation.CreatedAt = timestamppb.New(createdAt)
		invitations = append(invitations, &invitation)
	}
	return invitations, totalCount, nil
}

func (s *Storage) DeleteUnansweredSessionInvitations(ctx context.Context, sessionID string) error {
	_, err := s.db.ExecContext(ctx, "DELETE FROM session_invitations WHERE session_id = $1", sessionID)
	return err
}
