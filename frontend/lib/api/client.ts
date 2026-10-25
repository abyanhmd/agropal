// Browser client for the Agropal backend. The access token lives in memory only; the refresh token
// is the backend's httpOnly `rid` cookie, which is why every request sends credentials. A 401 is
// answered by rotating the pair once and retrying.

import {inDemo} from "@/lib/demo/base";

export const API_URL = process.env.NEXT_PUBLIC_API_URL ?? "http://localhost:8080/api/v1";

interface ErrorEnvelope {
    info?: {code?: string, key?: string, message?: string};
    // Validation failures by field, e.g. {"price": ["price must be greater than 0"]}
    errors?: Record<string, string[]> | null;
}

export class ApiError extends Error {
    constructor(
        readonly status: number,
        message: string,
        readonly key?: string,
        // One message per failed rule; each already names its field
        readonly fields: string[] = [],
    ) {
        super(message);
    }
}

let accessToken: string | null = null;
let refreshing: Promise<boolean> | null = null;
const sessionListeners = new Set<(signedIn: boolean) => void>();

export function getAccessToken() {
    return accessToken;
}

export function setAccessToken(token: string | null) {
    accessToken = token;
}

// Told when a refresh fails mid-session, so the app can send the user back to sign in
export function onSessionChange(listener: (signedIn: boolean) => void) {
    sessionListeners.add(listener);
    return () => {
        sessionListeners.delete(listener);
    };
}

// The token pair struct has no json tags, so its fields keep the Go names
interface TokenPair {
    AccessTokenStr: string;
}

async function parse(response: Response): Promise<unknown> {
    const text = await response.text();
    if (!text) return null;
    try {
        return JSON.parse(text);
    } catch {
        return text;
    }
}

function toError(response: Response, body: unknown): ApiError {
    const envelope = (typeof body === "object" && body ? body : {}) as ErrorEnvelope;
    return new ApiError(
        response.status,
        envelope.info?.message ?? `The server responded with ${response.status}`,
        envelope.info?.key,
        Object.values(envelope.errors ?? {}).flat(),
    );
}

// Rotates the token pair with the rid cookie. Concurrent callers share one request, since each
// refresh invalidates the previous refresh token.
export function refreshSession(): Promise<boolean> {
    refreshing ??= (async () => {
        try {
            const response = await fetch(`${API_URL}/auth/refresh`, {method: "POST", credentials: "include"});
            if (!response.ok) {
                accessToken = null;
                return false;
            }
            const body = await parse(response) as {data?: TokenPair} | null;
            accessToken = body?.data?.AccessTokenStr ?? null;
            return accessToken !== null;
        } catch {
            return false;
        } finally {
            refreshing = null;
        }
    })();
    return refreshing;
}

interface RequestOptions {
    method?: "GET" | "POST" | "PATCH" | "PUT" | "DELETE";
    body?: unknown;
    signal?: AbortSignal;
    // Auth endpoints don't carry a token and must not trigger a refresh on 401
    auth?: boolean;
}

// Like api, but also hands back the response headers, for the few endpoints that report extra
// information there (e.g. a list's total)
export async function apiWithHeaders<T = unknown>(path: string, {method = "GET", body, signal, auth = true}: RequestOptions = {}): Promise<{data: T, headers: Headers}> {
    // The public demo answers from sample data instead of the network
    if (inDemo()) {
        // Loaded on demand, so the sample data isn't part of the real app's bundle
        const {demoRequest} = await import("@/lib/demo/api");
        const demo = demoRequest(method, path);
        if (demo.status >= 400) throw new ApiError(demo.status, demo.message ?? "Not available in the demo", "DEMO.READ_ONLY");
        return {data: demo.data as T, headers: new Headers(demo.headers)};
    }

    const send = () => fetch(`${API_URL}${path}`, {
        method,
        signal,
        credentials: "include",
        headers: {
            ...(body !== undefined && {"Content-Type": "application/json"}),
            ...(auth && accessToken && {Authorization: `Bearer ${accessToken}`}),
        },
        body: body === undefined ? undefined : JSON.stringify(body),
    });

    let response = await send();

    if (response.status === 401 && auth) {
        if (await refreshSession()) {
            response = await send();
        } else {
            sessionListeners.forEach((listener) => listener(false));
        }
    }

    const data = await parse(response);
    if (!response.ok) throw toError(response, data);
    return {data: data as T, headers: response.headers};
}

export async function api<T = unknown>(path: string, options?: RequestOptions): Promise<T> {
    return (await apiWithHeaders<T>(path, options)).data;
}

// Mutating endpoints and /auth/* wrap their result as {info, data}; GETs return the bare DTO
export async function apiData<T>(path: string, options?: RequestOptions): Promise<T> {
    const body = await api<{data: T}>(path, options);
    return body.data;
}

// Sign-in verify endpoints: store the access token from the issued pair
export function setSession(pair: TokenPair) {
    accessToken = pair.AccessTokenStr;
}
