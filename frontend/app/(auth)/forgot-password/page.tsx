'use client';

import React, {useState} from 'react';
import Link from "next/link";
import {useRouter} from "next/navigation";
import {confirmPasswordReset, requestPasswordReset} from "@/lib/api/auth";
import {CodeField, errorMessage, Field, FormError, SubmitButton} from "@/app/(auth)/_components/auth-form";

export default function ForgotPasswordPage() {
    const router = useRouter();
    const [email, setEmail] = useState("");
    const [code, setCode] = useState("");
    const [newPassword, setNewPassword] = useState("");
    const [codeSent, setCodeSent] = useState(false);
    const [pending, setPending] = useState(false);
    const [error, setError] = useState<string | null>(null);

    const handleSubmit = async (e: React.FormEvent) => {
        e.preventDefault();
        setPending(true);
        setError(null);
        try {
            if (!codeSent) {
                await requestPasswordReset(email.trim());
                setCodeSent(true);
            } else {
                await confirmPasswordReset(email.trim(), code, newPassword);
                router.replace("/login");
            }
        } catch (err) {
            setError(errorMessage(err));
        } finally {
            setPending(false);
        }
    };

    return (
        <form onSubmit={handleSubmit} className="flex flex-col gap-4">
            <div className="flex flex-col gap-1">
                <h1 className="font-serif text-2xl font-medium">Reset password</h1>
                <p className="text-sm text-muted-foreground">
                    {codeSent ? "Enter the code we emailed you and choose a new password." : "We'll email you a code to reset it."}
                </p>
            </div>

            {codeSent
                ? <>
                    <CodeField value={code} onChange={setCode} />
                    <Field id="new-password" label="New password" type="password" autoComplete="new-password" required value={newPassword} onChange={(e) => setNewPassword(e.target.value)} />
                </>
                : <Field id="email" label="Email" type="email" autoComplete="email" required autoFocus value={email} onChange={(e) => setEmail(e.target.value)} />
            }

            <FormError message={error} />
            <SubmitButton pending={pending}>{codeSent ? "Reset password" : "Send code"}</SubmitButton>

            <Link href="/login" className="text-sm text-muted-foreground underline underline-offset-4 hover:text-foreground">Back to sign in</Link>
        </form>
    );
}
