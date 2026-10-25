'use client';

import React, {createContext, useContext} from 'react';

// The public demo lives under /demo and shows sample data, so the app's pages are reused as they
// are and only their links need to know where they sit: "" in the real app, "/demo" in the demo.

export const DEMO_BASE = "/demo";
export const DEMO_FARM_ID = "demo-farm";

const BaseContext = createContext("");

export function AppBaseProvider({base, children}: {base: string, children: React.ReactNode}) {
    return <BaseContext value={base}>{children}</BaseContext>;
}

export const useAppBase = () => useContext(BaseContext);

// Turns an in-app path like "/products/1" into the link for wherever the user currently is
export function useAppHref(): (path: string) => string {
    const base = useContext(BaseContext);
    return (path) => base + path;
}

// Requests made from the demo never reach the network; the path says which world we are in
export function inDemo(): boolean {
    return typeof window !== "undefined" && (window.location.pathname === DEMO_BASE || window.location.pathname.startsWith(`${DEMO_BASE}/`));
}
