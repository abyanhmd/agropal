package di

import (
	dAuth "agropal/internal/domain/auth"
	dFarm "agropal/internal/domain/farm"
	dNote "agropal/internal/domain/note"
	dPost "agropal/internal/domain/post"
	dProduct "agropal/internal/domain/product"
	dTransaction "agropal/internal/domain/transaction"
	dUser "agropal/internal/domain/user"
	dWeather "agropal/internal/domain/weather"
	rAuth "agropal/internal/repository/auth"
	rFarm "agropal/internal/repository/farm"
	rNote "agropal/internal/repository/note"
	rPost "agropal/internal/repository/post"
	rProduct "agropal/internal/repository/product"
	rTransaction "agropal/internal/repository/transaction"
	rUser "agropal/internal/repository/user"
	rWeather "agropal/internal/repository/weather"
	"agropal/internal/utils"
)

type Repository struct {
	Auth        dAuth.Repository
	User        dUser.Repository
	Note        dNote.Repository
	Farm        dFarm.Repository
	Weather     dWeather.Repository
	Product     dProduct.Repository
	Transaction dTransaction.Repository
	Post        dPost.Repository
}

func NewRepository(logger *utils.Logger) *Repository {
	return &Repository{
		Auth:        rAuth.NewRepository(logger),
		User:        rUser.NewRepository(logger),
		Note:        rNote.NewRepository(logger),
		Farm:        rFarm.NewRepository(logger),
		Weather:     rWeather.NewRepository(logger),
		Product:     rProduct.NewRepository(logger),
		Transaction: rTransaction.NewRepository(logger),
		Post:        rPost.NewRepository(logger),
	}
}
