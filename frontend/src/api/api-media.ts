// Media API

"use strict";

import { CommonAuthenticatedErrorHandler, RequestErrorHandler, RequestParams } from "@asanrom/request-browser";
import { MediaData, MediaSizeStats } from "./models";
import { API_PREFIX, getApiURL } from "@/utils/api";

const API_GROUP_PREFIX = "/media";

/**
 * Error handler for all media APIs
 */
export type MediaApiErrorHandler = CommonAuthenticatedErrorHandler & {
    /**
     * Error: Not found
     */
    notFound: () => void;
};

/**
 * Get media metadata
 * @param id Media ID
 * @returns The request parameters
 */
export function apiMediaGetMedia(id: number): RequestParams<MediaData, MediaApiErrorHandler> {
    return {
        method: "GET",
        url: getApiURL(`${API_PREFIX}${API_GROUP_PREFIX}/${encodeURIComponent(id + "")}`),
        handleError: (err, handler) => {
            new RequestErrorHandler()
                .add(401, "*", handler.unauthorized)
                .add(404, "*", handler.notFound)
                .add(500, "*", "serverError" in handler ? handler.serverError : handler.temporalError)
                .add("*", "*", "networkError" in handler ? handler.networkError : handler.temporalError)
                .handle(err);
        },
    };
}

/**
 * Gets list of albums the media is in
 * @param id Media ID
 * @returns The request parameters
 */
export function apiMediaGetMediaAlbums(id: number): RequestParams<number[], MediaApiErrorHandler> {
    return {
        method: "GET",
        url: getApiURL(`${API_PREFIX}${API_GROUP_PREFIX}/${encodeURIComponent(id + "")}/albums`),
        handleError: (err, handler) => {
            new RequestErrorHandler()
                .add(401, "*", handler.unauthorized)
                .add(404, "*", handler.notFound)
                .add(500, "*", "serverError" in handler ? handler.serverError : handler.temporalError)
                .add("*", "*", "networkError" in handler ? handler.networkError : handler.temporalError)
                .handle(err);
        },
    };
}

/**
 * Gets the media size stats
 * @param id Media ID
 * @returns The request parameters
 */
export function apiMediaGetMediaSizeStats(id: number): RequestParams<MediaSizeStats, MediaApiErrorHandler> {
    return {
        method: "GET",
        url: getApiURL(`${API_PREFIX}${API_GROUP_PREFIX}/${encodeURIComponent(id + "")}/size_stats`),
        handleError: (err, handler) => {
            new RequestErrorHandler()
                .add(401, "*", handler.unauthorized)
                .add(404, "*", handler.notFound)
                .add(500, "*", "serverError" in handler ? handler.serverError : handler.temporalError)
                .add("*", "*", "networkError" in handler ? handler.networkError : handler.temporalError)
                .handle(err);
        },
    };
}
