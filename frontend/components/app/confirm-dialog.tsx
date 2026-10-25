'use client';

import React, {useId, useState} from 'react';
import {Button} from "@/components/ui/button";
import {Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle, DialogTrigger} from "@/components/ui/dialog";
import {Input} from "@/components/ui/input";
import {Label} from "@/components/ui/label";
import {errorMessage} from "@/app/(auth)/_components/auth-form";

// The theme has no destructive tokens wired up, so destructive buttons use the danger palette
export const DANGER_BUTTON = "cursor-pointer bg-danger text-danger-foreground hover:bg-danger-hover";

export interface ConfirmDialogProps {
    title: string;
    description: React.ReactNode;
    confirmLabel: string;
    pendingLabel?: string;
    // Resolves when done; a rejection is shown in the dialog, which stays open
    onConfirm: () => Promise<void>;
    // For actions that can't be taken back: the confirm button stays off until this is typed out
    requireText?: string;
    // The element that opens the dialog, e.g. <Button variant="ghost" />, and its contents
    trigger: React.ReactElement;
    children: React.ReactNode;
}

function ConfirmBody({title, description, confirmLabel, pendingLabel, onConfirm, requireText, onDone}: Omit<ConfirmDialogProps, "trigger" | "children"> & {onDone: () => void}) {
    const [pending, setPending] = useState(false);
    const [error, setError] = useState<string | null>(null);
    const [typed, setTyped] = useState("");
    const id = useId();

    const confirm = async () => {
        setPending(true);
        setError(null);
        try {
            await onConfirm();
            onDone();
        } catch (err) {
            setError(errorMessage(err));
            setPending(false);
        }
    };

    return (
        <>
            <DialogHeader>
                <DialogTitle>{title}</DialogTitle>
                <DialogDescription>{description}</DialogDescription>
            </DialogHeader>
            {requireText &&
                <div className="flex flex-col gap-2 font-sans">
                    <Label htmlFor={id}>Type <span className="font-medium">{requireText}</span> to confirm</Label>
                    <Input id={id} autoComplete="off" value={typed} onChange={(e) => setTyped(e.target.value)} />
                </div>
            }
            {error && <p role="alert" className="text-sm text-danger-strong font-sans">{error}</p>}
            <DialogFooter showCloseButton>
                <Button className={DANGER_BUTTON} disabled={pending || (requireText !== undefined && typed.trim() !== requireText)} onClick={() => void confirm()}>
                    {pending ? (pendingLabel ?? "Deleting…") : confirmLabel}
                </Button>
            </DialogFooter>
        </>
    );
}

export default function ConfirmDialog({trigger, children, ...body}: ConfirmDialogProps) {
    const [open, setOpen] = useState(false);

    return (
        <Dialog open={open} onOpenChange={setOpen}>
            <DialogTrigger render={trigger}>{children}</DialogTrigger>
            <DialogContent>
                {/* Its own component, so it starts fresh (no old error) each time the dialog opens */}
                <ConfirmBody {...body} onDone={() => setOpen(false)} />
            </DialogContent>
        </Dialog>
    );
}
