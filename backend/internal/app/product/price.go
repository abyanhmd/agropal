package product

import (
	dProduct "agropal/internal/domain/product"
	"context"
	"strings"
	"time"

	"github.com/google/uuid"
)

func (s *Service) CreateProductPrice(ctx context.Context, farmID, productID uuid.UUID, cmd ProductPriceCommand) (uuid.UUID, error) {
	nowUTC := time.Now().UTC()
	db := s.transactor.DB()

	product, errGet := s.productRepo.FindProductByID(ctx, db, productID)
	if errGet != nil {
		return uuid.Nil, errGet
	}
	if product.FarmID != farmID {
		return uuid.Nil, dProduct.ErrInvalidFarmID
	}

	priceID, errGen := uuid.NewV7()
	if errGen != nil {
		return uuid.Nil, errGen
	}
	price := cmd.ToDomain(productID, priceID)

	if price.Price <= 0 {
		return uuid.Nil, dProduct.ErrInvalidPriceAmount
	}
	if strings.TrimSpace(price.Currency) == "" {
		return uuid.Nil, dProduct.ErrInvalidCurrency
	}
	if price.DateAdded.After(nowUTC) {
		return uuid.Nil, dProduct.ErrInvalidDate
	}

	if err := s.productRepo.AddProductPrice(ctx, db, price); err != nil {
		return uuid.Nil, err
	}

	return priceID, nil
}

func (s *Service) DeleteProductPrice(ctx context.Context, priceID uuid.UUID) error {
	return s.productRepo.DeleteProductPrice(ctx, s.transactor.DB(), priceID)
}
