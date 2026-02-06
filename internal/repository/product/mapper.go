package product

import (
	dProduct "agropal/internal/domain/product"
)

func (m ModelProduct) ToDomain() dProduct.Product {
	return dProduct.Product{
		ProductID:   m.ProductID,
		Name:        m.Name,
		Description: m.Description,
		Image:       m.Image,
		FarmID:      m.FarmID,
	}
}

func (m ModelProductStock) ToDomain() dProduct.ProductStock {
	return dProduct.ProductStock{
		StockID:   m.StockID,
		ProductID: m.ProductID,
		DateAdded: m.DateAdded,
		Type:      dProduct.StockType(m.Type),
		Stock:     m.Stock,
	}
}

func (m ModelProductPrice) ToDomain() dProduct.ProductPrice {
	return dProduct.ProductPrice{
		PriceID:   m.PriceID,
		ProductID: m.ProductID,
		DateAdded: m.DateAdded,
		Price:     m.Price,
	}
}

func DomainToModelProduct(domain dProduct.Product) ModelProduct {
	return ModelProduct{
		ProductID:   domain.ProductID,
		Name:        domain.Name,
		Description: domain.Description,
		Image:       domain.Image,
		FarmID:      domain.FarmID,
	}
}

func DomainToModelStock(domain dProduct.ProductStock) ModelProductStock {
	return ModelProductStock{
		StockID:   domain.StockID,
		ProductID: domain.ProductID,
		DateAdded: domain.DateAdded,
		Type:      string(domain.Type),
		Stock:     domain.Stock,
	}
}

func DomainToModelPrice(domain dProduct.ProductPrice) ModelProductPrice {
	return ModelProductPrice{
		PriceID:   domain.PriceID,
		ProductID: domain.ProductID,
		DateAdded: domain.DateAdded,
		Price:     domain.Price,
	}
}
