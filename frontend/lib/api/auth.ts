import {api, apiData, ApiError, refreshSession, setAccessToken, setSession} from "@/lib/api/client";

export interface RegisterInput {
    name: string;
    username: string;
    email: string;
    password: string;
}

export const register = (input: RegisterInput) =>
    api("/auth/register", {method: "POST", body: input, auth: false});

export async function verifyRegistration(email: string, code: string) {
    setSession(await apiData("/auth/register/verify", {method: "POST", body: {email, code}, auth: false}));
}

export const login = (email: string, password: string) =>
    api("/auth/login", {method: "POST", body: {email, password}, auth: false});

export async function verifyLogin(email: string, code: string) {
    setSession(await apiData("/auth/login/verify", {method: "POST", body: {email, code}, auth: false}));
}

export const resendCode = (purpose: "register" | "reset_password", email: string) =>
    api(`/auth/${purpose}/resend-code`, {method: "POST", body: {email}, auth: false});

export const requestPasswordReset = (email: string) =>
    api("/auth/reset-password/request", {method: "POST", body: {email}, auth: false});

export const confirmPasswordReset = (email: string, code: string, newPassword: string) =>
    api("/auth/reset-password/confirm", {method: "POST", body: {email, code, new_password: newPassword}, auth: false});

export async function logout() {
    try {
        await api("/auth/logout", {method: "POST", auth: false});
    } catch (error) {
        // Already signed out on the server (expired or missing cookie) is the same outcome
        if (!(error instanceof ApiError)) throw error;
    }
    setAccessToken(null);
}

// Restores a session from the rid cookie, e.g. after a page reload
export const restoreSession = refreshSession;
