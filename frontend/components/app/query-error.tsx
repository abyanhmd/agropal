import React from 'react';
import {RotateCw} from "lucide-react";
import {Button} from "@/components/ui/button";

// Shown when a page's data couldn't be loaded from the backend
export default function QueryError({error, onRetry}: {error: Error, onRetry: () => void}) {
    return (
        <div className="flex flex-col items-center gap-3 rounded-xl border border-dashed border-border py-10 text-center text-sm text-muted-foreground font-sans">
            <p>Couldn&apos;t load this: {error.message}</p>
            <Button variant="outline" size="sm" className="cursor-pointer" onClick={onRetry}>
                <RotateCw />
                Try again
            </Button>
        </div>
    );
}
