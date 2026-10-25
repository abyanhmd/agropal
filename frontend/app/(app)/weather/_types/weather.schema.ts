// Mirrors the backend Weather entity (location_info table): one land location per farm
export interface FarmLocationSchema {
    ID: string,
    Longitude: number,
    Latitude: number,
    FarmID: string,
    Name?: string, // frontend only: place name from Mapbox reverse geocoding, for display
}

export type FarmLocationSchemaType = FarmLocationSchema;

// Open-Meteo forecast response, limited to the fields this page requests.
// Times are local to the location (timezone=auto) and have no offset, e.g. "2026-10-19T14:00".
export interface ForecastSchema {
    timezone: string,
    elevation: number,
    current: {
        time: string,
        temperature_2m: number,
        relative_humidity_2m: number,
        apparent_temperature: number,
        precipitation: number,
        weather_code: number,
        wind_speed_10m: number,
        wind_direction_10m: number,
        is_day: number,
    },
    hourly: {
        time: string[],
        temperature_2m: number[],
        precipitation_probability: number[],
        weather_code: number[],
        is_day: number[],
        soil_temperature_0cm: number[],
        soil_moisture_0_to_1cm: number[],
    },
    daily: {
        time: string[],
        weather_code: number[],
        temperature_2m_max: number[],
        temperature_2m_min: number[],
        precipitation_sum: number[],
        precipitation_probability_max: number[],
        et0_fao_evapotranspiration: number[],
        sunrise: string[],
        sunset: string[],
        uv_index_max: number[],
        wind_speed_10m_max: number[],
    },
}

export type ForecastSchemaType = ForecastSchema;
