// Config API

"use strict";

import { CommonAuthenticatedErrorHandler, RequestErrorHandler, RequestParams } from "@asanrom/request-browser";
import { VaultUserConfig } from "./models";
import { API_PREFIX, getApiURL } from "@/utils/api";
import { ProvidedAuthConfirmation } from "./api-auth";

const API_GROUP_PREFIX = "/config";

/**
 * Gets vault configuration
 * @returns The request parameters
 */
export function apiConfigGetConfig(): RequestParams<VaultUserConfig, CommonAuthenticatedErrorHandler> {
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
 * Error handler for set config API
 */
export type SetConfigErrorHandler = CommonAuthenticatedErrorHandler & {
    /**
     * Error: Bad request
     */
    badRequest: () => void;

    /**
     * Error: Access denied
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
 * Sets vault configuration
 * @param config The configuration
 * @param providedAuthConfirmation Auth confirmation
 * @returns The request parameters
 */
export function apiConfigSetConfig(
    config: VaultUserConfig,
    providedAuthConfirmation: ProvidedAuthConfirmation,
): RequestParams<void, SetConfigErrorHandler> {
    return {
        method: "POST",
        url: getApiURL(`${API_PREFIX}${API_GROUP_PREFIX}`),
        json: config,
        headers: {
            "x-auth-confirmation-pw": providedAuthConfirmation.password || "",
            "x-auth-confirmation-tfa": providedAuthConfirmation.tfaCode || "",
        },
        handleError: (err, handler) => {
            new RequestErrorHandler()
                .add(401, "*", handler.unauthorized)
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
