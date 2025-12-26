// Semantic search API

"use strict";

import { API_PREFIX, getApiURL } from "@/utils/api";
import type { CommonAuthenticatedErrorHandler, RequestParams } from "@asanrom/request-browser";
import { RequestErrorHandler } from "@asanrom/request-browser";
import type { AdvancedSearchResults } from "./models";

const API_GROUP_PREFIX = "/search/semantic";

/**
 * Semantic search vector type (vector origin)
 */
export type SemanticSearchVectorType = "text" | "image" | "any";

/**
 * Semantic search query parameters
 */
export interface SearchMediaSemanticBody {
    /**
     * The vector to perform the search
     */
    vector: number[];

    /**
     * Vector type filter
     */
    vectorType?: SemanticSearchVectorType;

    /**
     * Max number of results
     * Hard limit: 256
     */
    limit?: number;

    /**
     * Continuation token
     */
    continuationToken?: number;
}

/**
 * Error handler for the semantic search query API
 */
export type SemanticSearchQueryErrorHandler = CommonAuthenticatedErrorHandler & {
    /**
     * Invalid vector size
     */
    invalidVectorSize: () => void;
};

/**
 * Performs a semantic search query
 * @param queryOptions The query options
 * @returns The request parameters
 */
export function apiSemanticSearch(
    queryOptions: SearchMediaSemanticBody,
): RequestParams<AdvancedSearchResults, SemanticSearchQueryErrorHandler> {
    return {
        method: "POST",
        url: getApiURL(`${API_PREFIX}${API_GROUP_PREFIX}`),
        json: queryOptions,
        handleError: (err, handler) => {
            new RequestErrorHandler()
                .add(401, "*", handler.unauthorized)
                .add(400, "*", handler.invalidVectorSize)
                .add(500, "*", "serverError" in handler ? handler.serverError : handler.temporalError)
                .add("*", "*", "networkError" in handler ? handler.networkError : handler.temporalError)
                .handle(err);
        },
    };
}

/**
 * Response for the semantic search encoding APIs
 */
export interface SearchMediaSemanticEncodeResponse {
    /**
     * The encoded vector
     * Can be used to perform a query
     */
    vector: number[];
}

/**
 * Error handler for the semantic search text encoding API
 */
export type SemanticSearchEncodeTextErrorHandler = CommonAuthenticatedErrorHandler & {
    /**
     * Empty text
     */
    emptyText: () => void;

    /**
     * Semantic search service not yet available
     */
    notAvailable: () => void;
};

/**
 * Encodes a text into a vector in order
 * to be used to perform a semantic search query
 * @param text The text to encode
 * @returns The request parameters
 */
export function apiSemanticSearchEncodeText(
    text: string,
): RequestParams<SearchMediaSemanticEncodeResponse, SemanticSearchEncodeTextErrorHandler> {
    return {
        method: "POST",
        url: getApiURL(`${API_PREFIX}${API_GROUP_PREFIX}/encoder/text`),
        json: {
            text,
        },
        handleError: (err, handler) => {
            new RequestErrorHandler()
                .add(401, "*", handler.unauthorized)
                .add(400, "*", handler.emptyText)
                .add(404, "*", handler.notAvailable)
                .add(500, "*", "serverError" in handler ? handler.serverError : handler.temporalError)
                .add("*", "*", "networkError" in handler ? handler.networkError : handler.temporalError)
                .handle(err);
        },
    };
}

/**
 * Error handler for the semantic search image encoding API
 */
export type SemanticSearchEncodeImageErrorHandler = CommonAuthenticatedErrorHandler & {
    /**
     * Invalid image
     */
    invalidImage: () => void;

    /**
     * Image too large
     */
    imageTooLarge: () => void;

    /**
     * Semantic search service not yet available
     */
    notAvailable: () => void;
};

/**
 * Encodes an image into a vector in order
 * to be used to perform a semantic search query
 * @param image The image to encode
 * @returns The request parameters
 */
export function apiSemanticSearchEncodeImage(
    image: File,
): RequestParams<SearchMediaSemanticEncodeResponse, SemanticSearchEncodeImageErrorHandler> {
    const form = new FormData();
    form.append("image", image);
    return {
        method: "POST",
        url: getApiURL(`${API_PREFIX}${API_GROUP_PREFIX}/encoder/image`),
        form: form,
        handleError: (err, handler) => {
            new RequestErrorHandler()
                .add(401, "*", handler.unauthorized)
                .add(400, "*", handler.invalidImage)
                .add(404, "*", handler.notAvailable)
                .add(413, "*", handler.imageTooLarge)
                .add(500, "*", "serverError" in handler ? handler.serverError : handler.temporalError)
                .add("*", "*", "networkError" in handler ? handler.networkError : handler.temporalError)
                .handle(err);
        },
    };
}
