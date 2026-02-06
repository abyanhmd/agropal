package farm

import (
	dFarm "agropal/internal/domain/farm"
	"agropal/internal/domain/shared"
	"agropal/internal/utils"
	"context"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type RepositoryImpl struct {
	DB     *sqlx.DB
	Logger *utils.Logger
}

func (r *RepositoryImpl) Create(ctx context.Context, farm dFarm.Farm) error {
	insertQuery := `INSERT INTO farms (farm_id, user_id, name) VALUES (:farm_id, :user_id, :name);`

	modelFarm := DomainToModelFarm(farm)
	if _, errInsert := r.DB.NamedExecContext(ctx, insertQuery, modelFarm); errInsert != nil {
		r.Logger.Error(ctx, shared.AssetFarm, errInsert.Error())
		return errInsert
	}

	return nil
}

func (r *RepositoryImpl) GetOne(ctx context.Context, farmID uuid.UUID) (dFarm.Farm, error) {
	getQuery := `SELECT * FROM farms WHERE farm_id = $1;`

	var modelFarm ModelFarm
	if errQuery := r.DB.GetContext(ctx, &modelFarm, getQuery, farmID); errQuery != nil {
		r.Logger.Error(ctx, shared.AssetFarm, errQuery.Error())
		return dFarm.Farm{}, errQuery
	}

	return modelFarm.ToDomain(), nil
}

func (r *RepositoryImpl) GetMany(ctx context.Context, userID uuid.UUID) ([]dFarm.Farm, error) {
	selectQuery := `SELECT * FROM farms WHERE user_id = $1;`

	var modelFarm []ModelFarm
	if errSelect := r.DB.SelectContext(ctx, &modelFarm, selectQuery, userID); errSelect != nil {
		r.Logger.Error(ctx, shared.AssetFarm, errSelect.Error())
		return nil, errSelect
	}

	farms := make([]dFarm.Farm, len(modelFarm))
	for i, m := range modelFarm {
		farms[i] = m.ToDomain()
	}

	return farms, nil
}

func (r *RepositoryImpl) Update(ctx context.Context, farm dFarm.Farm) error {
	updateQuery := `UPDATE farms SET name = $1 WHERE farm_id = %2;`

	if _, errUpdate := r.DB.ExecContext(ctx, updateQuery, farm.Name, farm.FarmID); errUpdate != nil {
		r.Logger.Error(ctx, shared.AssetFarm, errUpdate.Error())
		return errUpdate
	}

	return nil
}

func (r *RepositoryImpl) Delete(ctx context.Context, farmID uuid.UUID) error {
	deleteQuery := `DELETE FROM farms WHERE farm_id = $1;`

	if _, errDelete := r.DB.ExecContext(ctx, deleteQuery, farmID); errDelete != nil {
		r.Logger.Error(ctx, shared.AssetFarm, errDelete.Error())
		return errDelete
	}

	return nil
}

func NewRepository(db *sqlx.DB, logger *utils.Logger) dFarm.Repository {
	return &RepositoryImpl{
		DB:     db,
		Logger: logger,
	}
}
