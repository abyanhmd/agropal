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
	"agropal/pkg"
)

type Repository struct {
	AuthRepo        dAuth.Repository
	UserRepo        dUser.Repository
	NoteRepo        dNote.Repository
	FarmRepo        dFarm.Repository
	WeatherRepo     dWeather.Repository
	ProductRepo     dProduct.Repository
	TransactionRepo dTransaction.Repository
	PostRepo        dPost.Repository
}

func NewRepository(logger *pkg.Logger) *Repository {
	return &Repository{
		AuthRepo:        rAuth.NewRepository(logger),
		UserRepo:        rUser.NewRepository(logger),
		NoteRepo:        rNote.NewRepository(logger),
		FarmRepo:        rFarm.NewRepository(logger),
		WeatherRepo:     rWeather.NewRepository(logger),
		ProductRepo:     rProduct.NewRepository(logger),
		TransactionRepo: rTransaction.NewRepository(logger),
		PostRepo:        rPost.NewRepository(logger),
	}
}
