-- +goose Up
CREATE TABLE IF NOT EXISTS products (
    product_id UUID PRIMARY KEY,
    farm_id UUID REFERENCES farms(farm_id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    image TEXT
);

CREATE TABLE IF NOT EXISTS product_stocks (
    stock_id UUID PRIMARY KEY,
    product_id UUID REFERENCES products(product_id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    stock_type VARCHAR(255) NOT NULL,
    quantity INT NOT NULL
);

CREATE TABLE IF NOT EXISTS product_prices (
    price_id UUID PRIMARY KEY,
    product_id UUID REFERENCES products(product_id)  ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    price NUMERIC(10,2) NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_products_farm_id ON products (farm_id);
CREATE INDEX IF NOT EXISTS idx_product_stocks_product_id ON product_stocks (product_id);
CREATE INDEX IF NOT EXISTS idx_product_prices_product_id ON product_prices (product_id);

-- +goose Down
DROP INDEX IF EXISTS idx_product_prices_product_id;

DROP INDEX IF EXISTS idx_product_stocks_product_id;

DROP INDEX IF EXISTS idx_products_farm_id;

DROP TABLE IF EXISTS product_prices;

DROP TABLE IF EXISTS product_stocks;

DROP TABLE IF EXISTS products;
