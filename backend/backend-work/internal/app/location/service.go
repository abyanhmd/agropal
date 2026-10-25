package location

import (
	"agropal/config"
	dFarm "agropal/internal/domain/farm"
	dLocation "agropal/internal/domain/location"
	"agropal/internal/domain/shared"
	dUser "agropal/internal/domain/user"
	"agropal/internal/infra/database"
	"context"

	"github.com/google/uuid"
)

type Service struct {
	transactor   database.Transactor
	userRepo     dUser.Repository
	farmRepo     dFarm.Repository
	locationRepo dLocation.Repository
}

func NewService(
	t database.Transactor,
	ur dUser.Repository,
	fr dFarm.Repository,
	wr dLocation.Repository,
) *Service {
	return &Service{
		transactor:   t,
		userRepo:     ur,
		farmRepo:     fr,
		locationRepo: wr,
	}
}

func (s *Service) Create(ctx context.Context, farmID uuid.UUID, cmd LocationCommand) (uuid.UUID, error) {
	if err := s.verifyFarm(ctx, farmID); err != nil {
		return uuid.Nil, err
	}

	locationID, errGen := uuid.NewV7()
	if errGen != nil {
		return uuid.Nil, errGen
	}
	location := cmd.ToDomain(locationID, farmID)

	if err := s.locationRepo.CreateLocation(ctx, s.transactor.DB(), location); err != nil {
		return uuid.Nil, err
	}

	return locationID, nil
}

func (s *Service) GetByFarmID(ctx context.Context, farmID uuid.UUID) ([]dLocation.Location, error) {
	if err := s.verifyFarm(ctx, farmID); err != nil {
		return []dLocation.Location{}, err
	}

	return s.locationRepo.ListLocationsByFarmID(ctx, s.transactor.DB(), farmID)
}

func (s *Service) Update(ctx context.Context, farmID, locationID uuid.UUID, cmd LocationCommand) error {
	if err := s.verifyFarm(ctx, farmID); err != nil {
		return err
	}

	return s.transactor.WithTx(ctx, func(q database.Querier) error {
		exists, err := s.locationRepo.GetLocationByID(ctx, q, locationID)
		if err != nil {
			return err
		}
		requesterRole := shared.GetMetadata(ctx, shared.KeyRole)
		if exists.FarmID != farmID && requesterRole != string(config.RoleAdmin) {
			return shared.ErrForbidden
		}
		location := cmd.ToDomain(locationID, farmID)
		if err = s.locationRepo.UpdateLocation(ctx, q, location); err != nil {
			return err
		}
		return nil
	})
}

func (s *Service) Delete(ctx context.Context, farmID, locationID uuid.UUID) error {
	if err := s.verifyFarm(ctx, farmID); err != nil {
		return err
	}

	return s.transactor.WithTx(ctx, func(q database.Querier) error {
		exists, err := s.locationRepo.GetLocationByID(ctx, q, locationID)
		if err != nil {
			return err
		}
		requesterRole := shared.GetMetadata(ctx, shared.KeyRole)
		if exists.FarmID != farmID && requesterRole != string(config.RoleAdmin) {
			return shared.ErrForbidden
		}
		if err := s.locationRepo.DeleteLocation(ctx, q, locationID); err != nil {
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
