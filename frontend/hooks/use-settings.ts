'use client';

import {useSyncExternalStore} from 'react';

// Preferences kept in this browser's localStorage. The profile (name, farm) is saved through the backend.

export interface AppSettings {
    notifications: {
        lowStock: boolean;
        priceChanges: boolean;
        weeklySummary: boolean;
    };
}

export const DEFAULT_SETTINGS: AppSettings = {
    notifications: {lowStock: true, priceChanges: false, weeklySummary: true},
};

const STORAGE_KEY = "agropal.settings";

let current: AppSettings | null = null;
const listeners = new Set<() => void>();

// Fill in anything missing or mistyped from the defaults, so older saved shapes still load
function sanitize(value: unknown): AppSettings {
    const saved = (value ?? {}) as Partial<Record<keyof AppSettings, Record<string, unknown>>>;
    const pick = <T extends object>(defaults: T, source: Record<string, unknown> | undefined): T =>
        Object.fromEntries(Object.entries(defaults).map(([key, fallback]) => {
            const candidate = source?.[key];
            return [key, typeof candidate === typeof fallback ? candidate : fallback];
        })) as T;

    return {
        notifications: pick(DEFAULT_SETTINGS.notifications, saved.notifications),
    };
}

function load(): AppSettings {
    try {
        const raw = localStorage.getItem(STORAGE_KEY);
        return raw ? sanitize(JSON.parse(raw)) : DEFAULT_SETTINGS;
    } catch {
        // Storage blocked (private mode, disabled site data) or corrupt JSON
        return DEFAULT_SETTINGS;
    }
}

function notify() {
    listeners.forEach((listener) => listener());
}

function subscribe(listener: () => void) {
    listeners.add(listener);

    // Keep other open tabs in sync
    const onStorage = (event: StorageEvent) => {
        if (event.key === STORAGE_KEY) {
            current = load();
            notify();
        }
    };
    window.addEventListener("storage", onStorage);

    return () => {
        listeners.delete(listener);
        window.removeEventListener("storage", onStorage);
    };
}

function getSnapshot(): AppSettings {
    if (current === null) current = load();
    return current;
}

// The server has no localStorage, so it (and the first client render) use the defaults
function getServerSnapshot(): AppSettings {
    return DEFAULT_SETTINGS;
}

function saveSettings(next: AppSettings) {
    current = next;
    try {
        localStorage.setItem(STORAGE_KEY, JSON.stringify(next));
    } catch {
        // Still applied for this session, just not persisted
    }
    notify();
}

export function useSettings() {
    const settings = useSyncExternalStore(subscribe, getSnapshot, getServerSnapshot);

    // Builds on the latest stored value, so quick successive updates don't overwrite each other
    const update = <K extends keyof AppSettings>(section: K, patch: Partial<AppSettings[K]>) => {
        const latest = getSnapshot();
        saveSettings({...latest, [section]: {...latest[section], ...patch}});
    };

    return {settings, update};
}
