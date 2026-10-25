'use client';

import {useEffect, useState} from 'react';
import {useFarmLocation} from "@/app/(app)/weather/_lib/location-store";
import {dateInMonth, getPlanting, PlantingCalendar, UnsupportedLocationError} from "@/app/(app)/calendar/_lib/cropgraph";

export type PlantingState =
    | {status: "loading"}
    | {status: "ready", calendar: PlantingCalendar}
    | {status: "unsupported"}
    | {status: "error", message: string};

type KeyedState = PlantingState & {key: string};

// CropGraph's planting calendar for a point on a date (YYYY-MM-DD)
export function usePlanting(latitude: number, longitude: number, date: string): PlantingState & {retry: () => void} {
    const key = `${latitude},${longitude},${date}`;
    const [state, setState] = useState<KeyedState>({status: "loading", key});
    const [attempt, setAttempt] = useState(0);

    useEffect(() => {
        let cancelled = false;
        getPlanting(latitude, longitude, date)
            .then((calendar) => !cancelled && setState({status: "ready", key, calendar}))
            .catch((error: unknown) => {
                if (cancelled) return;
                if (error instanceof UnsupportedLocationError) setState({status: "unsupported", key});
                else setState({status: "error", key, message: error instanceof Error ? error.message : "Something went wrong"});
            });
        return () => {
            cancelled = true;
        };
    }, [latitude, longitude, date, key, attempt]);

    const retry = () => {
        setState({status: "loading", key});
        setAttempt((n) => n + 1);
    };

    // A new month or location reads as loading until its own data arrives, never the old one's
    return state.key === key ? {...state, retry} : {status: "loading", retry};
}

export type CalendarAvailability = "checking" | "available" | "unavailable";

// Whether the farm's weather location has crop calendar data. Same location the weather page uses.
// A failed request counts as available, so the page can show its own retry instead of a dead link.
export function useCalendarAvailability(): CalendarAvailability {
    const {location, ready: isClient} = useFarmLocation();
    const key = location ? `${location.Latitude},${location.Longitude}` : null;
    const [result, setResult] = useState<{key: string, available: boolean} | null>(null);

    useEffect(() => {
        if (!location || !key) return;
        let cancelled = false;
        const now = new Date();
        getPlanting(location.Latitude, location.Longitude, dateInMonth(now.getFullYear(), now.getMonth()))
            .then(() => !cancelled && setResult({key, available: true}))
            .catch((error: unknown) => !cancelled && setResult({key, available: !(error instanceof UnsupportedLocationError)}));
        return () => {
            cancelled = true;
        };
    }, [location, key]);

    if (!isClient) return "checking";
    if (!location) return "unavailable";
    if (result?.key !== key) return "checking";
    return result.available ? "available" : "unavailable";
}
