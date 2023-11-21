// Account API

"use strict";

import { RequestParams } from "./request";
import { CommonAuthenticatedErrorHandler, RequestErrorHandler } from "./request-error";
import { API_PREFIX, getApiURL } from "./utils";

const API_GROUP_PREFIX = "/account";

/**
 * Account context
 */
export interface AccountContext {
    /**
     * Current username
     */
    username: string;

    /**
     * True if the user has root access to the vault
     */
    root: boolean;

    /**
     * True if the user has write access to the vault
     */
    write: boolean;

    /**
     * Custom title
     */
    title: string;

    /**
     * Custom stylesheet
     */
    css: string;

    /**
     * Current PMV version
     */
    version: string;
}

/**
 * Gets account context
 * @returns The request object
 */
export function apiAccountGetContext(): RequestParams<AccountContext, CommonAuthenticatedErrorHandler> {
    return {
        method: "GET",
        url: getApiURL(`${API_PREFIX}${API_GROUP_PREFIX}`),
        handleError: (err, handler) => {
            new RequestErrorHandler()
                .add(401, "*", handler.unauthorized)
                .add(500, "*", "serverError" in handler ? handler.serverError : handler.temporalError)
                .add("*", "*", "networkError" in handler ? handler.networkError : handler.temporalError)
                .handle(err);
        },
    };
}

/**
 * Error handler for username change
 */
export type ChangeUsernameErrorHandler = CommonAuthenticatedErrorHandler & {
    /**
     * Handler for error: Username in use
     */
    usernameInUse: () => void;

    /**
     * Handler for error: Invalid username
     */
    invalidUsername: () => void;

    /**
     * Handler for error: Invalid password
     */
    invalidPassword: () => void;
};

/**
 * Changes account username
 * @param username New username
 * @param password Account password
 * @returns The request object
 */
export function apiAccountChangeUsername(username: string, password: string): RequestParams<void, ChangeUsernameErrorHandler> {
    return {
        method: "POST",
        url: getApiURL(`${API_PREFIX}${API_GROUP_PREFIX}/username`),
        json: {
            username: username,
            password: password,
        },
        handleError: (err, handler) => {
            new RequestErrorHandler()
                .add(401, "*", handler.unauthorized)
                .add(400, "USERNAME_IN_USE", handler.usernameInUse)
                .add(400, "*", handler.invalidUsername)
                .add(403, "*", handler.invalidPassword)
                .add(500, "*", "serverError" in handler ? handler.serverError : handler.temporalError)
                .add("*", "*", "networkError" in handler ? handler.networkError : handler.temporalError)
                .handle(err);
        },
    };
}

/**
 * Error handler for password change
 */
export type ChangePasswordErrorHandler = CommonAuthenticatedErrorHandler & {
    /**
     * Handler for error: Invalid password
     */
    invalidPassword: () => void;

    /**
     * Handler for error: Invalid new password
     */
    invalidNewPassword: () => void;
};

/**
 * Changes account password
 * @param password The old password
 * @param newPassword The new password
 * @returns The request object
 */
export function apiAccountChangePassword(password: string, newPassword: string): RequestParams<void, ChangePasswordErrorHandler> {
    return {
        method: "POST",
        url: getApiURL(`${API_PREFIX}${API_GROUP_PREFIX}/password`),
        json: {
            old_password: password,
            password: newPassword,
        },
        handleError: (err, handler) => {
            new RequestErrorHandler()
                .add(401, "*", handler.unauthorized)
                .add(400, "*", handler.invalidNewPassword)
                .add(403, "*", handler.invalidPassword)
                .add(500, "*", "serverError" in handler ? handler.serverError : handler.temporalError)
                .add("*", "*", "networkError" in handler ? handler.networkError : handler.temporalError)
                .handle(err);
        },
    };
}
