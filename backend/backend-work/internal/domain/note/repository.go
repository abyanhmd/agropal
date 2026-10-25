package note

import (
	"agropal/internal/infra/database"
	"context"

	"github.com/google/uuid"
)

type Repository interface {
	CreateNote(ctx context.Context, q database.Querier, note Note) error
	CreateManyNotes(ctx context.Context, q database.Querier, note []Note) error
	ListNotes(ctx context.Context, q database.Querier, noteType NoteType, assetIDs []uuid.UUID, fields map[string]string) ([]Note, error)
	DeleteNotes(ctx context.Context, q database.Querier, noteIDs []uuid.UUID) error
}
