CREATE TABLE IF NOT EXISTS session_invitations (
    id TEXT PRIMARY KEY,
    session_id TEXT NOT NULL REFERENCES sessions(id) ON DELETE CASCADE,
    sender_id TEXT NOT NULL REFERENCES users(id),
    receiver_id TEXT NOT NULL REFERENCES users(id),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(session_id, receiver_id)
);

CREATE INDEX idx_session_invitations_receiver_id ON session_invitations(receiver_id);
