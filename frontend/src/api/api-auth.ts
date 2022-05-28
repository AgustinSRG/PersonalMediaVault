// Authentication API

import { RequestParams } from "@/utils/request";

export class AuthAPI {
    public static Login(username: string, password: string): RequestParams {
        return {
            method: "POST",
            url: "/api/auth/login",
            json: {
                username: username,
                password: password,
            },
        };
    }

    public static Logout(): RequestParams {
        return {
            method: "POST",
            url: "/api/auth/logout",
        };
    }
}
