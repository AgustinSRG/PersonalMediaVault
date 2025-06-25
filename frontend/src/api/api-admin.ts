// Admin API

"use strict";

import type { CommonAuthenticatedErrorHandler, RequestParams } from "@asanrom/request-browser";
import { RequestErrorHandler } from "@asanrom/request-browser";
import { API_PREFIX, getApiURL } from "@/utils/api";
import type { ProvidedAuthConfirmation } from "./api-auth";

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
     * Error: Invalid user password
     */
    invalidUserPassword: () => void;

    /**
     * Error: Username in use
     */
    usernameInUse: () => void;

    /**
     * Error: Generic bad request
     */
    badRequest: () => void;

    /**
     * Required auth confirmation (two factor authentication)
     */
    requiredAuthConfirmationTfa: () => void;

    /**
     * Invalid two factor authentication code
     */
    invalidTfaCode: () => void;

    /**
     * Required auth confirmation (password)
     */
    requiredAuthConfirmationPassword: () => void;

    /**
     * Invalid password
     */
    invalidPassword: () => void;

    /**
     * When you fail a confirmation, there is a cooldown of 5 seconds.
     */
    cooldown: () => void;
};

/**
 * Creates vault account
 * @param username Username
 * @param password Password
 * @param write True to give the account write permissions
 * @param providedAuthConfirmation Auth confirmation
 * @returns The request parameters
 */
export function apiAdminCreateAccount(
    username: string,
    password: string,
    write: boolean,
    providedAuthConfirmation: ProvidedAuthConfirmation,
): RequestParams<void, CreateAccountErrorHandler> {
    return {
        method: "POST",
        url: getApiURL(`${API_PREFIX}${API_GROUP_PREFIX}/accounts`),
        json: {
            username: username,
            password: password,
            write: write,
        },
        headers: {
            "x-auth-confirmation-pw": providedAuthConfirmation.password || "",
            "x-auth-confirmation-tfa": providedAuthConfirmation.tfaCode || "",
        },
        handleError: (err, handler) => {
            new RequestErrorHandler()
                .add(401, "*", handler.unauthorized)
                .add(400, "USERNAME_INVALID", handler.invalidUsername)
                .add(400, "PASSWORD_INVALID", handler.invalidUserPassword)
                .add(400, "USERNAME_IN_USE", handler.usernameInUse)
                .add(400, "*", handler.badRequest)
                .add(403, "AUTH_CONFIRMATION_REQUIRED_TFA", handler.requiredAuthConfirmationTfa)
                .add(403, "INVALID_TFA_CODE", handler.invalidTfaCode)
                .add(403, "AUTH_CONFIRMATION_REQUIRED_PW", handler.requiredAuthConfirmationPassword)
                .add(403, "INVALID_PASSWORD", handler.invalidPassword)
                .add(403, "COOLDOWN", handler.cooldown)
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

    /**
     * Required auth confirmation (two factor authentication)
     */
    requiredAuthConfirmationTfa: () => void;

    /**
     * Invalid two factor authentication code
     */
    invalidTfaCode: () => void;

    /**
     * Required auth confirmation (password)
     */
    requiredAuthConfirmationPassword: () => void;

    /**
     * Invalid password
     */
    invalidPassword: () => void;

    /**
     * When you fail a confirmation, there is a cooldown of 5 seconds.
     */
    cooldown: () => void;
};

/**
 * Deletes vault account
 * @param username Account username
 * @param newUsername Account new username
 * @param write True to give the account write permissions
 * @param providedAuthConfirmation Auth confirmation
 * @returns The request parameters
 */
export function apiAdminUpdateAccount(
    username: string,
    newUsername: string,
    write: boolean,
    providedAuthConfirmation: ProvidedAuthConfirmation,
): RequestParams<void, UpdateAccountErrorHandler> {
    return {
        method: "POST",
        url: getApiURL(`${API_PREFIX}${API_GROUP_PREFIX}/accounts/update`),
        json: {
            username: username,
            newUsername: newUsername,
            write: write,
        },
        headers: {
            "x-auth-confirmation-pw": providedAuthConfirmation.password || "",
            "x-auth-confirmation-tfa": providedAuthConfirmation.tfaCode || "",
        },
        handleError: (err, handler) => {
            new RequestErrorHandler()
                .add(401, "*", handler.unauthorized)
                .add(400, "USERNAME_INVALID", handler.invalidUsername)
                .add(400, "USERNAME_IN_USE", handler.usernameInUse)
                .add(400, "*", handler.badRequest)
                .add(403, "AUTH_CONFIRMATION_REQUIRED_TFA", handler.requiredAuthConfirmationTfa)
                .add(403, "INVALID_TFA_CODE", handler.invalidTfaCode)
                .add(403, "AUTH_CONFIRMATION_REQUIRED_PW", handler.requiredAuthConfirmationPassword)
                .add(403, "INVALID_PASSWORD", handler.invalidPassword)
                .add(403, "COOLDOWN", handler.cooldown)
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

    /**
     * Required auth confirmation (two factor authentication)
     */
    requiredAuthConfirmationTfa: () => void;

    /**
     * Invalid two factor authentication code
     */
    invalidTfaCode: () => void;

    /**
     * Required auth confirmation (password)
     */
    requiredAuthConfirmationPassword: () => void;

    /**
     * Invalid password
     */
    invalidPassword: () => void;

    /**
     * When you fail a confirmation, there is a cooldown of 5 seconds.
     */
    cooldown: () => void;
};

/**
 * Deletes vault account
 * @param username Account username
 * @param providedAuthConfirmation Auth confirmation
 * @returns The request parameters
 */
export function apiAdminDeleteAccount(
    username: string,
    providedAuthConfirmation: ProvidedAuthConfirmation,
): RequestParams<void, DeleteAccountErrorHandler> {
    return {
        method: "POST",
        url: getApiURL(`${API_PREFIX}${API_GROUP_PREFIX}/accounts/delete`),
        json: {
            username: username,
        },
        headers: {
            "x-auth-confirmation-pw": providedAuthConfirmation.password || "",
            "x-auth-confirmation-tfa": providedAuthConfirmation.tfaCode || "",
        },
        handleError: (err, handler) => {
            new RequestErrorHandler()
                .add(401, "*", handler.unauthorized)
                .add(403, "AUTH_CONFIRMATION_REQUIRED_TFA", handler.requiredAuthConfirmationTfa)
                .add(403, "INVALID_TFA_CODE", handler.invalidTfaCode)
                .add(403, "AUTH_CONFIRMATION_REQUIRED_PW", handler.requiredAuthConfirmationPassword)
                .add(403, "INVALID_PASSWORD", handler.invalidPassword)
                .add(403, "COOLDOWN", handler.cooldown)
                .add(403, "*", handler.accessDenied)
                .add(404, "*", handler.accountNotFound)
                .add(500, "*", "serverError" in handler ? handler.serverError : handler.temporalError)
                .add("*", "*", "networkError" in handler ? handler.networkError : handler.temporalError)
                .handle(err);
        },
    };
}
