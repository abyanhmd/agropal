'use client';

import React, {useState} from 'react';
import {Check} from "lucide-react";
import {Button} from "@/components/ui/button";
import {Input} from "@/components/ui/input";
import {Label} from "@/components/ui/label";
import SettingsSection from "@/app/(app)/settings/_components/section.settings";
import {useAuth} from "@/lib/auth/auth-provider";
import {updateFarm, updateMe} from "@/lib/api/resources";
import {errorMessage} from "@/app/(auth)/_components/auth-form";

interface Profile {
    name: string;
    farmName: string;
}

function ProfileForm({saved, email, onSave}: {saved: Profile, email: string, onSave: (profile: Profile) => Promise<void>}) {
    const [draft, setDraft] = useState<Profile>(saved);
    const [pending, setPending] = useState(false);
    const [error, setError] = useState<string | null>(null);
    const isDirty = draft.name.trim() !== saved.name || draft.farmName.trim() !== saved.farmName;

    const handleSubmit = async (e: React.FormEvent) => {
        e.preventDefault();
        setPending(true);
        setError(null);
        try {
            await onSave({name: draft.name.trim(), farmName: draft.farmName.trim()});
        } catch (err) {
            setError(errorMessage(err));
            setPending(false);
        }
    };

    return (
        <form onSubmit={handleSubmit}>
            <SettingsSection
                title="Profile"
                description="Your account and farm."
                footer={
                    <>
                        {error && <p role="alert" className="mr-auto text-sm text-danger-strong">{error}</p>}
                        {isDirty &&
                            <Button type="button" variant="ghost" onClick={() => setDraft(saved)}>
                                Discard
                            </Button>
                        }
                        <Button type="submit" disabled={!isDirty || pending || !draft.name.trim() || !draft.farmName.trim()}>
                            {pending ? "Saving…" : "Save changes"}
                        </Button>
                    </>
                }
            >
                <div className="grid gap-4 sm:grid-cols-2">
                    <div className="flex flex-col gap-2">
                        <Label htmlFor="profile-name">Your name</Label>
                        <Input id="profile-name" required placeholder="Jane Farmer" value={draft.name} onChange={(e) => setDraft({...draft, name: e.target.value})} />
                    </div>
                    <div className="flex flex-col gap-2">
                        <Label htmlFor="profile-farmName">Farm name</Label>
                        <Input id="profile-farmName" required placeholder="Green Valley Farm" value={draft.farmName} onChange={(e) => setDraft({...draft, farmName: e.target.value})} />
                    </div>
                    <div className="flex flex-col gap-2 sm:col-span-2">
                        <Label htmlFor="profile-email">Email</Label>
                        <Input id="profile-email" type="email" value={email} readOnly disabled />
                    </div>
                </div>
            </SettingsSection>
        </form>
    );
}

export default function ProfileSettings() {
    const {session, refresh} = useAuth();
    const [justSaved, setJustSaved] = useState(false);

    if (session.status !== "ready") return null;
    const {user, farm} = session;

    const handleSave = async (profile: Profile) => {
        if (profile.name !== user.name) await updateMe({name: profile.name});
        if (profile.farmName !== farm.name) await updateFarm(farm.farm_id, profile.farmName);
        await refresh();
        setJustSaved(true);
        setTimeout(() => setJustSaved(false), 2000);
    };

    return (
        <div className="relative">
            {/* Remount the form when the saved profile changes, so it starts from the new values */}
            <ProfileForm key={`${user.name}|${farm.name}`} saved={{name: user.name, farmName: farm.name}} email={user.email} onSave={handleSave} />
            {justSaved &&
                <p role="status" className="absolute bottom-3.5 left-5 flex items-center gap-1.5 text-sm text-accent-strong font-sans">
                    <Check className="size-4" />
                    Saved
                </p>
            }
        </div>
    );
}
