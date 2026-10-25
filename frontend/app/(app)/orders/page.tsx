'use client';

import React, {Suspense, useEffect, useEffectEvent, useMemo, useState} from 'react';
import {usePathname, useRouter, useSearchParams} from "next/navigation";
import {useProducts, useTransactions} from "@/lib/api/hooks";
import QueryError from "@/components/app/query-error";
import OrderTable, {OrderRow, OrderSelection, OrderTableSkeleton} from "@/app/(app)/orders/_components/table.order";
import OrderBulkBar from "@/app/(app)/orders/_components/bulk.order";
import OrderToolbar, {OrderFilters} from "@/app/(app)/orders/_components/toolbar.order";
import ListPagination from "@/components/app/list-pagination";
import {useDebouncedValue} from "@/hooks/use-debounced-value";
import {filtersHref, parseFilters} from "@/app/(app)/orders/_lib/filters-url";
import {parsePage} from "@/lib/search-params";

const PAGE_SIZE = 15;

const time = (row: OrderRow) => new Date(row.transaction.Date).getTime();

function applyFilters(rows: OrderRow[], filters: OrderFilters): OrderRow[] {
    const query = filters.search.trim().toLowerCase();

    const filtered = rows.filter(({transaction, product}) => {
        if (query && !transaction.Description.toLowerCase().includes(query) && !product?.Name.toLowerCase().includes(query)) {
            return false;
        }
        if (filters.type.length > 0 && !filters.type.includes(transaction.Type)) {
            return false;
        }
        if (filters.currency.length > 0 && !filters.currency.includes(transaction.Currency)) {
            return false;
        }
        return true;
    });

    return filtered.sort((a, b) => {
        switch (filters.sort) {
            case "date-desc": return time(b) - time(a);
            case "date-asc": return time(a) - time(b);
            case "amount-desc": return b.transaction.Amount - a.transaction.Amount;
            case "amount-asc": return a.transaction.Amount - b.transaction.Amount;
        }
    });
}

function OrderPageContent() {
    const router = useRouter();
    const pathname = usePathname();
    const searchParams = useSearchParams();

    const {data: transactions, error: ordersError, reload} = useTransactions();
    const {data: products} = useProducts();
    const error = ordersError;

    const productOptions = useMemo(() => (products ?? []).map((p) => ({ID: p.ID, Name: p.Name})), [products]);
    const rows = useMemo<OrderRow[] | null>(() => {
        if (!transactions || !products) return null;
        const byID = new Map(productOptions.map((p) => [p.ID, p]));
        return transactions.map((transaction) => ({transaction, product: byID.get(transaction.ProductID)}));
    }, [transactions, products, productOptions]);

    // The URL is the source of truth for filters, sort, and page
    const urlFilters = useMemo(() => parseFilters(searchParams), [searchParams]);
    const page = parsePage(searchParams);
    const hrefFor = (filters: OrderFilters, target: number) => filtersHref(pathname, searchParams, filters, target);

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
    const handleFiltersChange = (next: OrderFilters) => {
        setSearchInput(next.search);

        const nonSearchChanged =
            next.sort !== urlFilters.sort ||
            next.type.join() !== urlFilters.type.join() ||
            next.currency.join() !== urlFilters.currency.join();
        const searchCleared = next.search === "" && urlFilters.search !== "";

        // Plain typing is handled by the debounced effect above
        if (!nonSearchChanged && !searchCleared) return;

        // Keep the URL's search unless it was cleared; any change starts again from page 1
        const href = hrefFor({...next, search: searchCleared ? "" : urlFilters.search}, 1);
        // push, so Back undoes a filter or sort change
        router.push(href, {scroll: false});
    };

    // push, so the browser back button steps through pages
    const setPage = (target: number) => router.push(hrefFor(urlFilters, target), {scroll: false});
    const pageHref = (target: number) => hrefFor(urlFilters, target);

    const currencies = useMemo(
        () => [...new Set((rows ?? []).map((r) => r.transaction.Currency).filter(Boolean))].sort(),
        [rows]
    );

    const visibleRows = useMemo(
        () => rows ? applyFilters(rows, urlFilters) : null,
        [rows, urlFilters]
    );

    // Clamp in case the result set shrank below the current page
    const pageCount = Math.max(1, Math.ceil((visibleRows?.length ?? 0) / PAGE_SIZE));
    const currentPage = Math.min(page, pageCount);
    const pageRows = visibleRows?.slice((currentPage - 1) * PAGE_SIZE, currentPage * PAGE_SIZE);

    // ── Selection for bulk delete: kept across pages and filter changes, until cleared ──
    const [selected, setSelected] = useState<ReadonlySet<string>>(new Set());
    // Only orders that still exist count (one may have been deleted elsewhere)
    const selectedIDs = useMemo(
        () => (transactions ?? []).map((t) => t.ID).filter((id) => selected.has(id)),
        [transactions, selected]
    );
    const matchingIDs = useMemo(() => (visibleRows ?? []).map((row) => row.transaction.ID), [visibleRows]);

    const selection: OrderSelection = {
        selected,
        onChange: (ids, checked) => setSelected((prev) => {
            const next = new Set(prev);
            for (const id of ids) {
                if (checked) next.add(id);
                else next.delete(id);
            }
            return next;
        }),
    };

    return (
        <div className="page-container flex flex-col gap-4">
            <OrderToolbar
                filters={{...urlFilters, search: searchInput}}
                onChange={handleFiltersChange}
                currencies={currencies}
                products={productOptions}
                orderCount={transactions?.length ?? 0}
            />
            {visibleRows && pageRows
                ? <>
                    {selectedIDs.length > 0 &&
                        <OrderBulkBar
                            selectedIDs={selectedIDs}
                            matchingIDs={matchingIDs}
                            onSelect={(ids) => selection.onChange(ids, true)}
                            onClear={() => setSelected(new Set())}
                        />
                    }
                    <OrderTable rows={pageRows} emptyMessage="No orders match your filters" selection={selection} />
                    <ListPagination
                        page={currentPage}
                        pageSize={PAGE_SIZE}
                        totalItems={visibleRows.length}
                        onPageChange={setPage}
                        hrefForPage={pageHref}
                    />
                </>
                : error ? <QueryError error={error} onRetry={reload} />
                : <OrderTableSkeleton />
            }
        </div>
    );
}

// useSearchParams needs a Suspense boundary so the rest of the route can still be prerendered
function OrderPage() {
    return (
        <Suspense fallback={<div className="page-container"><OrderTableSkeleton /></div>}>
            <OrderPageContent />
        </Suspense>
    );
}

export default OrderPage;
