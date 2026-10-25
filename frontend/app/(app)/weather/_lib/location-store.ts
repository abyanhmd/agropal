'use client';

import {useMemo, useSyncExternalStore} from 'react';
import {FarmLocationSchema} from "@/app/(app)/weather/_types/weather.schema";
import {useFarm} from "@/lib/auth/auth-provider";
import {invalidate, useQuery} from "@/lib/api/use-query";
import {createLocation, listLocations, updateLocation} from "@/lib/api/resources";
import {DEMO_FARM_ID} from "@/lib/demo/base";
import {DEMO_LOCATION_NAME} from "@/lib/demo/identity";

// The farm's land location, saved through the backend's /farm/{farm_id}/location endpoints. The
// backend keeps only coordinates, so the place name from Mapbox is remembered in this browser.

const NAME_KEY = "agropal.farm-location-name";

interface SavedName {
    latitude: number;
    longitude: number;
    name: string;
}

const noopSubscribe = (listener: () => void) => {
    window.addEventListener("storage", listener);
    return () => window.removeEventListener("storage", listener);
};

function readName(): string {
    try {
        return localStorage.getItem(NAME_KEY) ?? "";
    } catch {
        return "";
    }
}

function parseName(raw: string): SavedName | null {
    try {
        const parsed = raw ? JSON.parse(raw) as SavedName : null;
        return typeof parsed?.name === "string" ? parsed : null;
    } catch {
        return null;
    }
}

const sameSpot = (a: SavedName, latitude: number, longitude: number) =>
    Math.abs(a.latitude - latitude) < 1e-6 && Math.abs(a.longitude - longitude) < 1e-6;

export function useFarmLocation() {
    const farm = useFarm()?.farm;
    const farmID = farm?.farm_id ?? null;
    const {data, loading} = useQuery(farmID && `location:${farmID}`, () => listLocations(farmID!));
    // A raw string is a stable snapshot; it is parsed below
    const rawName = useSyncExternalStore(noopSubscribe, readName, () => "");

    const saved = data?.[0];
    const savedName = parseName(rawName);

    const name = farmID === DEMO_FARM_ID
        ? DEMO_LOCATION_NAME
        : saved && savedName && sameSpot(savedName, saved.latitude, saved.longitude) ? savedName.name : undefined;

    // Memoized: effects elsewhere depend on this object, and a new one every render would re-run
    // them (and refetch) on every render
    const location = useMemo<FarmLocationSchema | null>(() => saved
        ? {ID: saved.location_id, FarmID: farmID ?? "", Latitude: saved.latitude, Longitude: saved.longitude, Name: name}
        : null, [saved, farmID, name]);

    const save = async (latitude: number, longitude: number, name?: string) => {
        if (!farmID) return;
        if (saved) await updateLocation(farmID, saved.location_id, latitude, longitude);
        else await createLocation(farmID, latitude, longitude);

        try {
            if (name) localStorage.setItem(NAME_KEY, JSON.stringify({latitude, longitude, name} satisfies SavedName));
            else localStorage.removeItem(NAME_KEY);
            window.dispatchEvent(new StorageEvent("storage", {key: NAME_KEY}));
        } catch {
            // Unnamed is fine; the coordinates still show
        }
        invalidate("location:");
    };

    // `ready` is false until the farm's saved location has been looked up
    return {location, save, ready: !loading};
}
