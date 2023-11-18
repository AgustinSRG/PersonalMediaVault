// Authentication API

"use strict";

import { Request } from "./request";
import { CommonErrorHandler, RequestErrorHandler } from "./request-error";
import { API_PREFIX, getApiURL } from "./utils";

const API_GROUP_PREFIX = "/auth";

/**
 * Login result
 */
export interface LoginResult {
    /**
     * Session id / token
     */
    session_id: string;

    /**
     * Vault fingerprint
     */
    vault_fingerprint: string;
}

/**
 * Error handler for login
 */
export interface LoginErrorHandler extends CommonErrorHandler {
    invalidCredentials: () => void;
    wrongCredentials: () => void;
    cooldown: () => void;
}

/**
 * API call: Login
 * @param username Username
 * @param password Password
 * @returns The request object
 */
export function apiAuthLogin(username: string, password: string): Request<LoginResult, LoginErrorHandler> {
    return new Request({
        method: "POST",
        url: getApiURL(`${API_PREFIX}${API_GROUP_PREFIX}/login`),
        json: {
            username: username,
            password: password,
        },
        handleError: (err, handler) => {
            new RequestErrorHandler()
                .add(400, "*", handler.invalidCredentials)
                .add(403, "COOLDOWN", handler.cooldown)
                .add(403, "*", handler.wrongCredentials)
                .add(500, "*", handler.serverError)
                .add("*", "*", handler.networkError)
                .handle(err);
        },
    });
}

export function apiAuthLogout(): Request<void, CommonErrorHandler> {
    return new Request({
        method: "POST",
        url: getApiURL(`${API_PREFIX}${API_GROUP_PREFIX}/logout`),
        handleError: (err, handler) => {
            new RequestErrorHandler()
                .add(500, "*", handler.serverError)
                .add("*", "*", handler.networkError)
                .handle(err);
        },
    });
}
