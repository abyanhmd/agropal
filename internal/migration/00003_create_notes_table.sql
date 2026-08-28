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

CREATE INDEX IF NOT EXISTS idx_notes_asset ON notes (asset_type, asset_id);
CREATE INDEX IF NOT EXISTS idx_notes_author_id ON notes (author_id);

-- +goose Down
DROP INDEX IF EXISTS idx_notes_author_id;

DROP INDEX IF EXISTS idx_notes_asset;

DROP TABLE IF EXISTS notes;
