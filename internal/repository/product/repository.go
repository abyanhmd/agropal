package product

import (
	"agropal/internal/database"
	dProduct "agropal/internal/domain/product"
	"agropal/internal/domain/shared"
	"agropal/internal/utils"
	"context"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

/* ==================== Create ====================== */

type RepositoryImpl struct {
	Logger *utils.Logger
}

func NewRepository(logger *utils.Logger) dProduct.Repository {
	return &RepositoryImpl{Logger: logger}
}

func (r *RepositoryImpl) Create(ctx context.Context, q database.Querier, p dProduct.Product) error {
	query := `
		INSERT INTO products (
			product_id, farm_id, name,
			description, image
		) VALUES (
			:product_id, :farm_id, :name,
			:description, :image
		);`

	modelProduct := ToProductModel(p)
	if _, err := q.NamedExecContext(ctx, query, modelProduct); err != nil {
		r.Logger.Error(ctx, shared.AssetProduct, err.Error())
		return err
	}

	return nil
}

func (r *RepositoryImpl) Update(ctx context.Context, q database.Querier, p dProduct.Product) error {
	query := `
		UPDATE products SET
		    name = :name,
		    description = :description,
		    image = :image
		WHERE product_id = :product_id;`

	modelProduct := ToProductModel(p)
	if _, err := q.NamedExecContext(ctx, query, modelProduct); err != nil {
		r.Logger.Error(ctx, shared.AssetProduct, err.Error())
		return err
	}

	return nil
}

func (r *RepositoryImpl) Delete(ctx context.Context, q database.Querier, id uuid.UUID) error {
	query := `DELETE FROM products WHERE product_id = $1;`

	if _, err := q.ExecContext(ctx, query, id); err != nil {
		r.Logger.Error(ctx, shared.AssetProduct, err.Error())
		return err
	}

	return nil
}

func (r *RepositoryImpl) FindByID(ctx context.Context, q database.Querier, id uuid.UUID) (dProduct.Product, error) {
	query := `SELECT * FROM products WHERE product_id = $1;`

	var modelProduct ProductModel
	if err := q.GetContext(ctx, &modelProduct, query, id); err != nil {
		r.Logger.Error(ctx, shared.AssetProduct, err.Error())
		return dProduct.Product{}, err
	}

	return modelProduct.ToDomain(), nil
}

func (r *RepositoryImpl) FindDetailByID(ctx context.Context, q database.Querier, id uuid.UUID) (dProduct.ProductDetail, error) {
	productQuery := `SELECT * FROM products WHERE product_id = $1;`
	stockQuery := `SELECT * FROM product_stocks WHERE product_id = $1;`
	priceQuery := `SELECT * FROM product_prices WHERE product_id = $1;`

	var productModel ProductModel
	if err := q.GetContext(ctx, &productModel, productQuery, id); err != nil {
		r.Logger.Error(ctx, shared.AssetProduct, err.Error())
		return dProduct.ProductDetail{}, err
	}

	var stocks []dProduct.ProductStock
	var stockModels []ProductStockModel
	if err := q.SelectContext(ctx, &stockModels, stockQuery, id); err != nil {
		r.Logger.Error(ctx, shared.AssetProduct, err.Error())
		return dProduct.ProductDetail{}, err
	}

	var prices []dProduct.ProductPrice
	var priceModels []ProductPriceModel
	if err := q.SelectContext(ctx, &priceModels, priceQuery, id); err != nil {
		r.Logger.Error(ctx, shared.AssetProduct, err.Error())
		return dProduct.ProductDetail{}, err
	}

	for i, m := range priceModels {
		prices[i] = m.ToDomain()
	}

	for i, m := range stockModels {
		stocks[i] = m.ToDomain()
	}

	return dProduct.ProductDetail{
		Product: productModel.ToDomain(),
		Prices:  prices,
		Stocks:  stocks,
	}, nil
}

func (r *RepositoryImpl) ListByFarmID(ctx context.Context, q database.Querier, farmID uuid.UUID) ([]dProduct.ProductDetail, error) {
	var productModels []ProductModel
	if err := q.SelectContext(ctx, &productModels, `SELECT * FROM products WHERE farm_id = $1;`, farmID); err != nil {
		r.Logger.Error(ctx, shared.AssetProduct, err.Error())
		return nil, err
	}
	if len(productModels) == 0 {
		return []dProduct.ProductDetail{}, nil
	}

	productIDs := make([]uuid.UUID, len(productModels))
	for i, m := range productModels {
		productIDs[i] = m.ID
	}

	var stockModels []ProductStockModel
	var priceModels []ProductPriceModel

	queryStocks, argsStocks, _ := sqlx.In(`SELECT * FROM product_stocks WHERE product_id IN (?)`, productIDs)
	queryPrices, argsPrices, _ := sqlx.In(`SELECT * FROM product_prices WHERE product_id IN (?)`, productIDs)

	if err := q.SelectContext(ctx, &stockModels, q.Rebind(queryStocks), argsStocks...); err != nil {
		return nil, err
	}
	if err := q.SelectContext(ctx, &priceModels, q.Rebind(queryPrices), argsPrices...); err != nil {
		return nil, err
	}

	stockMap := make(map[uuid.UUID][]dProduct.ProductStock)
	for _, sm := range stockModels {
		stockMap[sm.ProductID] = append(stockMap[sm.ProductID], sm.ToDomain())
	}

	priceMap := make(map[uuid.UUID][]dProduct.ProductPrice)
	for _, pm := range priceModels {
		priceMap[pm.ProductID] = append(priceMap[pm.ProductID], pm.ToDomain())
	}

	results := make([]dProduct.ProductDetail, len(productModels))
	for i, m := range productModels {
		results[i] = dProduct.ProductDetail{
			Product: m.ToDomain(),
			Stocks:  stockMap[m.ID],
			Prices:  priceMap[m.ID],
		}
	}

	return results, nil
}

func (r *RepositoryImpl) AddStock(ctx context.Context, q database.Querier, s dProduct.ProductStock) error {
	query := `
		INSERT INTO product_stocks (
			stock_id, product_id,
			created_at, type,
			quantity
		) VALUES (
			:stock_id, :product_id, 
			:created_at, :type, 
		  	:quantity
		);`

	stockModel := ToStockModel(s)
	if _, err := q.NamedExecContext(ctx, query, stockModel); err != nil {
		r.Logger.Error(ctx, shared.AssetProduct, err.Error())
		return err
	}

	return nil
}

func (r *RepositoryImpl) AddPrice(ctx context.Context, q database.Querier, p dProduct.ProductPrice) error {
	query := `
		INSERT INTO product_prices (
			price_id, product_id,
			created_at, price
		) VALUES (
			:price_id, :product_id,
			:created_at, :price
		);`

	priceModel := ToPriceModel(p)
	if _, err := q.NamedExecContext(ctx, query, priceModel); err != nil {
		r.Logger.Error(ctx, shared.AssetProduct, err.Error())
		return err
	}

	return nil
}

func (r *RepositoryImpl) DeleteStock(ctx context.Context, q database.Querier, stockID uuid.UUID) error {
	query := `DELETE FROM product_stocks WHERE stock_id = $1;`

	if _, err := q.ExecContext(ctx, query, stockID); err != nil {
		r.Logger.Error(ctx, shared.AssetProduct, err.Error())
		return err
	}

	return nil
}

func (r *RepositoryImpl) DeletePrice(ctx context.Context, q database.Querier, priceID uuid.UUID) error {
	query := `DELETE FROM product_prices WHERE price_id = $1;`

	if _, err := q.ExecContext(ctx, query, priceID); err != nil {
		r.Logger.Error(ctx, shared.AssetProduct, err.Error())
		return err
	}

	return nil
}
