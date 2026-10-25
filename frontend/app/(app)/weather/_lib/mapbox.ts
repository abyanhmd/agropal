// Mapbox access token, inlined at build time. Set it in frontend/.env.local:
//   NEXT_PUBLIC_MAPBOX_TOKEN=pk.xxxx
// It ships to the browser, so use a public (pk.) token restricted to your site's URLs.
export const MAPBOX_TOKEN = process.env.NEXT_PUBLIC_MAPBOX_TOKEN ?? "";

const GEOCODE_URL = "https://api.mapbox.com/search/geocode/v6";

export interface PlaceResult {
    id: string;
    name: string;       // "Bogor"
    fullName: string;   // "Bogor, West Java, Indonesia"
    longitude: number;
    latitude: number;
}

interface GeocodeFeature {
    id: string;
    properties: {
        name: string;
        full_name?: string;
        place_formatted?: string;
        coordinates: {longitude: number, latitude: number};
    };
}

function toPlace(feature: GeocodeFeature): PlaceResult {
    const {name, full_name, place_formatted, coordinates} = feature.properties;
    return {
        id: feature.id,
        name,
        fullName: full_name ?? (place_formatted ? `${name}, ${place_formatted}` : name),
        longitude: coordinates.longitude,
        latitude: coordinates.latitude,
    };
}

async function geocode(path: string, params: Record<string, string>, signal?: AbortSignal): Promise<PlaceResult[]> {
    const query = new URLSearchParams({...params, access_token: MAPBOX_TOKEN});
    const response = await fetch(`${GEOCODE_URL}/${path}?${query}`, {signal});
    if (!response.ok) throw new Error(`Mapbox geocoding responded with ${response.status}`);
    const body = await response.json() as {features: GeocodeFeature[]};
    return body.features.map(toPlace);
}

// Search box suggestions: "bogor" → places
export function searchPlaces(text: string, signal?: AbortSignal): Promise<PlaceResult[]> {
    return geocode("forward", {q: text, limit: "5"}, signal);
}

// Pin → the nearest place name, for labeling the saved location
export async function placeAt(longitude: number, latitude: number, signal?: AbortSignal): Promise<PlaceResult | null> {
    const places = await geocode("reverse", {longitude: String(longitude), latitude: String(latitude), limit: "1"}, signal);
    return places[0] ?? null;
}
