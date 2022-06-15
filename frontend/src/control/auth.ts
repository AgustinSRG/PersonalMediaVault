// Auth controller

import { AccountAPI } from "@/api/api-account";
import { AuthAPI } from "@/api/api-auth";
import { getCookie, setCookie } from "@/utils/cookie";
import { Request } from "@/utils/request";
import { Timeouts } from "@/utils/timeout";
import { AppEvents } from "./app-events";

export class AuthController {
    public static Locked = true;
    public static Session = "";
    public static Username = "";
    public static Loading = true;

    public static Initialize() {
        AuthController.Session = getCookie("x-session-token");
        AuthController.CheckAuthStatus();
        AppEvents.AddEventListener("unauthorized", AuthController.ClearSession);
    }

    public static CheckAuthStatus() {
        AuthController.Loading = true;
        AppEvents.Emit("auth-status-loading", true);
        Timeouts.Abort("auth-control-check");
        Request.Pending("auth-control-check", AccountAPI.GetUsername()).onSuccess(response => {
            AuthController.Locked = false;
            AuthController.Username = response.username;
            AppEvents.Emit("auth-status-changed", AuthController.Locked, AuthController.Username);
            AuthController.Loading = false;
            AppEvents.Emit("auth-status-loading", false);

        }).onRequestError(err => {
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
        }).onUnexpectedError(err => {
            console.error(err);
            // We assume the credentials are invalid
            AuthController.Locked = true;
            AuthController.Username = "";
            AppEvents.Emit("auth-status-changed", AuthController.Locked, AuthController.Username);
            AuthController.Loading = false;
            AppEvents.Emit("auth-status-loading", false);
        });
    }

    public static UpdateUsername(username: string) {
        AuthController.Username = username;
        AppEvents.Emit("auth-status-changed", AuthController.Locked, AuthController.Username);
    }

    public static SetSession(session: string) {
        AuthController.Locked = true;
        AuthController.Session = session;
        setCookie("x-session-token", session);
        AuthController.Username = "";
        AppEvents.Emit("auth-status-changed", AuthController.Locked, AuthController.Username);
        AuthController.CheckAuthStatus();
    }

    public static Logout() {
        const currentSession = AuthController.Session;
        Request.Do(AuthAPI.Logout()).onSuccess(() => {
            if (AuthController.Session === currentSession) {
                AuthController.ClearSession();
            }
        }).onRequestError(() => {
            if (AuthController.Session === currentSession) {
                AuthController.ClearSession();
            }
        });
    }

    public static ClearSession() {
        AuthController.Locked = true;
        AuthController.Session = "";
        setCookie("x-session-token", "");
        AuthController.Username = "";
        AppEvents.Emit("auth-status-changed", AuthController.Locked, AuthController.Username);
    }
}
