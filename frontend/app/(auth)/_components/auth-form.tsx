'use client';

import React, {useState} from 'react';
import {Eye, EyeOff} from "lucide-react";
import {Button} from "@/components/ui/button";
import {Input} from "@/components/ui/input";
import {Label} from "@/components/ui/label";
import {ApiError} from "@/lib/api/client";

export function errorMessage(error: unknown): string {
    if (error instanceof ApiError) {
        return error.fields.length > 0 ? `${error.message}: ${error.fields.join(", ")}` : error.message;
    }
    return "Couldn't reach the server. Check your connection and try again.";
}

export function Field({id, label, type, ...props}: {id: string, label: string} & React.ComponentProps<typeof Input>) {
    const [visible, setVisible] = useState(false);

    if (type !== "password") {
        return (
            <div className="flex flex-col gap-2">
                <Label htmlFor={id}>{label}</Label>
                <Input id={id} type={type} {...props} />
            </div>
        );
    }

    return (
        <div className="flex flex-col gap-2">
            <Label htmlFor={id}>{label}</Label>
            <div className="relative">
                <Input id={id} type={visible ? "text" : "password"} className="pr-10" {...props} />
                <button
                    type="button"
                    onClick={() => setVisible((v) => !v)}
                    aria-label={visible ? "Hide password" : "Show password"}
                    aria-pressed={visible}
                    className="absolute inset-y-0 right-0 flex w-10 cursor-pointer items-center justify-center text-muted-foreground hover:text-foreground"
                >
                    {visible ? <EyeOff className="size-4" /> : <Eye className="size-4" />}
                </button>
            </div>
        </div>
    );
}

export function FormError({message}: {message: string | null}) {
    return message ? <p role="alert" className="text-sm text-danger-strong">{message}</p> : null;
}

export function SubmitButton({pending, children}: {pending: boolean, children: React.ReactNode}) {
    return <Button type="submit" disabled={pending} className="cursor-pointer">{pending ? "Please wait…" : children}</Button>;
}

// Step two of every flow: the code emailed to the user
export function CodeField({value, onChange}: {value: string, onChange: (code: string) => void}) {
    return (
        <Field
            id="code"
            label="Verification code"
            inputMode="numeric"
            autoComplete="one-time-code"
            autoFocus
            required
            placeholder="123456"
            value={value}
            onChange={(e) => onChange(e.target.value.trim())}
        />
    );
}
