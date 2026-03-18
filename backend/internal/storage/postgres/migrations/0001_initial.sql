-- 0001_initial.sql

CREATE TABLE IF NOT EXISTS users (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    identifier TEXT,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT unique_name_identifier UNIQUE (name, identifier)
);

CREATE TABLE IF NOT EXISTS revoked_tokens (
    token TEXT PRIMARY KEY,
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL
);

CREATE TABLE IF NOT EXISTS cards (
    id TEXT PRIMARY KEY,
    text TEXT NOT NULL,
    color INTEGER NOT NULL,
    owner_id TEXT REFERENCES users(id),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE
);

CREATE TABLE IF NOT EXISTS decks (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    owner_id TEXT NOT NULL REFERENCES users(id),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE
);

CREATE TABLE IF NOT EXISTS deck_contributors (
    deck_id TEXT NOT NULL REFERENCES decks(id),
    user_id TEXT NOT NULL REFERENCES users(id),
    PRIMARY KEY (deck_id, user_id)
);

CREATE TABLE IF NOT EXISTS deck_cards (
    deck_id TEXT NOT NULL REFERENCES decks(id),
    card_id TEXT NOT NULL REFERENCES cards(id) ON DELETE CASCADE,
    PRIMARY KEY (deck_id, card_id)
);

CREATE TABLE IF NOT EXISTS sessions (
    id TEXT PRIMARY KEY,
    owner_id TEXT NOT NULL REFERENCES users(id),
    name TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS session_players (
    session_id TEXT NOT NULL REFERENCES sessions(id),
    player_id TEXT NOT NULL REFERENCES users(id),
    PRIMARY KEY (session_id, player_id)
);

CREATE TABLE IF NOT EXISTS session_decks (
    session_id TEXT NOT NULL REFERENCES sessions(id),
    deck_id TEXT NOT NULL REFERENCES decks(id),
    PRIMARY KEY (session_id, deck_id)
);

CREATE TABLE IF NOT EXISTS games (
    id TEXT PRIMARY KEY,
    session_id TEXT NOT NULL,
    state INTEGER NOT NULL,
    scores JSONB NOT NULL DEFAULT '{}',
    hidden_hands JSONB NOT NULL DEFAULT '{}',
    hidden_black_deck JSONB NOT NULL DEFAULT '[]',
    hidden_white_deck JSONB NOT NULL DEFAULT '[]',
    rounds JSONB NOT NULL DEFAULT '[]',
    players JSONB NOT NULL DEFAULT '[]',
    player_ids JSONB NOT NULL DEFAULT '[]',
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);
