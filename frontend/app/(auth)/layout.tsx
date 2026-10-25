import type {Metadata} from "next";
import {Sprout} from "lucide-react";
import type {ReactNode} from "react";

export const metadata: Metadata = {
    title: "Sign in",
};

export default function AuthLayout({children}: {children: ReactNode}) {
    return (
        <main className="flex min-h-svh flex-col items-center justify-center gap-6 px-4 py-10">
            <div className="flex items-center gap-2.5">
                <span className="flex size-9 items-center justify-center rounded-lg bg-sidebar-primary text-sidebar-primary-foreground">
                    <Sprout className="size-5" />
                </span>
                <span className="font-logo text-2xl font-semibold">Agropal</span>
            </div>
            <div className="w-full max-w-sm rounded-xl border border-border bg-background p-6 font-sans">
                {children}
            </div>
        </main>
    );
}
