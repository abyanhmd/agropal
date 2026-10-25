'use client';

import React, {useState} from 'react';
import {ChevronRight, SlidersHorizontal} from "lucide-react";
import {Button} from "@/components/ui/button";
import {Badge} from "@/components/ui/badge";
import {Checkbox} from "@/components/ui/checkbox";
import {Collapsible, CollapsibleContent, CollapsibleTrigger} from "@/components/ui/collapsible";
import {
    Dialog,
    DialogClose,
    DialogContent,
    DialogDescription,
    DialogFooter,
    DialogHeader,
    DialogTitle,
    DialogTrigger
} from "@/components/ui/dialog";

export interface FilterOption {
    value: string;
    label: string;
}

export interface FilterGroup {
    id: string;
    label: string;
    options: FilterOption[];
}

// Selected option values per group id; an empty/missing list means "no filter" for that group
export type FilterSelection = Record<string, string[]>;

export interface FilterTreeDialogProps {
    groups: FilterGroup[];
    value: FilterSelection;
    onApply: (value: FilterSelection) => void;
    title?: string;
}

export function countSelected(selection: FilterSelection): number {
    return Object.values(selection).reduce((total, values) => total + values.length, 0);
}

function FilterTreeGroup({group, selected, onChange}: {
    group: FilterGroup,
    selected: string[],
    onChange: (values: string[]) => void,
}) {
    const allSelected = group.options.length > 0 && selected.length === group.options.length;
    const someSelected = selected.length > 0 && !allSelected;

    const toggleOption = (value: string, checked: boolean) =>
        onChange(checked ? [...selected, value] : selected.filter((v) => v !== value));

    return (
        <Collapsible defaultOpen>
            <div className="flex items-center gap-2 rounded-md px-1 py-1.5 hover:bg-muted">
                <CollapsibleTrigger
                    aria-label={`Toggle ${group.label}`}
                    className="group/trigger flex size-6 items-center justify-center rounded-sm text-muted-foreground outline-none hover:text-foreground focus-visible:ring-3 focus-visible:ring-ring/50"
                >
                    <ChevronRight className="size-4 transition-transform group-data-panel-open/trigger:rotate-90" />
                </CollapsibleTrigger>
                <label className="flex flex-1 cursor-pointer items-center gap-2.5">
                    <Checkbox
                        checked={allSelected}
                        indeterminate={someSelected}
                        onCheckedChange={(checked) => onChange(checked ? group.options.map((o) => o.value) : [])}
                    />
                    <span className="font-medium">{group.label}</span>
                </label>
                {selected.length > 0 &&
                    <span className="text-xs text-muted-foreground tabular-nums">
                        {selected.length}/{group.options.length}
                    </span>
                }
            </div>

            <CollapsibleContent>
                {/* Tree guide line on the left links children to their group */}
                <ul className="ml-4 border-l border-border pl-4">
                    {group.options.map((option) => (
                        <li key={option.value}>
                            <label className="flex cursor-pointer items-center gap-2.5 rounded-md px-2 py-1.5 hover:bg-muted">
                                <Checkbox
                                    checked={selected.includes(option.value)}
                                    onCheckedChange={(checked) => toggleOption(option.value, checked)}
                                />
                                <span>{option.label}</span>
                            </label>
                        </li>
                    ))}
                    {group.options.length === 0 &&
                        <li className="px-2 py-1.5 text-muted-foreground">No options</li>
                    }
                </ul>
            </CollapsibleContent>
        </Collapsible>
    );
}

export default function FilterTreeDialog({groups, value, onApply, title = "Filters"}: FilterTreeDialogProps) {
    const [open, setOpen] = useState(false);
    // Edits happen on a draft and only apply when the user presses Apply
    const [draft, setDraft] = useState<FilterSelection>(value);
    const activeCount = countSelected(value);

    const handleOpenChange = (next: boolean) => {
        if (next) setDraft(value);
        setOpen(next);
    };

    const apply = () => {
        onApply(draft);
        setOpen(false);
    };

    return (
        <Dialog open={open} onOpenChange={handleOpenChange}>
            <DialogTrigger render={<Button variant="outline" className="h-9 rounded-xl bg-background" />}>
                <SlidersHorizontal />
                Filter
                {activeCount > 0 && <Badge className="ml-0.5 h-5 min-w-5 px-1.5 tabular-nums">{activeCount}</Badge>}
            </DialogTrigger>

            <DialogContent className="font-sans">
                <DialogHeader>
                    <DialogTitle className="font-serif text-lg">{title}</DialogTitle>
                    <DialogDescription>Pick any options. Leaving a group empty shows everything.</DialogDescription>
                </DialogHeader>

                <div className="-mx-2 flex max-h-[60vh] flex-col gap-1 overflow-y-auto px-2">
                    {groups.map((group) => (
                        <FilterTreeGroup
                            key={group.id}
                            group={group}
                            selected={draft[group.id] ?? []}
                            onChange={(values) => setDraft({...draft, [group.id]: values})}
                        />
                    ))}
                </div>

                <DialogFooter className="sm:justify-between">
                    <Button variant="ghost" onClick={() => setDraft({})} disabled={countSelected(draft) === 0}>
                        Reset
                    </Button>
                    <div className="flex flex-col-reverse gap-2 sm:flex-row">
                        <DialogClose render={<Button variant="outline" />}>Cancel</DialogClose>
                        <Button onClick={apply}>Apply filters</Button>
                    </div>
                </DialogFooter>
            </DialogContent>
        </Dialog>
    );
}
