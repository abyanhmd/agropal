'use client';

import {useCallback, useEffect, useState, useSyncExternalStore} from 'react';

// A small fetch cache: the last result per key is shown immediately on the next visit while a
// fresh one loads, and invalidate() refetches every mounted query whose key starts with a prefix.

type Entry = {data: unknown};
const cache = new Map<string, Entry>();
const listeners = new Set<(prefix: string) => void>();

export function invalidate(prefix: string) {
    for (const key of cache.keys()) if (key.startsWith(prefix)) cache.delete(key);
    listeners.forEach((listener) => listener(prefix));
}

// Drops cached results without refetching what's on screen. For deletes that navigate away:
// a refetch would flash "not found" on the page being left, and the next page loads fresh anyway.
export function forget(prefix: string) {
    for (const key of cache.keys()) if (key.startsWith(prefix)) cache.delete(key);
}

export function clearQueries() {
    cache.clear();
}

export interface QueryState<T> {
    data: T | undefined;
    error: Error | null;
    // true until the first result for this key, not on background refetches
    loading: boolean;
    reload: () => void;
}

const subscribeNothing = () => () => {};

// key === null waits (e.g. until the farm is known)
export function useQuery<T>(key: string | null, fetcher: () => Promise<T>): QueryState<T> {
    const [state, setState] = useState<{key: string | null, data?: T, error: Error | null}>({key, error: null});
    const [tick, setTick] = useState(0);

    useEffect(() => {
        if (key === null) return;
        const handler = (prefix: string) => {
            if (key.startsWith(prefix)) setTick((n) => n + 1);
        };
        listeners.add(handler);
        return () => {
            listeners.delete(handler);
        };
    }, [key]);

    useEffect(() => {
        if (key === null) return;
        let cancelled = false;
        fetcher()
            .then((data) => {
                cache.set(key, {data});
                if (!cancelled) setState({key, data, error: null});
            })
            .catch((error: unknown) => {
                if (!cancelled) setState({key, error: error instanceof Error ? error : new Error(String(error))});
            });
        return () => {
            cancelled = true;
        };
        // The fetcher is recreated every render; the key identifies what it loads
        // eslint-disable-next-line react-hooks/exhaustive-deps
    }, [key, tick]);

    const reload = useCallback(() => setTick((n) => n + 1), []);

    // Results for another key are stale; show the cached one for this key if there is one
    const own = state.key === key ? state : undefined;
    // The cache only exists in the browser, so the server and the hydrating render can't read it:
    // they would disagree about what to show. False there, true on every later render.
    const hydrated = useSyncExternalStore(subscribeNothing, () => true, () => false);
    const cached = key === null || !hydrated ? undefined : (cache.get(key)?.data as T | undefined);
    const data = own?.data ?? cached;

    return {data, error: own?.error ?? null, loading: data === undefined && !own?.error, reload};
}
