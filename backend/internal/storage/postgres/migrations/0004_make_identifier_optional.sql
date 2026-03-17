-- 0004_make_identifier_optional.sql

-- Make identifier column nullable
ALTER TABLE users ALTER COLUMN identifier DROP NOT NULL;

-- Set default value to NULL (though it is by default if nullable)
ALTER TABLE users ALTER COLUMN identifier SET DEFAULT NULL;
