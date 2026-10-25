'use client';

import React, {Suspense, useEffect, useEffectEvent, useMemo, useState} from 'react';
import {usePathname, useRouter, useSearchParams} from "next/navigation";
import {usePostAuthors, usePosts} from "@/lib/api/hooks";
import QueryError from "@/components/app/query-error";
import {authorsFrom} from "@/app/(app)/discussions/_lib/authors";
import PostList, {PostListSkeleton} from "@/app/(app)/discussions/_components/list.discussion";
import DiscussionToolbar, {PostFilters} from "@/app/(app)/discussions/_components/toolbar.discussion";
import ListPagination from "@/components/app/list-pagination";
import {useDebouncedValue} from "@/hooks/use-debounced-value";
import {filtersHref, parseFilters} from "@/app/(app)/discussions/_lib/filters-url";
import {parsePage} from "@/lib/search-params";

const PAGE_SIZE = 10;

function DiscussionPageContent() {
    const router = useRouter();
    const pathname = usePathname();
    const searchParams = useSearchParams();

    // Everyone who has written a post, for the author filter
    const {data: authorList = []} = usePostAuthors();

    // The URL is the source of truth for filters, sort, and page
    const urlFilters = useMemo(() => parseFilters(searchParams), [searchParams]);
    const page = parsePage(searchParams);
    const {data, error, reload} = usePosts({
        sort: urlFilters.sort,
        authorIDs: urlFilters.author,
        search: urlFilters.search,
        page,
        pageSize: PAGE_SIZE,
    });
    // Posts on the page carry their authors' names; the author list covers the filter
    const authors = useMemo(
        () => new Map([...authorList, ...authorsFrom(data?.posts ?? [])].map((a) => [a.ID, a])),
        [authorList, data]
    );
    const hrefFor = (filters: PostFilters, target: number) => filtersHref(pathname, searchParams, filters, target);

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
    const handleFiltersChange = (next: PostFilters) => {
        setSearchInput(next.search);

        const nonSearchChanged =
            next.sort !== urlFilters.sort ||
            next.author.join() !== urlFilters.author.join();
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

    // The result set may have shrunk below the current page (e.g. an old link): go to the last one
    const total = data?.total ?? 0;
    const pageCount = Math.max(1, Math.ceil(total / PAGE_SIZE));
    const currentPage = Math.min(page, pageCount);
    const pastTheEnd = data !== undefined && page > pageCount;
    const lastPageHref = hrefFor(urlFilters, pageCount);
    useEffect(() => {
        if (pastTheEnd) router.replace(lastPageHref, {scroll: false});
    }, [pastTheEnd, lastPageHref, router]);

    return (
        <div className="page-container flex flex-col gap-4">
            <DiscussionToolbar
                filters={{...urlFilters, search: searchInput}}
                onChange={handleFiltersChange}
                authors={authorList}
            />
            {data && !pastTheEnd
                ? <>
                    <PostList posts={data.posts} authors={authors} emptyMessage="No discussions match your filters" />
                    <ListPagination
                        page={currentPage}
                        pageSize={PAGE_SIZE}
                        totalItems={total}
                        onPageChange={setPage}
                        hrefForPage={pageHref}
                    />
                </>
                : error ? <QueryError error={error} onRetry={reload} /> : <PostListSkeleton />
            }
        </div>
    );
}

// useSearchParams needs a Suspense boundary so the rest of the route can still be prerendered
function DiscussionPage() {
    return (
        <Suspense fallback={<div className="page-container"><PostListSkeleton /></div>}>
            <DiscussionPageContent />
        </Suspense>
    );
}

export default DiscussionPage;
