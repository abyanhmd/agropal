import React from 'react';
import {Droplets, Sprout, Sun, Sunrise, Sunset, Thermometer, Wind} from "lucide-react";
import {ForecastSchema} from "@/app/(app)/weather/_types/weather.schema";
import {compassOf, describeWeather, longDateOf, timeOf, weekdayOf} from "@/app/(app)/weather/_lib/open-meteo";
import {weeklyWaterBalance} from "@/app/(app)/weather/_lib/use-forecast";

const round = (value: number) => Math.round(value);
const oneDecimal = (value: number) => (Math.round(value * 10) / 10).toFixed(1);

function Panel({title, aside, children}: {title: string, aside?: React.ReactNode, children: React.ReactNode}) {
    return (
        <section className="overflow-hidden rounded-xl border border-border bg-background">
            <div className="flex items-center justify-between gap-2 border-b border-border bg-muted px-4 py-2.5">
                <h2 className="text-xs font-medium uppercase tracking-wide text-muted-foreground">{title}</h2>
                {aside}
            </div>
            {children}
        </section>
    );
}

function Stat({icon: Icon, label, value}: {icon: typeof Wind, label: string, value: React.ReactNode}) {
    return (
        <div className="flex items-center gap-2.5">
            <div className="flex size-8 shrink-0 items-center justify-center rounded-lg bg-muted text-muted-foreground">
                <Icon className="size-4" />
            </div>
            <div className="min-w-0 font-sans">
                <p className="text-xs text-muted-foreground">{label}</p>
                <p className="text-sm font-medium tabular-nums">{value}</p>
            </div>
        </div>
    );
}

// Index of the hourly entry for the current hour, so the strip starts at "now"
function currentHourIndex(forecast: ForecastSchema): number {
    const hour = forecast.current.time.slice(0, 13); // "2026-10-19T14"
    return Math.max(0, forecast.hourly.time.findIndex((t) => t.startsWith(hour)));
}

export function CurrentConditions({forecast}: {forecast: ForecastSchema}) {
    const {current, daily} = forecast;
    const {label, icon: Icon} = describeWeather(current.weather_code, current.is_day === 1);

    return (
        <Panel title="Now" aside={<span className="text-xs text-muted-foreground font-sans">{longDateOf(current.time)} · {timeOf(current.time)} local</span>}>
            <div className="grid gap-6 p-4 md:grid-cols-[auto_1fr] md:items-center">
                <div className="flex items-center gap-4">
                    <Icon className="size-14 text-accent-strong" strokeWidth={1.5} />
                    <div className="font-sans">
                        <p className="text-4xl font-semibold tabular-nums">{round(current.temperature_2m)}°C</p>
                        <p className="text-sm text-muted-foreground">
                            {label} · feels like {round(current.apparent_temperature)}°
                        </p>
                        <p className="text-sm text-muted-foreground tabular-nums">
                            H {round(daily.temperature_2m_max[0])}° · L {round(daily.temperature_2m_min[0])}°
                        </p>
                    </div>
                </div>

                <div className="grid grid-cols-2 gap-4 sm:grid-cols-4">
                    <Stat icon={Droplets} label="Humidity" value={`${current.relative_humidity_2m}%`} />
                    <Stat icon={Wind} label="Wind" value={`${round(current.wind_speed_10m)} km/h ${compassOf(current.wind_direction_10m)}`} />
                    <Stat icon={Sunrise} label="Sunrise" value={timeOf(daily.sunrise[0])} />
                    <Stat icon={Sunset} label="Sunset" value={timeOf(daily.sunset[0])} />
                </div>
            </div>
        </Panel>
    );
}

export function HourlyForecast({forecast}: {forecast: ForecastSchema}) {
    const {hourly} = forecast;
    const start = currentHourIndex(forecast);
    const hours = hourly.time.slice(start, start + 24).map((time, i) => start + i);

    return (
        <Panel title="Next 24 hours">
            {/* w-0 min-w-full: the 24 hours scroll inside the panel instead of counting toward the
                page's minimum width, which would push the whole layout wider than the screen */}
            <ol className="flex w-0 min-w-full overflow-x-auto">
                {hours.map((i, n) => {
                    const {label, icon: Icon} = describeWeather(hourly.weather_code[i], hourly.is_day[i] === 1);
                    const rain = hourly.precipitation_probability[i];

                    return (
                        <li key={hourly.time[i]} className="flex w-16 shrink-0 flex-col items-center gap-1.5 px-1 py-3 font-sans">
                            <span className="text-xs text-muted-foreground tabular-nums">{n === 0 ? "Now" : timeOf(hourly.time[i])}</span>
                            <Icon className="size-5" aria-label={label} />
                            <span className="text-sm font-medium tabular-nums">{round(hourly.temperature_2m[i])}°</span>
                            {/* Likely rain (30%+) is highlighted; lower chances stay visible but muted */}
                            <span className={`text-xs tabular-nums ${rain >= 30 ? "text-info" : "text-muted-foreground"}`}>{rain}%</span>
                        </li>
                    );
                })}
            </ol>
        </Panel>
    );
}

