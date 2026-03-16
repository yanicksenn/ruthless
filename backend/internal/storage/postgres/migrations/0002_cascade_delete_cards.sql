-- 0002_cascade_delete_cards.sql

ALTER TABLE deck_cards
DROP CONSTRAINT deck_cards_card_id_fkey,
ADD CONSTRAINT deck_cards_card_id_fkey
    FOREIGN KEY (card_id)
    REFERENCES cards(id)
    ON DELETE CASCADE;
