'use client';

import React from 'react';
import Link from "next/link";
import {ArrowRight, CloudSun, Droplets, MapPin, RotateCw} from "lucide-react";
import {Button, buttonVariants} from "@/components/ui/button";
import {useFarmLocation} from "@/app/(app)/weather/_lib/location-store";
import {useForecast, weeklyWaterBalance} from "@/app/(app)/weather/_lib/use-forecast";
import {describeWeather, weekdayOf} from "@/app/(app)/weather/_lib/open-meteo";
import {useAppHref} from "@/lib/demo/base";

// Days shown after today in the mini forecast
const UPCOMING_DAYS = 4;

function WeatherCardShell({subtitle, children}: {subtitle?: string, children: React.ReactNode}) {
    // @container: the card lays out by its own width, since the sidebar takes a quarter of the window
    const appHref = useAppHref();
    return (
        <section className="@container flex flex-col rounded-xl border border-border bg-background font-sans">
            <header className="flex items-center justify-between gap-3 border-b border-border px-5 py-3">
                <div className="flex min-w-0 items-baseline gap-2">
                    <h2 className="font-medium">Weather</h2>
                    {subtitle && <span className="truncate text-sm text-muted-foreground">{subtitle}</span>}
                </div>
                <Link href={appHref("/weather")} className="inline-flex shrink-0 items-center gap-1 text-sm text-muted-foreground transition-colors hover:text-foreground">
                    Full forecast <ArrowRight className="size-4" />
                </Link>
            </header>
            {children}
        </section>
    );
}

function WeatherCardSkeleton() {
    return (
        <div className="flex animate-pulse flex-col gap-4 p-5 @3xl:flex-row @3xl:items-center">
            <div className="flex items-center gap-3">
                <div className="size-12 rounded-full bg-muted" />
                <div className="flex flex-col gap-2">
                    <div className="h-7 w-20 rounded bg-muted" />
                    <div className="h-3 w-28 rounded bg-muted" />
                </div>
            </div>
            <div className="h-16 flex-1 rounded-lg bg-muted" />
        </div>
    );
}

function LocationForecast({latitude, longitude}: {latitude: number, longitude: number}) {
    const state = useForecast(latitude, longitude);

    if (state.status === "loading") return <WeatherCardSkeleton />;

    if (state.status === "error") {
        return (
            <div className="flex flex-col items-center gap-2 px-5 py-6 text-center text-sm text-muted-foreground sm:flex-row sm:justify-center">
                <span>Couldn&apos;t load the forecast.</span>
                <Button variant="ghost" size="sm" onClick={state.retry}>
                    <RotateCw />
                    Try again
                </Button>
            </div>
        );
    }

    const {current, daily} = state.forecast;
    const {label, icon: Icon} = describeWeather(current.weather_code, current.is_day === 1);
    const {balance} = weeklyWaterBalance(state.forecast);
    const upcoming = daily.time.slice(1, 1 + UPCOMING_DAYS).map((day, n) => ({day, i: n + 1}));

    return (
        <div className="grid gap-5 p-5 @3xl:grid-cols-[auto_1fr_auto] @3xl:items-center">
            <div className="flex items-center gap-3">
                <Icon className="size-12 shrink-0 text-accent-strong" strokeWidth={1.5} aria-hidden />
                <div>
                    {/* Proportional figures for big standalone numbers, like the stat tiles */}
                    <p className="font-serif text-3xl font-semibold">{Math.round(current.temperature_2m)}°C</p>
                    <p className="text-sm text-muted-foreground">
                        {label} · H {Math.round(daily.temperature_2m_max[0])}° L {Math.round(daily.temperature_2m_min[0])}°
                    </p>
                </div>
            </div>

            <ol className="grid grid-cols-4 gap-2 @3xl:border-x @3xl:border-border @3xl:px-5">
                {upcoming.map(({day, i}) => {
                    const dayWeather = describeWeather(daily.weather_code[i]);
                    return (
                        <li key={day} className="flex flex-col items-center gap-1 text-center">
                            <span className="text-xs text-muted-foreground">{weekdayOf(day)}</span>
                            <dayWeather.icon className="size-5" aria-label={dayWeather.label} />
                            <span className="text-sm tabular-nums">
                                {Math.round(daily.temperature_2m_max[i])}°
                                <span className="text-muted-foreground"> {Math.round(daily.temperature_2m_min[i])}°</span>
                            </span>
                        </li>
                    );
                })}
            </ol>

            <div className="flex flex-col gap-1 text-sm @3xl:items-end @3xl:text-right">
                <span className="inline-flex items-center gap-1.5 text-muted-foreground tabular-nums">
                    <Droplets className="size-4 text-info" aria-hidden />
                    {daily.precipitation_probability_max[0]}% chance of rain today
                </span>
                <span className={`tabular-nums ${balance < 0 ? "text-warning-strong" : "text-accent-strong"}`}>
                    {balance < 0
                        ? `${(-balance).toFixed(1)} mm water deficit this week`
                        : `Rain covers crop water use this week`}
                </span>
            </div>
        </div>
    );
}

export default function WeatherCard() {
    const appHref = useAppHref();
    const {location, ready: isClient} = useFarmLocation();

    if (!isClient) {
        return <WeatherCardShell><WeatherCardSkeleton /></WeatherCardShell>;
    }

    if (!location) {
        return (
            <WeatherCardShell>
                <div className="flex flex-col items-center gap-3 px-5 py-6 text-center text-sm text-muted-foreground sm:flex-row sm:justify-center sm:text-left">
                    <CloudSun className="size-6 shrink-0" aria-hidden />
                    <span>Set your land&apos;s location to see its weather here.</span>
                    <Link href={appHref("/settings")} className={buttonVariants({variant: "outline", size: "sm"})}>
                        <MapPin />
                        Set location
                    </Link>
                </div>
            </WeatherCardShell>
        );
    }

    return (
        <WeatherCardShell subtitle={location.Name}>
            <LocationForecast key={`${location.Latitude},${location.Longitude}`} latitude={location.Latitude} longitude={location.Longitude} />
        </WeatherCardShell>
    );
}
