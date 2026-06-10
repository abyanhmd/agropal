package product

import (
	aProduct "agropal/internal/application/product"
	"agropal/internal/utils"
)

type Handler struct {
	IoHandler      utils.IOHandler
	ProductService aProduct.Service
}

func NewHandler(
	tool utils.IOHandler,
	userService aProduct.Service,
) Handler {
	return Handler{
		IoHandler:      tool,
		ProductService: userService,
	}
}
