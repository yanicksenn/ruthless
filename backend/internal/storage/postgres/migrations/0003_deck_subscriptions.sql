-- 0003_deck_subscriptions.sql

CREATE TABLE IF NOT EXISTS deck_subscriptions (
    deck_id TEXT NOT NULL REFERENCES decks(id) ON DELETE CASCADE,
    user_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (deck_id, user_id)
);
