package product

import (
	aProduct "agropal/internal/app/product"
	"agropal/internal/utils"
)

type Handler struct {
	ioHandler      *utils.IOHandler
	productService *aProduct.Service
}

func NewHandler(
	tool *utils.IOHandler,
	productService *aProduct.Service,
) Handler {
	return Handler{
		ioHandler:      tool,
		productService: productService,
	}
}
