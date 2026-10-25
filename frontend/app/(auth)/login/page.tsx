'use client';

import React, {Suspense, useState} from 'react';
import Link from "next/link";
import {useRouter, useSearchParams} from "next/navigation";
import {login, verifyLogin} from "@/lib/api/auth";
import {CodeField, errorMessage, Field, FormError, SubmitButton} from "@/app/(auth)/_components/auth-form";

// Only same-site paths, so a crafted ?next= can't send the user elsewhere
const safeNext = (next: string | null) => (next && next.startsWith("/") && !next.startsWith("//") ? next : "/dashboard");

function LoginForm() {
    const router = useRouter();
    const next = safeNext(useSearchParams().get("next"));
    const [email, setEmail] = useState("");
    const [password, setPassword] = useState("");
    const [code, setCode] = useState("");
    const [codeSent, setCodeSent] = useState(false);
    const [pending, setPending] = useState(false);
    const [error, setError] = useState<string | null>(null);

    const handleSubmit = async (e: React.FormEvent) => {
        e.preventDefault();
        setPending(true);
        setError(null);
        try {
            if (!codeSent) {
                await login(email.trim(), password);
                setCodeSent(true);
            } else {
                await verifyLogin(email.trim(), code);
                router.replace(next);
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
                <h1 className="font-serif text-2xl font-medium">Sign in</h1>
                <p className="text-sm text-muted-foreground">
                    {codeSent ? `We emailed a code to ${email.trim()}.` : "Welcome back to your farm workspace."}
                </p>
            </div>

            {codeSent
                ? <CodeField value={code} onChange={setCode} />
                : <>
                    <Field id="email" label="Email" type="email" autoComplete="email" required autoFocus value={email} onChange={(e) => setEmail(e.target.value)} />
                    <Field id="password" label="Password" type="password" autoComplete="current-password" required value={password} onChange={(e) => setPassword(e.target.value)} />
                </>
            }

            <FormError message={error} />
            <SubmitButton pending={pending}>{codeSent ? "Verify and sign in" : "Continue"}</SubmitButton>

            <div className="flex justify-between text-sm text-muted-foreground">
                <Link href="/forgot-password" className="underline underline-offset-4 hover:text-foreground">Forgot password?</Link>
                <Link href="/register" className="underline underline-offset-4 hover:text-foreground">Create account</Link>
            </div>
        </form>
    );
}

// useSearchParams needs a Suspense boundary so the rest of the route can still be prerendered
export default function LoginPage() {
    return (
        <Suspense>
            <LoginForm />
        </Suspense>
    );
}
