// API utils

"use strict";

import type { AppStatusPage } from "@/control/app-status";
import { AuthController } from "@/control/auth";

/**
 * API Prefix
 */
export const API_PREFIX = "/api";

/**
 * Resolves URL
 * @param path The Path
 * @param base The base URL
 * @returns The resolved URL
 */
function resolveUrl(path: string, base: string): string {
    return new URL(path, base).toString();
}

/**
 * Gets API URL from the full path
 * @param path The full path
 * @returns The full URL
 */
export function getApiURL(path: string): string {
    if (import.meta.env.DEV) {
        return resolveUrl(path, import.meta.env.VITE_DEV_TEST_HOST || "http://localhost");
    } else {
        return resolveUrl("." + path, location.protocol + "//" + location.host + location.pathname);
    }
}

/**
 * Gets asset URL from the path
 * @param path The asset full path
 * @returns The full URL
 */
export function getAssetURL(path: string): string {
    if (import.meta.env.DEV) {
        const assetUrl = new URL(resolveUrl(path, import.meta.env.VITE_DEV_TEST_HOST || "http://localhost"));
        assetUrl.searchParams.set("session_token", AuthController.Session);
        return assetUrl.toString();
    } else {
        return resolveUrl("." + path, location.protocol + "//" + location.host + location.pathname);
    }
}

/**
 * Generates a query for a request
 * @param params The query parameters
 * @returns The query string
 */
export function generateURIQuery(params: { [key: string]: any | undefined }): string {
    const keys = Object.keys(params);
    if (keys.length === 0) {
        return "";
    }

    let result = "";

    for (const key of keys) {
        if (params[key] === undefined || params[key] === null) {
            continue;
        }

        if (result !== "") {
            result += "&";
        } else {
            result += "?";
        }

        result += encodeURIComponent(key) + "=" + encodeURIComponent(params[key] + "");
    }

    return result;
}

/**
 * Status params for a frontend URL
 */
export type FrontendUrlStatusParams = {
    // Current media
    media?: number | string;

    // Current album
    album?: number | string;

    // Current page
    page?: AppStatusPage;

    // Home page group
    g?: number;

    // Search tag
    search?: string;

    // Search params
    sp?: string;

    // Random seed
    seed?: number;

    // Split page and media?
    split?: "yes";
};

/**
 * Computes frontend URL from a list of status parameters
 * @param statusParams The list of status parameters
 * @returns The frontend URL
 */
export function getFrontendUrl(statusParams: FrontendUrlStatusParams): string {
    return location.protocol + "//" + location.host + location.pathname + generateURIQuery(statusParams);
}
