-- +goose Up
CREATE TABLE IF NOT EXISTS farms (
    farm_id UUID PRIMARY KEY,
    name TEXT NOT NULL,
    user_id UUID NOT NULL REFERENCES users(user_id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_farms_user_id ON farms (user_id);

-- +goose Down
DROP INDEX IF EXISTS idx_farms_user_id;

DROP TABLE IF EXISTS farms;
