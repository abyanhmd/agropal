package product

import (
	"context"
	"errors"

	"github.com/google/uuid"
)

func (s *Service) CreateProductStock(ctx context.Context, farmID, productID uuid.UUID, cmd ProductStockCommand) error {
	product, errGet := s.ProductRepository.GetOnlyProduct(ctx, productID)
	if errGet != nil {
		return errGet
	}
	if product.FarmID != farmID {
		return errors.New("product id and  farm id does not match")
	}

	stockID, errGen := uuid.NewV7()
	if errGen != nil {
		return errGen
	}

	stock := cmd.ToDomain(productID, stockID)

	// TODO: Validate stock input

	return s.ProductRepository.CreateProductStock(ctx, stock)
}

func (s *Service) UpdateProductStock(ctx context.Context, farmID, productID, stockID uuid.UUID, cmd ProductStockCommand) error {
	product, errGet := s.ProductRepository.GetOnlyProduct(ctx, productID)
	if errGet != nil {
		return errGet
	}
	if product.FarmID != farmID {
		return errors.New("product id and  farm id does not match")
	}

	stock := cmd.ToDomain(stockID, productID)

	// TODO: Validate stock input

	return s.ProductRepository.UpdateProductStock(ctx, stock)
}

func (s *Service) DeleteProductStock(ctx context.Context, stockID uuid.UUID) error {
	return s.ProductRepository.DeleteProductStock(ctx, stockID)
}
