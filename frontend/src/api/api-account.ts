// Account API

import { GetApiURL, RequestParams } from "@/utils/request";

export class AccountAPI {
    public static GetUsername(): RequestParams<{ username: string, root: boolean, write: boolean, title: string, css: string }> {
        return {
            method: "GET",
            url: GetApiURL("/api/account/username"),
        };
    }

    public static ChangeUsername(username: string, password: string): RequestParams<void> {
        return {
            method: "POST",
            url: GetApiURL("/api/account/username"),
            json: {
                username: username,
                password: password,
            },
        };
    }

    public static ChangePassword(password: string, newPassword: string): RequestParams<void> {
        return {
            method: "POST",
            url: GetApiURL("/api/account/password"),
            json: {
                old_password: password,
                password: newPassword,
            },
        };
    }
}
