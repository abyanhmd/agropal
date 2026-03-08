package user

import (
	"agropal/internal/database"
	"agropal/internal/domain/shared"
	"context"
	"time"

	"github.com/google/uuid"
)

type Repository interface {
	Create(ctx context.Context, q database.Querier, user User) error
	FindByID(ctx context.Context, q database.Querier, userID uuid.UUID) (User, error)
	List(ctx context.Context, q database.Querier) ([]User, error)
	Update(ctx context.Context, q database.Querier, user User) error
	UpdatePassword(ctx context.Context, q database.Querier, userID uuid.UUID, password string) error

	SoftDelete(ctx context.Context, q database.Querier, userID uuid.UUID, deletedAt time.Time) error
	Delete(ctx context.Context, q database.Querier, userID uuid.UUID) error

	ListByAsset(ctx context.Context, q database.Querier, assetType shared.AssetType, assetID uuid.UUID) ([]User, error)
}
