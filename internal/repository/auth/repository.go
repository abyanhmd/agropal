package auth

import (
	"agropal/internal/database"
	dAuth "agropal/internal/domain/auth"
	"agropal/internal/utils"
	"context"
)

type RepositoryImpl struct {
	Logger *utils.Logger
}

func NewRepository(logger *utils.Logger) dAuth.Repository {
	return &RepositoryImpl{Logger: logger}
}

func (r *RepositoryImpl) CreateRefreshToken(ctx context.Context, q database.Querier, refreshToken dAuth.RefreshToken) (string, error) {
	//TODO implement me
	panic("implement me")
}

func (r *RepositoryImpl) GetRefreshToken(ctx context.Context, q database.Querier, refreshTokenID string) (dAuth.RefreshToken, error) {
	//TODO implement me
	panic("implement me")
}

func (r *RepositoryImpl) DeleteRefreshToken(ctx context.Context, q database.Querier, refreshTokenID string) error {
	//TODO implement me
	panic("implement me")
}
