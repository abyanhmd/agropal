import {
    Cloud, CloudDrizzle, CloudFog, CloudHail, CloudLightning, CloudMoon, CloudRain, CloudSnow, CloudSun, Moon, Snowflake, Sun,
} from "lucide-react";
import {ForecastSchema} from "@/app/(app)/weather/_types/weather.schema";

// Open-Meteo is free and keyless for non-commercial use, and allows browser requests (CORS)
const FORECAST_URL = "https://api.open-meteo.com/v1/forecast";

const CURRENT = ["temperature_2m", "relative_humidity_2m", "apparent_temperature", "precipitation", "weather_code", "wind_speed_10m", "wind_direction_10m", "is_day"];
const HOURLY = ["temperature_2m", "precipitation_probability", "weather_code", "is_day", "soil_temperature_0cm", "soil_moisture_0_to_1cm"];
const DAILY = ["weather_code", "temperature_2m_max", "temperature_2m_min", "precipitation_sum", "precipitation_probability_max", "et0_fao_evapotranspiration", "sunrise", "sunset", "uv_index_max", "wind_speed_10m_max"];

export async function getForecast(latitude: number, longitude: number, signal?: AbortSignal): Promise<ForecastSchema> {
    const params = new URLSearchParams({
        latitude: latitude.toFixed(4),
        longitude: longitude.toFixed(4),
        current: CURRENT.join(","),
        hourly: HOURLY.join(","),
        daily: DAILY.join(","),
        timezone: "auto",
        forecast_days: "7",
    });

    const response = await fetch(`${FORECAST_URL}?${params}`, {signal});
    if (!response.ok) {
        // Open-Meteo explains bad requests in {"reason": "..."}
        const body = await response.json().catch(() => null) as {reason?: string} | null;
        throw new Error(body?.reason ?? `Open-Meteo responded with ${response.status}`);
    }
    return response.json();
}

// ── WMO weather codes → label + icon ──
// https://open-meteo.com/en/docs#weather_variable_documentation

type WeatherIcon = typeof Sun;

const CODES: Record<number, {label: string, icon: WeatherIcon}> = {
    0: {label: "Clear sky", icon: Sun},
    1: {label: "Mainly clear", icon: CloudSun},
    2: {label: "Partly cloudy", icon: CloudSun},
    3: {label: "Overcast", icon: Cloud},
    45: {label: "Fog", icon: CloudFog},
    48: {label: "Freezing fog", icon: CloudFog},
    51: {label: "Light drizzle", icon: CloudDrizzle},
    53: {label: "Drizzle", icon: CloudDrizzle},
    55: {label: "Heavy drizzle", icon: CloudDrizzle},
    56: {label: "Freezing drizzle", icon: CloudDrizzle},
    57: {label: "Freezing drizzle", icon: CloudDrizzle},
    61: {label: "Light rain", icon: CloudRain},
    63: {label: "Rain", icon: CloudRain},
    65: {label: "Heavy rain", icon: CloudRain},
    66: {label: "Freezing rain", icon: CloudRain},
    67: {label: "Freezing rain", icon: CloudRain},
    71: {label: "Light snow", icon: CloudSnow},
    73: {label: "Snow", icon: CloudSnow},
    75: {label: "Heavy snow", icon: CloudSnow},
    77: {label: "Snow grains", icon: Snowflake},
    80: {label: "Light showers", icon: CloudRain},
    81: {label: "Showers", icon: CloudRain},
    82: {label: "Violent showers", icon: CloudRain},
    85: {label: "Snow showers", icon: CloudSnow},
    86: {label: "Heavy snow showers", icon: CloudSnow},
    95: {label: "Thunderstorm", icon: CloudLightning},
    96: {label: "Thunderstorm with hail", icon: CloudHail},
    99: {label: "Thunderstorm with hail", icon: CloudHail},
};

export function describeWeather(code: number, isDay = true): {label: string, icon: WeatherIcon} {
    const entry = CODES[code] ?? {label: "Unknown", icon: Cloud};
    // Clear and partly cloudy skies get a moon at night
    if (!isDay && entry.icon === Sun) return {...entry, icon: Moon};
    if (!isDay && entry.icon === CloudSun) return {...entry, icon: CloudMoon};
    return entry;
}

// ── Local time strings ──
// Open-Meteo times are already local to the farm, so they're read as-is instead of through
// new Date(), which would shift them into the viewer's own timezone.

// "2026-10-19T14:05" → "14:05"
export const timeOf = (local: string) => local.slice(11, 16);

const dayFormat = new Intl.DateTimeFormat("en-US", {weekday: "short", timeZone: "UTC"});
const longDayFormat = new Intl.DateTimeFormat("en-US", {weekday: "long", month: "short", day: "numeric", timeZone: "UTC"});

function asUTC(local: string): Date {
    const [year, month, day] = local.slice(0, 10).split("-").map(Number);
    return new Date(Date.UTC(year, month - 1, day));
}

// "2026-10-19" → "Mon"
export const weekdayOf = (local: string) => dayFormat.format(asUTC(local));
// "2026-10-19T14:05" → "Monday, Oct 19"
export const longDateOf = (local: string) => longDayFormat.format(asUTC(local));

// 0° → "N", 135° → "SE"
export function compassOf(degrees: number): string {
    const points = ["N", "NE", "E", "SE", "S", "SW", "W", "NW"];
    return points[Math.round(degrees / 45) % 8];
}
