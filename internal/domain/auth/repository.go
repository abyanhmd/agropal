package auth

import (
	"agropal/internal/infra/database"
	"context"
)

type Repository interface {
	CreateRefreshToken(ctx context.Context, q database.Querier, refreshToken RefreshToken) error
	GetRefreshToken(ctx context.Context, q database.Querier, refreshTokenID string) (RefreshToken, error)
	DeleteRefreshToken(ctx context.Context, q database.Querier, refreshTokenID string) error
}
