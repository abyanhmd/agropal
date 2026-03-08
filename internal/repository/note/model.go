package note

import (
	"time"

	"github.com/google/uuid"
)

type NoteModel struct {
	ID          uuid.UUID `db:"note_id"`
	CreatedAt   time.Time `db:"created_at"`
	Type        string    `db:"type"`
	AssetType   string    `db:"asset_type"`
	AssetID     uuid.UUID `db:"asset_id"`
	Field       string    `db:"field"`
	AuthorID    uuid.UUID `db:"author_id"`
	AuthorName  string    `db:"author_name"`
	IsSystemLog bool      `db:"is_system_log"`
}
