package note

import (
	"agropal/internal/database"
	"context"

	"github.com/google/uuid"
)

type Repository interface {
	Create(ctx context.Context, q database.Querier, note Note) error
	List(ctx context.Context, q database.Querier, noteType NoteType, assetIDs []uuid.UUID, fields map[string]string) ([]Note, error)
	Delete(ctx context.Context, q database.Querier, noteIDs []uuid.UUID) error
}
