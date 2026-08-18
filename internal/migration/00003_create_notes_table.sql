-- +goose Up
CREATE TABLE IF NOT EXISTS notes (
    note_id UUID PRIMARY KEY,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    note_type VARCHAR(50) NOT NULL,
    asset_type VARCHAR(50) NOT NULL,
    asset_id UUID NOT NULL,
    update_field VARCHAR(50),
    old_value TEXT,
    new_value TEXT,
    author_id UUID NOT NULL REFERENCES users(user_id),
    is_system_log BOOLEAN NOT NULL DEFAULT FALSE
);

-- +goose Down
DROP TABLE IF EXISTS notes;
