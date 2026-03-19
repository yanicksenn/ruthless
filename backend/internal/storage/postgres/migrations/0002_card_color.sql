-- 0002_card_color.sql

-- Drop the existing color column and recreate it as a generated column.
-- We can do this safely because the color is fully derived from the text.
-- This combined migration fix uses the regex operator '~' to correctly identify the literal '___' sequence.
ALTER TABLE cards DROP COLUMN color;
ALTER TABLE cards ADD COLUMN color INTEGER GENERATED ALWAYS AS (
    CASE WHEN text ~ '___' THEN 1 ELSE 2 END
) STORED;
