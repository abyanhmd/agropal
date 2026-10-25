'use client';

import React, {useMemo} from 'react';
import Link from "next/link";
import {ArrowDownRight, ArrowRight, ArrowUpRight, Boxes, PackageOpen} from "lucide-react";
import EmptyState from "@/components/app/empty-state";
import {ProductSchema} from "@/app/(app)/products/_types/product.schema";
import {StockBadge} from "@/app/(app)/products/_components/table.product";
import {DashboardSummary, ProductNet, summarize, unitsOf} from "@/app/(app)/dashboard/_lib/dashboard";
import {numberFormat, shortDate} from "@/lib/format";
import MovementChart from "@/app/(app)/dashboard/_components/movement-chart.dashboard";
import WeatherCard from "@/app/(app)/dashboard/_components/weather.dashboard";
import {useAppHref} from "@/lib/demo/base";

function StatTile({label, value, hint, children}: {label: string, value: React.ReactNode, hint?: string, children?: React.ReactNode}) {
    return (
        <div className="flex flex-col gap-1 rounded-xl border border-border bg-background p-4 font-sans">
            <span className="text-sm text-muted-foreground">{label}</span>
            {/* Proportional figures for big standalone numbers */}
            <span className="text-3xl font-semibold font-serif">{value}</span>
            {hint && <span className="text-xs text-muted-foreground">{hint}</span>}
            {children}
        </div>
    );
}

function Card({title, action, children}: {title: string, action?: React.ReactNode, children: React.ReactNode}) {
    return (
        <section className="flex flex-col rounded-xl border border-border bg-background font-sans">
            <header className="flex items-center justify-between gap-3 border-b border-border px-5 py-3">
                <h2 className="font-medium">{title}</h2>
                {action}
            </header>
            {children}
        </section>
    );
}

function NetList({items, unit, empty}: {items: ProductNet[], unit: string, empty: string}) {
    const appHref = useAppHref();
    if (items.length === 0) {
        return <p className="px-5 py-6 text-center text-sm text-muted-foreground">{empty}</p>;
    }
    return (
        <ul className="divide-y divide-border">
            {items.map((item) => (
                <li key={item.id}>
                    <Link href={appHref(`/products/${item.id}`)} className="flex items-center justify-between gap-3 px-5 py-2.5 text-sm transition-colors hover:bg-muted">
                        <span className="truncate font-serif">{item.name}</span>
                        <span className="shrink-0 font-semibold tabular-nums">
                            {item.net > 0 ? "+" : "−"}{numberFormat.format(Math.abs(item.net))}
                            <span className="ml-1 font-normal text-muted-foreground">{unit}</span>
                        </span>
                    </Link>
                </li>
            ))}
        </ul>
    );
}

function Overview({summary, unit}: {summary: DashboardSummary, unit: string}) {
    const appHref = useAppHref();
    const NetIcon = summary.net >= 0 ? ArrowUpRight : ArrowDownRight;

    return (
        <>
            <div className="grid grid-cols-2 gap-3 lg:grid-cols-4">
                <StatTile label="Products" value={numberFormat.format(summary.productCount)} hint={`Sold by ${unit}`} />
                <StatTile label="Stock in" value={numberFormat.format(summary.stockIn)} hint={unit} />
                <StatTile label="Stock out" value={numberFormat.format(summary.stockOut)} hint={unit} />
                <StatTile
                    label="Net movement"
                    value={
                        <span className="inline-flex items-center gap-1">
                            {summary.net >= 0 ? "+" : "−"}{numberFormat.format(Math.abs(summary.net))}
                            <NetIcon className={`size-5 ${summary.net >= 0 ? "text-accent-strong" : "text-danger-strong"}`} aria-hidden />
                        </span>
                    }
                    hint={`${unit} · ${summary.net >= 0 ? "more came in than went out" : "more went out than came in"}`}
                />
            </div>

            <MovementChart data={summary.monthly} unit={unit} />

            <div className="grid gap-6 lg:grid-cols-3">
                {/* min-w-0: a grid item won't shrink below its content, and long product names would widen it */}
                <div className="min-w-0 lg:col-span-2">
                    <Card
                        title="Recent activity"
                        action={
                            <Link href={appHref("/products")} className="inline-flex items-center gap-1 text-sm text-muted-foreground transition-colors hover:text-foreground">
                                All products <ArrowRight className="size-4" />
                            </Link>
                        }
                    >
                        <ul className="divide-y divide-border">
                            {summary.recent.map(({product, stock}) => (
                                <li key={product.id + stock.ID}>
                                    <Link href={appHref(`/products/${product.id}`)} className="flex items-center gap-4 px-5 py-2.5 transition-colors hover:bg-muted">
                                        <span className="w-14 shrink-0 text-sm text-muted-foreground tabular-nums">{shortDate.format(new Date(stock.DateAdded))}</span>
                                        <span className="min-w-0 flex-1 truncate font-serif">{product.name}</span>
                                        <StockBadge stock={stock} unit={unit} />
                                    </Link>
                                </li>
                            ))}
                        </ul>
                    </Card>
                </div>

                <div className="flex min-w-0 flex-col gap-6">
                    <Card title="Most stocked">
                        <NetList items={summary.mostStocked} unit={unit} empty="No product gained stock" />
                    </Card>
                    <Card title="Most depleted">
                        <NetList items={summary.mostDepleted} unit={unit} empty="No product lost stock" />
                    </Card>
                </div>
            </div>
        </>
    );
}

