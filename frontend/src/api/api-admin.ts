// Admin API

"use strict";

import { RequestParams } from "./request";
import { getApiURL } from "./utils";

export class AdminAPI {
    public static ListAccounts(): RequestParams<{ username: string; write: boolean }[]> {
        return {
            method: "GET",
            url: getApiURL("/api/admin/accounts"),
        };
    }

    public static CreateAccount(username: string, password: string, write: boolean): RequestParams<void> {
        return {
            method: "POST",
            url: getApiURL("/api/admin/accounts"),
            json: {
                username: username,
                password: password,
                write: write,
            },
        };
    }

    public static DeleteAccount(username: string): RequestParams<void> {
        return {
            method: "POST",
            url: getApiURL("/api/admin/accounts/delete"),
            json: {
                username: username,
            },
        };
    }
}
