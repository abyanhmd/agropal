'use client';

import React, {useState} from 'react';
import Link from "next/link";
import {ArrowLeft} from "lucide-react";
import {ProductSchema} from "@/app/(app)/products/_types/product.schema";
import {TransactionSchema} from "@/app/(app)/orders/_types/order.schema";
import {formatOrderDate, formatSignedAmount} from "@/app/(app)/orders/_components/table.order";
import {StockBadge} from "@/app/(app)/products/_components/table.product";
import {formatPrice} from "@/lib/format";
import {useSetBreadcrumb} from "@/hooks/breadcrumb-state";
import {Select, SelectContent, SelectItem, SelectTrigger, SelectValue} from "@/components/ui/select";
import {UNITS} from "@/app/(app)/products/_lib/units";
import {useFarm} from "@/lib/auth/auth-provider";
import {updateProduct} from "@/lib/api/resources";
import {invalidate} from "@/lib/api/use-query";
import ProductImage from "@/app/(app)/products/_components/image.product";
import ProductActions from "@/app/(app)/products/_components/actions.product";
import {AddPriceDialog, AddStockDialog, DeleteEntryButton} from "@/app/(app)/products/_components/entries.product";
import {useAppHref} from "@/lib/demo/base";

// UTC to match how the mock dates are created, so server and browser render the same text
const dateFormat = new Intl.DateTimeFormat("en-US", {month: "short", day: "numeric", year: "numeric", timeZone: "UTC"});
const formatDate = (date: Date | string) => dateFormat.format(new Date(date));

function StatCard({label, children}: {label: string, children: React.ReactNode}) {
    return (
        <div className="flex flex-col gap-1.5 rounded-xl border border-border bg-background p-4">
            <span className="text-xs font-medium uppercase tracking-wide text-muted-foreground">{label}</span>
            <div className="font-sans">{children}</div>
        </div>
    );
}

function Panel({title, action, children}: {title: string, action?: React.ReactNode, children: React.ReactNode}) {
    return (
        <section className="overflow-hidden rounded-xl border border-border bg-background">
            <div className="flex min-h-10 items-center justify-between gap-2 border-b border-border bg-muted px-4 py-1.5">
                <h2 className="text-xs font-medium uppercase tracking-wide text-muted-foreground">{title}</h2>
                {action}
            </div>
            <ul className="max-h-96 divide-y divide-border overflow-y-auto">{children}</ul>
        </section>
    );
}

export function ProductDetailSkeleton() {
    return (
        <div className="page-container flex animate-pulse flex-col gap-6">
            <div className="h-4 w-32 rounded bg-muted" />
            <div className="grid gap-6 md:grid-cols-[minmax(0,2fr)_3fr]">
                <div className="aspect-square rounded-xl bg-accent-muted" />
                <div className="flex flex-col gap-3">
                    <div className="h-8 w-2/3 rounded bg-muted" />
                    <div className="h-4 w-full rounded bg-muted" />
                    <div className="mt-4 grid grid-cols-1 gap-3 sm:grid-cols-3">
                        {Array.from({length: 3}, (_, i) => <div key={i} className="h-20 rounded-xl bg-muted" />)}
                    </div>
                </div>
            </div>
        </div>
    );
}

// Changes what the product is sold by; products created before units existed start on "pcs"
function UnitPicker({product}: {product: ProductSchema}) {
    const farm = useFarm()?.farm;
    const [saving, setSaving] = useState(false);
    const [error, setError] = useState<string | null>(null);
    // A unit outside the list (set through the API) still shows up as the current choice
    const items = UNITS.some((u) => u.value === product.Unit) ? UNITS : [...UNITS, {value: product.Unit, label: product.Unit}];

    const change = async (unit: string | null) => {
        if (!unit || !farm || unit === product.Unit) return;
        setSaving(true);
        setError(null);
        try {
            await updateProduct(farm.farm_id, product.ID, {unit});
            invalidate("products:");
        } catch (err) {
            setError(err instanceof Error ? err.message : "Couldn't change the unit");
        } finally {
            setSaving(false);
        }
    };

    return (
        <div className="absolute right-3 top-3 flex flex-col items-end gap-1">
            <Select items={items} value={product.Unit || null} onValueChange={change} disabled={saving}>
                <SelectTrigger aria-label="Sold by" size="sm" className="cursor-pointer rounded-full bg-background/80 font-sans text-xs backdrop-blur">
                    <span className="text-muted-foreground">per</span>
                    <SelectValue placeholder="unit" />
                </SelectTrigger>
                <SelectContent>
                    {items.map((o) => <SelectItem key={o.value} value={o.value}>{o.label}</SelectItem>)}
                </SelectContent>
            </Select>
            {error && <p role="alert" className="rounded bg-background/90 px-2 py-0.5 text-xs text-danger-strong">{error}</p>}
        </div>
    );
}

export interface ProductDetailProps {
    product: ProductSchema;
    orders?: TransactionSchema[]; // this product's orders
}

