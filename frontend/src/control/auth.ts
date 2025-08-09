// Auth controller

"use strict";

import { apiAccountGetContext } from "@/api/api-account";
import { apiAuthLogout } from "@/api/api-auth";
import { setNamedTimeout, clearNamedTimeout } from "@/utils/named-timeouts";
import { AppEvents } from "./app-events";
import { fetchFromLocalStorage, saveIntoLocalStorage } from "../utils/local-storage";
import { setAssetsSessionCookie } from "@/utils/cookie";
import { abortNamedApiRequest, addRequestAuthenticationHandler, makeApiRequest, makeNamedApiRequest } from "@asanrom/request-browser";

const REQUEST_KEY = "auth-control-check";
const REQUEST_KEY_SILENT = "auth-control-check-silent";

/**
 * Event triggered when the auth status starts or stops loading
 */
export const EVENT_NAME_AUTH_LOADING = "auth-status-loading";

/**
 * Event triggered when the authentication status changes
 */
export const EVENT_NAME_AUTH_CHANGED = "auth-status-changed";

/**
 * Event triggered when the authentication status cannot be loaded due to an error
 */
export const EVENT_NAME_AUTH_ERROR = "auth-status-loading-error";

const LS_KEY_AUTH_TOKEN = "x-session-token";
const LS_KEY_VAULT_FINGERPRINT = "x-vault-fingerprint";

const SESSION_TOKEN_HEADER_NAME = "x-session-token";

/**
 * Event triggered when a new version is detected
 */
export const EVENT_NAME_APP_NEW_VERSION = "app-new-version";

/**
 * Event triggered when a request results in 401 - Unauthorized
 */
export const EVENT_NAME_UNAUTHORIZED = "unauthorized";

/**
 * Authentication status management object
 */
export class AuthController {
    /**
     * True if the vault is locked
     */
    public static Locked = true;

    /**
     * Session ID
     */
    public static Session = "";

    /**
     * Username
     */
    public static Username = "";

    /**
     * Vault fingerprint
     */
    public static Fingerprint = "";

    /**
     * True if loading
     */
    public static Loading = true;

    /**
     * True if the user has root permission on the vault
     */
    public static IsRoot = false;

    /**
     * True if the user can make changes to the vault
     */
    public static CanWrite = false;

    /**
     * Custom title of the vault
     */
    public static Title = "";

    /**
     * Custom logo text of the vault
     */
    public static Logo = "";

    /**
     * Custom CSS of the vault
     */
    public static CSS = "";

    /**
     * Initialization logic
     * Runs at app startup
     */
    public static Initialize() {
        addRequestAuthenticationHandler(() => {
            AuthController.RefreshAuthStatus();

            const authHeaders = Object.create(null);
            authHeaders[SESSION_TOKEN_HEADER_NAME] = AuthController.Session;

            return authHeaders;
        });

        AppEvents.AddEventListener(EVENT_NAME_UNAUTHORIZED, AuthController.OnUnauthorized);

        AuthController.LoadAuthStatus();
        AuthController.SetAssetsCookie();
        AuthController.CheckAuthStatus();
    }

    /**
     * Loads authentication status from local storage
     */
    public static LoadAuthStatus() {
        AuthController.Session = fetchFromLocalStorage(LS_KEY_AUTH_TOKEN, "");
        AuthController.Fingerprint = fetchFromLocalStorage(LS_KEY_VAULT_FINGERPRINT, "");
    }

    /**
     * Sets assets cookie for media to load
     */
    public static SetAssetsCookie() {
        const cookieName = "st-" + AuthController.Fingerprint;
        setAssetsSessionCookie(cookieName, AuthController.Session);
    }

    /**
     * Refreshes auth status from the local storage
     * @returns True if the status was not synced, false if the status was synced
     */
    public static RefreshAuthStatus(): boolean {
        const storedSession = fetchFromLocalStorage(LS_KEY_AUTH_TOKEN, "");

        if (storedSession !== AuthController.Session) {
            AuthController.LoadAuthStatus();
            AuthController.SetAssetsCookie();
            AuthController.CheckAuthStatus();
            return true;
        } else {
            return false;
        }
    }

