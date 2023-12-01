// Config API

"use strict";

import { CommonAuthenticatedErrorHandler, RequestErrorHandler, RequestParams } from "@asanrom/request-browser";
import { VaultUserConfig } from "./models";
import { API_PREFIX, getApiURL } from "@/utils/api";

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
};

/**
 * Sets vault configuration
 * @param config The configuration
 * @returns The request parameters
 */
export function apiConfigSetConfig(config: VaultUserConfig): RequestParams<void, SetConfigErrorHandler> {
    return {
        method: "POST",
        url: getApiURL(`${API_PREFIX}${API_GROUP_PREFIX}`),
        json: config,
        handleError: (err, handler) => {
            new RequestErrorHandler()
                .add(401, "*", handler.unauthorized)
                .add(400, "*", handler.badRequest)
                .add(403, "*", handler.accessDenied)
                .add(500, "*", "serverError" in handler ? handler.serverError : handler.temporalError)
                .add("*", "*", "networkError" in handler ? handler.networkError : handler.temporalError)
                .handle(err);
        },
    };
}
