'use client';

import React, {useState} from 'react';
import {Plus, Trash2} from "lucide-react";
import {Button} from "@/components/ui/button";
import {Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle, DialogTrigger} from "@/components/ui/dialog";
import {Input} from "@/components/ui/input";
import {Label} from "@/components/ui/label";
import {Select, SelectContent, SelectItem, SelectTrigger, SelectValue} from "@/components/ui/select";
import ConfirmDialog from "@/components/app/confirm-dialog";
import {ProductSchema} from "@/app/(app)/products/_types/product.schema";
import {addProductPrice, addProductStock, removeProductPrice, removeProductStock} from "@/lib/api/hooks";
import {StockType} from "@/lib/api/resources";
import {useFarm} from "@/lib/auth/auth-provider";
import {CURRENCIES} from "@/lib/currencies";
import {todayIso} from "@/lib/dates";
import {errorMessage} from "@/app/(auth)/_components/auth-form";

// Adding to and deleting from a product's price and stock history. Entries are append-only on the
// backend: a wrong one is deleted and added again rather than edited.

const STOCK_TYPES: {value: StockType, label: string}[] = [
    {value: "INCOME", label: "Stock in"},
    {value: "EXPENSE", label: "Stock out"},
];

// Shared shell: a small "Add" button in a panel header that opens a form
function EntryDialog({label, children}: {label: string, children: (close: () => void) => React.ReactNode}) {
    const [open, setOpen] = useState(false);

    return (
        <Dialog open={open} onOpenChange={setOpen}>
            <DialogTrigger render={<Button variant="ghost" size="xs" className="cursor-pointer normal-case tracking-normal" />}>
                <Plus />
                {label}
            </DialogTrigger>
            <DialogContent>{children(() => setOpen(false))}</DialogContent>
        </Dialog>
    );
}

function useSubmit(action: () => Promise<void>, onDone: () => void) {
    const [pending, setPending] = useState(false);
    const [error, setError] = useState<string | null>(null);

    const submit = async (e: React.FormEvent) => {
        e.preventDefault();
        setPending(true);
        setError(null);
        try {
            await action();
            onDone();
        } catch (err) {
            setError(errorMessage(err));
            setPending(false);
        }
    };

    return {pending, error, submit};
}

function AddPriceForm({product, onDone}: {product: ProductSchema, onDone: () => void}) {
    const farm = useFarm()?.farm;
    // Starts from the current currency, so a new price is usually just the number
    const current = product.Prices?.[product.Prices.length - 1];
    const [price, setPrice] = useState("");
    const [currency, setCurrency] = useState(current?.Currency ?? "USD");
    const [day, setDay] = useState(todayIso());
    const canSubmit = Number(price) > 0 && day !== "";

    const {pending, error, submit} = useSubmit(async () => {
        if (farm) await addProductPrice(farm.farm_id, product.ID, Number(price), currency, day);
    }, onDone);

    return (
        <form onSubmit={submit} className="grid gap-6">
            <DialogHeader>
                <DialogTitle>Add price</DialogTitle>
                <DialogDescription>The newest price is the product&apos;s current price.</DialogDescription>
            </DialogHeader>

            <div className="grid gap-4 font-sans">
                <div className="grid grid-cols-[1fr_7rem] gap-4">
                    <div className="flex flex-col gap-2">
                        <Label htmlFor="price-amount">Price{product.Unit && ` per ${product.Unit}`}</Label>
                        <Input id="price-amount" type="number" min={0} step="any" required autoFocus placeholder="0.00" value={price} onChange={(e) => setPrice(e.target.value)} />
                    </div>
                    <div className="flex flex-col gap-2">
                        <Label htmlFor="price-currency">Currency</Label>
                        <Select items={CURRENCIES} value={currency} onValueChange={(next) => next && setCurrency(next)}>
                            <SelectTrigger id="price-currency" className="w-full cursor-pointer"><SelectValue /></SelectTrigger>
                            <SelectContent>
                                {CURRENCIES.map((o) => <SelectItem key={o.value} value={o.value}>{o.label}</SelectItem>)}
                            </SelectContent>
                        </Select>
                    </div>
                </div>
                <div className="flex flex-col gap-2">
                    <Label htmlFor="price-date">Date</Label>
                    <Input id="price-date" type="date" required max={todayIso()} value={day} onChange={(e) => setDay(e.target.value)} />
                </div>
            </div>

            {error && <p role="alert" className="text-sm text-danger-strong font-sans">{error}</p>}

            <DialogFooter showCloseButton>
                <Button type="submit" className="cursor-pointer" disabled={!canSubmit || pending}>{pending ? "Adding…" : "Add price"}</Button>
            </DialogFooter>
        </form>
    );
}

