'use client';

import React, {createContext, useCallback, useContext, useEffect, useMemo, useState} from 'react';
import {useRouter} from "next/navigation";
import {onSessionChange} from "@/lib/api/client";
import {logout as apiLogout, restoreSession} from "@/lib/api/auth";
import {createFarm, FarmDto, getMe, listFarms, UserDto} from "@/lib/api/resources";
import {clearQueries} from "@/lib/api/use-query";
import {DEMO_FARM, DEMO_USER} from "@/lib/demo/identity";

// Signed-in user and their farm for everything under (app). The farm is the user's first one,
// created on first sign-in if there is none, since every product and order belongs to a farm.

export type Session =
    | {status: "loading"}
    | {status: "signed-out"}
    | {status: "error", message: string}
    | {status: "ready", user: UserDto, farm: FarmDto};

interface AuthContextValue {
    session: Session;
    refresh: () => Promise<void>;
    signOut: () => Promise<void>;
}

const AuthContext = createContext<AuthContextValue | null>(null);

async function loadSession(): Promise<Session> {
    if (!await restoreSession()) return {status: "signed-out"};

    const user = await getMe();
    const [existing] = await listFarms();
    if (existing) return {status: "ready", user, farm: existing};

    // First sign-in: the farm is created here, and the endpoint says which one it made
    const name = `${user.name || user.username}'s farm`;
    const farmID = await createFarm(name);
    return {status: "ready", user, farm: {farm_id: farmID, user_id: user.user_id, name}};
}

export function AuthProvider({children}: {children: React.ReactNode}) {
    const router = useRouter();
    const [session, setSession] = useState<Session>({status: "loading"});

    const refresh = useCallback(async () => {
        try {
            setSession(await loadSession());
        } catch (error) {
            setSession({status: "error", message: error instanceof Error ? error.message : "Something went wrong"});
        }
    }, []);

    useEffect(() => {
        // eslint-disable-next-line react-hooks/set-state-in-effect
        void refresh();
    }, [refresh]);

    // A refresh that fails mid-session means the cookie expired or was revoked
    useEffect(() => onSessionChange((signedIn) => {
        if (!signedIn) setSession({status: "signed-out"});
    }), []);

    useEffect(() => {
        if (session.status === "signed-out") {
            clearQueries();
            // Read at redirect time: usePathname would make every route wait for runtime data
            const here = window.location.pathname + window.location.search;
            router.replace(`/login?next=${encodeURIComponent(here)}`);
        }
    }, [session.status, router]);

    const signOut = useCallback(async () => {
        await apiLogout();
        setSession({status: "signed-out"});
    }, []);

    const value = useMemo(() => ({session, refresh, signOut}), [session, refresh, signOut]);
    return <AuthContext value={value}>{children}</AuthContext>;
}

// The demo's signed-in user: always ready, and signing out just leaves the demo
const DEMO_SESSION: Session = {status: "ready", user: DEMO_USER, farm: DEMO_FARM};

export function DemoAuthProvider({children}: {children: React.ReactNode}) {
    const router = useRouter();
    const value = useMemo(() => ({
        session: DEMO_SESSION,
        refresh: async () => {},
        signOut: async () => router.push("/"),
    }), [router]);
    return <AuthContext value={value}>{children}</AuthContext>;
}

export function useAuth() {
    const value = useContext(AuthContext);
    if (!value) throw new Error("useAuth must be used inside AuthProvider");
    return value;
}

// For pages that only render once the session is ready (AuthGate guarantees it)
export function useFarm() {
    const {session} = useAuth();
    return session.status === "ready" ? {user: session.user, farm: session.farm} : null;
}
