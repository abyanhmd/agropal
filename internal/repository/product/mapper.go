package product

import (
	dProduct "agropal/internal/domain/product"
)

func ToProductModel(domain dProduct.Product) ProductModel {
	return ProductModel{
		ID:          domain.ID,
		FarmID:      domain.FarmID,
		Name:        domain.Name,
		Description: domain.Description,
		Image:       domain.Image,
	}
}

func ToStockModel(domain dProduct.ProductStock) ProductStockModel {
	return ProductStockModel{
		ID:        domain.ID,
		ProductID: domain.ProductID,
		CreatedAt: domain.DateAdded,
		Type:      string(domain.Type),
		Quantity:  domain.Stock,
	}
}

func ToPriceModel(domain dProduct.ProductPrice) ProductPriceModel {
	return ProductPriceModel{
		ID:        domain.ID,
		ProductID: domain.ProductID,
		CreatedAt: domain.DateAdded,
		Price:     domain.Price,
	}
}

func (m ProductModel) ToDomain() dProduct.Product {
	return dProduct.Product{
		ID:          m.ID,
		FarmID:      m.FarmID,
		Name:        m.Name,
		Description: m.Description,
		Image:       m.Image,
	}
}

func (m ProductStockModel) ToDomain() dProduct.ProductStock {
	return dProduct.ProductStock{
		ID:        m.ID,
		ProductID: m.ProductID,
		DateAdded: m.CreatedAt,
		Type:      dProduct.StockMovementType(m.Type),
		Stock:     m.Quantity,
	}
}

func (m ProductPriceModel) ToDomain() dProduct.ProductPrice {
	return dProduct.ProductPrice{
		ID:        m.ID,
		ProductID: m.ProductID,
		DateAdded: m.CreatedAt,
		Price:     m.Price,
	}
}
