'use client';

import React from 'react';
import Link from "next/link";
import {ArrowRight, Eye} from "lucide-react";
import {Button} from "@/components/ui/button";

// Pinned above every demo page, so it's always clear this is sample data and how to get a real farm
export default function DemoBanner() {
    return (
        <div className="flex flex-wrap items-center justify-between gap-x-4 gap-y-2 border-b border-border bg-accent-muted px-4 py-2 font-sans text-sm">
            <span className="flex items-center gap-2 text-accent-strong">
                <Eye className="size-4 shrink-0" aria-hidden />
                You&apos;re looking at a demo with sample data. Nothing here is saved.
            </span>
            <span className="flex items-center gap-3">
                <Link href="/" className="text-muted-foreground underline-offset-4 hover:text-foreground hover:underline">Back to home</Link>
                <Button size="sm" nativeButton={false} render={<Link href="/register" />} className="rounded-lg">
                    Get started
                    <ArrowRight />
                </Button>
            </span>
        </div>
    );
}
