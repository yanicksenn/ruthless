-- 0005_user_last_active.sql
ALTER TABLE users ADD COLUMN last_active_at TIMESTAMP WITH TIME ZONE;
