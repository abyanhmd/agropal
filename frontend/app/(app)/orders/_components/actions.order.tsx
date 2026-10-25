'use client';

import React, {useState} from 'react';
import {useRouter} from "next/navigation";
import {Pencil, Trash2} from "lucide-react";
import {Button} from "@/components/ui/button";
import {Dialog, DialogContent, DialogTrigger} from "@/components/ui/dialog";
import ConfirmDialog from "@/components/app/confirm-dialog";
import OrderForm from "@/app/(app)/orders/_components/form.order";
import {TransactionSchema} from "@/app/(app)/orders/_types/order.schema";
import {editOrder, removeOrders} from "@/lib/api/hooks";
import {useFarm} from "@/lib/auth/auth-provider";
import {toDateInput} from "@/lib/dates";
import {useAppHref} from "@/lib/demo/base";

// Edit and delete for the order page

function EditOrderDialog({order}: {order: TransactionSchema}) {
    const farm = useFarm()?.farm;
    const [open, setOpen] = useState(false);

    return (
        <Dialog open={open} onOpenChange={setOpen}>
            <DialogTrigger render={<Button variant="outline" className="cursor-pointer" />}>
                <Pencil />
                Edit
            </DialogTrigger>
            <DialogContent>
                <OrderForm
                    title="Edit order"
                    description="The product can't be changed; delete the order and add it again instead."
                    submitLabel="Save changes"
                    pendingLabel="Saving…"
                    initial={{
                        productID: order.ProductID,
                        type: order.Type,
                        amount: order.Amount,
                        currency: order.Currency,
                        date: toDateInput(order.Date),
                        description: order.Description,
                    }}
                    onSubmit={async (details) => {
                        if (!farm) return;
                        await editOrder(farm.farm_id, order, details);
                        setOpen(false);
                    }}
                />
            </DialogContent>
        </Dialog>
    );
}

function DeleteOrderButton({order}: {order: TransactionSchema}) {
    const appHref = useAppHref();
    const router = useRouter();

    return (
        <ConfirmDialog
            title="Delete this order?"
            description={`"${order.Description || "Untitled order"}" is deleted. This can't be undone.`}
            confirmLabel="Delete order"
            onConfirm={async () => {
                await removeOrders([order.ID], {leaving: true});
                router.replace(appHref("/orders"));
            }}
            trigger={<Button variant="outline" className="cursor-pointer text-danger-strong hover:text-danger-strong" />}
        >
            <Trash2 />
            Delete
        </ConfirmDialog>
    );
}

export default function OrderActions({order}: {order: TransactionSchema}) {
    return (
        <div className="flex shrink-0 gap-2">
            <EditOrderDialog order={order} />
            <DeleteOrderButton order={order} />
        </div>
    );
}
