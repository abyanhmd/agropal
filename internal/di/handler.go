package di

import (
	"agropal/internal/interface/farm"
	"agropal/internal/interface/product"
	"agropal/internal/interface/user"
	"agropal/internal/interface/weather"
	"agropal/internal/utils"
)

type Handler struct {
	User    user.Handler
	Farm    farm.Handler
	Weather weather.Handler
	Product product.Handler
	//Transaction transaction.Handler
	//Post        post.Handler
}

func NewHandler(tool utils.IOHandler, service *Application) *Handler {
	return &Handler{
		User:    user.NewHandler(tool, service.UserService),
		Farm:    farm.NewHandler(tool, service.FarmService),
		Weather: weather.NewHandler(tool, service.WeatherService),
		Product: product.NewHandler(tool, service.ProductService),
		//Transaction: transaction.NewHandler(tool, service.Transaction),
		//Post:        post.NewHandler(tool, service.Post),
	}
}
