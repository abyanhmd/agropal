-- +goose Up
CREATE TABLE IF NOT EXISTS farms (
    farm_id UUID PRIMARY KEY,
    name TEXT NOT NULL,
    user_id UUID NOT NULL REFERENCES users(user_id)
);

-- +goose Down
DROP TABLE IF EXISTS farms;
