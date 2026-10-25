'use client';

import {useSyncExternalStore} from 'react';

const noopSubscribe = () => () => {};

// true only after hydration, for values that differ between server and browser
export function useIsClient(): boolean {
    return useSyncExternalStore(noopSubscribe, () => true, () => false);
}
