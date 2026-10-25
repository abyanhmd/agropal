'use client';

import React, {useState} from 'react';
import Link from "next/link";
import {CalendarDays, ChevronLeft, ChevronRight, List, MapPin, RotateCw} from "lucide-react";
import {Button, buttonVariants} from "@/components/ui/button";
import {useFarmLocation} from "@/app/(app)/weather/_lib/location-store";
import {usePlanting} from "@/app/(app)/calendar/_lib/use-calendar";
import {Tooltip, TooltipContent, TooltipTrigger} from "@/components/ui/tooltip";
import {ACTION_HELP, ACTION_LABELS, dateInMonth, PlantingCrop} from "@/app/(app)/calendar/_lib/cropgraph";
import {useAppHref} from "@/lib/demo/base";

const MONTHS = ["Jan", "Feb", "Mar", "Apr", "May", "Jun", "Jul", "Aug", "Sep", "Oct", "Nov", "Dec"];
const DAY_MS = 24 * 60 * 60 * 1000;

type View = "list" | "calendar";

const VIEWS: {value: View, label: string, icon: React.ComponentType<{className?: string}>}[] = [
    {value: "list", label: "List", icon: List},
    {value: "calendar", label: "Calendar", icon: CalendarDays},
];

function formatDay(iso: string) {
    return new Date(`${iso}T00:00:00Z`).toLocaleDateString("en", {month: "short", day: "numeric", timeZone: "UTC"});
}

// Frost dates come as "MM-DD" without a year
function formatFrost(monthDay: string) {
    return formatDay(`2000-${monthDay}`);
}

const isoDate = (year: number, month: number, day: number) =>
    `${year}-${String(month + 1).padStart(2, "0")}-${String(day).padStart(2, "0")}`;

function todayIso() {
    const now = new Date();
    return isoDate(now.getFullYear(), now.getMonth(), now.getDate());
}

function daysBetween(from: string, to: string) {
    return Math.round((Date.parse(`${to}T00:00:00Z`) - Date.parse(`${from}T00:00:00Z`)) / DAY_MS);
}

function EmptyState({icon: Icon, children}: {icon: React.ComponentType<{className?: string}>, children: React.ReactNode}) {
    return (
        <div className="flex flex-col items-center gap-2 rounded-xl border border-dashed border-border py-12 text-center text-muted-foreground">
            <Icon className="size-8" />
            <div className="flex flex-col items-center gap-3 font-sans text-sm">{children}</div>
        </div>
    );
}

// Loading, unsupported, error and empty states shared by both views
function CalendarStatus({state, empty}: {state: ReturnType<typeof usePlanting>, empty: string | null}) {
    if (state.status === "loading") return <div className="h-48 animate-pulse rounded-xl bg-muted" />;

    if (state.status === "unsupported") {
        return (
            <EmptyState icon={CalendarDays}>
                <p>No crop calendar data for this location yet. CropGraph covers the United States only.</p>
            </EmptyState>
        );
    }

    if (state.status === "error") {
        return (
            <EmptyState icon={CalendarDays}>
                <p>Couldn&apos;t load the calendar: {state.message}</p>
                <Button variant="outline" size="sm" className="cursor-pointer" onClick={state.retry}>
                    <RotateCw />
                    Try again
                </Button>
            </EmptyState>
        );
    }

    return empty ? <EmptyState icon={CalendarDays}><p>{empty}</p></EmptyState> : null;
}

function ChangeLocationLink() {
    const appHref = useAppHref();
    return (
        <>
            {" "}
            <Link href={appHref("/settings")} className="underline underline-offset-4 hover:text-foreground">Change location</Link>
        </>
    );
}

function Attribution() {
    return (
        <p className="text-xs text-muted-foreground font-sans">
            Crop calendar data by <a href="https://cropgraph.com/" target="_blank" rel="noreferrer" className="underline underline-offset-4 hover:text-foreground">CropGraph</a>
        </p>
    );
}

function CropCard({crop}: {crop: PlantingCrop}) {
    return (
        <li className="flex flex-col gap-1 rounded-lg border border-border bg-background p-3 font-sans">
            <div className="flex items-baseline justify-between gap-2">
                <span className="font-medium">{crop.commonName}</span>
                <span className="text-xs capitalize text-muted-foreground">{crop.category.replace("-", " ")}</span>
            </div>
            <span className="text-sm tabular-nums text-muted-foreground">
                {formatDay(crop.windowStart)} – {formatDay(crop.windowEnd)} · harvest in {crop.daysToHarvest.min}–{crop.daysToHarvest.max} days
            </span>
        </li>
    );
}