    /**
     * Checks auth status
     */
    public static CheckAuthStatus() {
        AuthController.Loading = true;
        AppEvents.Emit(EVENT_NAME_AUTH_LOADING, true);
        clearNamedTimeout(REQUEST_KEY);
        makeNamedApiRequest(REQUEST_KEY, apiAccountGetContext())
            .onSuccess((response) => {
                AuthController.Locked = false;
                AuthController.IsRoot = response.root;
                AuthController.CanWrite = response.write;
                AuthController.Username = response.username;
                AuthController.Title = response.title;
                AuthController.Logo = response.logo;
                AuthController.CSS = response.css;
                if (import.meta.env.VITE__VERSION !== response.version) {
                    AppEvents.Emit(EVENT_NAME_APP_NEW_VERSION);
                }
                AppEvents.Emit(EVENT_NAME_AUTH_CHANGED, AuthController.Locked, AuthController.Username);
                AuthController.Loading = false;
                AppEvents.Emit(EVENT_NAME_AUTH_LOADING, false);
                AuthController.UpdateCustomStyle();
            })
            .onRequestError((err, handleErr) => {
                handleErr(err, {
                    unauthorized: () => {
                        AuthController.Locked = true;
                        AuthController.Username = "";
                        AppEvents.Emit(EVENT_NAME_AUTH_CHANGED, AuthController.Locked, AuthController.Username);
                        AuthController.Loading = false;
                        AppEvents.Emit(EVENT_NAME_AUTH_LOADING, false);
                    },
                    temporalError: () => {
                        // Retry
                        AppEvents.Emit(EVENT_NAME_AUTH_ERROR);
                        setNamedTimeout(REQUEST_KEY, 1500, AuthController.CheckAuthStatus);
                    },
                });
            })
            .onUnexpectedError((err) => {
                console.error(err);
                // We assume the credentials are invalid
                AuthController.Locked = true;
                AuthController.Username = "";
                AppEvents.Emit(EVENT_NAME_AUTH_CHANGED, AuthController.Locked, AuthController.Username);
                AuthController.Loading = false;
                AppEvents.Emit(EVENT_NAME_AUTH_LOADING, false);
            });
    }

    public static CheckingAuthSilent = false;

    /**
     * Checks auth status silently (no loading events)
     */
    public static CheckAuthStatusSilent() {
        if (AuthController.CheckingAuthSilent) {
            return;
        }

        clearNamedTimeout(REQUEST_KEY_SILENT);

        if (AuthController.Loading) {
            abortNamedApiRequest(REQUEST_KEY_SILENT);
            AuthController.CheckingAuthSilent = false;
        }

        AuthController.CheckingAuthSilent = true;

        makeNamedApiRequest(REQUEST_KEY_SILENT, apiAccountGetContext())
            .onSuccess(() => {
                AuthController.CheckingAuthSilent = false;
            })
            .onRequestError((err, handleErr) => {
                handleErr(err, {
                    unauthorized: () => {
                        AuthController.CheckingAuthSilent = false;
                        AuthController.CheckAuthStatus();
                    },
                    temporalError: () => {
                        // Retry
                        setNamedTimeout(REQUEST_KEY_SILENT, 1500, AuthController.CheckAuthStatusSilent);
                    },
                });
            })
            .onUnexpectedError((err) => {
                console.error(err);
                AuthController.CheckingAuthSilent = false;
                AuthController.CheckAuthStatus();
            });
    }

    /**
     * Updates local status username
     * @param username New username
     */
    public static UpdateUsername(username: string) {
        AuthController.Username = username;
        AppEvents.Emit(EVENT_NAME_AUTH_CHANGED, AuthController.Locked, AuthController.Username);
    }

    /**
     * Sets session
     * @param session Session ID
     * @param fingerprint Vault fingerprint
     */
    public static SetSession(session: string, fingerprint: string) {
        AuthController.Locked = true;
        AuthController.Session = session;
        saveIntoLocalStorage(LS_KEY_AUTH_TOKEN, session);
        AuthController.Fingerprint = fingerprint;
        saveIntoLocalStorage(LS_KEY_VAULT_FINGERPRINT, fingerprint);
        AuthController.SetAssetsCookie();
        AuthController.Username = "";
        AppEvents.Emit(EVENT_NAME_AUTH_CHANGED, AuthController.Locked, AuthController.Username);
        AuthController.CheckAuthStatus();
    }

    /**
     * Logs out
     */
    public static Logout() {
        const currentSession = AuthController.Session;
        makeApiRequest(apiAuthLogout())
            .onSuccess(() => {
                if (AuthController.Session === currentSession) {
                    AuthController.ClearSession();
                }
            })
            .onRequestError(() => {
                if (AuthController.Session === currentSession) {
                    AuthController.ClearSession();
                }
            });
    }

    /**
     * Clears current session to after logging out
     */
    public static OnUnauthorized() {
        if (AuthController.Locked) {
            return;
        }
        AuthController.ClearSession();
    }

    /**
     * Clears current session to after logging out
     */
    public static ClearSession() {
        AuthController.Locked = true;
        AuthController.Session = "";
        saveIntoLocalStorage(LS_KEY_AUTH_TOKEN, "");
        AuthController.Username = "";
        AuthController.SetAssetsCookie();
        AppEvents.Emit(EVENT_NAME_AUTH_CHANGED, AuthController.Locked, AuthController.Username);
    }

    /**
     * Updates custom style
     */
    private static UpdateCustomStyle() {
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
        styleElement.appendChild(document.createTextNode(AuthController.CSS));

        head.appendChild(styleElement);
    }
}
