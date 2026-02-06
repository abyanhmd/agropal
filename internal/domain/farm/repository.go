package farm

import (
	"context"

	"github.com/google/uuid"
)

type Repository interface {
	Create(ctx context.Context, farm Farm) error
	GetOne(ctx context.Context, farmID uuid.UUID) (Farm, error)
	GetMany(ctx context.Context, userID uuid.UUID) ([]Farm, error)
	Update(ctx context.Context, farm Farm) error
	Delete(ctx context.Context, farmID uuid.UUID) error
}
