import React from 'react';
import type {Metadata} from "next";
import AppearanceSettings from "@/app/(app)/settings/_components/appearance.settings";
import ProfileSettings from "@/app/(app)/settings/_components/profile.settings";
import PasswordSettings from "@/app/(app)/settings/_components/password.settings";
import DangerSettings from "@/app/(app)/settings/_components/danger.settings";
import LocationSettings from "@/app/(app)/settings/_components/location.settings";

export const metadata: Metadata = {
    title: "Settings",
};

// Notification preferences are left off the page until the backend sends notifications: the switches
// (notifications.settings.tsx, hooks/use-settings.ts) only saved to this browser and nothing read them.
export default function SettingsPage() {
    return (
        <div className="page-container flex flex-col gap-6">
            <div className="flex flex-col gap-1">
                <h1 className="font-serif text-3xl font-medium">Settings</h1>
                <p className="text-muted-foreground font-sans">Manage your profile, land location, appearance, password and account.</p>
            </div>

            <ProfileSettings />
            <LocationSettings />
            <AppearanceSettings />
            <PasswordSettings />
            <DangerSettings />
        </div>
    );
}
