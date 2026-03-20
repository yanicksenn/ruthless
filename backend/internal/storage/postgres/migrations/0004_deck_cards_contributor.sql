-- 0004_deck_cards_contributor.sql

-- Add contributor_id column to deck_cards
ALTER TABLE deck_cards ADD COLUMN contributor_id TEXT REFERENCES users(id);

-- Backfill contributor_id using the card's owner_id
UPDATE deck_cards
SET contributor_id = cards.owner_id
FROM cards
WHERE deck_cards.card_id = cards.id;

-- Make contributor_id NOT NULL if we want to enforce it for all future entries
-- ALTER TABLE deck_cards ALTER COLUMN contributor_id SET NOT NULL;
