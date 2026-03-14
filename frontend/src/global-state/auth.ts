// Authentication global state

"use strict";

import { apiAccountGetContext } from "@/api/api-account";
import { apiAuthLogout } from "@/api/api-auth";
import { setNamedTimeout, clearNamedTimeout } from "@/utils/named-timeouts";
import {
    addAppEventListener,
    emitAppEvent,
    EVENT_NAME_APP_NEW_VERSION,
    EVENT_NAME_AUTH_CHANGED,
    EVENT_NAME_AUTH_ERROR,
    EVENT_NAME_AUTH_LOADING,
    EVENT_NAME_UNAUTHORIZED,
} from "./app-events";
import { fetchFromLocalStorage, saveIntoLocalStorage } from "@/local-storage/local-storage";
import { setAssetsSessionCookie } from "@/utils/cookie";
import { abortNamedApiRequest, addRequestAuthenticationHandler, makeApiRequest, makeNamedApiRequest } from "@asanrom/request-browser";
import { getUniqueStringId } from "@/utils/unique-id";
import { LOAD_RETRY_DELAY, SESSION_TOKEN_HEADER_NAME } from "@/constants";

// Local storage key to store the auth token
const LS_KEY_AUTH_TOKEN = "x-session-token";

// Local storage key to store the vault fingerprint
const LS_KEY_VAULT_FINGERPRINT = "x-vault-fingerprint";

/**
 * Authentication status type
 */
export type AuthStatus = {
    /**
     * True if the vault is locked
     */
    locked: boolean;

    /**
     * True if loading the auth state
     */
    loading: boolean;

    /**
     * Current session token
     */
    session: string;

    /**
     * Current username
     */
    username: string;

    /**
     * Vault fingerprint
     */
    fingerprint: string;

    /**
     * True if the user has root permissions
     */
    isRoot: boolean;

    /**
     * True if the user has write permissions
     */
    canWrite: boolean;

    /**
     * Custom vault title
     */
    title: string;

    /**
     * Custom logo
     */
    logo: string;

    /**
     * Custom CSS
     */
    css: string;

    /**
     * True if semantic search is available
     */
    semanticSearchAvailable: boolean;

    /**
     * True if the authentication status is being silently checked
     */
    checkingSilent: boolean;
};

/**
 * Authentication state
 */
const AuthState: AuthStatus = {
    locked: true,
    loading: true,
    session: "",
    username: "",
    fingerprint: "",
    isRoot: false,
    canWrite: false,
    title: "",
    logo: "",
    css: "",
    semanticSearchAvailable: false,
    checkingSilent: false,
};

/**
 * Gets the current auth status
 * @returns The current auth status
 */
export function getAuthStatus(): Readonly<AuthStatus> {
    return AuthState;
}

/**
 * Checks if the auth status is being loaded
 * @returns True if the status is being loaded
 */
export function isLoadingAuthStatus(): boolean {
    return AuthState.loading;
}

/**
 * Checks if the vault is locked
 * @returns True if the vault is locked
 */
export function isVaultLocked(): boolean {
    return AuthState.locked;
}

/**
 * Loads session token and fingerprint from
 * the local storage
 */
function loadSessionTokenAndFingerprint() {
    AuthState.session = fetchFromLocalStorage(LS_KEY_AUTH_TOKEN, "");
    AuthState.fingerprint = fetchFromLocalStorage(LS_KEY_VAULT_FINGERPRINT, "");
}

/**
 * Sets the cookie for the images/videos
 * that do not support headers
 */
function setAssetsCookie() {
    const cookieName = "st-" + AuthState.fingerprint;
    setAssetsSessionCookie(cookieName, AuthState.session);
}

/**
 * Refreshes authentication status from the local storage
 * @returns True if the status was not synced, false if the status was synced
 */
export function refreshAuthenticationStatus(): boolean {
    const storedSession = fetchFromLocalStorage(LS_KEY_AUTH_TOKEN, "");

    if (storedSession !== AuthState.session) {
        loadSessionTokenAndFingerprint();
        setAssetsCookie();
        checkAuthenticationStatus();
        return true;
    } else {
        return false;
    }
}

/**
 * Sets the authentication loading state
 * @param loading The loading state
 */
function setAuthenticationLoading(loading: boolean) {
    AuthState.loading = loading;
    emitAppEvent(EVENT_NAME_AUTH_LOADING, loading);
}

/**
 * Updates custom style for the page
 */
function updateCustomStyle() {
    const head = document.head || document.getElementsByTagName("head")[0];

    if (!head) {
        return;
    }

    let styleElement: any = document.querySelector("#custom-style-pmv");

    if (styleElement) {
        styleElement.remove();
    }

    styleElement = document.createElement("style");

    styleElement.id = "custom-style-pmv";
    styleElement.type = "text/css";
    styleElement.appendChild(document.createTextNode(AuthState.css));

    head.appendChild(styleElement);
}

// Request ID for loading the auth context
const REQUEST_KEY = getUniqueStringId();

/**
 * Loads the authentication status from the server
 */
