package product

import (
	dFarm "agropal/internal/domain/farm"
	dNote "agropal/internal/domain/note"
	dProduct "agropal/internal/domain/product"
	dUser "agropal/internal/domain/user"
	"agropal/internal/infra/database"
)

type Service struct {
	transactor  database.Transactor
	userRepo    dUser.Repository
	farmRepo    dFarm.Repository
	productRepo dProduct.Repository
	noteRepo    dNote.Repository
}

func NewService(
	t database.Transactor,
	ur dUser.Repository,
	fr dFarm.Repository,
	pr dProduct.Repository,
	nr dNote.Repository,
) *Service {
	return &Service{
		transactor:  t,
		userRepo:    ur,
		farmRepo:    fr,
		productRepo: pr,
		noteRepo:    nr,
	}
}
