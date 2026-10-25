'use client';

import React from 'react';
import {usePathname, useRouter, useSearchParams} from "next/navigation";
import {ProductSchema} from "@/app/(app)/products/_types/product.schema";
import Dashboard from "@/app/(app)/dashboard/_components/dashboard";
import {defaultUnit, unitsOf} from "@/app/(app)/dashboard/_lib/dashboard";

export default function DashboardFromUrl({products, loading}: {products: ProductSchema[], loading: boolean}) {
    const router = useRouter();
    const pathname = usePathname();
    const searchParams = useSearchParams();

    const units = unitsOf(products);
    const fallback = defaultUnit(products);
    const requested = searchParams.get("unit");
    // Unknown units (old links, typos) fall back to the default
    const unit = requested && units.includes(requested) ? requested : fallback;

    const setUnit = (next: string) => {
        const params = new URLSearchParams(searchParams.toString());
        if (next === fallback) params.delete("unit");
        else params.set("unit", next);
        const query = params.toString();
        // push, so Back returns to the previous unit
        router.push(query ? `${pathname}?${query}` : pathname, {scroll: false});
    };

    return <Dashboard products={products} unit={unit} loading={loading} onUnitChange={setUnit} />;
}
