-- +goose Up
-- Users are soft-deleted, but email and username were unique across every row, so a deleted account
-- kept its email and username forever and nobody could sign up with them again. Uniqueness now only
-- applies to accounts that have not been deleted. The index names keep "email" and "username" in them,
-- which is how a clash is reported against the right field.
ALTER TABLE users DROP CONSTRAINT IF EXISTS users_email_key;
ALTER TABLE users DROP CONSTRAINT IF EXISTS users_username_key;

CREATE UNIQUE INDEX IF NOT EXISTS idx_users_email_active ON users (email) WHERE deleted_at IS NULL;
CREATE UNIQUE INDEX IF NOT EXISTS idx_users_username_active ON users (username) WHERE deleted_at IS NULL;

-- +goose Down
DROP INDEX IF EXISTS idx_users_username_active;
DROP INDEX IF EXISTS idx_users_email_active;

-- Fails if a deleted account and a live one now share an email or username; remove one of them first.
ALTER TABLE users ADD CONSTRAINT users_email_key UNIQUE (email);
ALTER TABLE users ADD CONSTRAINT users_username_key UNIQUE (username);
