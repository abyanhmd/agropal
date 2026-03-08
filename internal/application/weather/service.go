package weather

import (
	dWeather "agropal/internal/domain/weather"
	"context"

	"github.com/google/uuid"
)

type Service struct {
	WeatherRepository dWeather.Repository
}

func NewService(weatherRepository dWeather.Repository) Service {
	return Service{
		WeatherRepository: weatherRepository,
	}
}

func (s *Service) Create(ctx context.Context, cmd WeatherCommand) error {
	weatherID, errGen := uuid.NewV7()
	if errGen != nil {
		return errGen
	}
	weather := cmd.ToDomain(weatherID)

	// TODO: Validate weather input

	return s.WeatherRepository.Create(ctx, weather)
}

func (s *Service) GetByFarmID(ctx context.Context, farmID uuid.UUID) ([]dWeather.Weather, error) {
	return s.WeatherRepository.ListByFarmID(ctx, farmID)
}

func (s *Service) Update(ctx context.Context, farmID uuid.UUID, cmd WeatherCommand) error {
	weather := cmd.ToDomain(farmID)

	// TODO: Validate update input

	return s.WeatherRepository.Update(ctx, weather)
}

func (s *Service) Delete(ctx context.Context, farmID uuid.UUID) error {
	return s.WeatherRepository.Delete(ctx, farmID)
}
