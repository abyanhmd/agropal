'use client';

import React, {useEffect, useRef, useState} from 'react';
import type {Map as MapboxMap, Marker} from "mapbox-gl";
import "mapbox-gl/dist/mapbox-gl.css";
import {useTheme} from "next-themes";
import {Layers, LoaderCircle, LocateFixed, MapPin, Pencil, Search} from "lucide-react";
import {Input} from "@/components/ui/input";
import {Button} from "@/components/ui/button";
import {FarmLocationSchema} from "@/app/(app)/weather/_types/weather.schema";
import {MAPBOX_TOKEN, PlaceResult, placeAt, searchPlaces} from "@/app/(app)/weather/_lib/mapbox";
import {useDebouncedValue} from "@/hooks/use-debounced-value";

export interface LocationPickerProps {
    saved: FarmLocationSchema | null;
    onSave: (latitude: number, longitude: number, name?: string) => void;
}

interface Pin {
    latitude: number;
    longitude: number;
}

type MapLook = "map" | "satellite";

const STYLES = {
    light: "mapbox://styles/mapbox/outdoors-v12",
    dark: "mapbox://styles/mapbox/dark-v11",
    satellite: "mapbox://styles/mapbox/satellite-streets-v12",
};

// Whole world until a location is saved
const WORLD_VIEW = {center: [0, 20] as [number, number], zoom: 1.2};
const FARM_ZOOM = 13;

export const formatCoordinates = ({latitude, longitude}: Pin) =>
    `${Math.abs(latitude).toFixed(4)}° ${latitude >= 0 ? "N" : "S"}, ${Math.abs(longitude).toFixed(4)}° ${longitude >= 0 ? "E" : "W"}`;

