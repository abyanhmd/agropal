import {DEFAULT_FILTERS, PostFilters, PostSort, SORT_OPTIONS} from "@/app/(app)/discussions/_components/toolbar.discussion";
import {SearchParamsLike} from "@/lib/search-params";

// URL shape: ?q=compost&author=user-3&sort=most-liked&page=2
// Default values are left out, so an unfiltered first page is just the bare path.
const KEYS = ["q", "author", "sort", "page"] as const;

export function parseFilters(params: SearchParamsLike): PostFilters {
    const sort = params.get("sort");

    return {
        search: params.get("q") ?? "",
        author: [...new Set(params.getAll("author").filter(Boolean))],
        sort: SORT_OPTIONS.some((o) => o.value === sort) ? sort as PostSort : DEFAULT_FILTERS.sort,
    };
}

// Builds the URL for a filter state + page, keeping any unrelated query params intact
export function filtersHref(pathname: string, current: SearchParamsLike, filters: PostFilters, page: number): string {
    const params = new URLSearchParams(current.toString());
    KEYS.forEach((key) => params.delete(key));

    const search = filters.search.trim();
    if (search) params.set("q", search);
    filters.author.forEach((author) => params.append("author", author));
    if (filters.sort !== DEFAULT_FILTERS.sort) params.set("sort", filters.sort);
    if (page > 1) params.set("page", String(page));

    const query = params.toString();
    return query ? `${pathname}?${query}` : pathname;
}