function AddStockForm({product, onDone}: {product: ProductSchema, onDone: () => void}) {
    const farm = useFarm()?.farm;
    const [type, setType] = useState<StockType>("INCOME");
    const [stock, setStock] = useState("");
    const [day, setDay] = useState(todayIso());
    const canSubmit = Number(stock) > 0 && day !== "";

    const {pending, error, submit} = useSubmit(async () => {
        if (farm) await addProductStock(farm.farm_id, product.ID, Number(stock), type, day);
    }, onDone);

    return (
        <form onSubmit={submit} className="grid gap-6">
            <DialogHeader>
                <DialogTitle>Record stock movement</DialogTitle>
                <DialogDescription>Stock coming in or going out of the farm.</DialogDescription>
            </DialogHeader>

            <div className="grid gap-4 font-sans">
                <div className="grid grid-cols-2 gap-4">
                    <div className="flex flex-col gap-2">
                        <Label htmlFor="stock-type">Movement</Label>
                        <Select items={STOCK_TYPES} value={type} onValueChange={(next) => next && setType(next)}>
                            <SelectTrigger id="stock-type" className="w-full cursor-pointer"><SelectValue /></SelectTrigger>
                            <SelectContent>
                                {STOCK_TYPES.map((o) => <SelectItem key={o.value} value={o.value}>{o.label}</SelectItem>)}
                            </SelectContent>
                        </Select>
                    </div>
                    <div className="flex flex-col gap-2">
                        <Label htmlFor="stock-amount">Amount{product.Unit && ` (${product.Unit})`}</Label>
                        <Input id="stock-amount" type="number" min={0} step="any" required autoFocus placeholder="0" value={stock} onChange={(e) => setStock(e.target.value)} />
                    </div>
                </div>
                <div className="flex flex-col gap-2">
                    <Label htmlFor="stock-date">Date</Label>
                    <Input id="stock-date" type="date" required max={todayIso()} value={day} onChange={(e) => setDay(e.target.value)} />
                </div>
            </div>

            {error && <p role="alert" className="text-sm text-danger-strong font-sans">{error}</p>}

            <DialogFooter showCloseButton>
                <Button type="submit" className="cursor-pointer" disabled={!canSubmit || pending}>{pending ? "Adding…" : "Record"}</Button>
            </DialogFooter>
        </form>
    );
}

export function AddPriceDialog({product}: {product: ProductSchema}) {
    return <EntryDialog label="Add">{(close) => <AddPriceForm product={product} onDone={close} />}</EntryDialog>;
}

export function AddStockDialog({product}: {product: ProductSchema}) {
    return <EntryDialog label="Add">{(close) => <AddStockForm product={product} onDone={close} />}</EntryDialog>;
}

// A trash button at the end of a history row
export function DeleteEntryButton({kind, product, entryID, summary}: {
    kind: "price" | "stock",
    product: ProductSchema,
    entryID: string,
    summary: string, // e.g. "$4.50 on Mar 3, 2026", for the confirmation
}) {
    const farm = useFarm()?.farm;
    const what = kind === "price" ? "price" : "stock movement";

    return (
        <ConfirmDialog
            title={`Delete this ${what}?`}
            description={`${summary} is removed from ${product.Name}'s history.`}
            confirmLabel="Delete"
            onConfirm={async () => {
                if (!farm) return;
                if (kind === "price") await removeProductPrice(farm.farm_id, product.ID, entryID);
                else await removeProductStock(farm.farm_id, product.ID, entryID);
            }}
            trigger={<Button variant="ghost" size="icon-xs" className="cursor-pointer text-muted-foreground hover:text-danger-strong" />}
        >
            <Trash2 />
            <span className="sr-only">Delete {what}</span>
        </ConfirmDialog>
    );
}
