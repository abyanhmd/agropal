'use client';

import React, {useState} from 'react';
import {CHART_YEAR, MonthlyMovement} from "@/app/(app)/dashboard/_lib/dashboard";
import {numberFormat} from "@/lib/format";

const PLOT_HEIGHT = 240; // px, excludes the month labels below so they always fit

// Round the axis maximum up to a clean number: 87 → 100, 430 → 500, 1,240 → 1,500
function niceMax(value: number): number {
    if (value <= 0) return 10;
    const magnitude = 10 ** Math.floor(Math.log10(value));
    const step = [1, 1.5, 2, 2.5, 3, 4, 5, 6, 8, 10].find((s) => s * magnitude >= value) ?? 10;
    return step * magnitude;
}

function LegendKey({className, label}: {className: string, label: string}) {
    return (
        <span className="inline-flex items-center gap-1.5 text-sm text-muted-foreground">
            <span className={`size-2.5 rounded-[3px] ${className}`} />
            {label}
        </span>
    );
}

function Tooltip({entry, unit, align}: {entry: MonthlyMovement, unit: string, align: "left" | "center" | "right"}) {
    const net = entry.stockIn - entry.stockOut;
    const position = align === "left" ? "left-0" : align === "right" ? "right-0" : "left-1/2 -translate-x-1/2";

    return (
        <div role="tooltip" className={`pointer-events-none absolute top-1 z-10 w-40 rounded-lg border border-border bg-popover px-3 py-2 text-left shadow-md ${position}`}>
            <p className="mb-1 text-xs text-muted-foreground">{entry.month} {CHART_YEAR}</p>
            {[
                {label: "Stock in", value: entry.stockIn, key: "bg-chart-in"},
                {label: "Stock out", value: entry.stockOut, key: "bg-chart-out"},
            ].map((row) => (
                <p key={row.label} className="flex items-center gap-2 text-sm">
                    <span className={`h-0.5 w-3 rounded-full ${row.key}`} />
                    <span className="font-semibold tabular-nums">{numberFormat.format(row.value)}</span>
                    <span className="text-muted-foreground">{row.label}</span>
                </p>
            ))}
            <p className="mt-1 border-t border-border pt-1 text-sm">
                <span className="font-semibold tabular-nums">{net >= 0 ? "+" : "−"}{numberFormat.format(Math.abs(net))}</span>
                <span className="text-muted-foreground"> {unit} net</span>
            </p>
        </div>
    );
}

function MovementTable({data, unit}: {data: MonthlyMovement[], unit: string}) {
    return (
        <div className="overflow-x-auto">
            <table className="w-full text-sm">
                <thead>
                    <tr className="border-b border-border text-left text-xs uppercase tracking-wide text-muted-foreground">
                        <th className="py-2 pr-4 font-medium">Month</th>
                        <th className="py-2 pr-4 text-right font-medium">Stock in ({unit})</th>
                        <th className="py-2 pr-4 text-right font-medium">Stock out ({unit})</th>
                        <th className="py-2 text-right font-medium">Net ({unit})</th>
                    </tr>
                </thead>
                <tbody className="divide-y divide-border tabular-nums">
                    {data.map((entry) => {
                        const net = entry.stockIn - entry.stockOut;
                        return (
                            <tr key={entry.month}>
                                <td className="py-2 pr-4">{entry.month}</td>
                                <td className="py-2 pr-4 text-right">{numberFormat.format(entry.stockIn)}</td>
                                <td className="py-2 pr-4 text-right">{numberFormat.format(entry.stockOut)}</td>
                                <td className="py-2 text-right">{net >= 0 ? "+" : "−"}{numberFormat.format(Math.abs(net))}</td>
                            </tr>
                        );
                    })}
                </tbody>
            </table>
        </div>
    );
}

