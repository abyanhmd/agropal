package persistence

import (
	dFarm "agropal/internal/domain/farm"
	"agropal/internal/domain/shared"
	"agropal/internal/infra/database"
	"agropal/pkg"
	"context"

	"github.com/google/uuid"
)

type farmRepo struct {
	logger *pkg.Logger
}

func NewFarmRepository(l *pkg.Logger) dFarm.Repository {
	return &farmRepo{logger: l}
}

func (r *farmRepo) CreateFarm(ctx context.Context, q database.Querier, farm dFarm.Farm) error {
	query := `INSERT INTO farms (farm_id, user_id, name) VALUES (:farm_id, :user_id, :name);`

	farmModel := ToFarmModel(farm)
	if _, err := q.NamedExecContext(ctx, query, farmModel); err != nil {
		r.logger.Error(ctx, shared.AssetFarm, err.Error())
		return err
	}

	return nil
}

func (r *farmRepo) FindFarmByID(ctx context.Context, q database.Querier, farmID uuid.UUID) (dFarm.Farm, error) {
	query := `SELECT * FROM farms WHERE farm_id = $1;`

	var farmModel FarmModel
	if err := q.GetContext(ctx, &farmModel, query, farmID); err != nil {
		r.logger.Error(ctx, shared.AssetFarm, err.Error())
		return dFarm.Farm{}, err
	}

	return farmModel.ToDomain(), nil
}

func (r *farmRepo) ListFarmByUserID(ctx context.Context, q database.Querier, userID uuid.UUID) ([]dFarm.Farm, error) {
	query := `SELECT * FROM farms WHERE user_id = $1;`

	var farmModels []FarmModel
	if err := q.SelectContext(ctx, &farmModels, query, userID); err != nil {
		r.logger.Error(ctx, shared.AssetFarm, err.Error())
		return nil, err
	}

	farms := make([]dFarm.Farm, len(farmModels))
	for i, m := range farmModels {
		farms[i] = m.ToDomain()
	}

	return farms, nil
}

func (r *farmRepo) UpdateFarm(ctx context.Context, q database.Querier, farm dFarm.Farm) error {
	query := `UPDATE farms SET name = :name WHERE farm_id = :farm_id;`

	farmModel := ToFarmModel(farm)
	if _, err := q.ExecContext(ctx, query, farmModel); err != nil {
		r.logger.Error(ctx, shared.AssetFarm, err.Error())
		return err
	}

	return nil
}

func (r *farmRepo) DeleteFarm(ctx context.Context, q database.Querier, farmID uuid.UUID) error {
	query := `DELETE FROM farms WHERE farm_id = $1;`

	if _, err := q.ExecContext(ctx, query, farmID); err != nil {
		r.logger.Error(ctx, shared.AssetFarm, err.Error())
		return err
	}

	return nil
}
