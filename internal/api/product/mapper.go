package product

import (
	dProduct "agropal/internal/domain/product"
	"agropal/internal/response"
)

func DomainToProductResponse(p dProduct.Product) response.ProductResponse {
	return response.ProductResponse{
		ProductID:   p.ID,
		FarmID:      p.FarmID,
		Name:        p.Name,
		Description: p.Description,
		Image:       p.Image,
	}
}

func DomainToProductStockResponse(s dProduct.ProductStock) response.ProductStockResponse {
	return response.ProductStockResponse{
		StockID:   s.ID,
		ProductID: s.ProductID,
		DateAdded: s.DateAdded,
		Type:      string(s.Type),
		Stock:     s.Stock,
	}
}

func DomainToProductPriceResponse(p dProduct.ProductPrice) response.ProductPriceResponse {
	return response.ProductPriceResponse{
		PriceID:   p.ID,
		ProductID: p.ProductID,
		DateAdded: p.DateAdded,
		Price:     p.Price,
	}
}

func DomainToProductDetailResponse(d dProduct.ProductDetail) response.ProductDetailResponse {
	stocks := make([]response.ProductStockResponse, len(d.Stocks))
	for i, s := range d.Stocks {
		stocks[i] = DomainToProductStockResponse(s)
	}

	prices := make([]response.ProductPriceResponse, len(d.Prices))
	for i, p := range d.Prices {
		prices[i] = DomainToProductPriceResponse(p)
	}

	return response.ProductDetailResponse{
		Product: DomainToProductResponse(d.Product),
		Stocks:  stocks,
		Prices:  prices,
	}
}
