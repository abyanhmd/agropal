import React from 'react';

// Stands in for a whole section that has nothing to show yet, instead of a grid of empty cards
export default function EmptyState({icon: Icon, title, children, action}: {
    icon: React.ComponentType<{className?: string}>;
    title: string;
    children?: React.ReactNode;
    action?: React.ReactNode;
}) {
    return (
        <div className="flex flex-col items-center gap-3 rounded-xl border border-dashed border-border px-6 py-14 text-center font-sans">
            <Icon className="size-8 text-muted-foreground" aria-hidden />
            <div className="flex flex-col gap-1">
                <h2 className="font-serif text-xl font-medium">{title}</h2>
                {children && <p className="max-w-md text-sm text-muted-foreground">{children}</p>}
            </div>
            {action}
        </div>
    );
}
