import React from 'react';
import Link from "next/link";
import {ProductSchema} from "@/app/(app)/products/_types/product.schema";
import {PriceBadge, ProductEmptyState, StockBadge} from "@/app/(app)/products/_components/table.product";
import ProductImage from "@/app/(app)/products/_components/image.product";
import {useAppHref} from "@/lib/demo/base";

export interface ProductGridProps {
    products: ProductSchema[];
    emptyMessage?: string;
}

const GRID = "grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4";

export function ProductGridSkeleton({count = 12}: {count?: number}) {
    return (
        <div className={`${GRID} animate-pulse`}>
            {Array.from({length: count}, (_, i) => (
                <div key={i} className="overflow-hidden rounded-xl border border-border bg-background">
                    <div className="aspect-[4/3] bg-accent-muted" />
                    <div className="flex flex-col gap-2 p-4">
                        <div className="h-4 w-2/3 rounded bg-muted" />
                        <div className="h-3 w-full rounded bg-muted" />
                        <div className="mt-2 flex justify-between">
                            <div className="h-6 w-20 rounded-full bg-muted" />
                            <div className="h-6 w-16 rounded-full bg-muted" />
                        </div>
                    </div>
                </div>
            ))}
        </div>
    );
}

function ProductCard({product}: {product: ProductSchema}) {
    const appHref = useAppHref();
    // Stocks and prices are stored oldest → newest, so the latest is the last entry
    const latestStock = product.Stocks?.[product.Stocks.length - 1];
    const latestPrice = product.Prices?.[product.Prices.length - 1];

    return (
        // The title link stretches over the whole card (after:inset-0), so the card is one click target
        // while screen readers still get a single, clearly named link
        <article className="group relative flex flex-col overflow-hidden rounded-xl border border-border bg-background transition-shadow hover:shadow-md focus-within:ring-3 focus-within:ring-ring/50">
            <div className="relative flex aspect-[4/3] items-center justify-center bg-accent-muted text-accent-strong">
                <ProductImage
                    src={product.Image}
                    sizes="(min-width: 1280px) 25vw, (min-width: 1024px) 33vw, (min-width: 640px) 50vw, 100vw"
                    iconClassName="size-10 transition-transform group-hover:scale-110"
                />
                {product.Unit &&
                    <span className="absolute right-3 top-3 rounded-full border border-border bg-background/80 px-2 py-0.5 text-xs font-sans text-muted-foreground backdrop-blur">
                        {product.Unit}
                    </span>
                }
            </div>

            <div className="flex flex-1 flex-col gap-3 p-4">
                <div className="min-w-0">
                    <h3 className="truncate font-serif font-medium">
                        <Link
                            href={appHref(`/products/${product.ID}`)}
                            className="outline-none after:absolute after:inset-0 after:content-[''] group-hover:underline underline-offset-4"
                        >
                            {product.Name}
                        </Link>
                    </h3>
                    <p className="line-clamp-2 text-sm text-muted-foreground font-sans">{product.Description}</p>
                </div>

                <div className="mt-auto flex flex-wrap items-center justify-between gap-2">
                    {latestPrice
                        ? <PriceBadge price={latestPrice} unit={product.Unit} />
                        : <span className="text-sm text-muted-foreground">—</span>
                    }
                    {latestStock && <StockBadge stock={latestStock} unit={product.Unit} />}
                </div>
            </div>
        </article>
    );
}

export default function ProductGrid({products, emptyMessage}: ProductGridProps) {
    if (products.length === 0) {
        return <ProductEmptyState message={emptyMessage} />;
    }

    return (
        <div className={GRID}>
            {products.map((product) => (
                <ProductCard key={product.ID} product={product} />
            ))}
        </div>
    );
}
