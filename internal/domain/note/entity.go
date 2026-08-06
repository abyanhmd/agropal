package note

import (
	"agropal/internal/domain/shared"
	"time"

	"github.com/google/uuid"
)

type Note struct {
	ID         uuid.UUID
	CreatedAt  time.Time
	Type       NoteType
	AssetType  shared.AssetType
	AssetID    uuid.UUID
	AuthorID   uuid.UUID
	AuthorName string

	UpdateField *string

	OldValue *string
	NewValue *string

	IsSystemLog bool
}

type NoteType string

const (
	TypeCreate  NoteType = "create"
	TypeUpdate  NoteType = "update"
	TypeArchive NoteType = "archive"
	TypeDelete  NoteType = "delete"
	TypeMerge   NoteType = "merge"
)
