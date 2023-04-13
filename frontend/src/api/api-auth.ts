// Authentication API

import { GetApiURL, RequestParams } from "@/utils/request";

export class AuthAPI {
    public static Login(username: string, password: string): RequestParams {
        return {
            method: "POST",
            url: GetApiURL("/api/auth/login"),
            json: {
                username: username,
                password: password,
            },
        };
    }

    public static Logout(): RequestParams {
        return {
            method: "POST",
            url: GetApiURL("/api/auth/logout"),
        };
    }
}
