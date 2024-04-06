// Search API

"use strict";

import { CommonAuthenticatedErrorHandler, RequestErrorHandler, RequestParams } from "@asanrom/request-browser";
import { SearchResults, RandomResults, AdvancedSearchResults } from "./models";
import { generateURIQuery, getApiURL } from "@/utils/api";

/**
 * Search for media in the vault
 * @param tag Tag to filter by
 * @param order Order direction
 * @param page Page number, starting at 0
 * @param pageSize Page size
 * @returns The request parameters
 */
export function apiSearch(
    tag: string,
    order: "asc" | "desc",
    page: number,
    pageSize: number,
): RequestParams<SearchResults, CommonAuthenticatedErrorHandler> {
    return {
        method: "GET",
        url: getApiURL(
            "/api/search" +
                generateURIQuery({
                    tag: tag,
                    order: order,
                    page_index: page + "",
                    page_size: pageSize + "",
                }),
        ),
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
 * Gets random results
 * @param tag Tag to filter by
 * @param seed RNG seed
 * @param pageSize Page size
 * @returns The request parameters
 */
export function apiSearchRandom(
    tag: string,
    seed: number,
    pageSize: number,
): RequestParams<RandomResults, CommonAuthenticatedErrorHandler> {
    return {
        method: "GET",
        url: getApiURL(
            "/api/random" +
                generateURIQuery({
                    tag: tag,
                    seed: seed + "",
                    page_size: pageSize + "",
                }),
        ),
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
 * Search for media in the vault
 * @param tagMode Filtering mode
 * @param tags List of tags
 * @param order Order direction
 * @param continueRef Reference to the last element, to get more items
 * @param limit Max number of items to return
 * @returns The request parameters
 */
export function apiAdvancedSearch(
    tagMode: "allof" | "anyof" | "noneof",
    tags: string[],
    order: "asc" | "desc",
    continueRef: number,
    limit: number,
): RequestParams<AdvancedSearchResults, CommonAuthenticatedErrorHandler> {
    return {
        method: "GET",
        url: getApiURL(
            "/api/search/advanced" +
                generateURIQuery({
                    tags_mode: tagMode,
                    tags: JSON.stringify(tags),
                    order: order,
                    continue: continueRef,
                    limit: limit + "",
                }),
        ),
        handleError: (err, handler) => {
            new RequestErrorHandler()
                .add(401, "*", handler.unauthorized)
                .add(500, "*", "serverError" in handler ? handler.serverError : handler.temporalError)
                .add("*", "*", "networkError" in handler ? handler.networkError : handler.temporalError)
                .handle(err);
        },
    };
}
