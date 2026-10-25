import {DEFAULT_FILTERS, ProductFilters, ProductSort, SORT_OPTIONS, StockTrend} from "@/app/(app)/products/_components/toolbar.product";
import {SearchParamsLike} from "@/lib/search-params";

// URL shape: ?q=tomato&unit=kg&unit=pcs&trend=INCOME&sort=price-asc&page=2
// Default values are left out, so an unfiltered first page is just the bare path.
const KEYS = ["q", "unit", "trend", "sort", "page"] as const;

const TRENDS: StockTrend[] = ["INCOME", "EXPENSE"];

export function parseFilters(params: SearchParamsLike): ProductFilters {
    const sort = params.get("sort");

    return {
        search: params.get("q") ?? "",
        unit: [...new Set(params.getAll("unit").filter(Boolean))],
        trend: [...new Set(params.getAll("trend").filter((t): t is StockTrend => TRENDS.includes(t as StockTrend)))],
        sort: SORT_OPTIONS.some((o) => o.value === sort) ? sort as ProductSort : DEFAULT_FILTERS.sort,
    };
}

// Builds the URL for a filter state + page, keeping any unrelated query params intact
export function filtersHref(pathname: string, current: SearchParamsLike, filters: ProductFilters, page: number): string {
    const params = new URLSearchParams(current.toString());
    KEYS.forEach((key) => params.delete(key));

    const search = filters.search.trim();
    if (search) params.set("q", search);
    filters.unit.forEach((unit) => params.append("unit", unit));
    filters.trend.forEach((trend) => params.append("trend", trend));
    if (filters.sort !== DEFAULT_FILTERS.sort) params.set("sort", filters.sort);
    if (page > 1) params.set("page", String(page));

    const query = params.toString();
    return query ? `${pathname}?${query}` : pathname;
}
