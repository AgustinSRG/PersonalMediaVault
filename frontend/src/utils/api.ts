// API utils

"use strict";

/**
 * API Prefix
 */
export const API_PREFIX = "/api";

/**
 * Gets API URL from the full path
 * @param path The full path
 * @returns The full URL
 */
export function getApiURL(path: string): string {
    if (import.meta.env.DEV) {
        return (import.meta.env.DEV_TEST_HOST || "http://localhost") + path;
    } else {
        return location.protocol + "//" + location.host + path;
    }
}

/**
 * Gets asset URL from the path
 * @param path The asset full path
 * @returns The full URL
 */
export function getAssetURL(path: string): string {
    if (import.meta.env.DEV) {
        return (import.meta.env.DEV_TEST_HOST || "http://localhost") + path;
    } else {
        return location.protocol + "//" + location.host + path;
    }
}

/**
 * Generates a query for a request
 * @param params The query parameters
 * @returns The query string
 */
export function generateURIQuery(params: { [key: string]: string }): string {
    const keys = Object.keys(params);
    if (keys.length === 0) {
        return "";
    }

    let result = "";

    for (const key of keys) {
        if (!params[key]) {
            continue;
        }

        if (result !== "") {
            result += "&";
        } else {
            result += "?";
        }

        result += encodeURIComponent(key) + "=" + encodeURIComponent(params[key]);
    }

    return result;
}
