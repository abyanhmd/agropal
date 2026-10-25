package auth

import (
	"agropal/internal/infra/database"
	"context"
	"time"

	"github.com/google/uuid"
)

type Repository interface {
	CreateRefreshToken(ctx context.Context, q database.Querier, refreshToken RefreshToken) error
	GetRefreshToken(ctx context.Context, q database.Querier, refreshTokenID string) (RefreshToken, error)
	DeleteRefreshToken(ctx context.Context, q database.Querier, refreshTokenID string) error
	InvalidateSessions(ctx context.Context, q database.Querier, userID uuid.UUID) error
	DeleteExpiredRefreshTokens(ctx context.Context, q database.Querier, now time.Time) error
}