// One month at a time, with the crops to plant that month grouped by action
function ListView({latitude, longitude}: {latitude: number, longitude: number}) {
    const year = new Date().getFullYear();
    const [month, setMonth] = useState(new Date().getMonth());
    const state = usePlanting(latitude, longitude, dateInMonth(year, month));

    const groups = state.status === "ready"
        ? Object.entries(Object.groupBy(state.calendar.plantNow, (crop) => crop.action))
        : [];

    return (
        <div className="flex flex-col gap-4">
            <div role="tablist" aria-label="Month" className="flex flex-wrap gap-1.5">
                {MONTHS.map((name, i) => (
                    <button
                        key={name}
                        role="tab"
                        aria-selected={i === month}
                        onClick={() => setMonth(i)}
                        className="cursor-pointer rounded-lg border border-border px-3 py-1.5 font-sans text-sm transition-colors hover:bg-muted aria-selected:border-transparent aria-selected:bg-primary aria-selected:text-primary-foreground"
                    >
                        {name}
                    </button>
                ))}
            </div>

            <CalendarStatus state={state} empty={state.status === "ready" && groups.length === 0 ? `Nothing to plant in ${MONTHS[month]}.` : null} />

            {state.status === "ready" &&
                <>
                    <p className="font-sans text-sm text-muted-foreground">
                        Zone {state.calendar.zone.zone} · {state.calendar.climateType.replace("_", " ")} climate · last spring frost {formatFrost(state.calendar.frostDates.lastSpring)}, first fall frost {formatFrost(state.calendar.frostDates.firstFall)}
                    </p>
                    {groups.map(([action, crops]) => (
                        <section key={action} className="flex flex-col gap-2">
                            <h2 className="font-serif text-xl font-medium">
                                <Tooltip>
                                    <TooltipTrigger render={<span tabIndex={0} className="cursor-help underline decoration-[1.8px] decoration-dotted decoration-muted-foreground/40 underline-offset-4" />}>
                                        {ACTION_LABELS[action] ?? action}
                                    </TooltipTrigger>
                                    <TooltipContent>
                                        {ACTION_HELP[action] ?? ACTION_LABELS[action] ?? action}
                                    </TooltipContent>
                                </Tooltip>{" "}
                                <span className="font-sans text-sm text-muted-foreground">({crops?.length})</span>
                            </h2>
                            <ul className="grid gap-2 sm:grid-cols-2 xl:grid-cols-3">
                                {crops?.map((crop) => <CropCard key={crop.slug} crop={crop} />)}
                            </ul>
                        </section>
                    ))}
                    <Attribution />
                </>
            }
        </div>
    );
}

function MonthGrid({year, month, selected, today, onSelect}: {year: number, month: number, selected: string, today: string, onSelect: (date: string) => void}) {
    const firstWeekday = new Date(Date.UTC(year, month, 1)).getUTCDay();
    const daysInMonth = new Date(Date.UTC(year, month + 1, 0)).getUTCDate();
    const cells = [...Array<null>(firstWeekday).fill(null), ...Array.from({length: daysInMonth}, (_, i) => i + 1)];

    return (
        <div className="grid grid-cols-7 gap-1 text-center font-sans text-sm">
            {["Su", "Mo", "Tu", "We", "Th", "Fr", "Sa"].map((d) => (
                <span key={d} className="py-1 text-xs text-muted-foreground">{d}</span>
            ))}
            {cells.map((day, i) => {
                if (day === null) return <span key={`blank-${i}`} />;
                const date = isoDate(year, month, day);
                return (
                    <button
                        key={date}
                        aria-pressed={date === selected}
                        aria-current={date === today ? "date" : undefined}
                        onClick={() => onSelect(date)}
                        className="cursor-pointer aspect-square rounded-lg tabular-nums transition-colors hover:bg-muted aria-[current=date]:font-semibold aria-[current=date]:aria-[pressed=false]:text-accent-strong aria-pressed:bg-primary aria-pressed:text-primary-foreground"
                    >
                        {day}
                    </button>
                );
            })}
        </div>
    );
}

function SeasonItem({crop, date}: {crop: PlantingCrop, date: string}) {
    const left = daysBetween(date, crop.windowEnd);
    return (
        <li className="flex items-center justify-between gap-3 rounded-lg border border-border bg-background p-3 font-sans">
            <div className="flex min-w-0 flex-col">
                <span className="truncate font-medium">{crop.commonName}</span>
                <span className="text-sm text-muted-foreground">
                    {ACTION_LABELS[crop.action] ?? crop.action} · {formatDay(crop.windowStart)} – {formatDay(crop.windowEnd)}
                </span>
            </div>
            <span className={`shrink-0 text-sm tabular-nums ${left <= 7 ? "text-warning-strong" : "text-muted-foreground"}`}>
                {left <= 0 ? "Last day" : `${left} ${left === 1 ? "day" : "days"} left`}
            </span>
        </li>
    );
}

