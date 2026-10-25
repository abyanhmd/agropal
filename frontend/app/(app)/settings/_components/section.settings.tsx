import React from 'react';

export interface SettingsSectionProps {
    title: string;
    description?: string;
    children: React.ReactNode;
    footer?: React.ReactNode;
}

export default function SettingsSection({title, description, children, footer}: SettingsSectionProps) {
    return (
        <section className="overflow-hidden rounded-xl border border-border bg-background">
            <header className="flex flex-col gap-1 border-b border-border px-5 py-4">
                <h2 className="font-serif text-lg font-medium">{title}</h2>
                {description && <p className="text-sm text-muted-foreground font-sans">{description}</p>}
            </header>

            <div className="px-5 py-5 font-sans">{children}</div>

            {footer &&
                <footer className="flex items-center justify-end gap-3 border-t border-border bg-muted px-5 py-3">
                    {footer}
                </footer>
            }
        </section>
    );
}
