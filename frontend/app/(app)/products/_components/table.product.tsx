import React from 'react';
import Link from "next/link";
import {ProductPriceSchema, ProductSchema, ProductStockSchema} from "@/app/(app)/products/_types/product.schema";
import {TrendingUp, TrendingDown, Package} from "lucide-react";
import {formatPrice} from "@/lib/format";
import ProductImage from "@/app/(app)/products/_components/image.product";
import {useAppHref} from "@/lib/demo/base";

export interface ProductTableProps {
    products : ProductSchema[];
    emptyMessage?: string;
}

// Columns follow the table's own width (container queries), not the window's: the sidebar takes a quarter of it
const ROW_GRID = "grid grid-cols-[1fr_auto] @3xl:grid-cols-[1fr_4rem_7rem_8rem] @5xl:grid-cols-[1fr_8rem_12rem_12rem] items-center gap-4 px-4";


function TableHeader() {
    return (
        <div className={`${ROW_GRID} py-2.5 bg-muted border-b border-border text-xs font-medium uppercase tracking-wide text-muted-foreground`}>
            <span>Product</span>
            <span className="text-center hidden @3xl:block">Unit</span>
            <span className="text-center hidden @3xl:block">Latest change</span>
            <span className="text-center hidden @3xl:block">Price</span>
        </div>
    )
}

export function ProductTableSkeleton() {
    return (
        <div className="@container w-full overflow-hidden rounded-xl border border-border bg-background">
            <TableHeader />
            <div className="divide-y divide-border animate-pulse">
                {Array.from({length: 10}, (_, i) => (
                    <div key={i} className={`${ROW_GRID} py-3`}>
                        <div className="flex items-center gap-3">
                            <div className="w-10 h-10 shrink-0 rounded-lg bg-accent-muted"></div>
                            <div className="flex flex-col gap-y-2">
                                <div className="h-4 w-32 rounded bg-muted"></div>
                                <div className="h-3 w-48 rounded bg-muted"></div>
                            </div>
                        </div>
                        <div className="hidden @3xl:block h-5 w-10 rounded-full bg-muted"></div>
                        <div className="h-6 w-20 rounded-full bg-muted justify-self-end @3xl:justify-self-start"></div>
                        <div className="hidden @3xl:block h-3 w-20 rounded bg-muted justify-self-end"></div>
                    </div>
                ))}
            </div>
        </div>
    )
}

export function StockBadge({stock, unit}: {stock: ProductStockSchema, unit?: string}) {
    const isIncome = stock.Type == "INCOME";
    const Icon = isIncome ? TrendingUp : TrendingDown;

    return (
        <span className={`inline-flex items-center gap-1.5 rounded-full px-2.5 py-1 text-sm font-sans ${isIncome ? "bg-accent-muted text-accent-strong" : "bg-danger-muted text-danger-strong"}`}>
            <Icon className="w-4 h-4" />
            <span className="font-semibold tabular-nums">{isIncome ? "+" : "−"}{stock.Stock}</span>
            {unit && <span className="opacity-80">{unit}</span>}
        </span>
    )
}

export function PriceBadge({price, unit}: {price: ProductPriceSchema, unit?: string}) {
    return (
        <span className="inline-flex items-baseline gap-1 rounded-full border border-border bg-muted px-2.5 py-1 text-sm font-sans">
            <span className="font-semibold tabular-nums text-foreground">{formatPrice(price.Price, price.Currency)}</span>
            {unit && <span className="text-xs text-muted-foreground">/ {unit}</span>}
        </span>
    )
}

export function ProductEmptyState({message}: {message?: string}) {
    return (
        <div className="w-full flex flex-col items-center gap-2 rounded-xl border border-dashed border-border py-12 text-muted-foreground">
            <Package className="w-8 h-8" />
            <p className="font-sans text-sm">{message ?? "No products yet"}</p>
        </div>
    )
}

export default function ProductTable(props: ProductTableProps) {
    const appHref = useAppHref();
    if (props.products.length === 0) {
        return <ProductEmptyState message={props.emptyMessage} />
    }

    return (
        <div className="@container w-full overflow-hidden rounded-xl border border-border bg-background">
            <TableHeader />
            <div className="divide-y divide-border">
                {props.products.map((product) => {
                    // Stocks are sorted oldest → newest, so the latest change is the last entry
                    const latestStock = product.Stocks?.[product.Stocks.length - 1];
                    const latestPrice = product.Prices?.[product.Prices.length - 1];

                    return (
                        // Same pattern as the grid cards: the name link stretches over the whole row
                        <div key={product.ID} className={`${ROW_GRID} group relative py-3 transition-colors hover:bg-muted/60 focus-within:bg-muted/60`}>
                            <div className="flex items-center gap-3 min-w-0">
                                <div className="relative w-10 h-10 shrink-0 flex items-center justify-center overflow-hidden rounded-lg bg-accent-muted text-accent-strong">
                                    <ProductImage src={product.Image} sizes="40px" iconClassName="w-5 h-5" />
                                </div>
                                <div className="min-w-0">
                                    <p className="font-medium font-serif truncate">
                                        <Link
                                            href={appHref(`/products/${product.ID}`)}
                                            className="outline-none after:absolute after:inset-0 after:content-[''] group-hover:underline group-focus-within:underline underline-offset-4"
                                        >
                                            {product.Name}
                                        </Link>
                                    </p>
                                    <p className="text-sm text-muted-foreground/60 font-sans truncate">{product.Description}</p>
                                </div>
                            </div>

                            <div className="justify-self-center hidden @3xl:block">
                                {product.Unit &&
                                    <span className="rounded-full border border-border px-2 py-0.5 text-xs font-sans text-muted-foreground">
                                        {product.Unit}
                                    </span>
                                }
                            </div>

                            <div className="justify-self-center">
                                {latestStock
                                    ? <StockBadge stock={latestStock} unit={product.Unit} />
                                    : <span className="text-sm text-muted-foreground">—</span>
                                }
                            </div>

                            <div className="justify-self-center hidden @3xl:block">
                                {latestPrice
                                    ? <PriceBadge price={latestPrice} unit={product.Unit} />
                                    : <span className="text-sm text-muted-foreground">—</span>
                                }
                            </div>
                        </div>
                    )
                })}
            </div>
        </div>
    );
}
