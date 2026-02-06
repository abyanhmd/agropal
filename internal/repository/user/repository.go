package user

import (
	dNote "agropal/internal/domain/note"
	"agropal/internal/domain/shared"
	dUser "agropal/internal/domain/user"
	rNote "agropal/internal/repository/note"
	"agropal/internal/utils"
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type RepositoryImpl struct {
	DB     *sqlx.DB
	Logger *utils.Logger
}

func (r *RepositoryImpl) Create(ctx context.Context, user dUser.User) error {
	insertQuery := `INSERT INTO users (
		user_id, name, email,
		username, avatar,
		password, role
	) VALUES (
		:user_id, :name, :email,
		:username, :avatar,
		:password, :role
	);`

	modelUser := DomainToModelUser(user)

	if _, errInsert := r.DB.NamedExecContext(ctx, insertQuery, modelUser); errInsert != nil {
		r.Logger.Error(ctx, shared.AssetUser, errInsert.Error())
		return errInsert
	}

	return nil
}

func (r *RepositoryImpl) GetOne(ctx context.Context, userID uuid.UUID) (dUser.User, error) {
	getQuery := `SELECT u.*
		LEFT JOIN notes n 
			ON n.asset_id = u.user_id 
			AND n.asset_type = 'user'
			AND n.note_type NOT LIKE 'delete'; 
		FROM users u
		WHERE u.user_id = $1;`

	var modelUser ModelUser
	if errQuery := r.DB.GetContext(ctx, &modelUser, getQuery, userID); errQuery != nil {
		r.Logger.Error(ctx, shared.AssetUser, errQuery.Error())
		return dUser.User{}, errQuery
	}

	return modelUser.ToDomain(), nil
}

func (r *RepositoryImpl) Update(ctx context.Context, user dUser.User) error {
	updateQuery := `UPDATE users SET
		name = :name,
		avatar = :avatar,
		WHERE user_id = :user_id;`

	modelUser := DomainToModelUser(user)

	if _, errUpdate := r.DB.NamedExecContext(ctx, updateQuery, modelUser); errUpdate != nil {
		r.Logger.Error(ctx, shared.AssetUser, errUpdate.Error())
		return errUpdate
	}

	return nil
}

func (r *RepositoryImpl) UpdatePassword(ctx context.Context, userID uuid.UUID, password string) error {
	updateQuery := `UPDATE users SET
		password = $1
		WHERE user_id = $2;`

	if _, errUpdate := r.DB.ExecContext(ctx, updateQuery, password, userID); errUpdate != nil {
		r.Logger.Error(ctx, shared.AssetUser, errUpdate.Error())
		return errUpdate
	}

	return nil
}

func (r *RepositoryImpl) SoftDelete(ctx context.Context, deleteNote dNote.Note) error {
	insertNoteQuery := `INSERT INTO notes (
		note_id, date_added, note_type, asset_type, asset_id
	) VALUES (
		:note_id, :date_added, :note_type, :asset_type, :asset_id
	);`

	modelNote := rNote.DomainToModelNote(deleteNote)
	if _, errInsertNote := r.DB.NamedExecContext(ctx, insertNoteQuery, modelNote); errInsertNote != nil {
		r.Logger.Error(ctx, shared.AssetUser, errInsertNote.Error())
		return errInsertNote
	}

	return nil
}

func (r *RepositoryImpl) HardDelete(ctx context.Context, userID uuid.UUID, deleteNote dNote.Note) (err error) {
	tx, err := r.DB.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() {
		if p := recover(); p != nil {
			tx.Rollback()
			panic(p)
		} else if err != nil {
			tx.Rollback()
		}
	}()

	deleteUserQuery := `DELETE FROM users WHERE user_id = $1;`
	insertNoteQuery := `INSERT INTO notes (
		note_id, date_added, note_type, asset_type, asset_id
	) VALUES (
		:note_id, :date_added, :note_type, :asset_type, :asset_id
	);`

	if _, errDelete := r.DB.ExecContext(ctx, deleteUserQuery, userID); errDelete != nil {
		r.Logger.Error(ctx, shared.AssetUser, errDelete.Error())
		return errDelete
	}

	modelNote := rNote.DomainToModelNote(deleteNote)
	if _, errInsertNote := r.DB.NamedExecContext(ctx, insertNoteQuery, modelNote); errInsertNote != nil {
		r.Logger.Error(ctx, shared.AssetUser, errInsertNote.Error())
		return errInsertNote
	}

	return nil
}

func (r *RepositoryImpl) GetByAdmin(ctx context.Context, assetType shared.AssetType, assetID uuid.UUID) ([]dUser.User, error) {
	var getQuery string
	assetTypeStr := string(assetType)

	if assetID != uuid.Nil {
		getQuery = fmt.Sprintf(`SELECT * FROM users WHERE %s_id = $1;`, strings.ToLower(assetTypeStr))
	} else {
		getQuery = `SELECT * FROM users;`
	}

	var modelUsers []ModelUser
	var domainUsers []dUser.User

	if assetID != uuid.Nil {
		if errQuery := r.DB.GetContext(ctx, &modelUsers, getQuery, assetID); errQuery != nil {
			r.Logger.Error(ctx, shared.AssetUser, errQuery.Error())
			return domainUsers, errQuery
		}
	} else {
		if errQuery := r.DB.SelectContext(ctx, &modelUsers, getQuery); errQuery != nil {
			r.Logger.Error(ctx, shared.AssetUser, errQuery.Error())
			return domainUsers, errQuery
		}
	}

	for i, modelUser := range modelUsers {
		domainUsers[i] = modelUser.ToDomain()
	}

	return domainUsers, nil
}

func NewRepository(db *sqlx.DB, logger *utils.Logger) dUser.Repository {
	return &RepositoryImpl{
		DB:     db,
		Logger: logger,
	}
}
