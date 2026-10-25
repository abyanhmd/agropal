-- +goose Up
-- A product is sold by one unit (kg, pcs, ...); rows from before this migration have none.
ALTER TABLE products ADD COLUMN IF NOT EXISTS unit VARCHAR(50) NOT NULL DEFAULT '';

-- Prices carry their currency like transactions do; earlier prices were all recorded in USD.
ALTER TABLE product_prices ADD COLUMN IF NOT EXISTS currency VARCHAR(10) NOT NULL DEFAULT 'USD';

-- Stock movements use the same INCOME/EXPENSE vocabulary as transactions.
UPDATE product_stocks SET stock_type = 'INCOME' WHERE stock_type = 'IN';
UPDATE product_stocks SET stock_type = 'EXPENSE' WHERE stock_type = 'OUT';

-- +goose Down
UPDATE product_stocks SET stock_type = 'OUT' WHERE stock_type = 'EXPENSE';
UPDATE product_stocks SET stock_type = 'IN' WHERE stock_type = 'INCOME';

ALTER TABLE product_prices DROP COLUMN IF EXISTS currency;

ALTER TABLE products DROP COLUMN IF EXISTS unit;