// Month grid on the left; what's in season on the picked day on the right, with the days left in each window
function CalendarView({latitude, longitude}: {latitude: number, longitude: number}) {
    const today = todayIso();
    const [selected, setSelected] = useState(today);
    const [year, month] = selected.split("-").map(Number);
    const state = usePlanting(latitude, longitude, selected);

    const shiftMonth = (by: number) => {
        const target = new Date(Date.UTC(year, month - 1 + by, 1));
        const y = target.getUTCFullYear();
        const m = target.getUTCMonth();
        // Back on the current month, land on today; otherwise on the 1st
        const isCurrent = today.startsWith(`${y}-${String(m + 1).padStart(2, "0")}`);
        setSelected(isoDate(y, m, isCurrent ? Number(today.slice(8)) : 1));
    };

    const crops = state.status === "ready"
        ? [...state.calendar.plantNow].sort((a, b) => a.windowEnd.localeCompare(b.windowEnd))
        : [];

    return (
        <div className="flex flex-col gap-4">
            <div className="grid gap-6 lg:grid-cols-[22rem_1fr]">
                <div className="flex flex-col gap-3 self-start rounded-xl border border-border bg-background p-4 lg:sticky lg:top-4">
                    <div className="flex items-center justify-between">
                        <Button variant="ghost" size="icon" className="cursor-pointer" aria-label="Previous month" onClick={() => shiftMonth(-1)}><ChevronLeft /></Button>
                        <h2 className="font-serif text-lg font-medium">
                            {new Date(Date.UTC(year, month - 1, 1)).toLocaleDateString("en", {month: "long", year: "numeric", timeZone: "UTC"})}
                        </h2>
                        <Button variant="ghost" size="icon" className="cursor-pointer" aria-label="Next month" onClick={() => shiftMonth(1)}><ChevronRight /></Button>
                    </div>
                    <MonthGrid year={year} month={month - 1} selected={selected} today={today} onSelect={setSelected} />
                </div>

                <section className="flex min-w-0 flex-col gap-3">
                    <h2 className="font-serif text-xl font-medium">
                        {selected === today ? "In season now" : `In season on ${formatDay(selected)}`}
                        {state.status === "ready" && <span className="font-sans text-sm text-muted-foreground"> ({crops.length})</span>}
                    </h2>
                    <CalendarStatus state={state} empty={state.status === "ready" && crops.length === 0 ? "Nothing to plant on this day." : null} />
                    {crops.length > 0 &&
                        <ul className="flex flex-col gap-2">
                            {crops.map((crop) => <SeasonItem key={`${crop.slug}-${crop.action}`} crop={crop} date={selected} />)}
                        </ul>
                    }
                </section>
            </div>
            {state.status === "ready" && <Attribution />}
        </div>
    );
}

function Calendar({latitude, longitude}: {latitude: number, longitude: number}) {
    const [view, setView] = useState<View>("list");

    return (
        <div className="flex flex-col gap-4">
            {/* Same segmented control as the dashboard's unit filter */}
            <div className="flex items-center gap-3 font-sans">
                <span className="text-sm text-muted-foreground">View</span>
                <div role="radiogroup" aria-label="View" className="inline-flex h-9 items-center gap-0.5 rounded-xl border border-border bg-muted p-0.5">
                    {VIEWS.map(({value, label, icon: Icon}) => (
                        <button
                            key={value}
                            type="button"
                            role="radio"
                            aria-checked={view === value}
                            onClick={() => setView(value)}
                            className="inline-flex h-[30px] cursor-pointer items-center gap-1.5 rounded-[10px] px-3 text-sm text-muted-foreground outline-none hover:text-foreground focus-visible:ring-3 focus-visible:ring-ring/50 aria-checked:bg-background aria-checked:text-foreground aria-checked:shadow-xs"
                        >
                            <Icon className="size-4" />
                            {label}
                        </button>
                    ))}
                </div>
            </div>
            {view === "list"
                ? <ListView latitude={latitude} longitude={longitude} />
                : <CalendarView latitude={latitude} longitude={longitude} />}
        </div>
    );
}

export default function CalendarPage() {
    const appHref = useAppHref();
    const {location, ready: isClient} = useFarmLocation();

    return (
        <div className="page-container flex flex-col gap-4">
            <div className="flex flex-col gap-1">
                <h1 className="font-serif text-3xl font-medium">Crop calendar</h1>
                <p className="font-sans text-muted-foreground font-normal">
                    What to plant{location?.Name ? ` near ${location.Name}` : ""}.
                    {location && <ChangeLocationLink />}
                </p>
            </div>

            {!isClient
                ? <div className="h-48 animate-pulse rounded-xl bg-muted" />
                : location
                    // keyed, so a new location starts from a fresh state
                    ? <Calendar key={`${location.Latitude},${location.Longitude}`} latitude={location.Latitude} longitude={location.Longitude} />
                    : <EmptyState icon={MapPin}>
                        <p>Set your land&apos;s location to see its crop calendar.</p>
                        <Link href={appHref("/settings")} className={buttonVariants({variant: "outline", size: "sm"})}>
                            <MapPin />
                            Set location
                        </Link>
                    </EmptyState>
            }
        </div>
    );
}
