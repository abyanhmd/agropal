package product

import (
	dProduct "agropal/internal/domain/product"
	"agropal/internal/domain/shared"
	"agropal/internal/utils"
	"context"
	"database/sql"
	"errors"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type RepositoryImpl struct {
	DB     *sqlx.DB
	Logger *utils.Logger
}

/* ==================== Create ====================== */

func (r *RepositoryImpl) CreateProduct(ctx context.Context, product dProduct.Product) error {
	insertQuery := `INSERT INTO products (
		product_id, name,
		description, image,
		farm_id
	) VALUES (
	  	:product_id, :name,
		:description, :image,
		:farm_id
	);`

	modelProduct := DomainToModelProduct(product)
	if _, errInsert := r.DB.NamedExecContext(ctx, insertQuery, modelProduct); errInsert != nil {
		r.Logger.Error(ctx, shared.AssetProduct, errInsert.Error())
		return errInsert
	}

	return nil
}

func (r *RepositoryImpl) CreateProductStock(ctx context.Context, stock dProduct.ProductStock) error {
	insertQuery := `INSERT INTO product_stocks (
		stock_id, product_id, date_added, type, stock 
	) VALUES (
	    :stock_id, :product_id, :date_added, :type, :stock
	);`

	modelStock := DomainToModelStock(stock)
	if _, errInsert := r.DB.ExecContext(ctx, insertQuery, modelStock); errInsert != nil {
		r.Logger.Error(ctx, shared.AssetProduct, errInsert.Error())
		return errInsert
	}

	return nil
}

func (r *RepositoryImpl) CreateProductPrice(ctx context.Context, price dProduct.ProductPrice) error {
	insertQuery := `INSERT INTO product_prices (
		price_id, product_id, date_added, price
	) VALUES (
		:price_id, :product_id, :date_added, :price
	);`

	modelPrice := DomainToModelPrice(price)
	if _, errInsert := r.DB.ExecContext(ctx, insertQuery, modelPrice); errInsert != nil {
		r.Logger.Error(ctx, shared.AssetProduct, errInsert.Error())
		return errInsert
	}

	return nil
}

/* ==================== Update ====================== */

func (r *RepositoryImpl) UpdateProduct(ctx context.Context, product dProduct.Product) error {
	updateQuery := `UPDATE products SET name = :name, description = :description, image = :image WHERE product_id = :product_id;`

	modelProduct := DomainToModelProduct(product)
	if _, errUpdate := r.DB.NamedExecContext(ctx, updateQuery, modelProduct); errUpdate != nil {
		r.Logger.Error(ctx, shared.AssetProduct, errUpdate.Error())
		return errUpdate
	}

	return nil
}

func (r *RepositoryImpl) UpdateProductStock(ctx context.Context, stock dProduct.ProductStock) error {
	updateQuery := `UPDATE product_stocks SET stock = :stock WHERE stock_id = :stock_id;`

	modelStock := DomainToModelStock(stock)
	if _, errUpdate := r.DB.NamedExecContext(ctx, updateQuery, modelStock); errUpdate != nil {
		r.Logger.Error(ctx, shared.AssetProduct, errUpdate.Error())
		return errUpdate
	}

	return nil
}

func (r *RepositoryImpl) UpdateProductPrice(ctx context.Context, price dProduct.ProductPrice) error {
	updateQuery := `UPDATE product_prices SET price = :price WHERE price_id = :price_id;`

	modelPrice := DomainToModelPrice(price)
	if _, errUpdate := r.DB.NamedExecContext(ctx, updateQuery, modelPrice); errUpdate != nil {
		r.Logger.Error(ctx, shared.AssetProduct, errUpdate.Error())
		return errUpdate
	}

	return nil
}

/* ==================== Fetch ====================== */

func (r *RepositoryImpl) GetOnlyProduct(ctx context.Context, productID uuid.UUID) (dProduct.Product, error) {
	getQuery := `SELECT * FROM products WHERE product_id = $1;`

	var modelProduct ModelProduct
	if errGet := r.DB.GetContext(ctx, &modelProduct, getQuery, productID); errGet != nil {
		r.Logger.Error(ctx, shared.AssetProduct, errGet.Error())
		return dProduct.Product{}, errGet
	}

	return modelProduct.ToDomain(), nil
}

func (r *RepositoryImpl) GetOneProduct(ctx context.Context, productID uuid.UUID) (dProduct.ProductDetail, error) {
	getProductQuery := `SELECT * FROM products WHERE product_id = $1;`
	getProductStockQuery := `SELECT * FROM product_stocks WHERE product_id = $1 ORDER BY date_added DESC;`
	getProductPriceQuery := `SELECT * FROM product_prices WHERE product_id = $1 ORDER BY date_added DESC;`

	var modelProduct ModelProduct
	if errGetProduct := r.DB.GetContext(ctx, &modelProduct, getProductQuery, productID); errGetProduct != nil {
		r.Logger.Error(ctx, shared.AssetProduct, errGetProduct.Error())
		return dProduct.ProductDetail{}, errGetProduct
	}

	var modelStocks []ModelProductStock
	if errGetStocks := r.DB.SelectContext(ctx, &modelStocks, getProductStockQuery, productID); errGetStocks != nil && !errors.Is(errGetStocks, sql.ErrNoRows) {
		r.Logger.Error(ctx, shared.AssetProduct, errGetStocks.Error())
		return dProduct.ProductDetail{}, errGetStocks
	}

	var modelPrices []ModelProductPrice
	if errGetPrices := r.DB.SelectContext(ctx, &modelPrices, getProductPriceQuery, productID); errGetPrices != nil && !errors.Is(errGetPrices, sql.ErrNoRows) {
		r.Logger.Error(ctx, shared.AssetProduct, errGetPrices.Error())
		return dProduct.ProductDetail{}, errGetPrices
	}

	product := modelProduct.ToDomain()

	stocks := make([]dProduct.ProductStock, len(modelStocks))
	for i, m := range modelStocks {
		stocks[i] = m.ToDomain()
	}

	prices := make([]dProduct.ProductPrice, len(modelPrices))
	for i, m := range modelPrices {
		prices[i] = m.ToDomain()
	}

	return dProduct.ProductDetail{
		Product: product,
		Stocks:  stocks,
		Prices:  prices,
	}, nil
}

func (r *RepositoryImpl) GetManyProducts(ctx context.Context, farmID uuid.UUID) ([]dProduct.ProductDetail, error) {
	getQuery := `SELECT 
    		p.*, 
    		ps.*, 
    		pp.* 
		FROM products p 
		WHERE p.farm_id = $1
		LEFT JOIN LATERAL (
			SELECT ps.* FROM product_stocks ps WHERE ps.product_id = p.product_id ORDER BY ps.date_added DESC LIMIT 1
		) ps ON TRUE 
		LEFT JOIN LATERAL (
		    SELECT pp.* FROM product_prices pp WHERE pp.product_id = p.product_id ORDER BY pp.date_added DESC LIMIT 1
		) pp ON TRUE
		WHERE p.farm_id = $1;`

	type scanModel struct {
		ModelProduct
		ModelProductStock
		ModelProductPrice
	}

	var scanModels []scanModel
	if errGet := r.DB.SelectContext(ctx, &scanModels, getQuery, farmID); errGet != nil {
		r.Logger.Error(ctx, shared.AssetProduct, errGet.Error())
		return nil, errGet
	}

	productDetails := make([]dProduct.ProductDetail, len(scanModels))
	for i, sm := range scanModels {
		productDetails[i] = dProduct.ProductDetail{
			Product: sm.ModelProduct.ToDomain(),
			Stocks: []dProduct.ProductStock{
				sm.ModelProductStock.ToDomain(),
			},
			Prices: []dProduct.ProductPrice{
				sm.ModelProductPrice.ToDomain(),
			},
		}
	}

	return productDetails, nil
}

/* ==================== Delete ====================== */

func (r *RepositoryImpl) DeleteProduct(ctx context.Context, productID uuid.UUID) error {
	deleteQuery := `DELETE FROM products WHERE product_id = $1;`

	if _, errDelete := r.DB.ExecContext(ctx, deleteQuery, productID); errDelete != nil {
		r.Logger.Error(ctx, shared.AssetProduct, errDelete.Error())
		return errDelete
	}

	return nil
}

func (r *RepositoryImpl) DeleteProductStock(ctx context.Context, stockID uuid.UUID) error {
	deleteQuery := `DELETE FROM product_stocks WHERE stock_id = $1;`

	if _, errDelete := r.DB.ExecContext(ctx, deleteQuery, stockID); errDelete != nil {
		r.Logger.Error(ctx, shared.AssetProduct, errDelete.Error())
		return errDelete
	}

	return nil
}

func (r *RepositoryImpl) DeleteProductPrice(ctx context.Context, priceID uuid.UUID) error {
	deleteQuery := `DELETE FROM product_prices WHERE price_id = $1;`

	if _, errDelete := r.DB.ExecContext(ctx, deleteQuery, priceID); errDelete != nil {
		r.Logger.Error(ctx, shared.AssetProduct, errDelete.Error())
		return errDelete
	}

	return nil
}

func NewRepository(db *sqlx.DB, logger *utils.Logger) dProduct.Repository {
	return &RepositoryImpl{
		DB:     db,
		Logger: logger,
	}
}