export default function MovementChart({data, unit}: {data: MonthlyMovement[], unit: string}) {
    const [view, setView] = useState<"chart" | "table">("chart");
    const [active, setActive] = useState<number | null>(null);

    const max = niceMax(Math.max(...data.map((d) => Math.max(d.stockIn, d.stockOut))));
    const ticks = [max, max / 2, 0, max / 2, max]; // shown as magnitudes: up = in, down = out
    // Direct-label only the two peaks; the axis, tooltip, and table carry the rest
    const peakIn = data.reduce((best, d, i) => d.stockIn > data[best].stockIn ? i : best, 0);
    const peakOut = data.reduce((best, d, i) => d.stockOut > data[best].stockOut ? i : best, 0);

    return (
        <section className="rounded-xl border border-border bg-background p-5 font-sans">
            <header className="mb-4 flex flex-wrap items-start justify-between gap-3">
                <div>
                    <h2 className="font-medium">Monthly stock movement</h2>
                    <p className="text-sm text-muted-foreground">{CHART_YEAR} · in {unit}</p>
                </div>
                <div className="flex flex-wrap items-center gap-4">
                    <LegendKey className="bg-chart-in" label="Stock in (up)" />
                    <LegendKey className="bg-chart-out" label="Stock out (down)" />
                    <div role="group" aria-label="Display as" className="inline-flex rounded-lg border border-border bg-muted p-0.5 text-sm">
                        {(["chart", "table"] as const).map((option) => (
                            <button
                                key={option}
                                type="button"
                                aria-pressed={view === option}
                                onClick={() => setView(option)}
                                className="rounded-md px-2.5 py-1 capitalize text-muted-foreground outline-none hover:text-foreground focus-visible:ring-3 focus-visible:ring-ring/50 aria-pressed:bg-background aria-pressed:text-foreground aria-pressed:shadow-xs"
                            >
                                {option}
                            </button>
                        ))}
                    </div>
                </div>
            </header>

            {view === "table"
                ? <MovementTable data={data} unit={unit} />
                : <div className="flex gap-2">
                    {/* Y axis */}
                    <div className="relative w-10 shrink-0 text-right text-xs text-muted-foreground tabular-nums" style={{height: PLOT_HEIGHT}}>
                        {ticks.map((tick, i) => (
                            <span key={i} className="absolute right-0 -translate-y-1/2" style={{top: `${(i / (ticks.length - 1)) * 100}%`}}>
                                {numberFormat.format(tick)}
                            </span>
                        ))}
                    </div>

                    <div className="min-w-0 flex-1">
                        <div className="relative" style={{height: PLOT_HEIGHT}} onPointerLeave={() => setActive(null)}>
                            {/* Gridlines: solid hairlines; the zero line one step stronger */}
                            {ticks.map((_, i) => (
                                <div
                                    key={i}
                                    className={`absolute inset-x-0 h-px ${i === 2 ? "bg-foreground/25" : "bg-border"}`}
                                    style={{top: `${(i / (ticks.length - 1)) * 100}%`}}
                                />
                            ))}

                            <div className="absolute inset-0 grid grid-cols-12">
                                {data.map((entry, i) => {
                                    const inHeight = (entry.stockIn / max) * 50;
                                    const outHeight = (entry.stockOut / max) * 50;
                                    const align = i < 2 ? "left" : i > 9 ? "right" : "center";
                                    const net = entry.stockIn - entry.stockOut;

                                    return (
                                        // Whole column is the hit target, much bigger than the bars
                                        <div
                                            key={entry.month}
                                            tabIndex={0}
                                            aria-label={`${entry.month}: ${entry.stockIn} ${unit} in, ${entry.stockOut} ${unit} out, net ${net} ${unit}`}
                                            onPointerEnter={() => setActive(i)}
                                            onFocus={() => setActive(i)}
                                            onBlur={() => setActive(null)}
                                            className={`relative outline-none transition-colors focus-visible:bg-muted ${active === i ? "bg-muted/70" : ""}`}
                                        >
                                            {/* Stock in: grows up from zero, rounded at the data end only */}
                                            <div
                                                className="absolute bottom-1/2 left-1/2 w-[min(60%,24px)] -translate-x-1/2 rounded-t-[4px] bg-chart-in"
                                                style={{height: `${inHeight}%`, marginBottom: 1}}
                                            >
                                                {i === peakIn && entry.stockIn > 0 &&
                                                    <span className="absolute bottom-full left-1/2 mb-1 -translate-x-1/2 text-xs font-medium tabular-nums">
                                                        {numberFormat.format(entry.stockIn)}
                                                    </span>
                                                }
                                            </div>
                                            {/* Stock out: grows down from zero */}
                                            <div
                                                className="absolute left-1/2 top-1/2 w-[min(60%,24px)] -translate-x-1/2 rounded-b-[4px] bg-chart-out"
                                                style={{height: `${outHeight}%`, marginTop: 1}}
                                            >
                                                {i === peakOut && entry.stockOut > 0 &&
                                                    <span className="absolute left-1/2 top-full mt-1 -translate-x-1/2 text-xs font-medium tabular-nums">
                                                        {numberFormat.format(entry.stockOut)}
                                                    </span>
                                                }
                                            </div>

                                            {active === i && <Tooltip entry={entry} unit={unit} align={align} />}
                                        </div>
                                    );
                                })}
                            </div>
                        </div>

                        {/* X axis */}
                        <div className="mt-2 grid grid-cols-12 text-center text-xs text-muted-foreground">
                            {data.map((entry) => (
                                <span key={entry.month}>
                                    <span className="sm:hidden">{entry.month[0]}</span>
                                    <span className="hidden sm:inline">{entry.month}</span>
                                </span>
                            ))}
                        </div>
                    </div>
                </div>
            }
        </section>
    );
}
