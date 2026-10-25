'use client';

import React from 'react';
import {Trash2} from "lucide-react";
import {Button} from "@/components/ui/button";
import ConfirmDialog from "@/components/app/confirm-dialog";
import {removeAllOrders} from "@/lib/api/hooks";

// Wipes every order on the farm, regardless of the filters on screen
export default function DeleteAllOrdersButton({count}: {count: number}) {
    return (
        <ConfirmDialog
            title="Delete all orders?"
            description={`All ${count} ${count === 1 ? "order" : "orders"} are deleted for good, not just the ones matching your filters. Products, prices and stock stay.`}
            confirmLabel="Delete all orders"
            requireText="DELETE"
            onConfirm={removeAllOrders}
            trigger={<Button variant="outline" className="h-9 w-full cursor-pointer rounded-xl text-danger-strong sm:w-auto" />}
        >
            <Trash2 />
            Delete all
        </ConfirmDialog>
    );
}
