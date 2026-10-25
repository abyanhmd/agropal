'use client';

import React, {useState} from 'react';
import Link from "next/link";
import {Home, Shovel, Sprout} from "lucide-react";
import {Button} from "@/components/ui/button";

const EXCUSES = [
    "A goat ate this page. We're negotiating its return.",
    "We planted this page, but it never sprouted.",
    "The tractor took a wrong turn somewhere around here.",
    "This page is lying fallow for the season.",
    "The scarecrow did its job a little too well.",
    "Our chickens crossed the road and took this page with them.",
    "We checked the barn, the silo, and under the hay bales. Nothing.",
    "This page was harvested early. Sorry about that.",
];

// Shared by the standalone 404 (unknown URLs) and the in-app one (notFound() inside the app shell)
export default function NotFoundContent({homeHref = "/"}: {homeHref?: string}) {
    const [excuse, setExcuse] = useState(0);

    return (
        <div className="flex flex-1 flex-col items-center justify-center gap-8 px-6 py-16 text-center">
            <div className="flex flex-col items-center">
                <div className="flex items-end gap-3 font-serif text-8xl sm:text-9xl font-semibold leading-none text-accent-strong select-none">
                    <span>4</span>
                    {/* The "0" is a little sprout that's trying its best */}
                    <span className="mb-1 flex size-20 sm:size-24 items-center justify-center rounded-full bg-accent-muted">
                        <Sprout className="size-10 sm:size-12 origin-bottom animate-bounce" />
                    </span>
                    <span>4</span>
                </div>
                {/* Soil line */}
                <div className="mt-3 h-1.5 w-64 rounded-full bg-[repeating-linear-gradient(90deg,var(--accent-strong)_0_10px,transparent_10px_16px)] opacity-40" />
            </div>

            <div className="flex max-w-md flex-col gap-3">
                <h1 className="font-serif text-2xl">Nothing grows here (yet)</h1>
                <p
                    key={excuse}
                    aria-live="polite"
                    className="min-h-12 font-sans text-muted-foreground"
                >
                    {EXCUSES[excuse]}
                </p>
            </div>

            <div className="flex flex-wrap items-center justify-center gap-3">
                <Button size="lg" nativeButton={false} render={<Link href={homeHref} />} className="rounded-xl">
                    <Home />
                    Back to the farm
                </Button>
                <Button
                    size="lg"
                    variant="outline"
                    className="rounded-xl cursor-pointer"
                    onClick={() => setExcuse((i) => (i + 1) % EXCUSES.length)}
                >
                    <Shovel />
                    Ask the farmer why
                </Button>
            </div>
        </div>
    );
}
