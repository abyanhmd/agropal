import React from 'react';
import {ArrowUpDown, Search, X} from "lucide-react";
import {Input} from "@/components/ui/input";
import {Button} from "@/components/ui/button";
import {Select, SelectContent, SelectItem, SelectTrigger, SelectValue} from "@/components/ui/select";
import FilterTreeDialog, {FilterGroup, FilterSelection} from "@/components/app/filter-tree-dialog";
import AddPostDialog from "@/app/(app)/discussions/_components/add.discussion";
import {PostAuthor} from "@/app/(app)/discussions/_types/post.schema";

export type PostSort = "newest" | "oldest" | "most-liked" | "most-commented" | "most-viewed";

export const SORT_OPTIONS: {value: PostSort, label: string}[] = [
    {value: "newest", label: "Newest first"},
    {value: "oldest", label: "Oldest first"},
    {value: "most-liked", label: "Most liked"},
    {value: "most-commented", label: "Most discussed"},
    {value: "most-viewed", label: "Most viewed"},
];

export interface PostFilters {
    search: string;
    author: string[]; // user ids; empty = everyone
    sort: PostSort;
}

export const DEFAULT_FILTERS: PostFilters = {search: "", author: [], sort: "newest"};

export interface DiscussionToolbarProps {
    filters: PostFilters;
    onChange: (filters: PostFilters) => void;
    authors: PostAuthor[];
}

export default function DiscussionToolbar({filters, onChange, authors}: DiscussionToolbarProps) {
    const update = (patch: Partial<PostFilters>) => onChange({...filters, ...patch});
    const isFiltered = filters.author.length > 0;

    const filterGroups: FilterGroup[] = [
        {id: "author", label: "Author", options: authors.map((a) => ({value: a.ID, label: a.Name}))},
    ];
    const selection: FilterSelection = {author: filters.author};

    const applySelection = (next: FilterSelection) => update({author: next.author ?? []});

    // One removable chip per active option, e.g. "Author: Grace Liu ×"
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
                        placeholder="Search discussions…"
                        aria-label="Search discussions"
                        value={filters.search}
                        onChange={(e) => update({search: e.target.value})}
                        className="h-9 pl-8 rounded-xl"
                    />
                </div>

                <FilterTreeDialog groups={filterGroups} value={selection} onApply={applySelection} title="Filter discussions" />

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

                <AddPostDialog />
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
