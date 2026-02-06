package product

import (
	dProduct "agropal/internal/domain/product"
)

type Service struct {
	ProductRepository dProduct.Repository
}

func NewService(productRepository dProduct.Repository) Service {
	return Service{
		ProductRepository: productRepository,
	}
}
