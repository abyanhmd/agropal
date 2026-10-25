'use client';

import React, {useState} from 'react';
import {useRouter} from "next/navigation";
import {Plus} from "lucide-react";
import {Button} from "@/components/ui/button";
import {Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle, DialogTrigger} from "@/components/ui/dialog";
import {Input} from "@/components/ui/input";
import {Label} from "@/components/ui/label";
import {Select, SelectContent, SelectItem, SelectTrigger, SelectValue} from "@/components/ui/select";
import {UNITS} from "@/app/(app)/products/_lib/units";
import {CURRENCIES} from "@/lib/currencies";
import {addProduct} from "@/lib/api/hooks";
import {useFarm} from "@/lib/auth/auth-provider";
import {errorMessage} from "@/app/(auth)/_components/auth-form";
import {useAppHref} from "@/lib/demo/base";

interface Draft {
    name: string;
    description: string;
    image: string;
    unit: string;
    price: string;
    currency: string;
    stock: string;
}

const EMPTY: Draft = {name: "", description: "", image: "", unit: "kg", price: "", currency: "USD", stock: ""};

// Form state lives in its own component, so it starts fresh each time the dialog opens
function AddProductForm({onDone}: {onDone: (id?: string) => void}) {
    const [draft, setDraft] = useState<Draft>(EMPTY);
    const set = (patch: Partial<Draft>) => setDraft({...draft, ...patch});

    const farm = useFarm()?.farm;
    const [pending, setPending] = useState(false);
    const [error, setError] = useState<string | null>(null);

    const handleSubmit = async (e: React.FormEvent) => {
        e.preventDefault();
        if (!farm) return;
        setPending(true);
        setError(null);
        try {
            const id = await addProduct(farm.farm_id, {
                name: draft.name.trim(),
                description: draft.description.trim(),
                image: draft.image.trim(),
                unit: draft.unit,
                price: Number(draft.price) || 0,
                currency: draft.currency,
                stock: Number(draft.stock) || 0,
            });
            onDone(id);
        } catch (err) {
            setError(errorMessage(err));
            setPending(false);
        }
    };

    return (
        <form onSubmit={handleSubmit} className="grid gap-6">
            <DialogHeader>
                <DialogTitle>Add product</DialogTitle>
                <DialogDescription>Added to your farm.</DialogDescription>
            </DialogHeader>

            <div className="grid gap-4 font-sans">
                <div className="flex flex-col gap-2">
                    <Label htmlFor="product-name">Name</Label>
                    <Input id="product-name" required autoFocus placeholder="Organic tomatoes" value={draft.name} onChange={(e) => set({name: e.target.value})} />
                </div>

                <div className="flex flex-col gap-2">
                    <Label htmlFor="product-description">Description</Label>
                    <Input id="product-description" placeholder="Optional" value={draft.description} onChange={(e) => set({description: e.target.value})} />
                </div>

                <div className="flex flex-col gap-2">
                    <Label htmlFor="product-image">Image URL</Label>
                    <Input id="product-image" type="url" placeholder="Optional, https://…" value={draft.image} onChange={(e) => set({image: e.target.value})} />
                </div>

                <div className="grid grid-cols-2 gap-4">
                    <div className="flex flex-col gap-2">
                        <Label htmlFor="product-unit">Sold by</Label>
                        <Select items={UNITS} value={draft.unit} onValueChange={(unit) => unit && set({unit})}>
                            <SelectTrigger id="product-unit" className="w-full cursor-pointer"><SelectValue /></SelectTrigger>
                            <SelectContent>
                                {UNITS.map((o) => <SelectItem key={o.value} value={o.value}>{o.label}</SelectItem>)}
                            </SelectContent>
                        </Select>
                    </div>
                    <div className="flex flex-col gap-2">
                        <Label htmlFor="product-stock">Starting stock</Label>
                        <Input id="product-stock" type="number" min={0} step="any" placeholder="0" value={draft.stock} onChange={(e) => set({stock: e.target.value})} />
                    </div>
                </div>

                <div className="grid grid-cols-[1fr_7rem] gap-4">
                    <div className="flex flex-col gap-2">
                        <Label htmlFor="product-price">Price</Label>
                        <Input id="product-price" type="number" min={0} step="any" required placeholder="0.00" value={draft.price} onChange={(e) => set({price: e.target.value})} />
                    </div>
                    <div className="flex flex-col gap-2">
                        <Label htmlFor="product-currency">Currency</Label>
                        <Select items={CURRENCIES} value={draft.currency} onValueChange={(currency) => currency && set({currency})}>
                            <SelectTrigger id="product-currency" className="w-full cursor-pointer"><SelectValue /></SelectTrigger>
                            <SelectContent>
                                {CURRENCIES.map((o) => <SelectItem key={o.value} value={o.value}>{o.label}</SelectItem>)}
                            </SelectContent>
                        </Select>
                    </div>
                </div>
            </div>

            {error && <p role="alert" className="text-sm text-danger-strong font-sans">{error}</p>}

            <DialogFooter showCloseButton>
                <Button type="submit" className="cursor-pointer" disabled={!draft.name.trim() || pending}>{pending ? "Adding…" : "Add product"}</Button>
            </DialogFooter>
        </form>
    );
}

export default function AddProductDialog() {
    const appHref = useAppHref();
    const router = useRouter();
    const [open, setOpen] = useState(false);

    return (
        <Dialog open={open} onOpenChange={setOpen}>
            <DialogTrigger render={<Button className="h-9 w-full cursor-pointer rounded-xl sm:w-auto" />}>
                <Plus />
                Add product
            </DialogTrigger>
            <DialogContent>
                <AddProductForm
                    onDone={(id) => {
                        setOpen(false);
                        if (id) router.push(appHref(`/products/${id}`));
                    }}
                />
            </DialogContent>
        </Dialog>
    );
}
