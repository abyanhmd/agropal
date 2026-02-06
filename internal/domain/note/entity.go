package note

import (
	"agropal/internal/domain/shared"
	"time"

	"github.com/google/uuid"
)

type Note struct {
	NoteID    uuid.UUID
	DateAdded time.Time
	NoteType  NoteType
	AssetType shared.AssetType
	AssetID   uuid.UUID
}

type NoteType string

const (
	NoteCreate     NoteType = "create"
	NoteUpdate     NoteType = "update"
	NoteSoftDelete NoteType = "soft_delete"
	NoteHardDelete NoteType = "hard_delete"
	NoteMerge      NoteType = "merge"
)
