// CropGraph is free and keyless, and allows browser requests (CORS). Its data only covers the US
// (incl. AK, HI, PR); any other coordinate answers 400 INVALID_COORDINATES.
const PLANTING_URL = "https://api.cropgraph.com/api/planting";

export interface PlantingCrop {
    slug: string;
    commonName: string;
    scientificName: string;
    category: string;
    action: string;
    windowStart: string;
    windowEnd: string;
    daysToHarvest: {min: number, max: number};
    expectedHarvestEarliest?: string;
    notes?: string;
}

export interface PlantingCalendar {
    zone: {zone: string};
    climateType: string;
    frostDates: {lastSpring: string, firstFall: string, seasonDays: number};
    asOf: string;
    plantNow: PlantingCrop[];
}

// The location is outside CropGraph's coverage, as opposed to a failed request
export class UnsupportedLocationError extends Error {
    constructor() {
        super("CropGraph has no data for this location");
    }
}

const cache = new Map<string, Promise<PlantingCalendar>>();

// What to plant at a point on a date (YYYY-MM-DD). Answers are cached per ~1 km cell, since the
// API is rate limited (60 requests a minute).
export function getPlanting(latitude: number, longitude: number, date: string): Promise<PlantingCalendar> {
    const lat = latitude.toFixed(2);
    const lng = longitude.toFixed(2);
    const key = `${lat},${lng},${date}`;

    const cached = cache.get(key);
    if (cached) return cached;

    const request = fetchPlanting(lat, lng, date);
    cache.set(key, request);
    // Failures aren't kept, so retrying asks again
    request.catch(() => cache.delete(key));
    return request;
}

async function fetchPlanting(lat: string, lng: string, date: string): Promise<PlantingCalendar> {
    const response = await fetch(`${PLANTING_URL}?${new URLSearchParams({lat, lng, date})}`);
    if (response.ok) return response.json();

    const body = await response.json().catch(() => null) as {code?: string, message?: string} | null;
    if (body?.code === "INVALID_COORDINATES") throw new UnsupportedLocationError();
    throw new Error(body?.message ?? `CropGraph responded with ${response.status}`);
}

export function dateInMonth(year: number, month: number): string {
    return `${year}-${String(month + 1).padStart(2, "0")}-15`;
}

export const ACTION_LABELS: Record<string, string> = {
    start_indoors: "Start indoors",
    direct_sow: "Direct sow",
    transplant: "Transplant",
    plant_now: "Plant",
};

// Shown in a tooltip on each group heading
export const ACTION_HELP: Record<string, string> = {
    start_indoors: "Sow seeds in trays indoors, then move the seedlings outside once the weather allows.",
    direct_sow: "Sow seeds straight into the ground or bed where they will grow.",
    transplant: "Move seedlings that were started indoors into the garden.",
    plant_now: "Good to plant right now, by seed or seedling.",
};
