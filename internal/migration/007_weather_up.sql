CREATE TABLE IF NOT EXISTS location_info (
    location_id UUID PRIMARY KEY,
    longitude NUMERIC NOT NULL,
    latitude NUMERIC NOT NULL,
    farm_id UUID REFERENCES farms(farm_id)
);