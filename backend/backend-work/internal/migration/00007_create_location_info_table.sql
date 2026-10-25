-- +goose Up
CREATE TABLE IF NOT EXISTS location_info (
    location_id UUID PRIMARY KEY,
    longitude NUMERIC NOT NULL,
    latitude NUMERIC NOT NULL,
    farm_id UUID REFERENCES farms(farm_id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_location_info_farm_id ON location_info (farm_id);

-- +goose Down
DROP INDEX IF EXISTS idx_location_info_farm_id;

DROP TABLE IF EXISTS location_info;
