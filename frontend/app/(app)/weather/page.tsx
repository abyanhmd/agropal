'use client';

import React from 'react';
import Link from "next/link";
import {CloudSun, MapPin, RotateCw} from "lucide-react";
import {Button, buttonVariants} from "@/components/ui/button";
import {useFarmLocation} from "@/app/(app)/weather/_lib/location-store";
import {useForecast} from "@/app/(app)/weather/_lib/use-forecast";
import {CurrentConditions, DailyForecast, FieldConditions, ForecastSkeleton, HourlyForecast} from "@/app/(app)/weather/_components/forecast.weather";
import {useAppHref} from "@/lib/demo/base";

function Forecast({latitude, longitude}: {latitude: number, longitude: number}) {
    const state = useForecast(latitude, longitude);

    if (state.status === "loading") return <ForecastSkeleton />;

    if (state.status === "error") {
        return (
            <div className="flex flex-col items-center gap-3 rounded-xl border border-dashed border-border py-10 text-center text-sm text-muted-foreground font-sans">
                <p>Couldn&apos;t load the forecast: {state.message}</p>
                <Button variant="outline" size="sm" onClick={state.retry}>
                    <RotateCw />
                    Try again
                </Button>
            </div>
        );
    }

    const {forecast} = state;
    return (
        <div className="flex flex-col gap-4">
            <CurrentConditions forecast={forecast} />
            <HourlyForecast forecast={forecast} />
            <div className="grid gap-4 lg:grid-cols-2">
                <DailyForecast forecast={forecast} />
                <FieldConditions forecast={forecast} />
            </div>
            <p className="text-xs text-muted-foreground font-sans">
                Weather data by <a href="https://open-meteo.com/" target="_blank" rel="noreferrer" className="underline underline-offset-4 hover:text-foreground">Open-Meteo</a>
                {" "}· {forecast.timezone} · {Math.round(forecast.elevation)} m above sea level
            </p>
        </div>
    );
}

export default function WeatherPage() {
    const appHref = useAppHref();
    const {location, ready: isClient} = useFarmLocation();

    return (
        <div className="page-container flex flex-col gap-4">
            <div className="flex flex-col gap-1">
                <h1 className="font-serif text-3xl font-medium">Weather</h1>
                <p className="font-sans text-muted-foreground">
                    {location ? `Forecast for ${location.Name ?? "your land"}.` : "Forecast for your land."}
                    {isClient && location && <>{" "}<Link href={appHref("/settings")} className="underline underline-offset-4 hover:text-foreground">Change location</Link></>}
                </p>
            </div>

            {!isClient
                ? <ForecastSkeleton />
                : location
                    // keyed, so a new location starts from a fresh loading state
                    ? <Forecast key={`${location.Latitude},${location.Longitude}`} latitude={location.Latitude} longitude={location.Longitude} />
                    : <div className="flex flex-col items-center gap-2 rounded-xl border border-dashed border-border py-12 text-center text-muted-foreground">
                        <CloudSun className="size-8" />
                        <p className="font-sans text-sm">Set your land&apos;s location to see its weather.</p>
                        <Link href={appHref("/settings")} className={buttonVariants({variant: "outline", size: "sm"})}>
                            <MapPin />
                            Set location
                        </Link>
                    </div>
            }
        </div>
    );
}
