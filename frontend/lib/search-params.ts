// Helpers shared by every list page that keeps its filters in the URL

// Both URLSearchParams and Next's ReadonlyURLSearchParams fit this
export interface SearchParamsLike {
    get(name: string): string | null;
    getAll(name: string): string[];
    toString(): string;
}

// "?page=3" → 3; missing or invalid values fall back to the first page
export function parsePage(params: SearchParamsLike): number {
    const page = Number(params.get("page"));
    return Number.isInteger(page) && page >= 1 ? page : 1;
}
