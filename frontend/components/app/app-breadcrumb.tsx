'use client';

import React from 'react';
import Link from "next/link";
import {
    Breadcrumb,
    BreadcrumbEllipsis,
    BreadcrumbItem,
    BreadcrumbLink,
    BreadcrumbList,
    BreadcrumbPage,
    BreadcrumbSeparator
} from "@/components/ui/breadcrumb";
import {BreadcrumbItemData, useBreadcrumb} from "@/hooks/breadcrumb-state";
import {useAppHref} from "@/lib/demo/base";

// Long trails collapse to: first › … › last two
const MAX_VISIBLE = 3;

// The last crumb is the long one (a post title, a product name), so it alone gives way and truncates
function Crumb({item, last}: {item: BreadcrumbItemData, last: boolean}) {
    // Crumb links are in-app paths; the demo keeps them inside /demo
    const appHref = useAppHref();

    return (
        <BreadcrumbItem className={last ? "min-w-0 font-serif text-lg" : "shrink-0 font-serif text-lg"}>
            {item.href
                ? <BreadcrumbLink className="truncate" render={<Link href={appHref(item.href)} />}>{item.label}</BreadcrumbLink>
                : <BreadcrumbPage className="truncate">{item.label}</BreadcrumbPage>
            }
        </BreadcrumbItem>
    )
}

function AppBreadcrumb() {
    const {items} = useBreadcrumb();
    // Fallback for a page with no segments of its own
    const crumbs: BreadcrumbItemData[] = items.length > 0 ? items : [{label: "Dashboard"}];

    const collapsed = crumbs.length > MAX_VISIBLE;
    const head = collapsed ? crumbs.slice(0, 1) : crumbs;
    const tail = collapsed ? crumbs.slice(-2) : [];

    return (
        <Breadcrumb className="min-w-0">
            <BreadcrumbList className="flex-nowrap">
                {head.map((item, i) => (
                    <React.Fragment key={item.label + i}>
                        {i > 0 && <BreadcrumbSeparator />}
                        <Crumb item={item} last={!collapsed && i === head.length - 1} />
                    </React.Fragment>
                ))}

                {collapsed &&
                    <>
                        <BreadcrumbSeparator />
                        <BreadcrumbItem>
                            <BreadcrumbEllipsis />
                        </BreadcrumbItem>
                        {tail.map((item, i) => (
                            <React.Fragment key={item.label + i}>
                                <BreadcrumbSeparator />
                                <Crumb item={item} last={i === tail.length - 1} />
                            </React.Fragment>
                        ))}
                    </>
                }
            </BreadcrumbList>
        </Breadcrumb>
    );
}

export default AppBreadcrumb;
