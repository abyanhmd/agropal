package persistence

import (
	"agropal/internal/domain/shared"
	dWeather "agropal/internal/domain/weather"
	"agropal/internal/infra/database"
	"agropal/pkg"
	"context"

	"github.com/google/uuid"
)

type weatherRepo struct {
	logger *pkg.Logger
}

func NewWeatherRepository(l *pkg.Logger) dWeather.Repository {
	return &weatherRepo{logger: l}
}

func (r *weatherRepo) CreateWeatherLocation(ctx context.Context, q database.Querier, weather dWeather.Weather) error {
	query := `
		INSERT INTO weathers (
			weather_id, longitude, latitude, farm_id
		) VALUES (
			:weather_id, :longitude, :latitude, :farm_id
		);`

	weatherModel := ToWeatherModel(weather)
	if _, err := q.NamedExecContext(ctx, query, weatherModel); err != nil {
		r.logger.Error(ctx, shared.AssetWeather, err.Error())
		return err
	}

	return nil
}

func (r *weatherRepo) GetWeatherLocationByID(ctx context.Context, q database.Querier, id uuid.UUID) (dWeather.Weather, error) {
	query := `SELECT * FROM weathers WHERE weather_id = $1;`

	var weatherModel WeatherModel
	if err := q.GetContext(ctx, &weatherModel, query, id); err != nil {
		r.logger.Error(ctx, shared.AssetWeather, err.Error())
		return dWeather.Weather{}, err
	}

	return weatherModel.ToDomain(), nil
}

func (r *weatherRepo) ListWeatherLocationsByFarmID(ctx context.Context, q database.Querier, farmID uuid.UUID) ([]dWeather.Weather, error) {
	query := `SELECT * FROM weathers WHERE farm_id = $1;`

	var weatherModels []WeatherModel
	if err := q.SelectContext(ctx, &weatherModels, query, farmID); err != nil {
		r.logger.Error(ctx, shared.AssetWeather, err.Error())
		return nil, err
	}

	weathers := make([]dWeather.Weather, len(weatherModels))
	for i, m := range weatherModels {
		weathers[i] = m.ToDomain()
	}

	return weathers, nil
}

func (r *weatherRepo) UpdateWeatherLocation(ctx context.Context, q database.Querier, weather dWeather.Weather) error {
	query := `
		UPDATE weathers 
		SET longitude = :longitude,
			latitude = :latitude
		WHERE weather_id = :weather_id;`

	weatherModel := ToWeatherModel(weather)
	if _, err := q.ExecContext(ctx, query, weatherModel); err != nil {
		r.logger.Error(ctx, shared.AssetWeather, err.Error())
		return err
	}

	return nil
}

func (r *weatherRepo) DeleteWeatherLocation(ctx context.Context, q database.Querier, id uuid.UUID) error {
	query := `DELETE FROM weathers WHERE weather_id = $1;`

	if _, err := q.ExecContext(ctx, query, id); err != nil {
		r.logger.Error(ctx, shared.AssetWeather, err.Error())
		return err
	}

	return nil
}
