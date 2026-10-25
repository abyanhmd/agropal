'use client';

import React from 'react';
import Link from "next/link";
import {ArrowLeft, ChevronRight, Package} from "lucide-react";
import {TransactionSchema} from "@/app/(app)/orders/_types/order.schema";
import {ProductSchema} from "@/app/(app)/products/_types/product.schema";
import {formatOrderDate, formatSignedAmount, TypeBadge} from "@/app/(app)/orders/_components/table.order";
import {useSetBreadcrumb} from "@/hooks/breadcrumb-state";
import OrderActions from "@/app/(app)/orders/_components/actions.order";
import {useAppHref} from "@/lib/demo/base";

export interface OrderDetailProps {
    transaction: TransactionSchema;
    product: ProductSchema | null;
    related: TransactionSchema[]; // other orders for the same product, newest first
}

function StatCard({label, children}: {label: string, children: React.ReactNode}) {
    return (
        <div className="flex flex-col gap-1.5 rounded-xl border border-border bg-background p-4">
            <span className="text-xs font-medium uppercase tracking-wide text-muted-foreground">{label}</span>
            <div className="font-sans">{children}</div>
        </div>
    );
}

export function OrderDetailSkeleton() {
    return (
        <div className="page-container flex animate-pulse flex-col gap-6">
            <div className="h-4 w-32 rounded bg-muted" />
            <div className="flex flex-col gap-3">
                <div className="h-8 w-1/2 rounded bg-muted" />
                <div className="h-4 w-40 rounded bg-muted" />
            </div>
            <div className="grid grid-cols-1 gap-3 sm:grid-cols-3">
                {Array.from({length: 3}, (_, i) => <div key={i} className="h-20 rounded-xl bg-muted" />)}
            </div>
            <div className="grid gap-6 lg:grid-cols-2">
                <div className="h-40 rounded-xl bg-muted" />
                <div className="h-40 rounded-xl bg-muted" />
            </div>
        </div>
    );
}

export default function OrderDetail({transaction, product, related}: OrderDetailProps) {
    const appHref = useAppHref();
    const title = transaction.Description || "Untitled order";
    useSetBreadcrumb([{label: "Orders", href: "/orders"}, {label: title}]);

    const isIncome = transaction.Type === "INCOME";

    return (
        <div className="page-container flex flex-col gap-6">
            <Link
                href={appHref("/orders")}
                className="inline-flex w-fit items-center gap-1.5 text-sm text-muted-foreground font-sans transition-colors hover:text-foreground"
            >
                <ArrowLeft className="size-4" />
                Back to orders
            </Link>

            <div className="flex flex-wrap items-start justify-between gap-3">
                <div className="flex min-w-0 flex-col gap-2">
                    <div className="flex flex-wrap items-center gap-3">
                        <h1 className="font-serif text-3xl font-medium">{title}</h1>
                        <TypeBadge type={transaction.Type} />
                    </div>
                    <p className="text-muted-foreground font-sans text-sm">Order {transaction.ID}</p>
                </div>
                <OrderActions order={transaction} />
            </div>

            <div className="grid grid-cols-1 gap-3 sm:grid-cols-3">
                <StatCard label="Amount">
                    <p className={`text-xl font-semibold tabular-nums ${isIncome ? "text-accent-strong" : "text-danger-strong"}`}>
                        {formatSignedAmount(transaction)}
                    </p>
                </StatCard>

                <StatCard label="Date">
                    <p className="text-xl font-semibold tabular-nums">{formatOrderDate(transaction.Date)}</p>
                </StatCard>

                <StatCard label="Currency">
                    <p className="text-xl font-semibold">{transaction.Currency}</p>
                </StatCard>
            </div>

            <div className="grid gap-6 lg:grid-cols-2">
                <section className="overflow-hidden rounded-xl border border-border bg-background">
                    <h2 className="border-b border-border bg-muted px-4 py-2.5 text-xs font-medium uppercase tracking-wide text-muted-foreground">
                        Product
                    </h2>
                    {product
                        ? <Link
                            href={appHref(`/products/${product.ID}`)}
                            className="group flex items-center gap-3 px-4 py-3 outline-none transition-colors hover:bg-muted/60 focus-visible:bg-muted/60"
                        >
                            <div className="w-10 h-10 shrink-0 flex items-center justify-center rounded-lg bg-accent-muted text-accent-strong">
                                <Package className="w-5 h-5" />
                            </div>
                            <div className="min-w-0 flex-1">
                                <p className="font-medium font-serif truncate group-hover:underline underline-offset-4">{product.Name}</p>
                                <p className="text-sm text-muted-foreground/60 font-sans truncate">{product.Description}</p>
                            </div>
                            {product.Unit &&
                                <span className="rounded-full border border-border px-2 py-0.5 text-xs font-sans text-muted-foreground">
                                    {product.Unit}
                                </span>
                            }
                            <ChevronRight className="size-4 text-muted-foreground" />
                        </Link>
                        : <p className="px-4 py-6 text-center text-sm text-muted-foreground font-sans">This product no longer exists</p>
                    }
                </section>

                <section className="overflow-hidden rounded-xl border border-border bg-background">
                    <h2 className="border-b border-border bg-muted px-4 py-2.5 text-xs font-medium uppercase tracking-wide text-muted-foreground">
                        Other orders for this product
                    </h2>
                    <ul className="max-h-96 divide-y divide-border overflow-y-auto">
                        {related.map((other) => (
                            <li key={other.ID}>
                                <Link
                                    href={appHref(`/orders/${other.ID}`)}
                                    className="flex items-center justify-between gap-4 px-4 py-2.5 font-sans outline-none transition-colors hover:bg-muted/60 focus-visible:bg-muted/60"
                                >
                                    <span className="min-w-0">
                                        <span className="block truncate text-sm">{other.Description || "Untitled order"}</span>
                                        <span className="block text-xs text-muted-foreground tabular-nums">{formatOrderDate(other.Date)}</span>
                                    </span>
                                    <span className={`shrink-0 font-semibold tabular-nums ${other.Type === "INCOME" ? "text-accent-strong" : "text-danger-strong"}`}>
                                        {formatSignedAmount(other)}
                                    </span>
                                </Link>
                            </li>
                        ))}
                        {related.length === 0 && <li className="px-4 py-6 text-center text-sm text-muted-foreground font-sans">No other orders yet</li>}
                    </ul>
                </section>
            </div>
        </div>
    );
}
