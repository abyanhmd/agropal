'use client';

import React, {Suspense, useEffect, useEffectEvent, useMemo, useState} from 'react';
import {usePathname, useRouter, useSearchParams} from "next/navigation";
import {ProductSchema} from "@/app/(app)/products/_types/product.schema";
import {useProducts} from "@/lib/api/hooks";
import QueryError from "@/components/app/query-error";
import ProductTable, {ProductTableSkeleton} from "@/app/(app)/products/_components/table.product";
import ProductToolbar, {ProductFilters, ProductView} from "@/app/(app)/products/_components/toolbar.product";
import ProductGrid, {ProductGridSkeleton} from "@/app/(app)/products/_components/grid.product";
import ListPagination from "@/components/app/list-pagination";
import {useDebouncedValue} from "@/hooks/use-debounced-value";
import {filtersHref, parseFilters} from "@/app/(app)/products/_lib/filters-url";
import {parsePage} from "@/lib/search-params";

// 12 fills grid rows evenly at 1, 2, 3, and 4 columns
const PAGE_SIZE = 12;

// Prices and stocks are stored oldest → newest, so the last entry is the latest
const latestPrice = (product: ProductSchema) => product.Prices?.[product.Prices.length - 1]?.Price ?? 0;
const latestStock = (product: ProductSchema) => product.Stocks?.[product.Stocks.length - 1];

function applyFilters(products: ProductSchema[], filters: ProductFilters): ProductSchema[] {
    const query = filters.search.trim().toLowerCase();

    const filtered = products.filter((product) => {
        if (query && !product.Name.toLowerCase().includes(query) && !product.Description.toLowerCase().includes(query)) {
            return false;
        }
        if (filters.unit.length > 0 && !filters.unit.includes(product.Unit)) {
            return false;
        }
        const trend = latestStock(product)?.Type;
        if (filters.trend.length > 0 && !(trend && (filters.trend as string[]).includes(trend))) {
            return false;
        }
        return true;
    });

    return filtered.sort((a, b) => {
        switch (filters.sort) {
            case "name-asc": return a.Name.localeCompare(b.Name, undefined, {numeric: true});
            case "name-desc": return b.Name.localeCompare(a.Name, undefined, {numeric: true});
            case "price-asc": return latestPrice(a) - latestPrice(b);
            case "price-desc": return latestPrice(b) - latestPrice(a);
            case "stock-desc": return (latestStock(b)?.Stock ?? 0) - (latestStock(a)?.Stock ?? 0);
            case "stock-asc": return (latestStock(a)?.Stock ?? 0) - (latestStock(b)?.Stock ?? 0);
        }
    });
}

function ProductPageContent() {
    const router = useRouter();
    const pathname = usePathname();
    const searchParams = useSearchParams();

    const {data: products, error, reload} = useProducts();

    // The URL is the source of truth for filters, sort, and page
    const urlFilters = useMemo(() => parseFilters(searchParams), [searchParams]);
    const page = parsePage(searchParams);
    const view: ProductView = searchParams.get("view") === "grid" ? "grid" : "list";
    const hrefFor = (filters: ProductFilters, target: number) => filtersHref(pathname, searchParams, filters, target);

    // ── Search: the input is local so typing stays instant; the URL catches up after a pause ──
    const [searchInput, setSearchInput] = useState(urlFilters.search);
    const debouncedSearch = useDebouncedValue(searchInput, 300);
    // An emptied box applies right away so clearing doesn't lag
    const settledSearch = searchInput === "" ? "" : debouncedSearch;

    // When the URL's q changes from outside (back/forward, clear), mirror it into the input.
    // Skipped when it matches what we just wrote, so in-flight typing isn't overwritten.
    const [prevUrlSearch, setPrevUrlSearch] = useState(urlFilters.search);
    if (urlFilters.search !== prevUrlSearch) {
        setPrevUrlSearch(urlFilters.search);
        if (urlFilters.search !== settledSearch.trim()) setSearchInput(urlFilters.search);
    }

    const writeSearchToUrl = useEffectEvent((value: string) => {
        if (value.trim() === urlFilters.search) return;
        // replace, so each search doesn't become its own history entry
        router.replace(hrefFor({...urlFilters, search: value}, 1), {scroll: false});
    });
    useEffect(() => {
        writeSearchToUrl(settledSearch);
    }, [settledSearch]);

    // ── Filters / sort / page ──
    const handleFiltersChange = (next: ProductFilters) => {
        setSearchInput(next.search);

        const nonSearchChanged =
            next.sort !== urlFilters.sort ||
            next.unit.join() !== urlFilters.unit.join() ||
            next.trend.join() !== urlFilters.trend.join();
        const searchCleared = next.search === "" && urlFilters.search !== "";

        // Plain typing is handled by the debounced effect above
        if (!nonSearchChanged && !searchCleared) return;

        // Keep the URL's search unless it was cleared; any change starts again from page 1
        const href = hrefFor({...next, search: searchCleared ? "" : urlFilters.search}, 1);
        // push, so Back undoes a filter or sort change
        router.push(href, {scroll: false});
    };

    // List is the default, so only ?view=grid is written; replace, since switching views isn't navigation
    const setView = (next: ProductView) => {
        const params = new URLSearchParams(searchParams.toString());
        if (next === "grid") params.set("view", "grid");
        else params.delete("view");
        const query = params.toString();
        router.replace(query ? `${pathname}?${query}` : pathname, {scroll: false});
    };

    // push, so the browser back button steps through pages
    const setPage = (target: number) => router.push(hrefFor(urlFilters, target), {scroll: false});
    const pageHref = (target: number) => hrefFor(urlFilters, target);

    const units = useMemo(() => [...new Set((products ?? []).map((p) => p.Unit).filter(Boolean))].sort(), [products]);

    const visibleProducts = useMemo(
        () => products ? applyFilters(products, urlFilters) : null,
        [products, urlFilters]
    );

    // Clamp in case the result set shrank below the current page
    const pageCount = Math.max(1, Math.ceil((visibleProducts?.length ?? 0) / PAGE_SIZE));
    const currentPage = Math.min(page, pageCount);
    const pageProducts = visibleProducts?.slice((currentPage - 1) * PAGE_SIZE, currentPage * PAGE_SIZE);

    return (
        <>
            <div>
                <div className="page-container flex flex-col gap-4">
                    <ProductToolbar
                        filters={{...urlFilters, search: searchInput}}
                        onChange={handleFiltersChange}
                        units={units}
                        view={view}
                        onViewChange={setView}
                    />
                    {visibleProducts && pageProducts
                        ? <>
                            {view === "grid"
                                ? <ProductGrid products={pageProducts} emptyMessage="No products match your filters" />
                                : <ProductTable products={pageProducts} emptyMessage="No products match your filters" />
                            }
                            <ListPagination
                                page={currentPage}
                                pageSize={PAGE_SIZE}
                                totalItems={visibleProducts.length}
                                onPageChange={setPage}
                                hrefForPage={pageHref}
                            />
                        </>
                        : error ? <QueryError error={error} onRetry={reload} />
                        : view === "grid" ? <ProductGridSkeleton count={PAGE_SIZE} /> : <ProductTableSkeleton />
                    }
                </div>
            </div>
        </>
    );
}

// useSearchParams needs a Suspense boundary so the rest of the route can still be prerendered
function ProductPage() {
    return (
        <Suspense fallback={<div className="page-container"><ProductTableSkeleton /></div>}>
            <ProductPageContent />
        </Suspense>
    );
}

export default ProductPage;