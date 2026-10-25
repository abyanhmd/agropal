'use client';

import React, {Suspense} from 'react';
import {useProducts} from "@/lib/api/hooks";
import QueryError from "@/components/app/query-error";
import Dashboard from "@/app/(app)/dashboard/_components/dashboard";
import DashboardFromUrl from "@/app/(app)/dashboard/_components/dashboard-url";
import {defaultUnit} from "@/app/(app)/dashboard/_lib/dashboard";

export default function DashboardLoader() {
    const {data: products, error, reload} = useProducts();

    if (error && !products) return <div className="page-container"><QueryError error={error} onRetry={reload} /></div>;

    const loaded = products ?? [];
    // ?unit= is only known at request time, so reading it needs Suspense. The fallback is the
    // dashboard for the default unit.
    return (
        <Suspense fallback={<Dashboard products={loaded} unit={defaultUnit(loaded)} loading={!products} />}>
            <DashboardFromUrl products={loaded} loading={!products} />
        </Suspense>
    );
}
