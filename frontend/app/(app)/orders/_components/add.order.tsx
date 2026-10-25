'use client';

import React, {useState} from 'react';
import {useRouter} from "next/navigation";
import {Plus} from "lucide-react";
import {Button} from "@/components/ui/button";
import {Dialog, DialogContent, DialogTrigger} from "@/components/ui/dialog";
import OrderForm from "@/app/(app)/orders/_components/form.order";
import type {OrderProduct} from "@/app/(app)/orders/_components/table.order";
import {addOrder} from "@/lib/api/hooks";
import {useFarm} from "@/lib/auth/auth-provider";
import {todayIso} from "@/lib/dates";
import {useAppHref} from "@/lib/demo/base";

export default function AddOrderDialog({products}: {products: OrderProduct[]}) {
    const appHref = useAppHref();
    const router = useRouter();
    const farm = useFarm()?.farm;
    const [open, setOpen] = useState(false);

    return (
        <Dialog open={open} onOpenChange={setOpen}>
            <DialogTrigger render={<Button className="h-9 w-full cursor-pointer rounded-xl sm:w-auto" />}>
                <Plus />
                Add order
            </DialogTrigger>
            <DialogContent>
                <OrderForm
                    title="Add order"
                    description="Recorded against the product you pick."
                    submitLabel="Add order"
                    pendingLabel="Adding…"
                    initial={{productID: "", type: "INCOME", amount: 0, currency: "USD", date: todayIso(), description: ""}}
                    products={products}
                    onSubmit={async (order) => {
                        if (!farm) return;
                        const id = await addOrder(farm.farm_id, order);
                        setOpen(false);
                        router.push(appHref(`/orders/${id}`));
                    }}
                />
            </DialogContent>
        </Dialog>
    );
}
