'use client';

import React, {useState} from 'react';
import {Button} from "@/components/ui/button";
import SettingsSection from "@/app/(app)/settings/_components/section.settings";
import {Field, FormError} from "@/app/(auth)/_components/auth-form";
import {errorMessage} from "@/app/(auth)/_components/auth-form";
import {useAuth} from "@/lib/auth/auth-provider";
import {changePassword} from "@/lib/api/resources";

const EMPTY = {current: "", next: "", confirm: ""};

export default function PasswordSettings() {
    const {signOut} = useAuth();
    const [form, setForm] = useState(EMPTY);
    const [pending, setPending] = useState(false);
    const [error, setError] = useState<string | null>(null);
    const [done, setDone] = useState(false);

    const mismatch = form.confirm !== "" && form.next !== form.confirm;
    const canSubmit = form.current !== "" && form.next !== "" && form.next === form.confirm && !pending && !done;

    const handleSubmit = async (e: React.FormEvent) => {
        e.preventDefault();
        if (!canSubmit) return;
        setPending(true);
        setError(null);
        try {
            await changePassword(form.current, form.next);
            setDone(true);
            // Every session ended with the old password, this one included
            setTimeout(() => void signOut(), 1500);
        } catch (err) {
            setError(errorMessage(err));
            setPending(false);
        }
    };

    return (
        <form onSubmit={handleSubmit}>
            <SettingsSection
                title="Password"
                description="Changing it signs you out everywhere, so you'll sign in again with the new one."
                footer={
                    <>
                        {done && <p role="status" className="mr-auto text-sm text-accent-strong">Password changed. Signing you out…</p>}
                        {!done && <FormError message={mismatch ? "The new passwords don't match." : error} />}
                        <Button type="submit" disabled={!canSubmit}>{pending ? "Changing…" : "Change password"}</Button>
                    </>
                }
            >
                <div className="grid gap-4 sm:grid-cols-3">
                    <Field id="current-password" label="Current password" type="password" autoComplete="current-password" required value={form.current} onChange={(e) => setForm({...form, current: e.target.value})} />
                    <Field id="new-password" label="New password" type="password" autoComplete="new-password" required value={form.next} onChange={(e) => setForm({...form, next: e.target.value})} />
                    <Field id="confirm-password" label="Confirm new password" type="password" autoComplete="new-password" required value={form.confirm} onChange={(e) => setForm({...form, confirm: e.target.value})} />
                </div>
                <p className="pt-3 text-sm text-muted-foreground">At least 8 characters, with a lowercase letter, an uppercase letter, a number, and one of @ $ ! % * ? &amp;</p>
            </SettingsSection>
        </form>
    );
}
