import React from 'react';
import {
    Pagination,
    PaginationContent,
    PaginationEllipsis,
    PaginationItem,
    PaginationLink,
    PaginationNext,
    PaginationPrevious
} from "@/components/ui/pagination";

export interface ListPaginationProps {
    page: number;       // 1-based
    pageSize: number;
    totalItems: number;
    onPageChange: (page: number) => void;
    hrefForPage?: (page: number) => string; // real link targets, so middle-click / open in new tab work
}

// First, last, and the pages around the current one; gaps become "…"
// e.g. page 6 of 12 → [1, "…", 5, 6, 7, "…", 12]
function pageList(page: number, pageCount: number): (number | "ellipsis")[] {
    const pages = new Set([1, pageCount, page - 1, page, page + 1]);
    const sorted = [...pages].filter((p) => p >= 1 && p <= pageCount).sort((a, b) => a - b);

    const result: (number | "ellipsis")[] = [];
    sorted.forEach((p, i) => {
        const prev = sorted[i - 1];
        if (prev !== undefined && p - prev === 2) result.push(prev + 1); // a gap of one page is just shown
        else if (prev !== undefined && p - prev > 2) result.push("ellipsis");
        result.push(p);
    });
    return result;
}

export default function ListPagination({page, pageSize, totalItems, onPageChange, hrefForPage}: ListPaginationProps) {
    const pageCount = Math.max(1, Math.ceil(totalItems / pageSize));
    if (totalItems === 0) return null;

    const from = (page - 1) * pageSize + 1;
    const to = Math.min(page * pageSize, totalItems);

    const href = (target: number) => hrefForPage?.(Math.min(Math.max(target, 1), pageCount)) ?? "#";

    const go = (target: number) => (e: React.MouseEvent) => {
        // Let modified clicks (new tab / window) use the real href
        if (e.metaKey || e.ctrlKey || e.shiftKey || e.button !== 0) return;
        e.preventDefault();
        if (target >= 1 && target <= pageCount && target !== page) onPageChange(target);
    };
    const disabledClass = "pointer-events-none opacity-50";

    return (
        <div className="flex flex-col items-center gap-3 sm:flex-row sm:justify-between">
            <p className="text-sm text-muted-foreground font-sans tabular-nums">
                Showing {from}–{to} of {totalItems}
            </p>

            {pageCount > 1 &&
                <Pagination className="mx-0 w-auto">
                    <PaginationContent>
                        <PaginationItem>
                            <PaginationPrevious
                                href={href(page - 1)}
                                onClick={go(page - 1)}
                                aria-disabled={page === 1}
                                className={page === 1 ? disabledClass : undefined}
                            />
                        </PaginationItem>

                        {pageList(page, pageCount).map((item, i) => (
                            <PaginationItem key={item === "ellipsis" ? `ellipsis-${i}` : item}>
                                {item === "ellipsis"
                                    ? <PaginationEllipsis />
                                    : <PaginationLink href={href(item)} isActive={item === page} onClick={go(item)}>
                                        {item}
                                    </PaginationLink>
                                }
                            </PaginationItem>
                        ))}

                        <PaginationItem>
                            <PaginationNext
                                href={href(page + 1)}
                                onClick={go(page + 1)}
                                aria-disabled={page === pageCount}
                                className={page === pageCount ? disabledClass : undefined}
                            />
                        </PaginationItem>
                    </PaginationContent>
                </Pagination>
            }
        </div>
    );
}
