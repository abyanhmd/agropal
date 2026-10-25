import React from 'react';
import type {Metadata} from "next";
import AppearanceSettings from "@/app/(app)/settings/_components/appearance.settings";
import ProfileSettings from "@/app/(app)/settings/_components/profile.settings";
import LocationSettings from "@/app/(app)/settings/_components/location.settings";

export const metadata: Metadata = {title: "Settings"};

// The settings a visitor can look at; password and account deletion mean nothing without an account
export default function DemoSettingsPage() {
    return (
        <div className="page-container flex flex-col gap-6">
            <div className="flex flex-col gap-1">
                <h1 className="font-serif text-3xl font-medium">Settings</h1>
                <p className="text-muted-foreground font-sans">Your profile, land location and appearance. Saving is switched off in the demo.</p>
            </div>

            <ProfileSettings />
            <LocationSettings />
            <AppearanceSettings />
        </div>
    );
}