// Search, device location, and the map. Mounted only while picking, so the map (and its Mapbox
// map load) isn't created just to view the forecast.
function LocationEditor({saved, onSave, onClose}: LocationPickerProps & {onClose?: () => void}) {
    const {resolvedTheme} = useTheme();
    const containerRef = useRef<HTMLDivElement>(null);
    const mapRef = useRef<MapboxMap | null>(null);
    const markerRef = useRef<Marker | null>(null);
    const [mapReady, setMapReady] = useState(false);
    const [look, setLook] = useState<MapLook>("map");

    // A pin being placed but not saved yet; null shows the saved location
    const [draft, setDraft] = useState<Pin | null>(null);
    const [draftPlace, setDraftPlace] = useState<{pin: Pin, name?: string} | null>(null);
    const [locating, setLocating] = useState(false);
    const [locateError, setLocateError] = useState<string | null>(null);

    const savedPin: Pin | null = saved ? {latitude: saved.Latitude, longitude: saved.Longitude} : null;
    const pin = draft ?? savedPin;
    const style = look === "satellite" ? STYLES.satellite : resolvedTheme === "dark" ? STYLES.dark : STYLES.light;

    // ── Map setup: mapbox-gl touches window, so it's loaded only in the browser ──
    useEffect(() => {
        if (!MAPBOX_TOKEN || !containerRef.current) return;
        let cancelled = false;

        (async () => {
            const {default: mapboxgl} = await import("mapbox-gl");
            if (cancelled || !containerRef.current) return;

            mapboxgl.accessToken = MAPBOX_TOKEN;
            const start = savedPin
                ? {center: [savedPin.longitude, savedPin.latitude] as [number, number], zoom: FARM_ZOOM}
                : WORLD_VIEW;
            const map = new mapboxgl.Map({container: containerRef.current, style, ...start});
            map.addControl(new mapboxgl.NavigationControl({showCompass: false}), "top-right");
            map.on("click", (e) => setDraft({latitude: e.lngLat.lat, longitude: e.lngLat.lng}));

            const marker = new mapboxgl.Marker({draggable: true, color: "#4a7a4d"});
            marker.on("dragend", () => {
                const {lat, lng} = marker.getLngLat();
                setDraft({latitude: lat, longitude: lng});
            });

            mapRef.current = map;
            markerRef.current = marker;
            map.on("load", () => setMapReady(true));
        })();

        return () => {
            cancelled = true;
            mapRef.current?.remove();
            mapRef.current = null;
            markerRef.current = null;
            setMapReady(false);
        };
        // Created once; style and pin changes are applied by the effects below
        // eslint-disable-next-line react-hooks/exhaustive-deps
    }, []);

    useEffect(() => {
        if (mapReady) mapRef.current?.setStyle(style);
    }, [style, mapReady]);

    // Keep the marker on the current pin
    const pinKey = pin ? `${pin.latitude},${pin.longitude}` : "";
    useEffect(() => {
        const map = mapRef.current;
        const marker = markerRef.current;
        if (!mapReady || !map || !marker) return;
        if (pin) marker.setLngLat([pin.longitude, pin.latitude]).addTo(map);
        else marker.remove();
        // eslint-disable-next-line react-hooks/exhaustive-deps
    }, [pinKey, mapReady]);

    // The saved location comes from localStorage after hydration, usually after the map was
    // created on the world view, so move to it once it's known (and whenever it changes)
    const savedKey = savedPin ? `${savedPin.latitude},${savedPin.longitude}` : "";
    useEffect(() => {
        const map = mapRef.current;
        if (!mapReady || !map || !savedPin) return;
        map.jumpTo({center: [savedPin.longitude, savedPin.latitude], zoom: Math.max(map.getZoom(), FARM_ZOOM)});
        // eslint-disable-next-line react-hooks/exhaustive-deps
    }, [savedKey, mapReady]);

    const flyTo = (target: Pin) => mapRef.current?.flyTo({center: [target.longitude, target.latitude], zoom: Math.max(mapRef.current.getZoom(), FARM_ZOOM)});

    // ── Name the draft pin (reverse geocoding) ──
    const draftKey = draft ? `${draft.latitude},${draft.longitude}` : "";
    const debouncedDraftKey = useDebouncedValue(draftKey, 400);
    useEffect(() => {
        if (!MAPBOX_TOKEN || !draft || debouncedDraftKey !== draftKey) return;
        if (draftPlace?.pin === draft) return; // already named, e.g. picked from search
        const controller = new AbortController();
        placeAt(draft.longitude, draft.latitude, controller.signal)
            .then((place) => setDraftPlace({pin: draft, name: place?.fullName}))
            .catch(() => {
                // Unnamed is fine; the coordinates still show
            });
        return () => controller.abort();
        // eslint-disable-next-line react-hooks/exhaustive-deps
    }, [debouncedDraftKey]);
    const draftName = draftPlace && draft && draftPlace.pin === draft ? draftPlace.name : undefined;

    // ── Search ──
    const [query, setQuery] = useState("");
    const [results, setResults] = useState<PlaceResult[]>([]);
    const [resultsOpen, setResultsOpen] = useState(false);
    const debouncedQuery = useDebouncedValue(query.trim(), 300);

    useEffect(() => {
        if (!MAPBOX_TOKEN || debouncedQuery.length < 2) return;
        const controller = new AbortController();
        searchPlaces(debouncedQuery, controller.signal)
            .then((places) => {
                setResults(places);
                setResultsOpen(true);
            })
            .catch(() => setResults([]));
        return () => controller.abort();
    }, [debouncedQuery]);

    const pickResult = (place: PlaceResult) => {
        const next = {latitude: place.latitude, longitude: place.longitude};
        setDraft(next);
        setDraftPlace({pin: next, name: place.fullName});
        setQuery(place.name);
        setResultsOpen(false);
        flyTo(next);
    };

    // ── Device location ──
    const locateMe = () => {
        if (!navigator.geolocation) {
            setLocateError("This browser can't share its location.");
            return;
        }
        setLocating(true);
        setLocateError(null);
        navigator.geolocation.getCurrentPosition(
            ({coords}) => {
                const next = {latitude: coords.latitude, longitude: coords.longitude};
                setDraft(next);
                flyTo(next);
                setLocating(false);
            },
            (error) => {
                setLocateError(error.code === error.PERMISSION_DENIED
                    ? "Location access was blocked. Allow it in the browser, or pick the spot on the map."
                    : "Couldn't get your location. Pick the spot on the map instead.");
                setLocating(false);
            },
            {enableHighAccuracy: true, timeout: 10_000},
        );
    };

    const save = () => {
        if (!draft) return;
        onSave(draft.latitude, draft.longitude, draftName);
        setDraft(null);
        onClose?.();
    };

    const discard = () => {
        setDraft(null);
        if (savedPin) flyTo(savedPin);
    };

    return (
        <>
            <p className="text-sm text-muted-foreground font-sans">
                Click the map or drag the pin to where your land is, then save. Weather and the crop calendar follow the saved pin.
            </p>

            <div className="flex flex-col gap-2 sm:flex-row">
                <div className="relative flex-1">
                    <Search className="pointer-events-none absolute left-2.5 top-1/2 size-4 -translate-y-1/2 text-muted-foreground" />
                    <Input
                        type="search"
                        placeholder="Search a village, town, or address…"
                        aria-label="Search places"
                        value={query}
                        disabled={!MAPBOX_TOKEN}
                        onChange={(e) => {
                            setQuery(e.target.value);
                            if (e.target.value.trim().length < 2) setResultsOpen(false);
                        }}
                        onFocus={() => results.length > 0 && setResultsOpen(true)}
                        onKeyDown={(e) => {
                            if (e.key === "Escape") setResultsOpen(false);
                            if (e.key === "Enter" && results[0]) pickResult(results[0]);
                        }}
                        className="h-9 pl-8 rounded-xl"
                    />
                    {resultsOpen && results.length > 0 &&
                        <ul className="absolute inset-x-0 top-full z-20 mt-1 overflow-hidden rounded-xl border border-border bg-popover py-1 shadow-md">
                            {results.map((place) => (
                                <li key={place.id}>
                                    <button
                                        type="button"
                                        onClick={() => pickResult(place)}
                                        className="flex w-full cursor-pointer items-start gap-2 px-3 py-2 text-left text-sm font-sans outline-none hover:bg-muted focus-visible:bg-muted"
                                    >
                                        <MapPin className="mt-0.5 size-4 shrink-0 text-muted-foreground" />
                                        <span className="min-w-0">
                                            <span className="block truncate font-medium">{place.name}</span>
                                            <span className="block truncate text-xs text-muted-foreground">{place.fullName}</span>
                                        </span>
                                    </button>
                                </li>
                            ))}
                        </ul>
                    }
                </div>

                <Button variant="outline" className="h-9 rounded-xl bg-background" onClick={locateMe} disabled={locating}>
                    {locating ? <LoaderCircle className="animate-spin" /> : <LocateFixed />}
                    Use my location
                </Button>
            </div>
            {locateError && <p className="text-sm text-danger-strong font-sans">{locateError}</p>}

            {MAPBOX_TOKEN
                ? <div className="relative h-80 overflow-hidden rounded-xl border border-border md:h-96">
                    {/* Sized with h-full, not absolute inset-0: mapbox-gl.css sets .mapboxgl-map to
                        position: relative, which would win over the utility and collapse it to 0px */}
                    <div ref={containerRef} className="h-full w-full" />
                    {!mapReady && <div className="absolute inset-0 animate-pulse bg-muted" />}
                    <Button
                        variant="outline"
                        size="sm"
                        aria-pressed={look === "satellite"}
                        onClick={() => setLook(look === "satellite" ? "map" : "satellite")}
                        className="absolute left-2 top-2 bg-background/90 backdrop-blur dark:bg-background/90"
                    >
                        <Layers />
                        {look === "satellite" ? "Map" : "Satellite"}
                    </Button>
                </div>
                : <div className="flex flex-col items-center gap-2 rounded-xl border border-dashed border-border px-4 py-10 text-center text-sm text-muted-foreground font-sans">
                    <MapPin className="size-7" />
                    <p>The map needs a Mapbox token.</p>
                    <p>
                        Add <code className="rounded bg-muted px-1 py-0.5 font-mono text-xs text-foreground">NEXT_PUBLIC_MAPBOX_TOKEN=pk.…</code> to{" "}
                        <code className="rounded bg-muted px-1 py-0.5 font-mono text-xs text-foreground">frontend/.env.local</code> and restart the dev server.
                        Until then, &ldquo;Use my location&rdquo; still works.
                    </p>
                </div>
            }

            {/* Pending pin: confirm or discard */}
            {draft &&
                <div className="flex flex-col gap-3 rounded-xl border border-accent/40 bg-accent-muted/50 p-3 sm:flex-row sm:items-center">
                    <div className="min-w-0 flex-1 font-sans">
                        <p className="truncate text-sm font-medium">{draftName ?? "New location"}</p>
                        <p className="text-xs text-muted-foreground tabular-nums">{formatCoordinates(draft)}</p>
                    </div>
                    <div className="flex gap-2">
                        <Button variant="ghost" size="sm" onClick={discard}>Discard pin</Button>
                        <Button size="sm" onClick={save}>Save location</Button>
                    </div>
                </div>
            }
        </>
    );
}

