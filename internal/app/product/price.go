package product

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
)

func (s *Service) CreateProductPrice(ctx context.Context, farmID, productID uuid.UUID, cmd ProductPriceCommand) error {
	nowUTC := time.Now().UTC()
	db := s.transactor.DB()

	product, errGet := s.productRepo.FindProductByID(ctx, db, productID)
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

	if price.Price <= 0 {
		return errors.New("price cannot be zero or negative")
	}
	if price.DateAdded.After(nowUTC) {
		return errors.New("date added cannot be in the future")
	}

	return s.productRepo.AddProductPrice(ctx, db, price)
}

func (s *Service) DeleteProductPrice(ctx context.Context, priceID uuid.UUID) error {
	return s.productRepo.DeleteProductPrice(ctx, s.transactor.DB(), priceID)
}
