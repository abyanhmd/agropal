package note

import (
	"time"

	"github.com/google/uuid"
)

type ModelNote struct {
	NoteID    uuid.UUID `db:"note_id"`
	DateAdded time.Time `db:"date_added"`
	NoteType  string    `db:"note_type"`
	AssetType string    `db:"asset_type"`
	AssetID   uuid.UUID `db:"asset_id"`
}
