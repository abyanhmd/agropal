package note

import (
	"agropal/internal/domain/shared"
	"time"

	"github.com/google/uuid"
)

type Note struct {
	ID          uuid.UUID
	CreatedAt   time.Time
	Type        NoteType
	AssetType   shared.AssetType
	AssetID     uuid.UUID
	Field       string
	AuthorID    uuid.UUID
	AuthorName  string
	IsSystemLog bool
}

type NoteType string

const (
	NoteCreate     NoteType = "CREATE"
	NoteUpdate     NoteType = "UPDATE"
	NoteSoftDelete NoteType = "SOFT_DELETE"
	NoteHardDelete NoteType = "HARD_DELETE"
	NoteMerge      NoteType = "MERGE"
)
