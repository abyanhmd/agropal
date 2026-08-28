-- +goose Up
CREATE TABLE IF NOT EXISTS transactions (
    transaction_id UUID PRIMARY KEY,
    transaction_date TIMESTAMPTZ NOT NULL,
    transaction_type VARCHAR(50) NOT NULL,
    amount NUMERIC(10,2) NOT NULL,
    currency VARCHAR(10) NOT NULL,
    description VARCHAR(255),
    product_id UUID REFERENCES products(product_id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_transactions_product_id ON transactions (product_id);

-- +goose Down
DROP INDEX IF EXISTS idx_transactions_product_id;

DROP TABLE IF EXISTS transactions;
