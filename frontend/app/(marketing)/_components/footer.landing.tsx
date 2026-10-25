import React from 'react';
import Logo from "@/app/(marketing)/_components/logo.landing";

export default function Footer() {
    return (
        <footer className="border-t border-border">
            <div className="mx-auto flex w-full max-w-6xl flex-col items-center justify-between gap-4 px-4 py-8 font-sans text-sm text-muted-foreground sm:flex-row">
                <Logo />
                <p>Made for the people who grow our food.</p>
            </div>
        </footer>
    );
}