function SavedLocation({location}: {location: FarmLocationSchema}) {
    return (
        <div className="flex min-w-0 items-center gap-2 font-sans">
            <MapPin className="size-4 shrink-0 text-accent-strong" />
            <div className="min-w-0">
                <p className="truncate text-sm font-medium">{location.Name ?? "Saved location"}</p>
                <p className="text-xs text-muted-foreground tabular-nums">{formatCoordinates({latitude: location.Latitude, longitude: location.Longitude})}</p>
            </div>
        </div>
    );
}

export function LocationPickerSkeleton() {
    return (
        <section className="flex animate-pulse flex-col gap-3 rounded-xl border border-border bg-background p-4">
            <div className="h-6 w-36 rounded bg-muted" />
            <div className="h-9 w-64 rounded bg-muted" />
        </section>
    );
}

// The map shows until a location is saved; after that it's a summary with an Edit button
export default function LocationPicker({saved, onSave}: LocationPickerProps) {
    const [editing, setEditing] = useState(false);

    return (
        <section className="flex flex-col gap-3 rounded-xl border border-border bg-background p-4">
            <div className="flex items-center justify-between gap-3">
                <h2 className="font-serif text-lg font-medium">Land location</h2>
                {saved && (editing
                    ? <Button variant="ghost" size="sm" onClick={() => setEditing(false)}>Cancel</Button>
                    : <Button variant="outline" size="sm" onClick={() => setEditing(true)}>
                        <Pencil />
                        Edit
                    </Button>
                )}
            </div>

            {saved && !editing
                ? <SavedLocation location={saved} />
                : <LocationEditor saved={saved} onSave={onSave} onClose={() => setEditing(false)} />
            }
        </section>
    );
}
