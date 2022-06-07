// Authentication API

import { GetAPIURL, RequestParams } from "@/utils/request";

export class AuthAPI {
    public static Login(username: string, password: string): RequestParams {
        return {
            method: "POST",
            url: GetAPIURL("/api/auth/login"),
            json: {
                username: username,
                password: password,
            },
        };
    }

    public static Logout(): RequestParams {
        return {
            method: "POST",
            url: GetAPIURL("/api/auth/logout"),
        };
    }
}
