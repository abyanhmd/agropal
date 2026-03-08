package weather

import (
	"agropal/internal/database"
	"agropal/internal/domain/shared"
	dWeather "agropal/internal/domain/weather"
	"agropal/internal/utils"
	"context"

	"github.com/google/uuid"
)

type RepositoryImpl struct {
	Logger *utils.Logger
}

func NewRepository(logger *utils.Logger) dWeather.Repository {
	return &RepositoryImpl{Logger: logger}
}

func (r *RepositoryImpl) Create(ctx context.Context, q database.Querier, weather dWeather.Weather) error {
	query := `
		INSERT INTO weathers (
			weather_id, longitude, latitude, farm_id
		) VALUES (
			:weather_id, :longitude, :latitude, :farm_id
		);`

	weatherModel := ToWeatherModel(weather)
	if _, err := q.NamedExecContext(ctx, query, weatherModel); err != nil {
		r.Logger.Error(ctx, shared.AssetWeather, err.Error())
		return err
	}

	return nil
}

func (r *RepositoryImpl) ListByFarmID(ctx context.Context, q database.Querier, farmID uuid.UUID) ([]dWeather.Weather, error) {
	query := `SELECT * FROM weathers WHERE farm_id = $1;`

	var weatherModels []WeatherModel
	if err := q.SelectContext(ctx, &weatherModels, query, farmID); err != nil {
		r.Logger.Error(ctx, shared.AssetWeather, err.Error())
		return nil, err
	}

	weathers := make([]dWeather.Weather, len(weatherModels))
	for i, m := range weatherModels {
		weathers[i] = m.ToDomain()
	}

	return weathers, nil
}

func (r *RepositoryImpl) Update(ctx context.Context, q database.Querier, weather dWeather.Weather) error {
	query := `
		UPDATE weathers 
		SET longitude = :longitude,
			latitude = :latitude
		WHERE weather_id = :weather_id;`

	weatherModel := ToWeatherModel(weather)
	if _, err := q.ExecContext(ctx, query, weatherModel); err != nil {
		r.Logger.Error(ctx, shared.AssetWeather, err.Error())
		return err
	}

	return nil
}

func (r *RepositoryImpl) Delete(ctx context.Context, q database.Querier, weatherID uuid.UUID) error {
	query := `DELETE FROM weathers WHERE weather_id = $1;`

	if _, err := q.ExecContext(ctx, query, weatherID); err != nil {
		r.Logger.Error(ctx, shared.AssetWeather, err.Error())
		return err
	}

	return nil
}
