package persistence

import (
	"time"
	dAuth "agropal/internal/domain/auth"
	"agropal/internal/domain/shared"
	"agropal/internal/infra/database"
	"agropal/pkg"
	"context"

	"github.com/google/uuid"
)

type authRepo struct {
	logger *pkg.Logger
}

func NewAuthRepository(l *pkg.Logger) dAuth.Repository {
	return &authRepo{logger: l}
}

func (r *authRepo) CreateRefreshToken(ctx context.Context, q database.Querier, token dAuth.RefreshToken) error {
	query := `
		INSERT INTO refresh_tokens (
			refresh_token_id, user_id, expires_at,
			token, user_agent,
			ip_address
		) VALUES (
			:refresh_token_id, :user_id, :expires_at,
			:token, :user_agent,
			:ip_address
		);`

	tokenModel := ToRefreshTokenModel(token)
	if _, err := q.NamedExecContext(ctx, query, tokenModel); err != nil {
		r.logger.Error(ctx, shared.AssetFarm, err.Error())
		return err
	}

	return nil
}

func (r *authRepo) GetRefreshToken(ctx context.Context, q database.Querier, refreshTokenID string) (dAuth.RefreshToken, error) {
	query := `SELECT * FROM refresh_tokens WHERE refresh_token_id = $1 LIMIT 1;`

	var tokenModel RefreshTokenModel
	if err := q.GetContext(ctx, &tokenModel, query, refreshTokenID); err != nil {
		r.logger.Error(ctx, shared.AssetFarm, err.Error())
		return dAuth.RefreshToken{}, err
	}

	return tokenModel.ToDomain(), nil
}

func (r *authRepo) DeleteRefreshToken(ctx context.Context, q database.Querier, refreshTokenID string) error {
	query := `DELETE FROM refresh_tokens WHERE refresh_token_id = $1;`

	if _, err := q.ExecContext(ctx, query, refreshTokenID); err != nil {
		r.logger.Error(ctx, shared.AssetFarm, err.Error())
		return err
	}

	return nil
}

func (r *authRepo) InvalidateSessions(ctx context.Context, q database.Querier, userID uuid.UUID) error {
	query := `DELETE FROM refresh_tokens WHERE user_id = $1;`

	if _, err := q.ExecContext(ctx, query, userID); err != nil {
		r.logger.Error(ctx, shared.AssetFarm, err.Error())
		return err
	}

	return nil
}

func (r *authRepo) DeleteExpiredRefreshTokens(ctx context.Context, q database.Querier, now time.Time) error {
	query := `DELETE FROM refresh_tokens WHERE expires_at < $1;`

	if _, err := q.ExecContext(ctx, query, now); err != nil {
		r.logger.Error(ctx, shared.AssetFarm, err.Error())
		return err
	}

	return nil
}
