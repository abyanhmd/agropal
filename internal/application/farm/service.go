package farm

import (
	dFarm "agropal/internal/domain/farm"
	"agropal/internal/utils"
	"context"

	"github.com/google/uuid"
)

type Service struct {
	FarmRepository dFarm.Repository
}

func NewService(farmRepository dFarm.Repository) Service {
	return Service{
		FarmRepository: farmRepository,
	}
}

func (s *Service) Create(ctx context.Context, cmd FarmCommand) error {
	userID, errParse := utils.ExtractUserID(ctx)
	if errParse != nil {
		return errParse
	}
	farmID, errGen := uuid.NewV7()
	if errGen != nil {
		return errGen
	}
	farm := cmd.ToDomain(userID)
	farm.ID = farmID

	// TODO: Validate farm input

	return s.FarmRepository.Create(ctx, farm)
}

func (s *Service) GetOne(ctx context.Context, farmID uuid.UUID) (dFarm.Farm, error) {
	return s.FarmRepository.FindByID(ctx, farmID)
}

func (s *Service) GetMany(ctx context.Context) ([]dFarm.Farm, error) {
	userID, errParse := utils.ExtractUserID(ctx)
	if errParse != nil {
		return nil, errParse
	}

	return s.FarmRepository.ListByUserID(ctx, userID)
}

func (s *Service) Update(ctx context.Context, farmID uuid.UUID, cmd FarmCommand) error {
	farm := cmd.ToDomain(farmID)

	// TODO: Validate update input

	return s.FarmRepository.Update(ctx, farm)
}

func (s *Service) Delete(ctx context.Context, farmID uuid.UUID) error {
	return s.FarmRepository.Delete(ctx, farmID)
}
