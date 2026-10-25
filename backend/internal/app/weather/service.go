package weather

import (
	"agropal/config"
	dFarm "agropal/internal/domain/farm"
	"agropal/internal/domain/shared"
	dUser "agropal/internal/domain/user"
	dWeather "agropal/internal/domain/weather"
	"agropal/internal/infra/database"
	"context"

	"github.com/google/uuid"
)

type Service struct {
	transactor  database.Transactor
	userRepo    dUser.Repository
	farmRepo    dFarm.Repository
	weatherRepo dWeather.Repository
}

func NewService(
	t database.Transactor,
	ur dUser.Repository,
	fr dFarm.Repository,
	wr dWeather.Repository,
) *Service {
	return &Service{
		transactor:  t,
		userRepo:    ur,
		farmRepo:    fr,
		weatherRepo: wr,
	}
}

func (s *Service) Create(ctx context.Context, farmID uuid.UUID, cmd WeatherCommand) error {
	if err := s.verifyFarm(ctx, farmID); err != nil {
		return err
	}

	weatherID, errGen := uuid.NewV7()
	if errGen != nil {
		return errGen
	}
	weather := cmd.ToDomain(weatherID, farmID)

	return s.weatherRepo.CreateWeatherLocation(ctx, s.transactor.DB(), weather)
}

func (s *Service) GetByFarmID(ctx context.Context, farmID uuid.UUID) ([]dWeather.Weather, error) {
	if err := s.verifyFarm(ctx, farmID); err != nil {
		return []dWeather.Weather{}, err
	}

	return s.weatherRepo.ListWeatherLocationsByFarmID(ctx, s.transactor.DB(), farmID)
}

func (s *Service) Update(ctx context.Context, farmID, weatherID uuid.UUID, cmd WeatherCommand) error {
	if err := s.verifyFarm(ctx, farmID); err != nil {
		return err
	}

	return s.transactor.WithTx(ctx, func(q database.Querier) error {
		exists, err := s.weatherRepo.GetWeatherLocationByID(ctx, q, weatherID)
		if err != nil {
			return err
		}
		requesterRole := shared.GetMetadata(ctx, shared.KeyRole)
		if exists.FarmID != farmID && requesterRole != string(config.RoleAdmin) {
			return shared.ErrForbidden
		}
		weather := cmd.ToDomain(weatherID, farmID)
		if err = s.weatherRepo.UpdateWeatherLocation(ctx, q, weather); err != nil {
			return err
		}
		return nil
	})
}

func (s *Service) Delete(ctx context.Context, farmID, weatherID uuid.UUID) error {
	if err := s.verifyFarm(ctx, farmID); err != nil {
		return err
	}

	return s.transactor.WithTx(ctx, func(q database.Querier) error {
		exists, err := s.weatherRepo.GetWeatherLocationByID(ctx, q, weatherID)
		if err != nil {
			return err
		}
		requesterRole := shared.GetMetadata(ctx, shared.KeyRole)
		if exists.FarmID != farmID && requesterRole != string(config.RoleAdmin) {
			return shared.ErrForbidden
		}
		if err := s.weatherRepo.DeleteWeatherLocation(ctx, q, weatherID); err != nil {
			return err
		}
		return nil
	})
}

func (s *Service) verifyFarm(ctx context.Context, farmID uuid.UUID) error {
	return s.transactor.WithTx(ctx, func(q database.Querier) error {
		userID, err := shared.ExtractUserID(ctx)
		if err != nil {
			return err
		}
		user, err := s.userRepo.FindUserByID(ctx, q, userID)
		if err != nil {
			return err
		}
		if user == (dUser.User{}) {
			return dUser.ErrInvalidUserID
		}
		farm, err := s.farmRepo.FindFarmByID(ctx, q, farmID)
		if err != nil {
			return err
		}
		requesterRole := shared.GetMetadata(ctx, shared.KeyRole)
		if farm.UserID != user.ID && requesterRole != string(config.RoleAdmin) {
			return shared.ErrForbidden
		}
		return nil
	})
}