export function DailyForecast({forecast}: {forecast: ForecastSchema}) {
    const {daily} = forecast;
    // Shared scale for the temperature bars, so days compare at a glance
    const weekMin = Math.min(...daily.temperature_2m_min);
    const weekMax = Math.max(...daily.temperature_2m_max);
    const span = Math.max(1, weekMax - weekMin);

    return (
        <Panel title="7-day forecast">
            <ul className="divide-y divide-border">
                {daily.time.map((day, i) => {
                    const {label, icon: Icon} = describeWeather(daily.weather_code[i]);
                    const low = daily.temperature_2m_min[i];
                    const high = daily.temperature_2m_max[i];

                    return (
                        <li key={day} className="grid grid-cols-[3rem_1.5rem_1fr_auto] items-center gap-3 px-4 py-2.5 font-sans sm:grid-cols-[3rem_1.5rem_8rem_1fr_auto]">
                            <span className="text-sm font-medium">{i === 0 ? "Today" : weekdayOf(day)}</span>
                            <Icon className="size-5" aria-label={label} />
                            <span className="hidden truncate text-sm text-muted-foreground sm:block">{label}</span>
                            <div className="flex items-center gap-2 text-sm tabular-nums">
                                <span className="w-8 text-right text-muted-foreground">{round(low)}°</span>
                                <div className="relative h-1.5 flex-1 rounded-full bg-muted">
                                    <div
                                        className="absolute inset-y-0 rounded-full bg-gradient-to-r from-info to-warning"
                                        style={{left: `${((low - weekMin) / span) * 100}%`, right: `${((weekMax - high) / span) * 100}%`}}
                                    />
                                </div>
                                <span className="w-8">{round(high)}°</span>
                            </div>
                            <span className="w-16 text-right text-xs text-info tabular-nums">
                                {daily.precipitation_sum[i] > 0 ? `${oneDecimal(daily.precipitation_sum[i])} mm` : ""}
                            </span>
                        </li>
                    );
                })}
            </ul>
        </Panel>
    );
}

// Numbers that matter for the land itself: soil, and how much water the crops lose vs. receive
export function FieldConditions({forecast}: {forecast: ForecastSchema}) {
    const {hourly} = forecast;
    const now = currentHourIndex(forecast);
    // ET₀: the water a well-watered grass crop would lose
    const {rain: weekRain, et0: weekEt0, balance} = weeklyWaterBalance(forecast);

    return (
        <Panel title="Field conditions">
            <div className="grid grid-cols-1 gap-4 p-4 sm:grid-cols-2">
                <Stat icon={Thermometer} label="Soil temperature (surface)" value={`${oneDecimal(hourly.soil_temperature_0cm[now])}°C`} />
                {/* m³/m³ volumetric water content, shown as a percentage */}
                <Stat icon={Sprout} label="Soil moisture (0–1 cm)" value={`${round(hourly.soil_moisture_0_to_1cm[now] * 100)}%`} />
                <Stat icon={Droplets} label="Rain this week" value={`${oneDecimal(weekRain)} mm`} />
                <Stat icon={Sun} label="Evaporation this week (ET₀)" value={`${oneDecimal(weekEt0)} mm`} />
            </div>
            <p className={`border-t border-border px-4 py-3 text-sm font-sans ${balance < 0 ? "text-warning-strong" : "text-accent-strong"}`}>
                {balance < 0
                    ? `Water deficit of ${oneDecimal(-balance)} mm this week: crops are expected to lose more water than rain brings.`
                    : `Rain should cover crop water use this week, with ${oneDecimal(balance)} mm to spare.`}
            </p>
        </Panel>
    );
}

export function ForecastSkeleton() {
    return (
        <div className="flex animate-pulse flex-col gap-4">
            <div className="h-40 rounded-xl bg-muted" />
            <div className="h-32 rounded-xl bg-muted" />
            <div className="grid gap-4 lg:grid-cols-2">
                <div className="h-80 rounded-xl bg-muted" />
                <div className="h-48 rounded-xl bg-muted" />
            </div>
        </div>
    );
}
