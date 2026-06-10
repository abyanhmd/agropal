package di

import (
	"agropal/internal/application/auth"
	"agropal/internal/application/farm"
	"agropal/internal/application/post"
	"agropal/internal/application/product"
	"agropal/internal/application/transaction"
	"agropal/internal/application/user"
	"agropal/internal/application/weather"
	"agropal/internal/infra/database"
)

type Application struct {
	AuthService        auth.Service
	UserService        user.Service
	FarmService        farm.Service
	WeatherService     weather.Service
	ProductService     product.Service
	TransactionService transaction.Service
	PostService        post.Service
}

func NewService(
	t *database.Transactor,
	r *Repository,
) *Application {
	return &Application{
		AuthService:        auth.NewService(r.AuthRepo),
		UserService:        user.NewService(r.UserRepo),
		FarmService:        farm.NewService(r.FarmRepo),
		WeatherService:     weather.NewService(r.WeatherRepo),
		ProductService:     product.NewService(r.ProductRepo),
		TransactionService: transaction.NewService(r.TransactionRepo),
		PostService:        post.NewService(r.PostRepo),
	}

}
