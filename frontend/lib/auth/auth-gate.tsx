'use client';

import React from 'react';
import {Button} from "@/components/ui/button";
import {useAuth} from "@/lib/auth/auth-provider";

// Holds the app back until the session is known; signed-out users are sent to /login by the provider
export default function AuthGate({children}: {children: React.ReactNode}) {
    const {session, refresh} = useAuth();

    if (session.status === "ready") return children;

    if (session.status === "error") {
        return (
            <div className="flex min-h-svh flex-col items-center justify-center gap-3 p-6 text-center font-sans">
                <p className="text-muted-foreground">Couldn&apos;t reach the server: {session.message}</p>
                <Button variant="outline" className="cursor-pointer" onClick={() => void refresh()}>Try again</Button>
            </div>
        );
    }

    return (
        <div className="flex min-h-svh items-center justify-center" role="status" aria-label="Loading">
            <div className="size-8 animate-spin rounded-full border-2 border-muted border-t-accent-strong" />
        </div>
    );
}
