'use client';

import React, {useState} from 'react';
import {useRouter} from "next/navigation";
import {Pencil, Trash2} from "lucide-react";
import {Button} from "@/components/ui/button";
import {Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle, DialogTrigger} from "@/components/ui/dialog";
import {Input} from "@/components/ui/input";
import {Label} from "@/components/ui/label";
import ConfirmDialog from "@/components/app/confirm-dialog";
import {ProductSchema} from "@/app/(app)/products/_types/product.schema";
import {editProduct, ProductDetails, removeProduct} from "@/lib/api/hooks";
import {useFarm} from "@/lib/auth/auth-provider";
import {errorMessage} from "@/app/(auth)/_components/auth-form";
import {useAppHref} from "@/lib/demo/base";

// Edit and delete for the product page. The unit has its own picker on the image.

// Form state lives in its own component, so it starts fresh each time the dialog opens
function EditProductForm({product, onDone}: {product: ProductSchema, onDone: () => void}) {
    const saved: ProductDetails = {name: product.Name, description: product.Description, image: product.Image};
    const [draft, setDraft] = useState<ProductDetails>(saved);
    const set = (patch: Partial<ProductDetails>) => setDraft({...draft, ...patch});

    const farm = useFarm()?.farm;
    const [pending, setPending] = useState(false);
    const [error, setError] = useState<string | null>(null);

    const trimmed: ProductDetails = {name: draft.name.trim(), description: draft.description.trim(), image: draft.image.trim()};
    const isDirty = trimmed.name !== saved.name || trimmed.description !== saved.description || trimmed.image !== saved.image;

    const handleSubmit = async (e: React.FormEvent) => {
        e.preventDefault();
        if (!farm) return;
        setPending(true);
        setError(null);
        try {
            await editProduct(farm.farm_id, product.ID, trimmed);
            onDone();
        } catch (err) {
            setError(errorMessage(err));
            setPending(false);
        }
    };

    return (
        <form onSubmit={handleSubmit} className="grid gap-6">
            <DialogHeader>
                <DialogTitle>Edit product</DialogTitle>
                <DialogDescription>Prices and stock are kept as history; add new entries on the product page.</DialogDescription>
            </DialogHeader>

            <div className="grid gap-4 font-sans">
                <div className="flex flex-col gap-2">
                    <Label htmlFor="edit-product-name">Name</Label>
                    <Input id="edit-product-name" required autoFocus value={draft.name} onChange={(e) => set({name: e.target.value})} />
                </div>
                <div className="flex flex-col gap-2">
                    <Label htmlFor="edit-product-description">Description</Label>
                    <Input id="edit-product-description" placeholder="Optional" value={draft.description} onChange={(e) => set({description: e.target.value})} />
                </div>
                <div className="flex flex-col gap-2">
                    <Label htmlFor="edit-product-image">Image URL</Label>
                    <Input id="edit-product-image" type="url" placeholder="Optional, https://…" value={draft.image} onChange={(e) => set({image: e.target.value})} />
                </div>
            </div>

            {error && <p role="alert" className="text-sm text-danger-strong font-sans">{error}</p>}

            <DialogFooter showCloseButton>
                <Button type="submit" className="cursor-pointer" disabled={!isDirty || !trimmed.name || pending}>
                    {pending ? "Saving…" : "Save changes"}
                </Button>
            </DialogFooter>
        </form>
    );
}

function EditProductDialog({product}: {product: ProductSchema}) {
    const [open, setOpen] = useState(false);

    return (
        <Dialog open={open} onOpenChange={setOpen}>
            <DialogTrigger render={<Button variant="outline" className="cursor-pointer" />}>
                <Pencil />
                Edit
            </DialogTrigger>
            <DialogContent>
                <EditProductForm product={product} onDone={() => setOpen(false)} />
            </DialogContent>
        </Dialog>
    );
}

function DeleteProductButton({product, orderCount}: {product: ProductSchema, orderCount: number}) {
    const appHref = useAppHref();
    const router = useRouter();
    const farm = useFarm()?.farm;

    const orders = orderCount === 1 ? "1 order" : `${orderCount} orders`;

    return (
        <ConfirmDialog
            title={`Delete ${product.Name}?`}
            description={orderCount > 0
                ? `Its stock and price history and its ${orders} are deleted with it. This can't be undone.`
                : "Its stock and price history are deleted with it. This can't be undone."}
            confirmLabel="Delete product"
            onConfirm={async () => {
                if (!farm) return;
                await removeProduct(farm.farm_id, product.ID);
                router.replace(appHref("/products"));
            }}
            trigger={<Button variant="outline" className="cursor-pointer text-danger-strong hover:text-danger-strong" />}
        >
            <Trash2 />
            Delete
        </ConfirmDialog>
    );
}

export default function ProductActions({product, orderCount}: {product: ProductSchema, orderCount: number}) {
    return (
        <div className="flex shrink-0 gap-2">
            <EditProductDialog product={product} />
            <DeleteProductButton product={product} orderCount={orderCount} />
        </div>
    );
}
