import React from 'react';
import Link from "next/link";

export default function NotFoundState({what, backHref}: {what: string, backHref: string}) {
    return (
        <div className="page-container flex flex-col items-start gap-3 font-sans">
            <h1 className="font-serif text-3xl font-medium">Not found</h1>
            <p className="text-muted-foreground">This {what} doesn&apos;t exist, or it isn&apos;t part of your farm.</p>
            <Link href={backHref} className="underline underline-offset-4">Back to {what}s</Link>
        </div>
    );
}
