import {DEFAULT_FILTERS, OrderFilters, OrderSort, SORT_OPTIONS, TYPE_OPTIONS} from "@/app/(app)/orders/_components/toolbar.order";
import {TransactionType} from "@/app/(app)/orders/_types/order.schema";
import {SearchParamsLike} from "@/lib/search-params";

// URL shape: ?q=seed&type=EXPENSE&currency=USD&sort=amount-desc&page=2
// Default values are left out, so an unfiltered first page is just the bare path.
const KEYS = ["q", "type", "currency", "sort", "page"] as const;

const TYPES = TYPE_OPTIONS.map((o) => o.value);

export function parseFilters(params: SearchParamsLike): OrderFilters {
    const sort = params.get("sort");

    return {
        search: params.get("q") ?? "",
        type: [...new Set(params.getAll("type").filter((t): t is TransactionType => TYPES.includes(t as TransactionType)))],
        currency: [...new Set(params.getAll("currency").filter(Boolean))],
        sort: SORT_OPTIONS.some((o) => o.value === sort) ? sort as OrderSort : DEFAULT_FILTERS.sort,
    };
}

// Builds the URL for a filter state + page, keeping any unrelated query params intact
export function filtersHref(pathname: string, current: SearchParamsLike, filters: OrderFilters, page: number): string {
    const params = new URLSearchParams(current.toString());
    KEYS.forEach((key) => params.delete(key));

    const search = filters.search.trim();
    if (search) params.set("q", search);
    filters.type.forEach((type) => params.append("type", type));
    filters.currency.forEach((currency) => params.append("currency", currency));
    if (filters.sort !== DEFAULT_FILTERS.sort) params.set("sort", filters.sort);
    if (page > 1) params.set("page", String(page));

    const query = params.toString();
    return query ? `${pathname}?${query}` : pathname;
}
