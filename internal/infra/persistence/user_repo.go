package persistence

import (
	"agropal/internal/domain/shared"
	dUser "agropal/internal/domain/user"
	"agropal/internal/infra/database"
	"agropal/pkg"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

type userRepo struct {
	logger *pkg.Logger
}

func NewUserRepository(l *pkg.Logger) dUser.Repository {
	return &userRepo{logger: l}
}

func (r *userRepo) CreateUser(ctx context.Context, q database.Querier, user dUser.User) error {
	query := `
		INSERT INTO users (
			user_id, name, email, username,
			avatar, password, role,
			created_at, updated_at, deleted_at, last_login_at
		) VALUES (
			:user_id, :name, :email, :username,
			:avatar, :password, :role, 
			:created_at, :updated_at, :deleted_at, :last_login_at
		);`

	userModel := ToUserModel(user)
	if _, err := q.NamedExecContext(ctx, query, userModel); err != nil {
		r.logger.Error(ctx, shared.AssetUser, err.Error())
		return err
	}

	return nil
}

func (r *userRepo) FindUserByID(ctx context.Context, q database.Querier, userID uuid.UUID) (dUser.User, error) {
	query := `
		SELECT *
		FROM users
		WHERE user_id = $1
			AND deleted_at IS NULL;`

	var userModel UserModel
	if err := q.GetContext(ctx, &userModel, query, userID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return dUser.User{}, dUser.ErrNotFound
		}
		r.logger.Error(ctx, shared.AssetUser, err.Error())
		return dUser.User{}, err
	}

	return userModel.ToDomain(), nil
}

func (r *userRepo) FindUserByEmail(ctx context.Context, q database.Querier, email string) (dUser.User, error) {
	query := `
		SELECT *
		FROM users
		WHERE email = $1
			AND deleted_at IS NULL;`

	var userModel UserModel
	if err := q.GetContext(ctx, &userModel, query, email); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return dUser.User{}, dUser.ErrNotFound
		}
		r.logger.Error(ctx, shared.AssetUser, err.Error())
		return dUser.User{}, err
	}

	return userModel.ToDomain(), nil
}

func (r *userRepo) FindUserByUsername(ctx context.Context, q database.Querier, username string) (dUser.User, error) {
	query := `
		SELECT *
		FROM users
		WHERE username = $1
			AND deleted_at IS NULL;`

	var userModel UserModel
	if err := q.GetContext(ctx, &userModel, query, username); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return dUser.User{}, dUser.ErrNotFound
		}
		r.logger.Error(ctx, shared.AssetUser, err.Error())
		return dUser.User{}, err
	}

	return userModel.ToDomain(), nil
}

func (r *userRepo) ListUsers(ctx context.Context, q database.Querier) ([]dUser.User, error) {
	//TODO implement me
	panic("implement me")
}

func (r *userRepo) UpdateUser(ctx context.Context, q database.Querier, user dUser.User) error {
	query := `
		UPDATE users SET
			name = :name,
			avatar = :avatar
		WHERE user_id = :user_id;`

	userModel := ToUserModel(user)
	if _, err := q.NamedExecContext(ctx, query, userModel); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return dUser.ErrNotFound
		}
		r.logger.Error(ctx, shared.AssetUser, err.Error())
		return err
	}

	return nil
}

func (r *userRepo) UpdateUserPassword(ctx context.Context, q database.Querier, userID uuid.UUID, password string) error {
	updateQuery := `
		UPDATE users SET
			password = $1
		WHERE user_id = $2;`

	if _, err := q.ExecContext(ctx, updateQuery, password, userID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return dUser.ErrNotFound
		}
		r.logger.Error(ctx, shared.AssetUser, err.Error())
		return err
	}

	return nil
}

func (r *userRepo) SoftDeleteUser(ctx context.Context, q database.Querier, userID uuid.UUID, deletedAt time.Time) error {
	query := `UPDATE users SET deleted_at = $1 WHERE user_id = $2;`

	if _, err := q.ExecContext(ctx, query, deletedAt, userID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return dUser.ErrNotFound
		}
		r.logger.Error(ctx, shared.AssetUser, err.Error())
		return err
	}

	return nil
}

func (r *userRepo) DeleteUser(ctx context.Context, q database.Querier, userID uuid.UUID) error {
	query := `DELETE FROM users WHERE user_id = $1;`

	if _, err := q.ExecContext(ctx, query, userID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return dUser.ErrNotFound
		}
		r.logger.Error(ctx, shared.AssetUser, err.Error())
		return err
	}

	return nil
}

func (r *userRepo) ListUsersByAsset(ctx context.Context, q database.Querier, assetType shared.AssetType, assetID uuid.UUID) ([]dUser.User, error) {
	var getQuery string
	assetTypeStr := string(assetType)

	if assetID != uuid.Nil {
		getQuery = fmt.Sprintf(`SELECT * FROM users WHERE %s_id = $1;`, strings.ToLower(assetTypeStr))
	} else {
		getQuery = `SELECT * FROM users;`
	}

	var modelUsers []UserModel
	var domainUsers []dUser.User

	if assetID != uuid.Nil {
		if errQuery := q.GetContext(ctx, &modelUsers, getQuery, assetID); errQuery != nil {
			r.logger.Error(ctx, shared.AssetUser, errQuery.Error())
			return domainUsers, errQuery
		}
	} else {
		if errQuery := q.SelectContext(ctx, &modelUsers, getQuery); errQuery != nil {
			r.logger.Error(ctx, shared.AssetUser, errQuery.Error())
			return domainUsers, errQuery
		}
	}

	for i, modelUser := range modelUsers {
		domainUsers[i] = modelUser.ToDomain()
	}

	return domainUsers, nil
}
