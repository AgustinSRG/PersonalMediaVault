// Account API

import { GetAPIURL, RequestParams } from "@/utils/request";

export class AccountAPI {
    public static GetUsername(): RequestParams {
        return {
            method: "GET",
            url: GetAPIURL("/api/account/username"),
        };
    }

    public static ChangeUsername(username: string, password: string): RequestParams {
        return {
            method: "POST",
            url: GetAPIURL("/api/account/username"),
            json: {
                username: username,
                password: password,
            },
        };
    }

    public static ChangePassword(password: string, newPassword: string): RequestParams {
        return {
            method: "POST",
            url: GetAPIURL("/api/account/username"),
            json: {
                old_password: password,
                password: newPassword,
            },
        };
    }
}
