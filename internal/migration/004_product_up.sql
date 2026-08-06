CREATE TABLE IF NOT EXISTS products (
    product_id UUID PRIMARY KEY,
    farm_id UUID REFERENCES farms(farm_id),
    name VARCHAR(255) NOT NULL,
    description TEXT,
    image TEXT
);

CREATE TABLE IF NOT EXISTS product_stocks (
    stock_id UUID PRIMARY KEY,
    product_id UUID REFERENCES products(product_id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    stock_type VARCHAR(255) NOT NULL,
    quantity INT NOT NULL
)

CREATE TABLE IF NOT EXISTS product_prices (
    price_id UUID PRIMARY KEY,
    product_id UUID REFERENCES products(product_id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    price NUMERIC(10,2) NOT NULL
)