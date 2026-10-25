'use client';

import React from 'react';
import {useTheme} from "next-themes";
import {Monitor, Moon, Sun} from "lucide-react";
import {RadioGroup, RadioGroupItem} from "@/components/ui/radio-group";
import {useIsClient} from "@/hooks/use-is-client";
import SettingsSection from "@/app/(app)/settings/_components/section.settings";

const THEMES = [
    {value: "light", label: "Light", icon: Sun},
    {value: "dark", label: "Dark", icon: Moon},
    {value: "system", label: "System", icon: Monitor},
] as const;

// Tiny window mockup in the theme's colors
function ThemePreview({theme}: {theme: (typeof THEMES)[number]["value"]}) {
    const light = <div className="flex h-full flex-col gap-1.5 bg-white p-2">
        <div className="h-1.5 w-1/2 rounded-full bg-[#8cb88e]" />
        <div className="h-1.5 w-3/4 rounded-full bg-neutral-200" />
        <div className="h-1.5 w-2/3 rounded-full bg-neutral-200" />
    </div>;
    const dark = <div className="flex h-full flex-col gap-1.5 bg-neutral-900 p-2">
        <div className="h-1.5 w-1/2 rounded-full bg-[#8cb88e]" />
        <div className="h-1.5 w-3/4 rounded-full bg-neutral-700" />
        <div className="h-1.5 w-2/3 rounded-full bg-neutral-700" />
    </div>;

    return (
        <div className="h-16 overflow-hidden rounded-md border border-border">
            {theme === "light" && light}
            {theme === "dark" && dark}
            {theme === "system" &&
                <div className="grid h-full grid-cols-2">
                    <div className="overflow-hidden">{light}</div>
                    <div className="overflow-hidden">{dark}</div>
                </div>
            }
        </div>
    );
}

export default function AppearanceSettings() {
    const {theme, setTheme} = useTheme();
    // The saved theme is only known in the browser; select nothing until then to avoid a mismatch
    const isClient = useIsClient();

    return (
        <SettingsSection title="Appearance" description="Choose how Agropal looks. System follows your device setting.">
            <RadioGroup
                aria-label="Theme"
                value={isClient ? (theme ?? "system") : ""}
                onValueChange={(value) => setTheme(value as string)}
                className="grid-cols-1 sm:grid-cols-3"
            >
                {THEMES.map(({value, label, icon: Icon}) => (
                    <label
                        key={value}
                        className="flex cursor-pointer flex-col gap-3 rounded-lg border border-border p-3 transition-colors hover:bg-muted has-data-checked:border-accent has-data-checked:bg-accent-muted/50"
                    >
                        <ThemePreview theme={value} />
                        <span className="flex items-center gap-2 text-sm">
                            <RadioGroupItem value={value} />
                            <Icon className="size-4 text-muted-foreground" />
                            {label}
                        </span>
                    </label>
                ))}
            </RadioGroup>
        </SettingsSection>
    );
}
