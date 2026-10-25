import React from 'react';
import {ArrowUpDown, LayoutGrid, List, Search, X} from "lucide-react";
import {Input} from "@/components/ui/input";
import {Button} from "@/components/ui/button";
import {Select, SelectContent, SelectItem, SelectTrigger, SelectValue} from "@/components/ui/select";
import AddProductDialog from "@/app/(app)/products/_components/add.product";
import FilterTreeDialog, {FilterGroup, FilterSelection} from "@/components/app/filter-tree-dialog";

export type StockTrend = "INCOME" | "EXPENSE";

export type ProductView = "list" | "grid";

export type ProductSort = "name-asc" | "name-desc" | "price-asc" | "price-desc" | "stock-desc" | "stock-asc";

export const SORT_OPTIONS: {value: ProductSort, label: string}[] = [
    {value: "name-asc", label: "Name (A–Z)"},
    {value: "name-desc", label: "Name (Z–A)"},
    {value: "price-asc", label: "Price: low to high"},
    {value: "price-desc", label: "Price: high to low"},
    {value: "stock-desc", label: "Latest change: biggest"},
    {value: "stock-asc", label: "Latest change: smallest"},
];

const TREND_OPTIONS: {value: StockTrend, label: string}[] = [
    {value: "INCOME", label: "Income"},
    {value: "EXPENSE", label: "Expense"},
];

export interface ProductFilters {
    search: string;
    unit: string[];      // empty = all units
    trend: StockTrend[]; // empty = any latest change
    sort: ProductSort;
}

export const DEFAULT_FILTERS: ProductFilters = {search: "", unit: [], trend: [], sort: "name-asc"};

export interface ProductToolbarProps {
    filters: ProductFilters;
    onChange: (filters: ProductFilters) => void;
    units: string[];
    view: ProductView;
    onViewChange: (view: ProductView) => void;
}

const VIEW_OPTIONS: {value: ProductView, label: string, icon: typeof List}[] = [
    {value: "list", label: "List view", icon: List},
    {value: "grid", label: "Grid view", icon: LayoutGrid},
];

export default function ProductToolbar({filters, onChange, units, view, onViewChange}: ProductToolbarProps) {
    const update = (patch: Partial<ProductFilters>) => onChange({...filters, ...patch});
    const isFiltered = filters.unit.length > 0 || filters.trend.length > 0;

    const filterGroups: FilterGroup[] = [
        {id: "unit", label: "Unit", options: units.map((unit) => ({value: unit, label: unit}))},
        {id: "trend", label: "Latest change", options: TREND_OPTIONS},
    ];
    const selection: FilterSelection = {unit: filters.unit, trend: filters.trend};

    const applySelection = (next: FilterSelection) => update({
        unit: next.unit ?? [],
        trend: (next.trend ?? []) as StockTrend[],
    });

    // One removable chip per active option, e.g. "Unit: kg ×"
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
                        placeholder="Search products…"
                        aria-label="Search products"
                        value={filters.search}
                        onChange={(e) => update({search: e.target.value})}
                        className="h-9 pl-8 rounded-xl"
                    />
                </div>

                <FilterTreeDialog groups={filterGroups} value={selection} onApply={applySelection} title="Filter products" />

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

                <div role="group" aria-label="View" className="inline-flex h-9 shrink-0 items-center gap-0.5 self-start rounded-xl border border-border bg-muted p-0.5 sm:self-auto">
                    {VIEW_OPTIONS.map(({value, label, icon: Icon}) => (
                        <Button
                            key={value}
                            variant="ghost"
                            size="icon-sm"
                            aria-label={label}
                            aria-pressed={view === value}
                            onClick={() => onViewChange(value)}
                            className="size-[30px] cursor-pointer rounded-[10px] text-muted-foreground hover:bg-transparent hover:text-foreground aria-pressed:bg-background aria-pressed:text-foreground aria-pressed:shadow-xs"
                        >
                            <Icon />
                        </Button>
                    ))}
                </div>

                <AddProductDialog />
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
