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
	"github.com/lib/pq"
)

// uniqueViolationField maps a unique-constraint violation on the users
// table to the domain error the caller should see, by inspecting which
// column the constraint covers. Postgres names an inline `UNIQUE` column
// constraint "<table>_<column>_key" by default, so a substring match on the
// column name is enough without hardcoding the exact constraint name.
func uniqueViolationField(pqErr *pq.Error) error {
	switch {
	case strings.Contains(pqErr.Constraint, "email"):
		return dUser.ErrEmailExists
	case strings.Contains(pqErr.Constraint, "username"):
		return dUser.ErrUsernameExists
	default:
		return nil
	}
}

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
			created_at, deleted_at, last_login_at
		) VALUES (
			:user_id, :name, :email, :username,
			:avatar, :password, :role,
			:created_at, :deleted_at, :last_login_at
		);`

	userModel := ToUserModel(user)
	if _, err := q.NamedExecContext(ctx, query, userModel); err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) && pqErr.Code == "23505" {
			if domainErr := uniqueViolationField(pqErr); domainErr != nil {
				return domainErr
			}
		}
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
	query := `
		SELECT *
		FROM users
		WHERE deleted_at IS NULL
		ORDER BY created_at;`

	var userModels []UserModel
	if err := q.SelectContext(ctx, &userModels, query); err != nil {
		r.logger.Error(ctx, shared.AssetUser, err.Error())
		return nil, err
	}

	users := make([]dUser.User, len(userModels))
	for i, m := range userModels {
		users[i] = m.ToDomain()
	}

	return users, nil
}

func (r *userRepo) UpdateUser(ctx context.Context, q database.Querier, user dUser.User) error {
	query := `
		UPDATE users SET
			name = :name,
			avatar = :avatar
		WHERE user_id = :user_id;`

	userModel := ToUserModel(user)
	result, err := q.NamedExecContext(ctx, query, userModel)
	if err != nil {
		r.logger.Error(ctx, shared.AssetUser, err.Error())
		return err
	}
	if rows, _ := result.RowsAffected(); rows == 0 {
		return dUser.ErrNotFound
	}

	return nil
}

func (r *userRepo) UpdateUserPassword(ctx context.Context, q database.Querier, userID uuid.UUID, password string) error {
	updateQuery := `
		UPDATE users SET password = $1
		WHERE user_id = $2 AND deleted_at IS NULL;`

	result, err := q.ExecContext(ctx, updateQuery, password, userID)
	if err != nil {
		r.logger.Error(ctx, shared.AssetUser, err.Error())
		return err
	}
	if rows, _ := result.RowsAffected(); rows == 0 {
		return dUser.ErrNotFound
	}

	return nil
}

func (r *userRepo) VerifyUserEmail(ctx context.Context, q database.Querier, userID uuid.UUID, verifiedAt time.Time) error {
	query := `UPDATE users SET email_verified_at = $1 WHERE user_id = $2 AND deleted_at IS NULL;`

	result, err := q.ExecContext(ctx, query, verifiedAt, userID)
	if err != nil {
		r.logger.Error(ctx, shared.AssetUser, err.Error())
		return err
	}
	if rows, _ := result.RowsAffected(); rows == 0 {
		return dUser.ErrNotFound
	}

	return nil
}

func (r *userRepo) SoftDeleteUser(ctx context.Context, q database.Querier, userID uuid.UUID, deletedAt time.Time) error {
	query := `UPDATE users SET deleted_at = $1 WHERE user_id = $2 AND deleted_at IS NULL;`

	result, err := q.ExecContext(ctx, query, deletedAt, userID)
	if err != nil {
		r.logger.Error(ctx, shared.AssetUser, err.Error())
		return err
	}
	if rows, _ := result.RowsAffected(); rows == 0 {
		return dUser.ErrNotFound
	}

	return nil
}

func (r *userRepo) DeleteUser(ctx context.Context, q database.Querier, userID uuid.UUID) error {
	query := `DELETE FROM users WHERE user_id = $1;`

	result, err := q.ExecContext(ctx, query, userID)
	if err != nil {
		r.logger.Error(ctx, shared.AssetUser, err.Error())
		return err
	}
	if rows, _ := result.RowsAffected(); rows == 0 {
		return dUser.ErrNotFound
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

	if assetID != uuid.Nil {
		if errQuery := q.SelectContext(ctx, &modelUsers, getQuery, assetID); errQuery != nil {
			r.logger.Error(ctx, shared.AssetUser, errQuery.Error())
			return nil, errQuery
		}
	} else {
		if errQuery := q.SelectContext(ctx, &modelUsers, getQuery); errQuery != nil {
			r.logger.Error(ctx, shared.AssetUser, errQuery.Error())
			return nil, errQuery
		}
	}

	domainUsers := make([]dUser.User, len(modelUsers))
	for i, modelUser := range modelUsers {
		domainUsers[i] = modelUser.ToDomain()
	}

	return domainUsers, nil
}
