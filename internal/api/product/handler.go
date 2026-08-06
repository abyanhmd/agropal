package product

import (
	aProduct "agropal/internal/app/product"
	"agropal/internal/utils"
)

type Handler struct {
	IoHandler      *utils.IOHandler
	ProductService *aProduct.Service
}

func NewHandler(
	tool *utils.IOHandler,
	productService *aProduct.Service,
) Handler {
	return Handler{
		IoHandler:      tool,
		ProductService: productService,
	}
}
