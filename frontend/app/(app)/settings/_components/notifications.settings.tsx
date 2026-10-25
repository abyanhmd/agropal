'use client';

import React from 'react';
import {Label} from "@/components/ui/label";
import {Switch} from "@/components/ui/switch";
import {AppSettings, useSettings} from "@/hooks/use-settings";
import SettingsSection from "@/app/(app)/settings/_components/section.settings";

const OPTIONS: {key: keyof AppSettings["notifications"], label: string, description: string}[] = [
    {key: "lowStock", label: "Low stock alerts", description: "When a product's stock keeps dropping."},
    {key: "priceChanges", label: "Price changes", description: "When a product's price is updated."},
    {key: "weeklySummary", label: "Weekly summary", description: "A recap of stock and prices every Monday."},
];

export default function NotificationSettings() {
    const {settings, update} = useSettings();

    return (
        <SettingsSection title="Notifications" description="Changes save as soon as you flip a switch.">
            <ul className="-my-3 divide-y divide-border">
                {OPTIONS.map(({key, label, description}) => (
                    <li key={key} className="flex items-center justify-between gap-4 py-3">
                        <div className="flex flex-col gap-0.5">
                            <Label htmlFor={`notify-${key}`}>{label}</Label>
                            <span className="text-sm text-muted-foreground">{description}</span>
                        </div>
                        <Switch
                            id={`notify-${key}`}
                            checked={settings.notifications[key]}
                            onCheckedChange={(checked) => update("notifications", {[key]: checked})}
                        />
                    </li>
                ))}
            </ul>
        </SettingsSection>
    );
}
