'use client';

import React from 'react';
import {useProduct, useProductTransactions} from "@/lib/api/hooks";
import {ApiError} from "@/lib/api/client";
import QueryError from "@/components/app/query-error";
import NotFoundState from "@/components/app/not-found-state";
import ProductDetail, {ProductDetailSkeleton} from "@/app/(app)/products/_components/detail.product";
import {useAppHref} from "@/lib/demo/base";

// Loads a product and its orders for the signed-in user's farm
export default function ProductLoader({id}: {id: string}) {
    const appHref = useAppHref();
    const {data: product, error, reload} = useProduct(id);
    const {data: orders} = useProductTransactions(id);

    if (error instanceof ApiError && error.status === 404) {
        return <NotFoundState what="product" backHref={appHref("/products")} />;
    }
    if (error && !product) return <div className="page-container"><QueryError error={error} onRetry={reload} /></div>;
    if (!product) return <ProductDetailSkeleton />;

    return <ProductDetail product={product} orders={orders ?? []} />;
}
