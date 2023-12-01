// Tags API

"use strict";

import { CommonAuthenticatedErrorHandler, RequestErrorHandler, RequestParams } from "@asanrom/request-browser";
import { MediaTag } from "./models";
import { getApiURL } from "@/utils/api";

/**
 * Gets list of tags
 * @returns The request parameters
 */
export function apiTagsGetTags(): RequestParams<MediaTag[], CommonAuthenticatedErrorHandler> {
    return {
        method: "GET",
        url: getApiURL("/api/tags"),
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
 * Error handler for tag media API
 */
export type TagMediaErrorHandler = CommonAuthenticatedErrorHandler & {
    /**
     * Error: Invalid tag name
     */
    invalidTagName: () => void;

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
 * Tags media
 * @param media Media ID
 * @param tagName Tag name
 * @returns The request parameters
 */
export function apiTagsTagMedia(media: number, tagName: string): RequestParams<MediaTag, TagMediaErrorHandler> {
    return {
        method: "POST",
        url: getApiURL("/api/tags/add"),
        json: {
            media_id: media,
            tag_name: tagName,
        },
        handleError: (err, handler) => {
            new RequestErrorHandler()
                .add(401, "*", handler.unauthorized)
                .add(400, "INVALID_NAME", handler.invalidTagName)
                .add(400, "*", handler.badRequest)
                .add(403, "*", handler.accessDenied)
                .add(500, "*", "serverError" in handler ? handler.serverError : handler.temporalError)
                .add("*", "*", "networkError" in handler ? handler.networkError : handler.temporalError)
                .handle(err);
        },
    };
}

/**
 * Response for media untag API
 */
export interface UntagMediaResponse {
    /**
     * True if the tag was removed
     */
    removed: boolean;
}

/**
 * Error handler for untag media API
 */
export type UntagMediaErrorHandler = CommonAuthenticatedErrorHandler & {
    /**
     * Error: Access denied
     */
    accessDenied: () => void;
};

/**
 * Removes a tag
 * @param media Media ID
 * @param tagId Tag ID
 * @returns The request parameters
 */
export function apiTagsUntagMedia(media: number, tagId: number): RequestParams<UntagMediaResponse, UntagMediaErrorHandler> {
    return {
        method: "POST",
        url: getApiURL("/api/tags/remove"),
        json: {
            media_id: media,
            tag_id: tagId,
        },
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
