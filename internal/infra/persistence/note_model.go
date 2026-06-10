package persistence

import (
	dNote "agropal/internal/domain/note"
	"agropal/internal/domain/shared"
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

func ToNoteModel(domain dNote.Note) NoteModel {
	return NoteModel{
		ID:          domain.ID,
		CreatedAt:   domain.CreatedAt,
		Type:        string(domain.Type),
		AssetType:   string(domain.AssetType),
		AssetID:     domain.AssetID,
		Field:       domain.UpdateField,
		AuthorID:    domain.AuthorID,
		AuthorName:  domain.AuthorName,
		IsSystemLog: domain.IsSystemLog,
	}
}

func (n NoteModel) ToDomain() dNote.Note {
	return dNote.Note{
		ID:          n.ID,
		CreatedAt:   n.CreatedAt,
		Type:        dNote.NoteType(n.Type),
		AssetType:   shared.AssetType(n.AssetType),
		AssetID:     n.AssetID,
		UpdateField: n.Field,
		AuthorID:    n.AuthorID,
		AuthorName:  n.AuthorName,
		IsSystemLog: n.IsSystemLog,
	}
}
