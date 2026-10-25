// Package note is a write-only audit trail. Users, products and posts record a note when they are
// created, changed or removed, and nothing reads the notes back through the API; that is deliberate
// for now. Reading them would take an app/note service and an api/note router.
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

var (
	SuccessCreated shared.SuccessKey = "NOTE.CREATE_SUCCESS"
	SuccessUpdated shared.SuccessKey = "NOTE.UPDATE_SUCCESS"
	SuccessDeleted shared.SuccessKey = "NOTE.DELETE_SUCCESS"
)
