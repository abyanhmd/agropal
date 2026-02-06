package product

import (
	dProduct "agropal/internal/domain/product"
	"context"

	"github.com/google/uuid"
)

func (s *Service) CreateProduct(ctx context.Context, cmd ProductCommand) error {
	productID, errGen := uuid.NewV7()
	if errGen != nil {
		return errGen
	}
	product := cmd.ToDomain(productID)

	// TODO: Validate product input

	return s.ProductRepository.CreateProduct(ctx, product)
}

func (s *Service) UpdateProduct(ctx context.Context, productID uuid.UUID, cmd ProductCommand) error {
	product := cmd.ToDomain(productID)

	// TODO: Validate product input

	return s.ProductRepository.UpdateProduct(ctx, product)
}

func (s *Service) GetOneProduct(ctx context.Context, productID uuid.UUID) (dProduct.ProductDetail, error) {
	return s.ProductRepository.GetOneProduct(ctx, productID)
}

func (s *Service) GetManyProducts(ctx context.Context, farmID uuid.UUID) ([]dProduct.ProductDetail, error) {
	return s.ProductRepository.GetManyProducts(ctx, farmID)
}

func (s *Service) DeleteProduct(ctx context.Context, productID uuid.UUID) error {
	return s.ProductRepository.DeleteProduct(ctx, productID)
}
