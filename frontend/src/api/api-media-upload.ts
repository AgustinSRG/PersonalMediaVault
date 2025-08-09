// Upload media API

"use strict";

import type { CommonAuthenticatedErrorHandler, RequestParams } from "@asanrom/request-browser";
import { RequestErrorHandler } from "@asanrom/request-browser";
import { API_PREFIX, getApiURL } from "@/utils/api";

/**
 * Upload API response
 */
export interface UploadResponse {
    /**
     * Created media ID
     */
    media_id: number;
}

/**
 * Error handler for all media APIs
 */
export type UploadApiErrorHandler = CommonAuthenticatedErrorHandler & {
    /**
     * Error: Invalid media file
     */
    invalidMediaFile: () => void;

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
 * Uploads media file
 * @param title The title
 * @param file The media file
 * @param album The album to include it into or -1 for no album
 * @returns The request parameters
 */
export function apiUploadMedia(title: string, file: File, album: number): RequestParams<UploadResponse, UploadApiErrorHandler> {
    const form = new FormData();
    form.append("file", file);

    return {
        method: "POST",
        url: getApiURL(`${API_PREFIX}/upload?title=${encodeURIComponent(title)}&album=${encodeURIComponent(album + "")}`),
        form: form,
        handleError: (err, handler) => {
            new RequestErrorHandler()
                .add(401, "*", handler.unauthorized)
                .add(400, "INVALID_MEDIA", handler.invalidMediaFile)
                .add(400, "*", handler.badRequest)
                .add(403, "*", handler.accessDenied)
                .add(500, "*", "serverError" in handler ? handler.serverError : handler.temporalError)
                .add("*", "*", "networkError" in handler ? handler.networkError : handler.temporalError)
                .handle(err);
        },
    };
}
