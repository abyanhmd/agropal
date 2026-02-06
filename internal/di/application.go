package di

import (
	"agropal/internal/application/farm"
	"agropal/internal/application/post"
	"agropal/internal/application/product"
	"agropal/internal/application/transaction"
	"agropal/internal/application/user"
	"agropal/internal/application/weather"
)

type Application struct {
	User        user.Service
	Farm        farm.Service
	Weather     weather.Service
	Product     product.Service
	Transaction transaction.Service
	Post        post.Service
}

func NewService(repository *Repository) *Application {
	return &Application{
		User:        user.NewService(repository.User),
		Farm:        farm.NewService(repository.Farm),
		Weather:     weather.NewService(repository.Weather),
		Product:     product.NewService(repository.Product),
		Transaction: transaction.NewService(repository.Transaction),
		Post:        post.NewService(repository.Post),
	}

}
