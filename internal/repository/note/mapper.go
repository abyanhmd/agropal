package note

import (
	dNote "agropal/internal/domain/note"
	"agropal/internal/domain/shared"
)

func ToNoteModel(domain dNote.Note) NoteModel {
	return NoteModel{
		ID:          domain.ID,
		CreatedAt:   domain.CreatedAt,
		Type:        string(domain.Type),
		AssetType:   string(domain.AssetType),
		AssetID:     domain.AssetID,
		Field:       domain.Field,
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
		Field:       n.Field,
		AuthorID:    n.AuthorID,
		AuthorName:  n.AuthorName,
		IsSystemLog: n.IsSystemLog,
	}
}
