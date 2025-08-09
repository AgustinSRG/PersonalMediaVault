// Account API

"use strict";

import type { CommonAuthenticatedErrorHandler, RequestParams } from "@asanrom/request-browser";
import { RequestErrorHandler } from "@asanrom/request-browser";
import { API_PREFIX, generateURIQuery, getApiURL } from "@/utils/api";
import type { ProvidedAuthConfirmation } from "./api-auth";

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
     * Custom logotype text
     */
    logo: string;

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
 * @returns The request parameters
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

    /**
     * Required auth confirmation (two factor authentication)
     */
    requiredAuthConfirmationTfa: () => void;

    /**
     * Invalid two factor authentication code
     */
    invalidTfaCode: () => void;

    /**
     * When you fail a confirmation, there is a cooldown of 5 seconds.
     */
    cooldown: () => void;
};

/**
 * Changes account username
 * @param username New username
 * @param password Account password
 * @param providedAuthConfirmation Auth confirmation
 * @returns The request parameters
 */
export function apiAccountChangeUsername(
    username: string,
    password: string,
    providedAuthConfirmation: ProvidedAuthConfirmation,
): RequestParams<void, ChangeUsernameErrorHandler> {
    return {
        method: "POST",
        url: getApiURL(`${API_PREFIX}${API_GROUP_PREFIX}/username`),
        json: {
            username: username,
            password: password,
        },
        headers: {
            "x-auth-confirmation-pw": providedAuthConfirmation.password || "",
            "x-auth-confirmation-tfa": providedAuthConfirmation.tfaCode || "",
        },
        handleError: (err, handler) => {
            new RequestErrorHandler()
                .add(401, "*", handler.unauthorized)
                .add(400, "USERNAME_IN_USE", handler.usernameInUse)
                .add(400, "*", handler.invalidUsername)
                .add(403, "AUTH_CONFIRMATION_REQUIRED_TFA", handler.requiredAuthConfirmationTfa)
                .add(403, "INVALID_TFA_CODE", handler.invalidTfaCode)
                .add(403, "COOLDOWN", handler.cooldown)
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

    /**
     * Required auth confirmation (two factor authentication)
     */
    requiredAuthConfirmationTfa: () => void;

    /**
     * Invalid two factor authentication code
     */
    invalidTfaCode: () => void;

    /**
     * When you fail a confirmation, there is a cooldown of 5 seconds.
     */
    cooldown: () => void;
};

/**
 * Changes account password
 * @param password The old password
 * @param newPassword The new password
 * @param providedAuthConfirmation Auth confirmation
 * @returns The request parameters
 */
export function apiAccountChangePassword(
    password: string,
    newPassword: string,
    providedAuthConfirmation: ProvidedAuthConfirmation,
): RequestParams<void, ChangePasswordErrorHandler> {
    return {
        method: "POST",
        url: getApiURL(`${API_PREFIX}${API_GROUP_PREFIX}/password`),
        json: {
            old_password: password,
            password: newPassword,
        },
        headers: {
            "x-auth-confirmation-pw": providedAuthConfirmation.password || "",
            "x-auth-confirmation-tfa": providedAuthConfirmation.tfaCode || "",
        },
        handleError: (err, handler) => {
            new RequestErrorHandler()
                .add(401, "*", handler.unauthorized)
                .add(400, "*", handler.invalidNewPassword)
                .add(403, "AUTH_CONFIRMATION_REQUIRED_TFA", handler.requiredAuthConfirmationTfa)
                .add(403, "INVALID_TFA_CODE", handler.invalidTfaCode)
                .add(403, "COOLDOWN", handler.cooldown)
                .add(403, "*", handler.invalidPassword)
                .add(500, "*", "serverError" in handler ? handler.serverError : handler.temporalError)
                .add("*", "*", "networkError" in handler ? handler.networkError : handler.temporalError)
                .handle(err);
        },
    };
}

/**
 * Auth confirmation method
 */
export type AuthConfirmationMethod = "tfa" | "pw";

/**
 * Security settings of an account
 */
export interface AccountSecuritySettings {
    /**
     * True if two factor authentication is enabled
     */
    tfa: boolean;

    /**
     * Two factor authentication method
     */
    tfaMethod?: string;

    /**
     * True if auth confirmation is enabled
     */
    authConfirmation: boolean;

    /**
     * Preferred method for auth confirmation
     */
    authConfirmationMethod: AuthConfirmationMethod;

    /**
     * Auth confirmation period (seconds)
     */
    authConfirmationPeriodSeconds: number;
}

/**
 * Error handler for security settings API
 */
export type AccountSecurityErrorHandler = CommonAuthenticatedErrorHandler & {
    /**
     * Access denied to the API
     */
    accessDenied: () => void;
};

/**
 * Gets account security settings
 * @returns The request parameters
 */
export function apiAccountGetSecuritySettings(): RequestParams<AccountSecuritySettings, AccountSecurityErrorHandler> {
    return {
        method: "GET",
        url: getApiURL(`${API_PREFIX}${API_GROUP_PREFIX}/security`),
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
 * Error handler for security settings set API
 */
export type AccountSecuritySetSettingsErrorHandler = CommonAuthenticatedErrorHandler & {
    /**
     * Invalid settings
     */
    invalidSettings: () => void;

    /**
     * Access denied to the API
     */
    accessDenied: () => void;

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
 * Sets account security settings
 * @param authConfirmation True if auth confirmation enabled
 * @param authConfirmationMethod Auth confirmation method
 * @param authConfirmationPeriodSeconds Auth confirmation period (seconds)
 * @param providedAuthConfirmation Auth confirmation
 * @returns The request parameters
 */
export function apiAccountSetSecuritySettings(
    authConfirmation: boolean,
    authConfirmationMethod: AuthConfirmationMethod,
    authConfirmationPeriodSeconds: number,
    providedAuthConfirmation: ProvidedAuthConfirmation,
): RequestParams<void, AccountSecuritySetSettingsErrorHandler> {
    return {
        method: "POST",
        url: getApiURL(`${API_PREFIX}${API_GROUP_PREFIX}/security`),
        json: {
            authConfirmation,
            authConfirmationMethod,
            authConfirmationPeriodSeconds,
        },
        headers: {
            "x-auth-confirmation-pw": providedAuthConfirmation.password || "",
            "x-auth-confirmation-tfa": providedAuthConfirmation.tfaCode || "",
        },
        handleError: (err, handler) => {
            new RequestErrorHandler()
                .add(401, "*", handler.unauthorized)
                .add(400, "*", handler.invalidSettings)
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
 * HMAC algorithm for TOTP
 */
export type TimeOtpAlgorithm = "sha1" | "sha256" | "sha512";

/**
 * TOTP period
 */
export type TimeOtpPeriod = "30" | "60" | "120";

/**
 * TOTP settings
 */
export interface TimeOtpSettings {
    /**
     * Issuer
     */
    issuer: string;

    /**
     * Account
     */
    account: string;

    /**
     * Algorithm
     */
    algorithm: TimeOtpAlgorithm;

    /**
     * Period (seconds)
     */
    period: TimeOtpPeriod;

    /**
     * Clock skew
     */
    skew: "allow" | "disallow";
}

/**
 * TOTP result
 */
export interface TimeOtpResult {
    /**
     * Secret
     */
    secret: string;

    /**
     * Method
     */
    method: string;

    /**
     * URL
     */
    url: string;

    /**
     * QR image
     */
    qr: string;
}

/**
 * Error handler for TOTP settings API
 */
export type TimeOtpSettingsErrorHandler = CommonAuthenticatedErrorHandler & {
    /**
     * Invalid settings
     */
    invalidSettings: () => void;

    /**
     * Access denied to the API
     */
    accessDenied: () => void;
};

/**
 * Gets the necessary parameters to enable TOTP for two factor authentication
 * @returns The request parameters
 */
export function apiAccountTimeOtpSettings(settings: TimeOtpSettings): RequestParams<TimeOtpResult, TimeOtpSettingsErrorHandler> {
    return {
        method: "GET",
        url: getApiURL(`${API_PREFIX}${API_GROUP_PREFIX}/security/tfa/totp${generateURIQuery(settings)}`),
        handleError: (err, handler) => {
            new RequestErrorHandler()
                .add(401, "*", handler.unauthorized)
                .add(400, "*", handler.invalidSettings)
                .add(403, "*", handler.accessDenied)
                .add(500, "*", "serverError" in handler ? handler.serverError : handler.temporalError)
                .add("*", "*", "networkError" in handler ? handler.networkError : handler.temporalError)
                .handle(err);
        },
    };
}

/**
 * Parameters to enable two factor authentication with TOTP
 */
export interface TfaEnableTimeOtpBody {
    /**
     * Secret
     */
    secret: string;

    /**
     * Method
     */
    method: string;

    /**
     * Account password
     */
    password: string;

    /**
     * One-time code
     */
    code: string;
}

/**
 * Error handler for TOTP enable API
 */
export type TimeOtpEnableErrorHandler = CommonAuthenticatedErrorHandler & {
    /**
     * Two factor authentication already enabled
     */
    tfaAlreadyEnabled: () => void;

    /**
     * Invalid secret or method
     */
    invalidSecretOrMethod: () => void;

    /**
     * Invalid one-time code
     */
    invalidCode: () => void;

    /**
     * Invalid password
     */
    invalidPassword: () => void;

    /**
     * Access denied to the API
     */
    accessDenied: () => void;
};

/**
 * Enables two factor authentication (TOTP)
 * @returns The request parameters
 */
export function apiAccountTimeOtpEnable(options: TfaEnableTimeOtpBody): RequestParams<void, TimeOtpEnableErrorHandler> {
    return {
        method: "POST",
        url: getApiURL(`${API_PREFIX}${API_GROUP_PREFIX}/security/tfa/totp`),
        json: options,
        handleError: (err, handler) => {
            new RequestErrorHandler()
                .add(401, "*", handler.unauthorized)
                .add(400, "TFA_ENABLED", handler.tfaAlreadyEnabled)
                .add(400, "INVALID_TOTP_CODE", handler.invalidCode)
                .add(400, "*", handler.invalidSecretOrMethod)
                .add(403, "INVALID_PASSWORD", handler.invalidPassword)
                .add(403, "*", handler.accessDenied)
                .add(500, "*", "serverError" in handler ? handler.serverError : handler.temporalError)
                .add("*", "*", "networkError" in handler ? handler.networkError : handler.temporalError)
                .handle(err);
        },
    };
}

/**
 * Error handler for TFA disable API
 */
export type TfaDisableErrorHandler = CommonAuthenticatedErrorHandler & {
    /**
     * Two factor authentication is not enabled
     */
    tfaNotEnabled: () => void;

    /**
     * Invalid one-time code
     */
    invalidCode: () => void;

    /**
     * There is a 5 second cooldown
     */
    cooldown: () => void;

    /**
     * Access denied to the API
     */
    accessDenied: () => void;
};

/**
 * Disables two factor authentication
 * @returns The request parameters
 */
export function apiAccountTfaDisable(code: string): RequestParams<void, TfaDisableErrorHandler> {
    return {
        method: "POST",
        url: getApiURL(`${API_PREFIX}${API_GROUP_PREFIX}/security/tfa/disable`),
        json: {
            code,
        },
        handleError: (err, handler) => {
            new RequestErrorHandler()
                .add(401, "*", handler.unauthorized)
                .add(400, "*", handler.tfaNotEnabled)
                .add(403, "COOLDOWN", handler.cooldown)
                .add(403, "INVALID_CODE", handler.invalidCode)
                .add(403, "*", handler.accessDenied)
                .add(500, "*", "serverError" in handler ? handler.serverError : handler.temporalError)
                .add("*", "*", "networkError" in handler ? handler.networkError : handler.temporalError)
                .handle(err);
        },
    };
}
