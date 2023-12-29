// Authentication API

"use strict";

import { CommonAuthenticatedErrorHandler, CommonErrorHandler, RequestErrorHandler, RequestParams } from "@asanrom/request-browser";
import { API_PREFIX, getApiURL } from "@/utils/api";

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
export type LoginErrorHandler = CommonErrorHandler & {
    invalidCredentials: () => void;
    wrongCredentials: () => void;
    cooldown: () => void;
};

/**
 * Session duration
 */
export type SessionDuration = "day" | "week" | "month" | "year";

/**
 * API call: Login
 * @param username Username
 * @param password Password
 * @param duration Session duration
 * @returns The request parameters
 */
export function apiAuthLogin(
    username: string,
    password: string,
    duration?: SessionDuration,
): RequestParams<LoginResult, LoginErrorHandler> {
    return {
        method: "POST",
        url: getApiURL(`${API_PREFIX}${API_GROUP_PREFIX}/login`),
        json: {
            username: username,
            password: password,
            duration: duration,
        },
        handleError: (err, handler) => {
            new RequestErrorHandler()
                .add(400, "*", handler.invalidCredentials)
                .add(403, "COOLDOWN", handler.cooldown)
                .add(403, "*", handler.wrongCredentials)
                .add(500, "*", "serverError" in handler ? handler.serverError : handler.temporalError)
                .add("*", "*", "networkError" in handler ? handler.networkError : handler.temporalError)
                .handle(err);
        },
    };
}

/**
 * Closes the current session
 * @returns The request parameters
 */
export function apiAuthLogout(): RequestParams<void, CommonAuthenticatedErrorHandler> {
    return {
        method: "POST",
        url: getApiURL(`${API_PREFIX}${API_GROUP_PREFIX}/logout`),
        handleError: (err, handler) => {
            new RequestErrorHandler()
                .add(401, "*", handler.unauthorized)
                .add(500, "*", "serverError" in handler ? handler.serverError : handler.temporalError)
                .add("*", "*", "networkError" in handler ? handler.networkError : handler.temporalError)
                .handle(err);
        },
    };
}
