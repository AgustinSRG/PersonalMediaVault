// Auth controller

"use strict";

import { AccountAPI } from "@/api/api-account";
import { AuthAPI } from "@/api/api-auth";
import { Request } from "@/utils/request";
import { Timeouts } from "@/utils/timeout";
import { AppEvents } from "./app-events";
import { LocalStorage } from "./local-storage";
import { setAssetsSessionCookie } from "@/utils/cookie";

const EVENT_NAME_LOADING = "auth-status-loading";
const EVENT_NAME_CHANGED = "auth-status-changed";
const EVENT_NAME_ERROR = "auth-status-loading-error";

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
     * Custom CSS of the vault
     */
    public static CSS = "";

    /**
     * Initialization logic
     * Runs at app startup
     */
    public static Initialize() {
        AppEvents.AddEventListener("unauthorized", AuthController.ClearSession);

        AuthController.LoadAuthStatus();
        AuthController.SetAssetsCookie();
        AuthController.CheckAuthStatus();
    }

    /**
     * Loads authentication status from local storage
     */
    public static LoadAuthStatus() {
        AuthController.Session = LocalStorage.Get("x-session-token", "");
        AuthController.Fingerprint = LocalStorage.Get("x-vault-fingerprint", "");
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
        const storedSession = LocalStorage.Get("x-session-token", "");

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
        AppEvents.Emit(EVENT_NAME_LOADING, true);
        Timeouts.Abort("auth-control-check");
        Request.Pending("auth-control-check", AccountAPI.GetContext())
            .onSuccess((response) => {
                AuthController.Locked = false;
                AuthController.IsRoot = response.root;
                AuthController.CanWrite = response.write;
                AuthController.Username = response.username;
                AuthController.Title = response.title;
                AuthController.CSS = response.css;
                if (import.meta.env.VITE__VERSION !== response.version) {
                    AppEvents.Emit("app-new-version");
                }
                AppEvents.Emit(EVENT_NAME_CHANGED, AuthController.Locked, AuthController.Username);
                AuthController.Loading = false;
                AppEvents.Emit(EVENT_NAME_LOADING, false);
                AuthController.UpdateCustomStyle();
            })
            .onRequestError((err) => {
                Request.ErrorHandler()
                    .add(401, "*", () => {
                        AuthController.Locked = true;
                        AuthController.Username = "";
                        AppEvents.Emit(EVENT_NAME_CHANGED, AuthController.Locked, AuthController.Username);
                        AuthController.Loading = false;
                        AppEvents.Emit(EVENT_NAME_LOADING, false);
                    })
                    .add("*", "*", () => {
                        // Retry
                        AppEvents.Emit(EVENT_NAME_ERROR);
                        Timeouts.Set("auth-control-check", 1500, AuthController.CheckAuthStatus);
                    })
                    .handle(err);
            })
            .onUnexpectedError((err) => {
                console.error(err);
                // We assume the credentials are invalid
                AuthController.Locked = true;
                AuthController.Username = "";
                AppEvents.Emit(EVENT_NAME_CHANGED, AuthController.Locked, AuthController.Username);
                AuthController.Loading = false;
                AppEvents.Emit(EVENT_NAME_LOADING, false);
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

        Timeouts.Abort("auth-control-check-silent");

        if (AuthController.Loading) {
            Request.Abort("auth-control-check-silent");
            AuthController.CheckingAuthSilent = false;
        }

        AuthController.CheckingAuthSilent = true;

        Request.Pending("auth-control-check-silent", AccountAPI.GetContext())
            .onSuccess(() => {
                AuthController.CheckingAuthSilent = false;
            })
            .onRequestError((err) => {
                Request.ErrorHandler()
                    .add(401, "*", () => {
                        AuthController.CheckingAuthSilent = false;
                        AuthController.CheckAuthStatus();
                    })
                    .add("*", "*", () => {
                        // Retry
                        Timeouts.Set("auth-control-check-silent", 1500, AuthController.CheckAuthStatusSilent);
                    })
                    .handle(err);
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
        AppEvents.Emit(EVENT_NAME_CHANGED, AuthController.Locked, AuthController.Username);
    }

    /**
     * Sets session
     * @param session Session ID
     * @param fingerprint Vault fingerprint
     */
    public static SetSession(session: string, fingerprint: string) {
        AuthController.Locked = true;
        AuthController.Session = session;
        LocalStorage.Set("x-session-token", session);
        AuthController.Fingerprint = fingerprint;
        LocalStorage.Set("x-vault-fingerprint", fingerprint);
        AuthController.SetAssetsCookie();
        AuthController.Username = "";
        AppEvents.Emit(EVENT_NAME_CHANGED, AuthController.Locked, AuthController.Username);
        AuthController.CheckAuthStatus();
    }

    /**
     * Logs out
     */
    public static Logout() {
        const currentSession = AuthController.Session;
        Request.Do(AuthAPI.Logout())
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
    public static ClearSession() {
        AuthController.Locked = true;
        AuthController.Session = "";
        LocalStorage.Set("x-session-token", "");
        AuthController.Username = "";
        AuthController.SetAssetsCookie();
        AppEvents.Emit(EVENT_NAME_CHANGED, AuthController.Locked, AuthController.Username);
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

    /**
     * Adds event listener to check for loading status updates
     * @param handler Event handler
     */
    public static AddLoadingEventListener(handler: (loading: boolean) => void) {
        AppEvents.AddEventListener(EVENT_NAME_LOADING, handler);
    }

    /**
     * Removes event listener (Loading event)
     * @param handler Event handler
     */
    public static RemoveLoadingEventListener(handler: (loading: boolean) => void) {
        AppEvents.RemoveEventListener(EVENT_NAME_LOADING, handler);
    }

    /**
     * Adds event listener to check for auth status changes
     * @param handler Event handler
     */
    public static AddChangeEventListener(handler: (locked: boolean, username: string) => void) {
        AppEvents.AddEventListener(EVENT_NAME_CHANGED, handler);
    }

    /**
     * Removes event listener (Changed event)
     * @param handler Event handler
     */
    public static RemoveChangeEventListener(handler: (locked: boolean, username: string) => void) {
        AppEvents.RemoveEventListener(EVENT_NAME_CHANGED, handler);
    }

    /**
     * Adds event listener to check for loading errors
     * @param handler Event handler
     */
    public static AddErrorEventListener(handler: () => void) {
        AppEvents.AddEventListener(EVENT_NAME_ERROR, handler);
    }

    /**
     * Removes event listener (Error event)
     * @param handler Event handler
     */
    public static RemoveErrorEventListener(handler: () => void) {
        AppEvents.RemoveEventListener(EVENT_NAME_ERROR, handler);
    }
}
