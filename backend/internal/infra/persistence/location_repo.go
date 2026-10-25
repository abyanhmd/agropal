package persistence

import (
	dLocation "agropal/internal/domain/location"
	"agropal/internal/domain/shared"
	"agropal/internal/infra/database"
	"agropal/pkg"
	"context"

	"github.com/google/uuid"
)

type locationRepo struct {
	logger *pkg.Logger
}

func NewLocationRepository(l *pkg.Logger) dLocation.Repository {
	return &locationRepo{logger: l}
}

func (r *locationRepo) CreateLocation(ctx context.Context, q database.Querier, location dLocation.Location) error {
	query := `
		INSERT INTO location_info (
			location_id, longitude, latitude, farm_id
		) VALUES (
			:location_id, :longitude, :latitude, :farm_id
		);`

	locationModel := ToLocationModel(location)
	if _, err := q.NamedExecContext(ctx, query, locationModel); err != nil {
		r.logger.Error(ctx, shared.AssetLocation, err.Error())
		return err
	}

	return nil
}

func (r *locationRepo) GetLocationByID(ctx context.Context, q database.Querier, id uuid.UUID) (dLocation.Location, error) {
	query := `SELECT * FROM location_info WHERE location_id = $1;`

	var locationModel LocationModel
	if err := q.GetContext(ctx, &locationModel, query, id); err != nil {
		r.logger.Error(ctx, shared.AssetLocation, err.Error())
		return dLocation.Location{}, err
	}

	return locationModel.ToDomain(), nil
}

func (r *locationRepo) ListLocationsByFarmID(ctx context.Context, q database.Querier, farmID uuid.UUID) ([]dLocation.Location, error) {
	query := `SELECT * FROM location_info WHERE farm_id = $1;`

	var locationModels []LocationModel
	if err := q.SelectContext(ctx, &locationModels, query, farmID); err != nil {
		r.logger.Error(ctx, shared.AssetLocation, err.Error())
		return nil, err
	}

	locations := make([]dLocation.Location, len(locationModels))
	for i, m := range locationModels {
		locations[i] = m.ToDomain()
	}

	return locations, nil
}

func (r *locationRepo) UpdateLocation(ctx context.Context, q database.Querier, location dLocation.Location) error {
	query := `
		UPDATE location_info
		SET longitude = :longitude,
			latitude = :latitude
		WHERE location_id = :location_id;`

	locationModel := ToLocationModel(location)
	if _, err := q.NamedExecContext(ctx, query, locationModel); err != nil {
		r.logger.Error(ctx, shared.AssetLocation, err.Error())
		return err
	}

	return nil
}

func (r *locationRepo) DeleteLocation(ctx context.Context, q database.Querier, id uuid.UUID) error {
	query := `DELETE FROM location_info WHERE location_id = $1;`

	if _, err := q.ExecContext(ctx, query, id); err != nil {
		r.logger.Error(ctx, shared.AssetLocation, err.Error())
		return err
	}

	return nil
}
