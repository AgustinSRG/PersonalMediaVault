// Home API

"use strict";

import { API_PREFIX, getApiURL } from "@/utils/api";
import type { CommonAuthenticatedErrorHandler, RequestParams } from "@asanrom/request-browser";
import { RequestErrorHandler } from "@asanrom/request-browser";
import type { AlbumListItem, MediaListItem } from "./models";
import type { ProvidedAuthConfirmation } from "./api-auth";

const API_GROUP_PREFIX = "/home";

/**
 * Types of groups in the home page
 */
export const HomePageGroupTypes = {
    /**
     * Custom group with custom elements
     */
    Custom: 0,

    /**
     * Recently uploaded media
     */
    RecentMedia: 1,

    /**
     * Recently modified albums
     */
    RecentAlbums: 2,
};

/**
 * A group in the home page
 */
export interface HomePageGroup {
    /**
     * Group ID
     */
    id: number;

    /**
     * Group type (HomePageGroupTypes)
     */
    type?: number;

    /**
     * Group name
     */
    name?: string;

    /**
     * Number of custom elements
     */
    elementsCount?: number;
}

/**
 * Gets the list of groups for the home page
 * @returns The request parameters
 */
export function apiHomeGetGroups(): RequestParams<HomePageGroup[], CommonAuthenticatedErrorHandler> {
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
 * Options to add a group to the home page
 */
export interface HomeAddGroupBody {
    /**
     * Group name
     */
    name: string;

    /**
     * Group type (HomePageGroupTypes)
     */
    type: number;

    /**
     * True to prepend the group, false to append it
     */
    prepend: boolean;
}

/**
 * Error handler for the APIs that modify the home page
 */
export type HomeApiWriteErrorHandler = CommonAuthenticatedErrorHandler & {
    /**
     * Error: Access denied
     */
    accessDenied: () => void;
};

/**
 * Error handler for the APIs that adds a group to the home page
 */
export type HomeApiAddGroupErrorHandler = HomeApiWriteErrorHandler & {
    /**
     * Error: Invalid name
     */
    invalidName: () => void;

    /**
     * Error: Invalid group type
     */
    invalidGroupType: () => void;

    /**
     * Error: Too many groups
     */
    tooManyGroups: () => void;
};

/**
 * Adds a group to the home page
 * @param options Request options
 * @returns The request parameters
 */
export function apiHomeAddGroup(options: HomeAddGroupBody): RequestParams<HomePageGroup, HomeApiAddGroupErrorHandler> {
    return {
        method: "POST",
        url: getApiURL(`${API_PREFIX}${API_GROUP_PREFIX}`),
        json: options,
        handleError: (err, handler) => {
            new RequestErrorHandler()
                .add(400, "INVALID_NAME", handler.invalidName)
                .add(400, "TOO_MANY_GROUPS", handler.tooManyGroups)
                .add(400, "*", handler.invalidGroupType)
                .add(401, "*", handler.unauthorized)
                .add(403, "*", handler.accessDenied)
                .add(500, "*", "serverError" in handler ? handler.serverError : handler.temporalError)
                .add("*", "*", "networkError" in handler ? handler.networkError : handler.temporalError)
                .handle(err);
        },
    };
}

/**
 * Custom element in a home page group
 * Only one (media, or album) will be defined
 */
export interface HomePageElement {
    /**
     * Media element
     */
    media?: MediaListItem;

    /**
     * Album element
     */
    album?: AlbumListItem;
}

/**
 * Error handler for the APIs that handle a group
 */
export type HomeApiGroupErrorHandler = CommonAuthenticatedErrorHandler & {
    /**
     * Error: Group not found
     */
    notFound: () => void;
};

/**
 * Gets the list of elements of a home group
 * @param id The home group ID
 * @returns The request parameters
 */
export function apiHomeGetGroupElements(id: number): RequestParams<HomePageElement[], HomeApiGroupErrorHandler> {
    return {
        method: "GET",
        url: getApiURL(`${API_PREFIX}${API_GROUP_PREFIX}/${encodeURIComponent(id + "")}/elements`),
        handleError: (err, handler) => {
            new RequestErrorHandler()
                .add(400, "*", handler.notFound)
                .add(401, "*", handler.unauthorized)
                .add(404, "*", handler.notFound)
                .add(500, "*", "serverError" in handler ? handler.serverError : handler.temporalError)
                .add("*", "*", "networkError" in handler ? handler.networkError : handler.temporalError)
                .handle(err);
        },
    };
}

/**
 * Reference to an element
 */
export interface HomePageElementRef {
    /**
     * Element type (0 = media, 1 = album)
     */
    t: 0 | 1;

    /**
     * ID of the media or album
     */
    i: number;
}

/**
 * Gets references from a list of elements of the home page
 * @param elements The elements
 * @returns The references
 */
export function getHomePageElementReferences(elements: HomePageElement[]): HomePageElementRef[] {
    return elements
        .map((e) => {
            if (e.media) {
                return {
                    t: 0,
                    i: e.media.id,
                } as HomePageElementRef;
            } else if (e.album) {
                return {
                    t: 1,
                    i: e.album.id,
                } as HomePageElementRef;
            } else {
                return null;
            }
        })
        .filter((e) => e !== null);
}

/**
 * Error handler for the APIs that modify the home page groups
 */
export type HomeApiGroupWriteErrorHandler = HomeApiGroupErrorHandler & {
    /**
     * Error: Access denied
     */
    accessDenied: () => void;
};

/**
 * Error handler for the API that changes the elements of a group in the home page
 */
export type HomeApiGroupSetElementsErrorHandler = HomeApiGroupWriteErrorHandler & {
    /**
     * Error: Not a custom group
     */
    notCustomGroup: () => void;
};

/**
 * Error handler for the API that changes the elements of a group in the home page
 */
export type HomeApiGroupAddElementsErrorHandler = HomeApiGroupSetElementsErrorHandler & {
    /**
     * Error: Too many elements
     */
    tooManyElements: () => void;
};

/**
 * Add element to a group in the home page
 * @param id The home group ID
 * @param element Element to add
 * @returns The request parameters
 */
export function apiHomeGroupAddElement(id: number, element: HomePageElementRef): RequestParams<void, HomeApiGroupAddElementsErrorHandler> {
    return {
        method: "POST",
        url: getApiURL(`${API_PREFIX}${API_GROUP_PREFIX}/${encodeURIComponent(id + "")}/elements`),
        json: {
            add: element,
        },
        handleError: (err, handler) => {
            new RequestErrorHandler()
                .add(400, "NOT_CUSTOM_GROUP", handler.notCustomGroup)
                .add(400, "TOO_MANY_ELEMENTS", handler.tooManyElements)
                .add(401, "*", handler.unauthorized)
                .add(403, "*", handler.accessDenied)
                .add(404, "*", handler.notFound)
                .add(500, "*", "serverError" in handler ? handler.serverError : handler.temporalError)
                .add("*", "*", "networkError" in handler ? handler.networkError : handler.temporalError)
                .handle(err);
        },
    };
}

/**
 * Deletes element from a group in the home page
 * @param id The home group ID
 * @param element Element to delete
 * @returns The request parameters
 */
export function apiHomeGroupDeleteElement(
    id: number,
    element: HomePageElementRef,
): RequestParams<void, HomeApiGroupSetElementsErrorHandler> {
    return {
        method: "POST",
        url: getApiURL(`${API_PREFIX}${API_GROUP_PREFIX}/${encodeURIComponent(id + "")}/elements`),
        json: {
            delete: element,
        },
        handleError: (err, handler) => {
            new RequestErrorHandler()
                .add(400, "NOT_CUSTOM_GROUP", handler.notCustomGroup)
                .add(401, "*", handler.unauthorized)
                .add(403, "*", handler.accessDenied)
                .add(404, "*", handler.notFound)
                .add(500, "*", "serverError" in handler ? handler.serverError : handler.temporalError)
                .add("*", "*", "networkError" in handler ? handler.networkError : handler.temporalError)
                .handle(err);
        },
    };
}

/**
 * Moves element from a group in the home page
 * @param id The home group ID
 * @param element Element to move
 * @param position The position to move the element
 * @returns The request parameters
 */
export function apiHomeGroupMoveElement(
    id: number,
    element: HomePageElementRef,
    position: number,
): RequestParams<void, HomeApiGroupSetElementsErrorHandler> {
    return {
        method: "POST",
        url: getApiURL(`${API_PREFIX}${API_GROUP_PREFIX}/${encodeURIComponent(id + "")}/elements`),
        json: {
            move: {
                t: element.t,
                i: element.i,
                position,
            },
        },
        handleError: (err, handler) => {
            new RequestErrorHandler()
                .add(400, "NOT_CUSTOM_GROUP", handler.notCustomGroup)
                .add(401, "*", handler.unauthorized)
                .add(403, "*", handler.accessDenied)
                .add(404, "*", handler.notFound)
                .add(500, "*", "serverError" in handler ? handler.serverError : handler.temporalError)
                .add("*", "*", "networkError" in handler ? handler.networkError : handler.temporalError)
                .handle(err);
        },
    };
}

/**
 * Error handler for the API that renames a group in the home page
 */
export type HomeApiGroupRenameErrorHandler = HomeApiGroupWriteErrorHandler & {
    /**
     * Error: Invalid name
     */
    invalidName: () => void;
};

/**
 * Renames a home page group
 * @param id The home group ID
 * @param newName New group name
 * @returns The request parameters
 */
export function apiHomeGroupRename(id: number, newName: string): RequestParams<void, HomeApiGroupRenameErrorHandler> {
    return {
        method: "POST",
        url: getApiURL(`${API_PREFIX}${API_GROUP_PREFIX}/${encodeURIComponent(id + "")}/name`),
        json: {
            name: newName,
        },
        handleError: (err, handler) => {
            new RequestErrorHandler()
                .add(400, "*", handler.invalidName)
                .add(401, "*", handler.unauthorized)
                .add(403, "*", handler.accessDenied)
                .add(404, "*", handler.notFound)
                .add(500, "*", "serverError" in handler ? handler.serverError : handler.temporalError)
                .add("*", "*", "networkError" in handler ? handler.networkError : handler.temporalError)
                .handle(err);
        },
    };
}

/**
 * Moves a home page group
 * @param id The home group ID
 * @param newPosition New group position
 * @returns The request parameters
 */
export function apiHomeGroupMove(id: number, newPosition: number): RequestParams<void, HomeApiGroupWriteErrorHandler> {
    return {
        method: "POST",
        url: getApiURL(`${API_PREFIX}${API_GROUP_PREFIX}/${encodeURIComponent(id + "")}/move`),
        json: {
            position: newPosition,
        },
        handleError: (err, handler) => {
            new RequestErrorHandler()
                .add(401, "*", handler.unauthorized)
                .add(403, "*", handler.accessDenied)
                .add(404, "*", handler.notFound)
                .add(500, "*", "serverError" in handler ? handler.serverError : handler.temporalError)
                .add("*", "*", "networkError" in handler ? handler.networkError : handler.temporalError)
                .handle(err);
        },
    };
}

/**
 * Error handler for the API that deletes a group in the home page
 */
export type HomeApiGroupDeleteErrorHandler = HomeApiGroupWriteErrorHandler & {
    /**
     * Required auth confirmation (two factor authentication)
     */
    requiredAuthConfirmationTfa: () => void;

    /**
     * Invalid two factor authentication code
     */
    invalidTfaCode: () => void;

    /**
     * Required auth confirmation (password)
     */
    requiredAuthConfirmationPassword: () => void;

    /**
     * Invalid password
     */
    invalidPassword: () => void;

    /**
     * When you fail a confirmation, there is a cooldown of 5 seconds.
     */
    cooldown: () => void;
};

/**
 * Deletes a home page group
 * @param id The home group ID
 * @param providedAuthConfirmation Auth confirmation
 * @returns The request parameters
 */
export function apiHomeGroupDelete(
    id: number,
    providedAuthConfirmation: ProvidedAuthConfirmation,
): RequestParams<void, HomeApiGroupDeleteErrorHandler> {
    return {
        method: "DELETE",
        url: getApiURL(`${API_PREFIX}${API_GROUP_PREFIX}/${encodeURIComponent(id + "")}`),
        headers: {
            "x-auth-confirmation-pw": providedAuthConfirmation.password || "",
            "x-auth-confirmation-tfa": providedAuthConfirmation.tfaCode || "",
        },
        handleError: (err, handler) => {
            new RequestErrorHandler()
                .add(401, "*", handler.unauthorized)
                .add(403, "AUTH_CONFIRMATION_REQUIRED_TFA", handler.requiredAuthConfirmationTfa)
                .add(403, "INVALID_TFA_CODE", handler.invalidTfaCode)
                .add(403, "AUTH_CONFIRMATION_REQUIRED_PW", handler.requiredAuthConfirmationPassword)
                .add(403, "INVALID_PASSWORD", handler.invalidPassword)
                .add(403, "COOLDOWN", handler.cooldown)
                .add(403, "*", handler.accessDenied)
                .add(500, "*", "serverError" in handler ? handler.serverError : handler.temporalError)
                .add("*", "*", "networkError" in handler ? handler.networkError : handler.temporalError)
                .handle(err);
        },
    };
}
