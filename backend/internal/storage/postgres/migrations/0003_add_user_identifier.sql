-- 0003_add_user_identifier.sql

-- Add identifier column
ALTER TABLE users ADD COLUMN identifier TEXT;

-- For existing users, we could generate random identifiers, 
-- but since this is a greenfield/development project, 
-- we can just leave them null or set a default.
-- Let's set a dummy one for now to avoid nulls if we want to enforce it later.
UPDATE users SET identifier = '00000000' WHERE identifier IS NULL;

-- Make it NOT NULL
ALTER TABLE users ALTER COLUMN identifier SET NOT NULL;

-- Add unique constraint on (name, identifier)
ALTER TABLE users ADD CONSTRAINT unique_name_identifier UNIQUE (name, identifier);
