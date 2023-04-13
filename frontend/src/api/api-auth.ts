// Authentication API

import { GetApiURL, RequestParams } from "@/utils/request";

export class AuthAPI {
    public static Login(username: string, password: string): RequestParams<{session_id: string, vault_fingerprint: string}> {
        return {
            method: "POST",
            url: GetApiURL("/api/auth/login"),
            json: {
                username: username,
                password: password,
            },
        };
    }

    public static Logout(): RequestParams<void> {
        return {
            method: "POST",
            url: GetApiURL("/api/auth/logout"),
        };
    }
}
