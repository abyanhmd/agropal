import React from 'react';
import Link from "next/link";
import {ArrowDownLeft, ArrowUpRight, ShoppingCart} from "lucide-react";
import {TransactionSchema} from "@/app/(app)/orders/_types/order.schema";
import {formatPrice} from "@/lib/format";
import {Checkbox} from "@/components/ui/checkbox";
import {useAppHref} from "@/lib/demo/base";

// The product an order belongs to; the API only returns its id, so the page joins in the name
export interface OrderProduct {
    ID: string;
    Name: string;
}

export interface OrderRow {
    transaction: TransactionSchema;
    product?: OrderProduct;
}

// Checkboxes for bulk actions; the header box covers the rows shown
export interface OrderSelection {
    selected: ReadonlySet<string>;
    onChange: (ids: string[], checked: boolean) => void;
}

export interface OrderTableProps {
    rows: OrderRow[];
    emptyMessage?: string;
    selection?: OrderSelection;
}

// Columns follow the table's own width (container queries), not the window's: the sidebar takes a quarter of it
const ROW_GRID = "grid grid-cols-[1fr_auto] @3xl:grid-cols-[1fr_7rem_7rem_8rem] @5xl:grid-cols-[1fr_9rem_9rem_10rem] items-center gap-4 px-4";
// Same columns after a leading checkbox
const SELECTABLE_ROW_GRID = "grid grid-cols-[auto_1fr_auto] @3xl:grid-cols-[auto_1fr_7rem_7rem_8rem] @5xl:grid-cols-[auto_1fr_9rem_9rem_10rem] items-center gap-4 px-4";

// UTC to match how the mock dates are created, so server and browser render the same text
const dateFormat = new Intl.DateTimeFormat("en-US", {month: "short", day: "numeric", year: "numeric", timeZone: "UTC"});
export const formatOrderDate = (date: Date | string) => dateFormat.format(new Date(date));

// "+$120.50" for income, "−£40.00" for expense
export function formatSignedAmount(transaction: TransactionSchema): string {
    return (transaction.Type === "INCOME" ? "+" : "−") + formatPrice(transaction.Amount, transaction.Currency);
}

function TableHeader({rows, selection}: {rows?: OrderRow[], selection?: OrderSelection}) {
    const ids = rows?.map((row) => row.transaction.ID) ?? [];
    const checkedCount = selection ? ids.filter((id) => selection.selected.has(id)).length : 0;

    return (
        <div className={`${selection ? SELECTABLE_ROW_GRID : ROW_GRID} py-2.5 bg-muted border-b border-border text-xs font-medium uppercase tracking-wide text-muted-foreground`}>
            {selection &&
                <Checkbox
                    aria-label="Select the orders shown"
                    className="cursor-pointer bg-background"
                    checked={ids.length > 0 && checkedCount === ids.length}
                    indeterminate={checkedCount > 0 && checkedCount < ids.length}
                    onCheckedChange={(checked) => selection.onChange(ids, checked)}
                />
            }
            <span>Order</span>
            <span className="text-center hidden @3xl:block">Date</span>
            <span className="text-center hidden @3xl:block">Type</span>
            <span className="text-right">Amount</span>
        </div>
    )
}

export function OrderTableSkeleton() {
    return (
        <div className="@container w-full overflow-hidden rounded-xl border border-border bg-background">
            <TableHeader />
            <div className="divide-y divide-border animate-pulse">
                {Array.from({length: 10}, (_, i) => (
                    <div key={i} className={`${ROW_GRID} py-3`}>
                        <div className="flex items-center gap-3">
                            <div className="w-10 h-10 shrink-0 rounded-lg bg-accent-muted"></div>
                            <div className="flex flex-col gap-y-2">
                                <div className="h-4 w-32 rounded bg-muted"></div>
                                <div className="h-3 w-24 rounded bg-muted"></div>
                            </div>
                        </div>
                        <div className="hidden @3xl:block h-3 w-20 rounded bg-muted justify-self-center"></div>
                        <div className="hidden @3xl:block h-6 w-20 rounded-full bg-muted justify-self-center"></div>
                        <div className="h-4 w-16 rounded bg-muted justify-self-end"></div>
                    </div>
                ))}
            </div>
        </div>
    )
}

