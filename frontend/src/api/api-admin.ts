// Admin API

"use strict";

import { CommonAuthenticatedErrorHandler, RequestErrorHandler, RequestParams } from "@asanrom/request-browser";
import { API_PREFIX, getApiURL } from "@/utils/api";

const API_GROUP_PREFIX = "/admin";

/**
 * Vault extra account
 */
export interface VaultAccount {
    /**
     * Account username
     */
    username: string;

    /**
     * True if the account has write permissions
     */
    write: boolean;
}

/**
 * Error handler for admins APIs
 */
export type AdminApiErrorHandler = CommonAuthenticatedErrorHandler & {
    /**
     * Error: Access denied
     */
    accessDenied: () => void;
};

/**
 * List vault accounts
 * @returns The request parameters
 */
export function apiAdminListAccounts(): RequestParams<VaultAccount[], AdminApiErrorHandler> {
    return {
        method: "GET",
        url: getApiURL(`${API_PREFIX}${API_GROUP_PREFIX}/accounts`),
        handleError: (err, handler) => {
            new RequestErrorHandler()
                .add(401, "*", handler.unauthorized)
                .add(403, "*", handler.accessDenied)
                .add(500, "*", "serverError" in handler ? handler.serverError : handler.temporalError)
                .add("*", "*", "networkError" in handler ? handler.networkError : handler.temporalError)
                .handle(err);
        },
    };
}

/**
 * Error handler for account creation API
 */
export type CreateAccountErrorHandler = AdminApiErrorHandler & {
    /**
     * Error: Invalid username
     */
    invalidUsername: () => void;

    /**
     * Error: Invalid password
     */
    invalidPassword: () => void;

    /**
     * Error: Username in use
     */
    usernameInUse: () => void;

    /**
     * Error: Generic bad request
     */
    badRequest: () => void;
};

/**
 * Creates vault account
 * @param username Username
 * @param password Password
 * @param write True to give the account write permissions
 * @returns The request parameters
 */
export function apiAdminCreateAccount(username: string, password: string, write: boolean): RequestParams<void, CreateAccountErrorHandler> {
    return {
        method: "POST",
        url: getApiURL(`${API_PREFIX}${API_GROUP_PREFIX}/accounts`),
        json: {
            username: username,
            password: password,
            write: write,
        },
        handleError: (err, handler) => {
            new RequestErrorHandler()
                .add(401, "*", handler.unauthorized)
                .add(400, "USERNAME_INVALID", handler.invalidUsername)
                .add(400, "PASSWORD_INVALID", handler.invalidPassword)
                .add(400, "USERNAME_IN_USE", handler.usernameInUse)
                .add(400, "*", handler.badRequest)
                .add(403, "*", handler.accessDenied)
                .add(500, "*", "serverError" in handler ? handler.serverError : handler.temporalError)
                .add("*", "*", "networkError" in handler ? handler.networkError : handler.temporalError)
                .handle(err);
        },
    };
}

/**
 * Error handler for update account API
 */
export type UpdateAccountErrorHandler = AdminApiErrorHandler & {
    /**
     * Error: Invalid username
     */
    invalidUsername: () => void;

    /**
     * Error: Username in use
     */
    usernameInUse: () => void;

    /**
     * Error: Generic bad request
     */
    badRequest: () => void;

    /**
     * Error: Account not found
     */
    accountNotFound: () => void;
};

/**
 * Deletes vault account
 * @param username Account username
 * @param newUsername Account new username
 * @param write True to give the account write permissions
 * @returns The request parameters
 */
export function apiAdminUpdateAccount(
    username: string,
    newUsername: string,
    write: boolean,
): RequestParams<void, UpdateAccountErrorHandler> {
    return {
        method: "POST",
        url: getApiURL(`${API_PREFIX}${API_GROUP_PREFIX}/accounts/update`),
        json: {
            username: username,
            newUsername: newUsername,
            write: write,
        },
        handleError: (err, handler) => {
            new RequestErrorHandler()
                .add(401, "*", handler.unauthorized)
                .add(400, "USERNAME_INVALID", handler.invalidUsername)
                .add(400, "USERNAME_IN_USE", handler.usernameInUse)
                .add(400, "*", handler.badRequest)
                .add(403, "*", handler.accessDenied)
                .add(404, "*", handler.accountNotFound)
                .add(500, "*", "serverError" in handler ? handler.serverError : handler.temporalError)
                .add("*", "*", "networkError" in handler ? handler.networkError : handler.temporalError)
                .handle(err);
        },
    };
}

/**
 * Error handler for delete account API
 */
export type DeleteAccountErrorHandler = AdminApiErrorHandler & {
    /**
     * Error: Account not found
     */
    accountNotFound: () => void;
};

/**
 * Deletes vault account
 * @param username Account username
 * @returns The request parameters
 */
export function apiAdminDeleteAccount(username: string): RequestParams<void, DeleteAccountErrorHandler> {
    return {
        method: "POST",
        url: getApiURL(`${API_PREFIX}${API_GROUP_PREFIX}/accounts/delete`),
        json: {
            username: username,
        },
        handleError: (err, handler) => {
            new RequestErrorHandler()
                .add(401, "*", handler.unauthorized)
                .add(403, "*", handler.accessDenied)
                .add(404, "*", handler.accountNotFound)
                .add(500, "*", "serverError" in handler ? handler.serverError : handler.temporalError)
                .add("*", "*", "networkError" in handler ? handler.networkError : handler.temporalError)
                .handle(err);
        },
    };
}
