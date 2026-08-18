package farm

import (
	"agropal/config"
	dFarm "agropal/internal/domain/farm"
	"agropal/internal/domain/shared"
	dUser "agropal/internal/domain/user"
	"agropal/internal/infra/database"
	"context"

	"github.com/google/uuid"
)

type Service struct {
	transactor database.Transactor
	userRepo   dUser.Repository
	farmRepo   dFarm.Repository
}

func NewService(
	t database.Transactor,
	ur dUser.Repository,
	fr dFarm.Repository,
) *Service {
	return &Service{
		transactor: t,
		userRepo:   ur,
		farmRepo:   fr,
	}
}

func (s *Service) Create(ctx context.Context, cmd FarmCommand) error {
	user, err := s.verifyUser(ctx)
	if err != nil {
		return err
	}

	farmID, err := uuid.NewV7()
	if err != nil {
		return err
	}
	farm := cmd.ToDomain(user.ID)
	farm.ID = farmID

	return s.farmRepo.CreateFarm(ctx, s.transactor.DB(), farm)
}

func (s *Service) GetOne(ctx context.Context, farmID uuid.UUID) (dFarm.Farm, error) {
	user, err := s.verifyUser(ctx)
	if err != nil {
		return dFarm.Farm{}, err
	}

	farm, err := s.farmRepo.FindFarmByID(ctx, s.transactor.DB(), farmID)
	if err != nil {
		return dFarm.Farm{}, err
	}

	if farm.UserID != user.ID {
		return dFarm.Farm{}, shared.ErrForbidden
	}

	return farm, nil
}

func (s *Service) GetMany(ctx context.Context) ([]dFarm.Farm, error) {
	user, err := s.verifyUser(ctx)
	if err != nil {
		return []dFarm.Farm{}, err
	}

	return s.farmRepo.ListFarmByUserID(ctx, s.transactor.DB(), user.ID)
}

func (s *Service) Update(ctx context.Context, farmID uuid.UUID, cmd FarmCommand) error {
	user, err := s.verifyUser(ctx)
	if err != nil {
		return err
	}

	return s.transactor.WithTx(ctx, func(q database.Querier) error {
		farms, err := s.farmRepo.FindFarmByID(ctx, q, farmID)
		if err != nil {
			return err
		}
		requesterRole := shared.GetMetadata(ctx, shared.KeyRole)
		if farms.UserID != user.ID && requesterRole != string(config.RoleAdmin) {
			return shared.ErrForbidden
		}

		if err = s.farmRepo.UpdateFarm(ctx, q, cmd.ToDomain(farmID)); err != nil {
			return err
		}
		return nil
	})

}

func (s *Service) Delete(ctx context.Context, farmID uuid.UUID) error {
	user, err := s.verifyUser(ctx)
	if err != nil {
		return err
	}

	return s.transactor.WithTx(ctx, func(q database.Querier) error {
		farms, err := s.farmRepo.FindFarmByID(ctx, q, farmID)
		if err != nil {
			return err
		}
		requesterRole := shared.GetMetadata(ctx, shared.KeyRole)
		if farms.UserID != user.ID && requesterRole != string(config.RoleAdmin) {
			return shared.ErrForbidden
		}

		if err = s.farmRepo.DeleteFarm(ctx, q, farmID); err != nil {
			return err
		}
		return nil
	})
}

/* ==================== Helper ====================== */

func (s *Service) verifyUser(ctx context.Context) (dUser.User, error) {
	userID, err := shared.ExtractUserID(ctx)
	if err != nil {
		return dUser.User{}, err
	}
	user, err := s.userRepo.FindUserByID(ctx, s.transactor.DB(), userID)
	if err != nil {
		return dUser.User{}, err
	}
	if user == (dUser.User{}) {
		return dUser.User{}, dUser.ErrInvalidUserID
	}
	return user, nil
}
