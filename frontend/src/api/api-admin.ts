// Admin API

import { GetAPIURL, RequestParams } from "@/utils/request";

export class AdminAPI {
    public static ListAccounts(): RequestParams {
        return {
            method: "GET",
            url: GetAPIURL("/api/admin/accounts"),
        };
    }

    public static CreateAccount(username: string, password: string, write: boolean): RequestParams {
        return {
            method: "POST",
            url: GetAPIURL("/api/admin/accounts"),
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
            url: GetAPIURL("/api/admin/accounts/delete"),
            json: {
                username: username,
            },
        };
    }
}
