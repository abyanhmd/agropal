'use client';

import React from 'react';
import {useProducts, useTransactions} from "@/lib/api/hooks";
import QueryError from "@/components/app/query-error";
import NotFoundState from "@/components/app/not-found-state";
import OrderDetail, {OrderDetailSkeleton} from "@/app/(app)/orders/_components/detail.order";
import {useAppHref} from "@/lib/demo/base";

// Loads an order, its product, and the product's other orders for the signed-in farm
export default function OrderLoader({id}: {id: string}) {
    const appHref = useAppHref();
    const {data: transactions, error, reload} = useTransactions();
    const {data: products} = useProducts();

    if (error && !transactions) return <div className="page-container"><QueryError error={error} onRetry={reload} /></div>;
    if (!transactions || !products) return <OrderDetailSkeleton />;

    const transaction = transactions.find((t) => t.ID === id);
    if (!transaction) return <NotFoundState what="order" backHref={appHref("/orders")} />;

    const product = products.find((p) => p.ID === transaction.ProductID) ?? null;
    // Already newest first
    const related = transactions.filter((t) => t.ProductID === transaction.ProductID && t.ID !== transaction.ID);

    return <OrderDetail transaction={transaction} product={product} related={related} />;
}
