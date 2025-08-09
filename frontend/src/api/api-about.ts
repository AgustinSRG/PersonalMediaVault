// About API

"use strict";

import type { CommonAuthenticatedErrorHandler, RequestParams } from "@asanrom/request-browser";
import { RequestErrorHandler } from "@asanrom/request-browser";
import { getApiURL, API_PREFIX } from "@/utils/api";

/**
 * Response of about API
 */
export interface AboutResponse {
    /**
     * Current server version
     */
    version: string;

    /**
     * Last available release
     */
    last_release: string;

    /**
     * Version of FFmpeg being used
     */
    ffmpeg_version: string;
}

/**
 * Gets version information about the server
 * @returns The request parameters
 */
export function apiAbout(): RequestParams<AboutResponse, CommonAuthenticatedErrorHandler> {
    return {
        method: "GET",
        url: getApiURL(`${API_PREFIX}/about`),
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
 * Response of the disk usage API
 */
export interface DiskUsageResponse {
    /**
     * Usage (percent)
     */
    usage: number;

    /**
     * Available bytes
     */
    available: number;

    /**
     * Free bytes
     */
    free: number;

    /**
     * Total bytes
     */
    total: number;
}

/**
 * Gets disk usage of the server
 * @returns The request parameters
 */
export function apiDiskUsage(): RequestParams<DiskUsageResponse, CommonAuthenticatedErrorHandler> {
    return {
        method: "GET",
        url: getApiURL(`${API_PREFIX}/about/disk_usage`),
        handleError: (err, handler) => {
            new RequestErrorHandler()
                .add(401, "*", handler.unauthorized)
                .add(500, "*", "serverError" in handler ? handler.serverError : handler.temporalError)
                .add("*", "*", "networkError" in handler ? handler.networkError : handler.temporalError)
                .handle(err);
        },
    };
}
