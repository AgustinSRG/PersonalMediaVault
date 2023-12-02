// About API

"use strict";

import { CommonAuthenticatedErrorHandler, RequestErrorHandler, RequestParams } from "@asanrom/request-browser";
import { getApiURL } from "@/utils/api";

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
        url: getApiURL("/api/about"),
        handleError: (err, handler) => {
            new RequestErrorHandler()
                .add(401, "*", handler.unauthorized)
                .add(500, "*", "serverError" in handler ? handler.serverError : handler.temporalError)
                .add("*", "*", "networkError" in handler ? handler.networkError : handler.temporalError)
                .handle(err);
        },
    };
}
