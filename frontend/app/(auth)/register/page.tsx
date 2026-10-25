'use client';

import React, {useState} from 'react';
import Link from "next/link";
import {useRouter} from "next/navigation";
import {register, resendCode, verifyRegistration} from "@/lib/api/auth";
import {Button} from "@/components/ui/button";
import {CodeField, errorMessage, Field, FormError, SubmitButton} from "@/app/(auth)/_components/auth-form";

export default function RegisterPage() {
    const router = useRouter();
    const [form, setForm] = useState({name: "", username: "", email: "", password: ""});
    const [code, setCode] = useState("");
    const [codeSent, setCodeSent] = useState(false);
    const [pending, setPending] = useState(false);
    const [error, setError] = useState<string | null>(null);
    const [resent, setResent] = useState(false);

    const set = (patch: Partial<typeof form>) => setForm({...form, ...patch});

    const handleSubmit = async (e: React.FormEvent) => {
        e.preventDefault();
        setPending(true);
        setError(null);
        try {
            if (!codeSent) {
                await register({...form, name: form.name.trim(), username: form.username.trim(), email: form.email.trim()});
                setCodeSent(true);
            } else {
                await verifyRegistration(form.email.trim(), code);
                router.replace("/dashboard");
            }
        } catch (err) {
            setError(errorMessage(err));
        } finally {
            setPending(false);
        }
    };

    const handleResend = async () => {
        setError(null);
        setResent(false);
        try {
            await resendCode("register", form.email.trim());
            setResent(true);
        } catch (err) {
            setError(errorMessage(err));
        }
    };

    return (
        <form onSubmit={handleSubmit} className="flex flex-col gap-4">
            <div className="flex flex-col gap-1">
                <h1 className="font-serif text-2xl font-medium">Create account</h1>
                <p className="text-sm text-muted-foreground">
                    {codeSent ? `We emailed a code to ${form.email.trim()}.` : "Start tracking your farm's products, orders, and weather."}
                </p>
            </div>

            {codeSent
                ? <>
                    <CodeField value={code} onChange={setCode} />
                    <Button type="button" variant="ghost" size="sm" className="cursor-pointer self-start" onClick={handleResend}>
                        {resent ? "Code sent again" : "Resend code"}
                    </Button>
                </>
                : <>
                    <Field id="name" label="Your name" autoComplete="name" required autoFocus value={form.name} onChange={(e) => set({name: e.target.value})} />
                    <Field id="username" label="Username" autoComplete="username" required value={form.username} onChange={(e) => set({username: e.target.value})} />
                    <Field id="email" label="Email" type="email" autoComplete="email" required value={form.email} onChange={(e) => set({email: e.target.value})} />
                    <Field id="password" label="Password" type="password" autoComplete="new-password" required value={form.password} onChange={(e) => set({password: e.target.value})} />
                </>
            }

            <FormError message={error} />
            <SubmitButton pending={pending}>{codeSent ? "Verify and create account" : "Continue"}</SubmitButton>

            <p className="text-sm text-muted-foreground">
                Already have an account?{" "}
                <Link href="/login" className="underline underline-offset-4 hover:text-foreground">Sign in</Link>
            </p>
        </form>
    );
}
