package farm

import (
	"agropal/internal/database"
	dFarm "agropal/internal/domain/farm"
	"agropal/internal/domain/shared"
	"agropal/internal/utils"
	"context"

	"github.com/google/uuid"
)

type RepositoryImpl struct {
	Logger *utils.Logger
}

func NewRepository(logger *utils.Logger) dFarm.Repository {
	return &RepositoryImpl{Logger: logger}
}

func (r *RepositoryImpl) Create(ctx context.Context, q database.Querier, farm dFarm.Farm) error {
	query := `INSERT INTO farms (farm_id, user_id, name) VALUES (:farm_id, :user_id, :name);`

	farmModel := ToFarmModel(farm)
	if _, err := q.NamedExecContext(ctx, query, farmModel); err != nil {
		r.Logger.Error(ctx, shared.AssetFarm, err.Error())
		return err
	}

	return nil
}

func (r *RepositoryImpl) FindByID(ctx context.Context, q database.Querier, farmID uuid.UUID) (dFarm.Farm, error) {
	query := `SELECT * FROM farms WHERE farm_id = $1;`

	var farmModel FarmModel
	if err := q.GetContext(ctx, &farmModel, query, farmID); err != nil {
		r.Logger.Error(ctx, shared.AssetFarm, err.Error())
		return dFarm.Farm{}, err
	}

	return farmModel.ToDomain(), nil
}

func (r *RepositoryImpl) ListByUserID(ctx context.Context, q database.Querier, userID uuid.UUID) ([]dFarm.Farm, error) {
	query := `SELECT * FROM farms WHERE user_id = $1;`

	var farmModels []FarmModel
	if err := q.SelectContext(ctx, &farmModels, query, userID); err != nil {
		r.Logger.Error(ctx, shared.AssetFarm, err.Error())
		return nil, err
	}

	farms := make([]dFarm.Farm, len(farmModels))
	for i, m := range farmModels {
		farms[i] = m.ToDomain()
	}

	return farms, nil
}

func (r *RepositoryImpl) Update(ctx context.Context, q database.Querier, farm dFarm.Farm) error {
	query := `UPDATE farms SET name = :name WHERE farm_id = :farm_id;`

	farmModel := ToFarmModel(farm)
	if _, err := q.ExecContext(ctx, query, farmModel); err != nil {
		r.Logger.Error(ctx, shared.AssetFarm, err.Error())
		return err
	}

	return nil
}

func (r *RepositoryImpl) Delete(ctx context.Context, q database.Querier, farmID uuid.UUID) error {
	query := `DELETE FROM farms WHERE farm_id = $1;`

	if _, err := q.ExecContext(ctx, query, farmID); err != nil {
		r.Logger.Error(ctx, shared.AssetFarm, err.Error())
		return err
	}

	return nil
}
