// Auth controller

import { AccountAPI } from "@/api/api-account";
import { AuthAPI } from "@/api/api-auth";
import { Request } from "@/utils/request";
import { Timeouts } from "@/utils/timeout";
import { AppEvents } from "./app-events";
import { LocalStorage } from "./local-storage";
import { setAssetsSessionCookie } from "@/utils/cookie";

export class AuthController {
    public static Locked = true;
    public static Session = "";
    public static Username = "";
    public static Fingerprint = "";
    public static Loading = true;

    public static IsRoot = false;
    public static CanWrite = false;

    public static Title = "";
    public static CSS = "";

    public static Initialize() {
        AppEvents.AddEventListener("unauthorized", AuthController.ClearSession);

        AuthController.LoadAuthStatus();
        AuthController.SetAssetsCookie();
        AuthController.CheckAuthStatus();
    }

    public static LoadAuthStatus() {
        AuthController.Session = LocalStorage.Get("x-session-token", "");
        AuthController.Fingerprint = LocalStorage.Get("x-vault-fingerprint", "");
    }

    public static SetAssetsCookie() {
        const cookieName = "st-" + AuthController.Fingerprint;
        setAssetsSessionCookie(cookieName, AuthController.Session);
    }

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

    public static CheckAuthStatus() {
        AuthController.Loading = true;
        AppEvents.Emit("auth-status-loading", true);
        Timeouts.Abort("auth-control-check");
        Request.Pending("auth-control-check", AccountAPI.GetContext())
            .onSuccess((response) => {
                AuthController.Locked = false;
                AuthController.IsRoot = response.root;
                AuthController.CanWrite = response.write;
                AuthController.Username = response.username;
                AuthController.Title = response.title;
                AuthController.CSS = response.css;
                AppEvents.Emit("auth-status-changed", AuthController.Locked, AuthController.Username);
                AuthController.Loading = false;
                AppEvents.Emit("auth-status-loading", false);
                AuthController.UpdateCustomStyle();
            })
            .onRequestError((err) => {
                Request.ErrorHandler()
                    .add(401, "*", () => {
                        AuthController.Locked = true;
                        AuthController.Username = "";
                        AppEvents.Emit("auth-status-changed", AuthController.Locked, AuthController.Username);
                        AuthController.Loading = false;
                        AppEvents.Emit("auth-status-loading", false);
                    })
                    .add("*", "*", () => {
                        // Retry
                        Timeouts.Set("auth-control-check", 1500, AuthController.CheckAuthStatus);
                    })
                    .handle(err);
            })
            .onUnexpectedError((err) => {
                console.error(err);
                // We assume the credentials are invalid
                AuthController.Locked = true;
                AuthController.Username = "";
                AppEvents.Emit("auth-status-changed", AuthController.Locked, AuthController.Username);
                AuthController.Loading = false;
                AppEvents.Emit("auth-status-loading", false);
            });
    }

    public static CheckingAuthSilent = false;

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

    public static UpdateUsername(username: string) {
        AuthController.Username = username;
        AppEvents.Emit("auth-status-changed", AuthController.Locked, AuthController.Username);
    }

    public static SetSession(session: string, fingerprint: string) {
        AuthController.Locked = true;
        AuthController.Session = session;
        LocalStorage.Set("x-session-token", session);
        AuthController.Fingerprint = fingerprint;
        LocalStorage.Set("x-vault-fingerprint", fingerprint);
        AuthController.SetAssetsCookie();
        AuthController.Username = "";
        AppEvents.Emit("auth-status-changed", AuthController.Locked, AuthController.Username);
        AuthController.CheckAuthStatus();
    }

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

    public static ClearSession() {
        AuthController.Locked = true;
        AuthController.Session = "";
        LocalStorage.Set("x-session-token", "");
        AuthController.Username = "";
        AuthController.SetAssetsCookie();
        AppEvents.Emit("auth-status-changed", AuthController.Locked, AuthController.Username);
    }

    public static UpdateCustomStyle() {
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
