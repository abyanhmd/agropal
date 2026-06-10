package user

import (
	"agropal/internal/domain/shared"
	"agropal/internal/infra/database"
	"context"
	"time"

	"github.com/google/uuid"
)

type Repository interface {
	CreateUser(ctx context.Context, q database.Querier, user User) error
	FindUserByID(ctx context.Context, q database.Querier, userID uuid.UUID) (User, error)
	FindUserByEmail(ctx context.Context, q database.Querier, email string) (User, error)
	ListUsers(ctx context.Context, q database.Querier) ([]User, error)
	UpdateUser(ctx context.Context, q database.Querier, user User) error
	UpdateUserPassword(ctx context.Context, q database.Querier, userID uuid.UUID, password string) error

	SoftDeleteUser(ctx context.Context, q database.Querier, userID uuid.UUID, deletedAt time.Time) error
	DeleteUser(ctx context.Context, q database.Querier, userID uuid.UUID) error

	ListUsersByAsset(ctx context.Context, q database.Querier, assetType shared.AssetType, assetID uuid.UUID) ([]User, error)
}
