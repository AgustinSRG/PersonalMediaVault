// Admin API

import { GetApiURL, RequestParams } from "@/utils/request";

export class AdminAPI {
    public static ListAccounts(): RequestParams<{ username: string, write: boolean }[]> {
        return {
            method: "GET",
            url: GetApiURL("/api/admin/accounts"),
        };
    }

    public static CreateAccount(username: string, password: string, write: boolean): RequestParams {
        return {
            method: "POST",
            url: GetApiURL("/api/admin/accounts"),
            json: {
                username: username,
                password: password,
                write: write,
            },
        };
    }

    public static DeleteAccount(username: string): RequestParams {
        return {
            method: "POST",
            url: GetApiURL("/api/admin/accounts/delete"),
            json: {
                username: username,
            },
        };
    }
}
