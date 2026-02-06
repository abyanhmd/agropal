package product

import (
	"context"
	"errors"

	"github.com/google/uuid"
)

func (s *Service) CreateProductPrice(ctx context.Context, farmID, productID uuid.UUID, cmd ProductPriceCommand) error {
	product, errGet := s.ProductRepository.GetOnlyProduct(ctx, productID)
	if errGet != nil {
		return errGet
	}
	if product.FarmID != farmID {
		return errors.New("product id and  farm id does not match")
	}

	priceID, errGen := uuid.NewV7()
	if errGen != nil {
		return errGen
	}

	price := cmd.ToDomain(productID, priceID)

	// TODO: Validate price input

	return s.ProductRepository.CreateProductPrice(ctx, price)
}

func (s *Service) UpdateProductPrice(ctx context.Context, farmID, productID, priceID uuid.UUID, cmd ProductPriceCommand) error {
	product, errGet := s.ProductRepository.GetOnlyProduct(ctx, productID)
	if errGet != nil {
		return errGet
	}
	if product.FarmID != farmID {
		return errors.New("product id and  farm id does not match")
	}

	price := cmd.ToDomain(productID, priceID)

	// TODO: Validate price input

	return s.ProductRepository.UpdateProductPrice(ctx, price)
}

func (s *Service) DeleteProductPrice(ctx context.Context, priceID uuid.UUID) error {
	return s.ProductRepository.DeleteProductPrice(ctx, priceID)
}