export interface DashboardProps {
    products: ProductSchema[];
    unit: string;
    // Products are still being fetched, so "no products" isn't known yet
    loading?: boolean;
    // Absent in the prerendered fallback, where the unit can't change yet
    onUnitChange?: (unit: string) => void;
}

export default function Dashboard({products, unit, loading = false, onUnitChange}: DashboardProps) {
    const appHref = useAppHref();
    const units = useMemo(() => unitsOf(products), [products]);
    const summary = useMemo(() => summarize(products, unit), [products, unit]);

    return (
        <div className="page-container flex flex-col gap-6">
            <div className="flex flex-col gap-1">
                <h1 className="font-serif text-3xl font-medium">Dashboard</h1>
                <p className="text-muted-foreground font-lato">Stock activity across your products.</p>
            </div>

            {/* Weather isn't per unit, so it sits above the unit filter */}
            <WeatherCard />

            {loading && products.length === 0
                ? <div className="h-64 animate-pulse rounded-xl bg-muted" />
                : products.length === 0
                    ? (
                        <EmptyState
                            icon={Boxes}
                            title="No products yet"
                            action={<Link href={appHref("/products")} className="text-sm underline underline-offset-4">Add your first product</Link>}
                        >
                            The dashboard summarises stock movement across your products. Add one and record some stock to see it here.
                        </EmptyState>
                    )
                    : !products.some((p) => p.Stocks?.length)
                        ? (
                            <EmptyState
                                icon={PackageOpen}
                                title="No stock recorded yet"
                                action={<Link href={appHref("/products")} className="text-sm underline underline-offset-4">Go to products</Link>}
                            >
                                Record stock coming in or going out on a product and its movement will show up here.
                            </EmptyState>
                        )
                        : (
                            <>
                                {/* Only worth choosing when there is more than one unit */}
                                {units.length > 1 && (
                                    <div className="flex items-center gap-3 font-sans">
                                        <span className="text-sm text-muted-foreground">Unit</span>
                                        <div role="radiogroup" aria-label="Unit" className="inline-flex h-9 items-center gap-0.5 rounded-xl border border-border bg-muted p-0.5">
                                            {units.map((u) => (
                                                <button
                                                    key={u}
                                                    type="button"
                                                    role="radio"
                                                    aria-checked={unit === u}
                                                    disabled={!onUnitChange}
                                                    onClick={() => onUnitChange?.(u)}
                                                    className="h-[30px] rounded-[10px] px-3 text-sm text-muted-foreground outline-none hover:text-foreground focus-visible:ring-3 focus-visible:ring-ring/50 aria-checked:bg-background aria-checked:text-foreground aria-checked:shadow-xs"
                                                >
                                                    {u}
                                                </button>
                                            ))}
                                        </div>
                                    </div>
                                )}
                                <Overview summary={summary} unit={unit} />
                            </>
                        )}
        </div>
    );
}
