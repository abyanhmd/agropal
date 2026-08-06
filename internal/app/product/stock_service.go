package product

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
)

func (s *Service) CreateProductStock(ctx context.Context, farmID, productID uuid.UUID, cmd ProductStockCommand) error {
	nowUTC := time.Now().UTC()

	product, errGet := s.productRepo.FindProductByID(ctx, s.transactor.DB(), productID)
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

	if stock.Stock <= 0 {
		return errors.New("stock cannot be zero or negative")
	}
	if stock.DateAdded.After(nowUTC) {
		return errors.New("date added cannot be in the future")
	}

	return s.productRepo.AddProductStock(ctx, s.transactor.DB(), stock)
}

func (s *Service) DeleteProductStock(ctx context.Context, stockID uuid.UUID) error {
	return s.productRepo.DeleteProductStock(ctx, s.transactor.DB(), stockID)
}
