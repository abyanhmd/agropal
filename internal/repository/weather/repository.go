package weather

import (
	"agropal/internal/domain/shared"
	dWeather "agropal/internal/domain/weather"
	"agropal/internal/utils"
	"context"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type RepositoryImpl struct {
	DB     *sqlx.DB
	Logger *utils.Logger
}

func (r *RepositoryImpl) Create(ctx context.Context, weather dWeather.Weather) error {
	insertQuery := `INSERT INTO weathers (
		weather_id, longitude, latitude, farm_id
	) VALUES (
		:weather_id, :longitude, :latitude, :farm_id
	);`

	modelWeather := DomainToModelWeather(weather)
	if _, errInsert := r.DB.NamedExecContext(ctx, insertQuery, modelWeather); errInsert != nil {
		r.Logger.Error(ctx, shared.AssetWeather, errInsert.Error())
		return errInsert
	}

	return nil
}

func (r *RepositoryImpl) GetByFarmID(ctx context.Context, farmID uuid.UUID) ([]dWeather.Weather, error) {
	selectQuery := `SELECT * FROM weathers WHERE farm_id = $1;`

	var modelWeathers []ModelWeather
	if errSelect := r.DB.SelectContext(ctx, &modelWeathers, selectQuery, farmID); errSelect != nil {
		r.Logger.Error(ctx, shared.AssetWeather, errSelect.Error())
		return nil, errSelect
	}

	weathers := make([]dWeather.Weather, len(modelWeathers))
	for i, m := range modelWeathers {
		weathers[i] = m.ToDomain()
	}

	return weathers, nil
}

func (r *RepositoryImpl) Update(ctx context.Context, weather dWeather.Weather) error {
	updateQuery := `UPDATE weathers 
	SET longitude = $1,
		latitude = $2
		WHERE weather_id = $3;`

	if _, errUpdate := r.DB.ExecContext(ctx, updateQuery, weather.Longitude, weather.Latitude, weather.WeatherID); errUpdate != nil {
		r.Logger.Error(ctx, shared.AssetWeather, errUpdate.Error())
		return errUpdate
	}

	return nil
}

func (r *RepositoryImpl) Delete(ctx context.Context, weatherID uuid.UUID) error {
	deleteQuery := `DELETE FROM weathers WHERE weather_id = $1;`

	if _, errDelete := r.DB.ExecContext(ctx, deleteQuery, weatherID); errDelete != nil {
		r.Logger.Error(ctx, shared.AssetWeather, errDelete.Error())
		return errDelete
	}

	return nil
}

func NewRepository(db *sqlx.DB, logger *utils.Logger) dWeather.Repository {
	return &RepositoryImpl{
		DB:     db,
		Logger: logger,
	}
}