export function TypeBadge({type}: {type: TransactionSchema["Type"]}) {
    const isIncome = type === "INCOME";
    const Icon = isIncome ? ArrowDownLeft : ArrowUpRight;

    return (
        <span className={`inline-flex items-center gap-1.5 rounded-full px-2.5 py-1 text-sm font-sans ${isIncome ? "bg-accent-muted text-accent-strong" : "bg-danger-muted text-danger-strong"}`}>
            <Icon className="w-4 h-4" />
            {isIncome ? "Income" : "Expense"}
        </span>
    )
}

export function OrderEmptyState({message}: {message?: string}) {
    return (
        <div className="w-full flex flex-col items-center gap-2 rounded-xl border border-dashed border-border py-12 text-muted-foreground">
            <ShoppingCart className="w-8 h-8" />
            <p className="font-sans text-sm">{message ?? "No orders yet"}</p>
        </div>
    )
}

export default function OrderTable({rows, emptyMessage, selection}: OrderTableProps) {
    const appHref = useAppHref();
    if (rows.length === 0) {
        return <OrderEmptyState message={emptyMessage} />
    }

    return (
        <div className="@container w-full overflow-hidden rounded-xl border border-border bg-background">
            <TableHeader rows={rows} selection={selection} />
            <div className="divide-y divide-border">
                {rows.map(({transaction, product}) => {
                    const isIncome = transaction.Type === "INCOME";
                    const isSelected = selection?.selected.has(transaction.ID) ?? false;

                    return (
                        // Same pattern as the product table: the description link stretches over the whole row
                        <div
                            key={transaction.ID}
                            className={`${selection ? SELECTABLE_ROW_GRID : ROW_GRID} group relative py-3 transition-colors hover:bg-muted/60 focus-within:bg-muted/60 ${isSelected ? "bg-accent-muted/40" : ""}`}
                        >
                            {selection &&
                                // Above the stretched link, so ticking it doesn't open the order
                                <Checkbox
                                    aria-label={`Select ${transaction.Description || "untitled order"}`}
                                    className="relative z-10 cursor-pointer"
                                    checked={isSelected}
                                    onCheckedChange={(checked) => selection.onChange([transaction.ID], checked)}
                                />
                            }
                            <div className="flex items-center gap-3 min-w-0">
                                {/* The icon gives its room to the description on phones */}
                                <div className={`w-10 h-10 shrink-0 hidden @3xl:flex items-center justify-center rounded-lg ${isIncome ? "bg-accent-muted text-accent-strong" : "bg-danger-muted text-danger-strong"}`}>
                                    <ShoppingCart className="w-5 h-5" />
                                </div>
                                <div className="min-w-0">
                                    {/* Wraps to two lines on phones, where a single truncated line shows barely a word */}
                                    <p className="font-medium font-serif line-clamp-2 @3xl:line-clamp-1">
                                        <Link
                                            href={appHref(`/orders/${transaction.ID}`)}
                                            className="outline-none after:absolute after:inset-0 after:content-[''] group-hover:underline group-focus-within:underline underline-offset-4"
                                        >
                                            {transaction.Description || "Untitled order"}
                                        </Link>
                                    </p>
                                    <p className="text-sm text-muted-foreground/60 font-sans line-clamp-2 @3xl:line-clamp-1">
                                        {product?.Name ?? "Unknown product"}
                                        {/* Date moves under the name when its column is hidden */}
                                        <span className="@3xl:hidden"> · {formatOrderDate(transaction.Date)}</span>
                                    </p>
                                </div>
                            </div>

                            <span className="justify-self-center hidden @3xl:block text-sm text-muted-foreground font-sans tabular-nums">
                                {formatOrderDate(transaction.Date)}
                            </span>

                            <div className="justify-self-center hidden @3xl:block">
                                <TypeBadge type={transaction.Type} />
                            </div>

                            <span className={`justify-self-end font-sans font-semibold tabular-nums ${isIncome ? "text-accent-strong" : "text-danger-strong"}`}>
                                {formatSignedAmount(transaction)}
                            </span>
                        </div>
                    )
                })}
            </div>
        </div>
    );
}
