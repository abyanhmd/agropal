package product

import (
	dProduct "agropal/internal/domain/product"
	"context"
	"time"

	"github.com/google/uuid"
)

func (s *Service) CreateProductStock(ctx context.Context, farmID, productID uuid.UUID, cmd ProductStockCommand) (uuid.UUID, error) {
	nowUTC := time.Now().UTC()

	product, errGet := s.productRepo.FindProductByID(ctx, s.transactor.DB(), productID)
	if errGet != nil {
		return uuid.Nil, errGet
	}
	if product.FarmID != farmID {
		return uuid.Nil, dProduct.ErrInvalidFarmID
	}

	stockID, errGen := uuid.NewV7()
	if errGen != nil {
		return uuid.Nil, errGen
	}

	stock := cmd.ToDomain(productID, stockID)

	if stock.Stock <= 0 {
		return uuid.Nil, dProduct.ErrInvalidStockAmount
	}
	if stock.DateAdded.After(nowUTC) {
		return uuid.Nil, dProduct.ErrInvalidDate
	}

	if err := s.productRepo.AddProductStock(ctx, s.transactor.DB(), stock); err != nil {
		return uuid.Nil, err
	}

	return stockID, nil
}

func (s *Service) DeleteProductStock(ctx context.Context, stockID uuid.UUID) error {
	return s.productRepo.DeleteProductStock(ctx, s.transactor.DB(), stockID)
}
