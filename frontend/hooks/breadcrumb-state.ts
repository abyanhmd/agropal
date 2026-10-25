'use client';

import React, {createContext, useContext, useEffect, useMemo, useState} from 'react';
import {useSelectedLayoutSegments} from "next/navigation";

export interface BreadcrumbItemData {
    label: string;
    href?: string; // omit for the current page (rendered as BreadcrumbPage)
}

interface BreadcrumbContextValue {
    customItems: BreadcrumbItemData[] | null;
    setItems: (items: BreadcrumbItemData[] | null) => void;
}

const BreadcrumbContext = createContext<BreadcrumbContextValue | null>(null);

export const NOT_FOUND_LABEL = "Lost in the fields";

// Next renders unknown URLs as its built-in "/_not-found" route
const NOT_FOUND_SEGMENT = "/_not-found";

// ["products", "new-item"] → [{label: "Products", href: "/products"}, {label: "New Item"}]
function itemsFromSegments(routeSegments: string[]): BreadcrumbItemData[] {
    if (routeSegments.includes(NOT_FOUND_SEGMENT)) return [{label: NOT_FOUND_LABEL}];
    // Route groups like "(marketing)" aren't part of the URL
    const segments = routeSegments.filter((segment) => !segment.startsWith("("));

    return segments.map((segment, index) => {
        const label = decodeURIComponent(segment)
            .replace(/[-_]/g, " ")
            .replace(/\b\w/g, (c) => c.toUpperCase());
        const isLast = index === segments.length - 1;

        return isLast ? {label} : {label, href: "/" + segments.slice(0, index + 1).join("/")};
    });
}

// The provider doesn't read the route itself: it wraps the whole app, and reading the
// route there would block prerendering of every dynamic route below it
export function BreadcrumbProvider({children}: {children: React.ReactNode}) {
    const [customItems, setItems] = useState<BreadcrumbItemData[] | null>(null);
    const value = useMemo<BreadcrumbContextValue>(() => ({customItems, setItems}), [customItems]);

    return React.createElement(BreadcrumbContext.Provider, {value}, children);
}

function useBreadcrumbContext(): BreadcrumbContextValue {
    const context = useContext(BreadcrumbContext);
    if (!context) {
        throw new Error("Breadcrumb hooks must be used inside <BreadcrumbProvider>");
    }
    return context;
}

// Read the current breadcrumb items (use this in the breadcrumb UI component, from the app layout).
// It reads the route, so render its component inside <Suspense>.
// Built from the rendered route's segments rather than usePathname: on a 404 the server renders
// "/_not-found" while the browser's pathname is the real URL, so a pathname-based crumb
// wouldn't match between the two and would break hydration.
export function useBreadcrumb(): {items: BreadcrumbItemData[]} {
    const {customItems} = useBreadcrumbContext();
    const segments = useSelectedLayoutSegments();
    return {items: customItems ?? itemsFromSegments(segments)};
}

// Override the breadcrumb from a page; it resets to the route default when the page unmounts
export function useSetBreadcrumb(items: BreadcrumbItemData[]) {
    const {setItems} = useBreadcrumbContext();
    // Compare by content so an inline array literal doesn't re-run the effect every render
    const key = JSON.stringify(items);

    useEffect(() => {
        setItems(JSON.parse(key));
        return () => setItems(null);
    }, [key, setItems]);
}
