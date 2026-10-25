'use client';

import React from 'react';
import {Trash2, X} from "lucide-react";
import {Button} from "@/components/ui/button";
import ConfirmDialog, {DANGER_BUTTON} from "@/components/app/confirm-dialog";
import {removeOrders} from "@/lib/api/hooks";

const plural = (n: number) => (n === 1 ? "1 order" : `${n} orders`);

export interface OrderBulkBarProps {
    selectedIDs: string[];
    matchingIDs: string[]; // every order the current filters show, across pages
    onSelect: (ids: string[]) => void;
    onClear: () => void;
}

// Shown above the table while orders are ticked
export default function OrderBulkBar({selectedIDs, matchingIDs, onSelect, onClear}: OrderBulkBarProps) {
    const count = selectedIDs.length;
    const canSelectMore = count < matchingIDs.length;

    return (
        <div role="region" aria-label="Selected orders" className="flex flex-wrap items-center gap-x-3 gap-y-2 rounded-xl border border-border bg-muted px-4 py-2 font-sans text-sm">
            <span className="font-medium tabular-nums">{plural(count)} selected</span>
            {canSelectMore &&
                <Button variant="link" size="sm" className="h-auto cursor-pointer px-0" onClick={() => onSelect(matchingIDs)}>
                    Select all {matchingIDs.length} matching
                </Button>
            }
            <div className="ml-auto flex gap-2">
                <Button variant="ghost" size="sm" className="cursor-pointer" onClick={onClear}>
                    <X />
                    Clear
                </Button>
                <ConfirmDialog
                    title={`Delete ${plural(count)}?`}
                    description="They're deleted for good. This can't be undone."
                    confirmLabel={`Delete ${plural(count)}`}
                    onConfirm={async () => {
                        await removeOrders(selectedIDs);
                        onClear();
                    }}
                    trigger={<Button size="sm" className={DANGER_BUTTON} />}
                >
                    <Trash2 />
                    Delete
                </ConfirmDialog>
            </div>
        </div>
    );
}
