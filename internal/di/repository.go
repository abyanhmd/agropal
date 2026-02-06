package di

import (
	dFarm "agropal/internal/domain/farm"
	dPost "agropal/internal/domain/post"
	dProduct "agropal/internal/domain/product"
	dTransaction "agropal/internal/domain/transaction"
	dUser "agropal/internal/domain/user"
	dWeather "agropal/internal/domain/weather"
	rFarm "agropal/internal/repository/farm"
	rPost "agropal/internal/repository/post"
	rProduct "agropal/internal/repository/product"
	rTransaction "agropal/internal/repository/transaction"
	rUser "agropal/internal/repository/user"
	rWeather "agropal/internal/repository/weather"
	"agropal/internal/utils"

	"github.com/jmoiron/sqlx"
)

type Repository struct {
	User        dUser.Repository
	Farm        dFarm.Repository
	Weather     dWeather.Repository
	Product     dProduct.Repository
	Transaction dTransaction.Repository
	Post        dPost.Repository
}

func NewRepository(db *sqlx.DB, logger *utils.Logger) *Repository {
	return &Repository{
		User:        rUser.NewRepository(db, logger),
		Farm:        rFarm.NewRepository(db, logger),
		Weather:     rWeather.NewRepository(db, logger),
		Product:     rProduct.NewRepository(db, logger),
		Transaction: rTransaction.NewRepository(db, logger),
		Post:        rPost.NewRepository(db, logger),
	}
}