export function checkAuthenticationStatus() {
    setAuthenticationLoading(true);

    clearNamedTimeout(REQUEST_KEY);

    makeNamedApiRequest(REQUEST_KEY, apiAccountGetContext())
        .onSuccess((response) => {
            AuthState.locked = false;
            AuthState.isRoot = response.root;
            AuthState.canWrite = response.write;
            AuthState.username = response.username;
            AuthState.title = response.title;
            AuthState.logo = response.logo;
            AuthState.css = response.css;
            AuthState.semanticSearchAvailable = response.semanticSearch || false;

            if (import.meta.env.VITE__VERSION !== response.version) {
                emitAppEvent(EVENT_NAME_APP_NEW_VERSION);
            }

            emitAppEvent(EVENT_NAME_AUTH_CHANGED, AuthState);

            setAuthenticationLoading(false);

            updateCustomStyle();
        })
        .onRequestError((err, handleErr) => {
            handleErr(err, {
                unauthorized: () => {
                    AuthState.locked = true;
                    AuthState.username = "";

                    emitAppEvent(EVENT_NAME_AUTH_CHANGED, AuthState);

                    setAuthenticationLoading(false);
                },
                temporalError: () => {
                    // Retry
                    emitAppEvent(EVENT_NAME_AUTH_ERROR);
                    setNamedTimeout(REQUEST_KEY, LOAD_RETRY_DELAY, checkAuthenticationStatus);
                },
            });
        })
        .onUnexpectedError((err) => {
            console.error(err);
            // We assume the credentials are invalid
            AuthState.locked = true;
            AuthState.username = "";

            emitAppEvent(EVENT_NAME_AUTH_CHANGED, AuthState);

            setAuthenticationLoading(false);
        });
}

// Request ID for silently loading the context
const REQUEST_KEY_SILENT = getUniqueStringId();

/**
 * Checks auth status silently (no loading events)
 */
export function checkAuthenticationStatusSilent() {
    if (AuthState.checkingSilent) {
        return;
    }

    clearNamedTimeout(REQUEST_KEY_SILENT);

    if (AuthState.loading) {
        abortNamedApiRequest(REQUEST_KEY_SILENT);
        AuthState.checkingSilent = false;
    }

    AuthState.checkingSilent = true;

    makeNamedApiRequest(REQUEST_KEY_SILENT, apiAccountGetContext())
        .onSuccess(() => {
            AuthState.checkingSilent = false;
        })
        .onRequestError((err, handleErr) => {
            handleErr(err, {
                unauthorized: () => {
                    AuthState.checkingSilent = false;
                    checkAuthenticationStatus();
                },
                temporalError: () => {
                    // Retry
                    setNamedTimeout(REQUEST_KEY_SILENT, LOAD_RETRY_DELAY, checkAuthenticationStatusSilent);
                },
            });
        })
        .onUnexpectedError((err) => {
            console.error(err);
            AuthState.checkingSilent = false;
            checkAuthenticationStatus();
        });
}

/**
 * Handles new authenticated session.
 * Call when the user logs into the vault.
 * @param session The session token
 * @param fingerprint The vault fingerprint
 */
export function handleAuthenticatedNewSession(session: string, fingerprint: string) {
    AuthState.locked = true;

    AuthState.session = session;
    saveIntoLocalStorage(LS_KEY_AUTH_TOKEN, session);

    AuthState.fingerprint = fingerprint;
    saveIntoLocalStorage(LS_KEY_VAULT_FINGERPRINT, fingerprint);

    setAssetsCookie();

    AuthState.username = "";

    emitAppEvent(EVENT_NAME_AUTH_CHANGED, AuthState);

    checkAuthenticationStatus();
}

/**
 * Updates currently authenticated username.
 * Call when the user updates their username.
 * @param username New username
 */
export function updateAuthenticatedUsername(username: string) {
    AuthState.username = username;
    emitAppEvent(EVENT_NAME_AUTH_CHANGED, AuthState);
}

/**
 * Clears current session token after logging out
 */
function clearSession() {
    AuthState.locked = true;
    AuthState.session = "";

    saveIntoLocalStorage(LS_KEY_AUTH_TOKEN, "");

    AuthState.username = "";

    setAssetsCookie();

    emitAppEvent(EVENT_NAME_AUTH_CHANGED, AuthState);
}

/**
 * Closes current session in order to log out
 */
export function closeCurrentAuthenticatedSession() {
    const currentSession = AuthState.session;
    makeApiRequest(apiAuthLogout())
        .onSuccess(() => {
            if (AuthState.session === currentSession) {
                clearSession();
            }
        })
        .onRequestError(() => {
            if (AuthState.session === currentSession) {
                clearSession();
            }
        });
}
/**
 * Initializes the authentication state
 */
export function initializeAuthentication() {
    // Add the session header to all API requests
    addRequestAuthenticationHandler(() => {
        refreshAuthenticationStatus();

        const authHeaders = Object.create(null);
        authHeaders[SESSION_TOKEN_HEADER_NAME] = AuthState.session;

        return authHeaders;
    });

    // Handle unauthorized signal, indicating the token is no longer valid
    addAppEventListener(EVENT_NAME_UNAUTHORIZED, () => {
        if (AuthState.locked) {
            return;
        }

        clearSession();
    });

    loadSessionTokenAndFingerprint();
    setAssetsCookie();
    checkAuthenticationStatus();
}
