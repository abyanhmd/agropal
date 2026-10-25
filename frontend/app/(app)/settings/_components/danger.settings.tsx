'use client';

import React from 'react';
import {Trash2} from "lucide-react";
import {Button} from "@/components/ui/button";
import ConfirmDialog, {DANGER_BUTTON} from "@/components/app/confirm-dialog";
import SettingsSection from "@/app/(app)/settings/_components/section.settings";
import {useAuth} from "@/lib/auth/auth-provider";
import {clearQueries} from "@/lib/api/use-query";
import {deleteAccount, deleteFarm} from "@/lib/api/resources";

function DangerRow({title, description, children}: {title: string, description: string, children: React.ReactNode}) {
    return (
        <li className="flex flex-col gap-3 py-4 first:pt-0 last:pb-0 sm:flex-row sm:items-center sm:justify-between">
            <div className="flex flex-col gap-0.5">
                <span className="text-sm font-medium">{title}</span>
                <span className="text-sm text-muted-foreground">{description}</span>
            </div>
            {children}
        </li>
    );
}

export default function DangerSettings() {
    const {session, refresh, signOut} = useAuth();
    if (session.status !== "ready") return null;
    const {user, farm} = session;

    return (
        <SettingsSection title="Danger zone" description="These can't be undone.">
            <ul className="divide-y divide-border">
                <DangerRow
                    title="Delete this farm"
                    description="Removes its products, prices, stock, locations and orders. You'll start again with a new, empty farm."
                >
                    <ConfirmDialog
                        title="Delete this farm?"
                        description={<>Everything in <span className="font-medium">{farm.name}</span> is deleted for good: products, prices, stock, locations and orders. A new, empty farm takes its place.</>}
                        confirmLabel="Delete farm"
                        requireText="DELETE"
                        onConfirm={async () => {
                            await deleteFarm(farm.farm_id);
                            // The old farm's data is gone; the session reload sets up the replacement
                            clearQueries();
                            await refresh();
                        }}
                        trigger={<Button variant="outline" className="cursor-pointer text-danger-strong" />}
                    >
                        <Trash2 />
                        Delete farm
                    </ConfirmDialog>
                </DangerRow>

                <DangerRow
                    title="Delete your account"
                    description="Closes your account and signs you out. Your email and username become free to use again."
                >
                    <ConfirmDialog
                        title="Delete your account?"
                        description="You'll be signed out and won't be able to sign in again. Your discussion posts stay up."
                        confirmLabel="Delete account"
                        requireText={user.username}
                        onConfirm={async () => {
                            await deleteAccount();
                            await signOut();
                        }}
                        trigger={<Button className={DANGER_BUTTON} />}
                    >
                        <Trash2 />
                        Delete account
                    </ConfirmDialog>
                </DangerRow>
            </ul>
        </SettingsSection>
    );
}
