package note

import (
	dNote "agropal/internal/domain/note"
	"agropal/internal/domain/shared"
)

func DomainToModelNote(domain dNote.Note) ModelNote {
	return ModelNote{
		NoteID:    domain.NoteID,
		DateAdded: domain.DateAdded,
		NoteType:  string(domain.NoteType),
		AssetType: string(domain.AssetType),
		AssetID:   domain.AssetID,
	}
}

func (n ModelNote) ToDomain() dNote.Note {
	return dNote.Note{
		NoteID:    n.NoteID,
		DateAdded: n.DateAdded,
		NoteType:  dNote.NoteType(n.NoteType),
		AssetType: shared.AssetType(n.AssetType),
		AssetID:   n.AssetID,
	}
}