export default function ProductDetail({product, orders = []}: ProductDetailProps) {
    const appHref = useAppHref();
    // Newest first
    const allOrders = [...orders].sort((a, b) => new Date(b.Date).getTime() - new Date(a.Date).getTime());
    useSetBreadcrumb([{label: "Products", href: "/products"}, {label: product.Name}]);

    // Stored oldest → newest; show newest first
    const stocks = [...(product.Stocks ?? [])].reverse();
    const prices = [...(product.Prices ?? [])].reverse();
    const latestStock = stocks[0];
    const latestPrice = prices[0];

    const netMovement = stocks.reduce((total, s) => total + (s.Type === "INCOME" ? s.Stock : -s.Stock), 0);

    return (
        <div className="page-container flex flex-col gap-6">
            <Link
                href={appHref("/products")}
                className="inline-flex w-fit items-center gap-1.5 text-sm text-muted-foreground font-sans transition-colors hover:text-foreground"
            >
                <ArrowLeft className="size-4" />
                Back to products
            </Link>

            <div className="grid gap-6 md:grid-cols-[minmax(0,2fr)_3fr]">
                <div className="relative flex aspect-square items-center justify-center overflow-hidden rounded-xl bg-accent-muted text-accent-strong">
                    <ProductImage src={product.Image} sizes="(min-width: 768px) 40vw, 100vw" iconClassName="size-16" />
                    <UnitPicker product={product} />
                </div>

                <div className="flex flex-col gap-4">
                    <div className="flex flex-wrap items-start justify-between gap-3">
                        <div className="flex min-w-0 flex-col gap-1">
                            <h1 className="font-serif text-3xl font-medium">{product.Name}</h1>
                            <p className="text-muted-foreground font-sans">{product.Description}</p>
                        </div>
                        <ProductActions product={product} orderCount={orders.length} />
                    </div>

                    <div className="grid grid-cols-1 gap-3 sm:grid-cols-3">
                        <StatCard label="Current price">
                            {latestPrice
                                ? <p className="text-xl font-semibold tabular-nums">
                                    {formatPrice(latestPrice.Price, latestPrice.Currency)}
                                    {product.Unit && <span className="ml-1 text-sm font-normal text-muted-foreground">/ {product.Unit}</span>}
                                </p>
                                : <p className="text-muted-foreground">—</p>
                            }
                        </StatCard>

                        <StatCard label="Latest change">
                            {latestStock
                                ? <div className="flex flex-col items-start gap-1">
                                    <StockBadge stock={latestStock} unit={product.Unit} />
                                    <span className="text-xs text-muted-foreground">{formatDate(latestStock.DateAdded)}</span>
                                </div>
                                : <p className="text-muted-foreground">—</p>
                            }
                        </StatCard>

                        <StatCard label="Net movement">
                            <p className={`text-xl font-semibold tabular-nums ${netMovement >= 0 ? "text-accent-strong" : "text-danger-strong"}`}>
                                {netMovement >= 0 ? "+" : "−"}{Math.abs(netMovement)}
                                {product.Unit && <span className="ml-1 text-sm font-normal text-muted-foreground">{product.Unit}</span>}
                            </p>
                        </StatCard>
                    </div>
                </div>
            </div>

            <div className="grid gap-6 lg:grid-cols-2">
                <Panel title="Stock movements" action={<AddStockDialog product={product} />}>
                    {stocks.map((stock) => (
                        <li key={stock.ID} className="flex items-center gap-4 px-4 py-2.5 font-sans">
                            <span className="mr-auto text-sm text-muted-foreground tabular-nums">{formatDate(stock.DateAdded)}</span>
                            <StockBadge stock={stock} unit={product.Unit} />
                            <DeleteEntryButton
                                kind="stock"
                                product={product}
                                entryID={stock.ID}
                                summary={`${stock.Type === "INCOME" ? "+" : "−"}${stock.Stock}${product.Unit ? ` ${product.Unit}` : ""} on ${formatDate(stock.DateAdded)}`}
                            />
                        </li>
                    ))}
                    {stocks.length === 0 && <li className="px-4 py-6 text-center text-sm text-muted-foreground">No stock movements yet</li>}
                </Panel>

                <Panel title="Price history" action={<AddPriceDialog product={product} />}>
                    {prices.map((price) => (
                        <li key={price.ID} className="flex items-center gap-4 px-4 py-2.5 font-sans">
                            <span className="mr-auto text-sm text-muted-foreground tabular-nums">{formatDate(price.DateAdded)}</span>
                            <span className="font-semibold tabular-nums">{formatPrice(price.Price, price.Currency)}</span>
                            <DeleteEntryButton
                                kind="price"
                                product={product}
                                entryID={price.ID}
                                summary={`${formatPrice(price.Price, price.Currency)} on ${formatDate(price.DateAdded)}`}
                            />
                        </li>
                    ))}
                    {prices.length === 0 && <li className="px-4 py-6 text-center text-sm text-muted-foreground">No prices yet</li>}
                </Panel>
            </div>

            <Panel title={`Orders (${allOrders.length})`}>
                {allOrders.map((order) => (
                    <li key={order.ID}>
                        <Link
                            href={appHref(`/orders/${order.ID}`)}
                            className="flex items-center justify-between gap-4 px-4 py-2.5 font-sans outline-none transition-colors hover:bg-muted/60 focus-visible:bg-muted/60"
                        >
                            <span className="min-w-0">
                                <span className="block truncate text-sm">{order.Description || "Untitled order"}</span>
                                <span className="block text-xs text-muted-foreground tabular-nums">{formatOrderDate(order.Date)}</span>
                            </span>
                            <span className={`shrink-0 font-semibold tabular-nums ${order.Type === "INCOME" ? "text-accent-strong" : "text-danger-strong"}`}>
                                {formatSignedAmount(order)}
                            </span>
                        </Link>
                    </li>
                ))}
                {allOrders.length === 0 && <li className="px-4 py-6 text-center text-sm text-muted-foreground">No orders yet</li>}
            </Panel>
        </div>
    );
}
