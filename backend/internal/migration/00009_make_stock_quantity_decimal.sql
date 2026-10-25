-- +goose Up
-- Stock is measured in kg, litres and so on, so it can be fractional. Whole numbers keep their value.
ALTER TABLE product_stocks ALTER COLUMN quantity TYPE NUMERIC(14,3) USING quantity::NUMERIC(14,3);

-- +goose Down
ALTER TABLE product_stocks ALTER COLUMN quantity TYPE INT USING ROUND(quantity)::INT;
