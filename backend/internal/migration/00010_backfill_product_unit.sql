-- +goose Up
-- Products created before units existed have none, and nothing recorded which unit each was sold by.
-- 'pcs' is the neutral choice; correct individual products with PATCH /farm/{farm_id}/product/{product_id}.
-- A product has exactly one unit, so a product sold both ways should be two products.
UPDATE products SET unit = 'pcs' WHERE unit = '';

-- +goose Down
-- The backfilled rows can't be told apart from products that were given 'pcs' on purpose, so nothing is undone.
SELECT 1;
