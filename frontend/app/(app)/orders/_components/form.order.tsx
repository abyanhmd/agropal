'use client';

import React, {useState} from 'react';
import {DialogDescription, DialogFooter, DialogHeader, DialogTitle} from "@/components/ui/dialog";
import {Button} from "@/components/ui/button";
import {Input} from "@/components/ui/input";
import {Label} from "@/components/ui/label";
import {Select, SelectContent, SelectItem, SelectTrigger, SelectValue} from "@/components/ui/select";
import {TransactionType} from "@/app/(app)/orders/_types/order.schema";
import type {OrderProduct} from "@/app/(app)/orders/_components/table.order";
import {NewOrder} from "@/lib/api/hooks";
import {CURRENCIES} from "@/lib/currencies";
import {errorMessage} from "@/app/(auth)/_components/auth-form";

// The add and edit order dialogs share this form. Editing can't move an order to another product
// (the backend keys it by product), so the product picker is only shown when adding.

const TYPES: {value: TransactionType, label: string}[] = [
    {value: "INCOME", label: "Income"},
    {value: "EXPENSE", label: "Expense"},
];

interface Draft {
    productID: string;
    type: TransactionType;
    amount: string;
    currency: string;
    date: string;
    description: string;
}

export interface OrderFormProps {
    title: string;
    description: string;
    submitLabel: string;
    pendingLabel: string;
    initial: NewOrder;
    // Omitted when editing: the product is fixed
    products?: OrderProduct[];
    onSubmit: (order: NewOrder) => Promise<void>;
}

// Form state lives in its own component, so it starts fresh each time its dialog opens
export default function OrderForm({title, description, submitLabel, pendingLabel, initial, products, onSubmit}: OrderFormProps) {
    const [draft, setDraft] = useState<Draft>({...initial, amount: initial.amount > 0 ? String(initial.amount) : ""});
    const set = (patch: Partial<Draft>) => setDraft({...draft, ...patch});

    const productItems = (products ?? []).map((p) => ({value: p.ID, label: p.Name}));
    const canSubmit = draft.productID !== "" && Number(draft.amount) > 0 && draft.date !== "";

    const [pending, setPending] = useState(false);
    const [error, setError] = useState<string | null>(null);

    const handleSubmit = async (e: React.FormEvent) => {
        e.preventDefault();
        if (!canSubmit) return;
        setPending(true);
        setError(null);
        try {
            await onSubmit({
                productID: draft.productID,
                type: draft.type,
                amount: Number(draft.amount),
                currency: draft.currency,
                date: draft.date,
                description: draft.description.trim(),
            });
        } catch (err) {
            setError(errorMessage(err));
            setPending(false);
        }
    };

    return (
        <form onSubmit={handleSubmit} className="grid gap-6">
            <DialogHeader>
                <DialogTitle>{title}</DialogTitle>
                <DialogDescription>{description}</DialogDescription>
            </DialogHeader>

            <div className="grid gap-4 font-sans">
                {products &&
                    <div className="flex flex-col gap-2">
                        <Label htmlFor="order-product">Product</Label>
                        <Select items={productItems} value={draft.productID || null} onValueChange={(productID) => productID && set({productID})}>
                            <SelectTrigger id="order-product" className="w-full cursor-pointer">
                                <SelectValue placeholder="Choose a product" />
                            </SelectTrigger>
                            <SelectContent>
                                {productItems.map((o) => <SelectItem key={o.value} value={o.value}>{o.label}</SelectItem>)}
                            </SelectContent>
                        </Select>
                    </div>
                }

                <div className="grid grid-cols-2 gap-4">
                    <div className="flex flex-col gap-2">
                        <Label htmlFor="order-type">Type</Label>
                        <Select items={TYPES} value={draft.type} onValueChange={(type) => type && set({type})}>
                            <SelectTrigger id="order-type" className="w-full cursor-pointer"><SelectValue /></SelectTrigger>
                            <SelectContent>
                                {TYPES.map((o) => <SelectItem key={o.value} value={o.value}>{o.label}</SelectItem>)}
                            </SelectContent>
                        </Select>
                    </div>
                    <div className="flex flex-col gap-2">
                        <Label htmlFor="order-date">Date</Label>
                        <Input id="order-date" type="date" required value={draft.date} onChange={(e) => set({date: e.target.value})} />
                    </div>
                </div>

                <div className="grid grid-cols-[1fr_7rem] gap-4">
                    <div className="flex flex-col gap-2">
                        <Label htmlFor="order-amount">Amount</Label>
                        <Input id="order-amount" type="number" min={0} step="any" required placeholder="0.00" value={draft.amount} onChange={(e) => set({amount: e.target.value})} />
                    </div>
                    <div className="flex flex-col gap-2">
                        <Label htmlFor="order-currency">Currency</Label>
                        <Select items={CURRENCIES} value={draft.currency} onValueChange={(currency) => currency && set({currency})}>
                            <SelectTrigger id="order-currency" className="w-full cursor-pointer"><SelectValue /></SelectTrigger>
                            <SelectContent>
                                {CURRENCIES.map((o) => <SelectItem key={o.value} value={o.value}>{o.label}</SelectItem>)}
                            </SelectContent>
                        </Select>
                    </div>
                </div>

                <div className="flex flex-col gap-2">
                    <Label htmlFor="order-description">Description</Label>
                    <Input id="order-description" placeholder="Wholesale order" value={draft.description} onChange={(e) => set({description: e.target.value})} />
                </div>
            </div>

            {error && <p role="alert" className="text-sm text-danger-strong font-sans">{error}</p>}

            <DialogFooter showCloseButton>
                <Button type="submit" className="cursor-pointer" disabled={!canSubmit || pending}>{pending ? pendingLabel : submitLabel}</Button>
            </DialogFooter>
        </form>
    );
}
