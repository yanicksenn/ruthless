-- 0001_initial.sql

CREATE TABLE IF NOT EXISTS cards (
    id TEXT PRIMARY KEY,
    text TEXT NOT NULL,
    blanks INTEGER NOT NULL
);

CREATE TABLE IF NOT EXISTS users (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS decks (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    owner_id TEXT NOT NULL,
    contributors JSONB NOT NULL DEFAULT '[]',
    card_ids JSONB NOT NULL DEFAULT '[]'
);

CREATE TABLE IF NOT EXISTS sessions (
    id TEXT PRIMARY KEY,
    player_ids JSONB NOT NULL DEFAULT '[]',
    deck_ids JSONB NOT NULL DEFAULT '[]'
);

CREATE TABLE IF NOT EXISTS games (
    id TEXT PRIMARY KEY,
    session_id TEXT NOT NULL,
    state INTEGER NOT NULL,
    scores JSONB NOT NULL DEFAULT '{}',
    hidden_hands JSONB NOT NULL DEFAULT '{}',
    hidden_black_deck JSONB NOT NULL DEFAULT '[]',
    hidden_white_deck JSONB NOT NULL DEFAULT '[]',
    rounds JSONB NOT NULL DEFAULT '[]'
);
