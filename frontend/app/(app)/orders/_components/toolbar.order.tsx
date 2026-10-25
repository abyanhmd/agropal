import React from 'react';
import {ArrowUpDown, Search, X} from "lucide-react";
import {Input} from "@/components/ui/input";
import {Button} from "@/components/ui/button";
import {Select, SelectContent, SelectItem, SelectTrigger, SelectValue} from "@/components/ui/select";
import AddOrderDialog from "@/app/(app)/orders/_components/add.order";
import DeleteAllOrdersButton from "@/app/(app)/orders/_components/delete-all.order";
import type {OrderProduct} from "@/app/(app)/orders/_components/table.order";
import FilterTreeDialog, {FilterGroup, FilterSelection} from "@/components/app/filter-tree-dialog";
import {TransactionType} from "@/app/(app)/orders/_types/order.schema";

export type OrderSort = "date-desc" | "date-asc" | "amount-desc" | "amount-asc";

export const SORT_OPTIONS: {value: OrderSort, label: string}[] = [
    {value: "date-desc", label: "Newest first"},
    {value: "date-asc", label: "Oldest first"},
    {value: "amount-desc", label: "Amount: high to low"},
    {value: "amount-asc", label: "Amount: low to high"},
];

export const TYPE_OPTIONS: {value: TransactionType, label: string}[] = [
    {value: "INCOME", label: "Income"},
    {value: "EXPENSE", label: "Expense"},
];

export interface OrderFilters {
    search: string;
    type: TransactionType[]; // empty = income and expense
    currency: string[];      // empty = all currencies
    sort: OrderSort;
}

export const DEFAULT_FILTERS: OrderFilters = {search: "", type: [], currency: [], sort: "date-desc"};

export interface OrderToolbarProps {
    filters: OrderFilters;
    onChange: (filters: OrderFilters) => void;
    currencies: string[];
    products: OrderProduct[]; // for the add dialog
    orderCount: number;       // every order on the farm, for "delete all"
}

export default function OrderToolbar({filters, onChange, currencies, products, orderCount}: OrderToolbarProps) {
    const update = (patch: Partial<OrderFilters>) => onChange({...filters, ...patch});
    const isFiltered = filters.type.length > 0 || filters.currency.length > 0;

    const filterGroups: FilterGroup[] = [
        {id: "type", label: "Type", options: TYPE_OPTIONS},
        {id: "currency", label: "Currency", options: currencies.map((currency) => ({value: currency, label: currency}))},
    ];
    const selection: FilterSelection = {type: filters.type, currency: filters.currency};

    const applySelection = (next: FilterSelection) => update({
        type: (next.type ?? []) as TransactionType[],
        currency: next.currency ?? [],
    });

    // One removable chip per active option, e.g. "Type: Income ×"
    const activeChips = filterGroups.flatMap((group) =>
        (selection[group.id] ?? []).map((value) => ({
            group,
            value,
            label: group.options.find((o) => o.value === value)?.label ?? value,
        }))
    );
    const removeChip = (groupId: string, value: string) =>
        applySelection({...selection, [groupId]: (selection[groupId] ?? []).filter((v) => v !== value)});

    return (
        <div className="flex flex-col gap-2">
            <div className="flex flex-col gap-3 sm:flex-row sm:flex-wrap sm:items-center">
                <div className="relative min-w-48 flex-1">
                    <Search className="pointer-events-none absolute left-2.5 top-1/2 size-4 -translate-y-1/2 text-muted-foreground" />
                    <Input
                        type="search"
                        placeholder="Search orders…"
                        aria-label="Search orders"
                        value={filters.search}
                        onChange={(e) => update({search: e.target.value})}
                        className="h-9 pl-8 rounded-xl"
                    />
                </div>

                <FilterTreeDialog groups={filterGroups} value={selection} onApply={applySelection} title="Filter orders" />

                <Select
                    items={SORT_OPTIONS}
                    value={filters.sort}
                    onValueChange={(sort) => {
                        if (sort) update({sort});
                    }}
                >
                    <SelectTrigger aria-label="Sort by" className="h-9 data-[size=default]:h-9 w-full sm:w-56 bg-background font-sans rounded-xl">
                        <ArrowUpDown className="text-muted-foreground" />
                        <SelectValue className="flex-1 text-left" />
                    </SelectTrigger>
                    <SelectContent>
                        {SORT_OPTIONS.map((option) => (
                            <SelectItem key={option.value} value={option.value}>{option.label}</SelectItem>
                        ))}
                    </SelectContent>
                </Select>

                {orderCount > 0 && <DeleteAllOrdersButton count={orderCount} />}
                <AddOrderDialog products={products} />
            </div>

            <div className="flex flex-wrap items-center gap-2">
                {activeChips.map((chip) => (
                    <span
                        key={chip.group.id + chip.value}
                        className="inline-flex items-center gap-1 rounded-full border border-border bg-muted py-0.5 pl-2.5 pr-1 text-sm font-sans"
                    >
                        <span className="text-muted-foreground">{chip.group.label}:</span>
                        {chip.label}
                        <button
                            type="button"
                            aria-label={`Remove ${chip.group.label} ${chip.label} filter`}
                            onClick={() => removeChip(chip.group.id, chip.value)}
                            className="flex size-5 items-center justify-center rounded-full text-muted-foreground outline-none hover:bg-background hover:text-foreground focus-visible:ring-3 focus-visible:ring-ring/50"
                        >
                            <X className="size-3.5" />
                        </button>
                    </span>
                ))}

                {isFiltered &&
                    <Button variant="ghost" size="sm" onClick={() => onChange({...DEFAULT_FILTERS, sort: filters.sort})}>
                        <X />
                        Clear filters
                    </Button>
                }
            </div>
        </div>
    );
}
