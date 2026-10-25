import React from 'react';
import {ArrowUpRight, CloudSun, Droplets, Thermometer} from "lucide-react";

// A static, hand-drawn miniature of the dashboard for the hero. Decorative only, so the
// numbers are illustrative and the whole thing is hidden from screen readers.

const BARS = [
    {stockIn: 40, stockOut: 22},
    {stockIn: 55, stockOut: 30},
    {stockIn: 35, stockOut: 38},
    {stockIn: 70, stockOut: 34},
    {stockIn: 62, stockOut: 45},
    {stockIn: 85, stockOut: 40},
    {stockIn: 58, stockOut: 52},
    {stockIn: 76, stockOut: 36},
];

function MiniStat({label, value, accent}: {label: string, value: string, accent?: boolean}) {
    return (
        <div className="flex flex-col gap-0.5 rounded-lg border border-border bg-background p-3">
            <span className="text-[11px] text-muted-foreground">{label}</span>
            <span className={`font-serif text-base sm:text-lg font-semibold ${accent ? "text-accent-strong" : ""}`}>{value}</span>
        </div>
    );
}

export default function LandingPreview() {
    return (
        <div aria-hidden className="relative isolate select-none">
            {/* Soft glow behind the card */}
            <div className="absolute -inset-6 -z-10 rounded-[2rem] bg-accent-muted blur-2xl opacity-70" />

            <div className="overflow-hidden rounded-2xl border border-border bg-sidebar shadow-xl shadow-black/5">
                {/* Window chrome */}
                <div className="flex items-center gap-1.5 border-b border-border px-4 py-2.5">
                    <span className="size-2.5 rounded-full bg-danger/60" />
                    <span className="size-2.5 rounded-full bg-warning/60" />
                    <span className="size-2.5 rounded-full bg-accent/70" />
                    <span className="ml-3 font-serif text-sm text-muted-foreground">Dashboard</span>
                </div>

                <div className="flex flex-col gap-3 p-4 font-sans">
                    <div className="flex items-center gap-3 rounded-lg border border-border bg-background p-3">
                        <CloudSun className="size-7 text-warning-strong" />
                        <div className="flex flex-1 flex-col">
                            <span className="font-serif text-lg font-semibold leading-tight">24°</span>
                            <span className="text-[11px] text-muted-foreground">Partly cloudy</span>
                        </div>
                        <div className="flex flex-col gap-1 text-[11px] text-muted-foreground">
                            <span className="flex items-center gap-1"><Thermometer className="size-3" /> Soil 19°</span>
                            <span className="flex items-center gap-1"><Droplets className="size-3" /> Moisture 31%</span>
                        </div>
                    </div>

                    <div className="grid grid-cols-3 gap-3">
                        <MiniStat label="Stock in" value="1,240 kg" />
                        <MiniStat label="Stock out" value="860 kg" />
                        <MiniStat label="Net" value="+380 kg" accent />
                    </div>

                    <div className="rounded-lg border border-border bg-background p-3">
                        <div className="mb-3 flex items-center justify-between">
                            <span className="text-xs font-medium">Monthly movement</span>
                            <span className="flex items-center gap-0.5 text-[11px] text-accent-strong">
                                <ArrowUpRight className="size-3" /> trending up
                            </span>
                        </div>
                        <div className="flex h-28 items-end gap-2">
                            {BARS.map((bar, i) => (
                                <div key={i} className="flex h-full flex-1 items-end gap-0.5">
                                    <div className="flex-1 rounded-t-sm bg-chart-in" style={{height: `${bar.stockIn}%`}} />
                                    <div className="flex-1 rounded-t-sm bg-chart-out" style={{height: `${bar.stockOut}%`}} />
                                </div>
                            ))}
                        </div>
                    </div>
                </div>
            </div>
        </div>
    );
}
