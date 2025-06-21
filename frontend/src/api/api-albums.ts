// Albums API

"use strict";

import { CommonAuthenticatedErrorHandler, RequestErrorHandler, RequestParams } from "@asanrom/request-browser";
import { AlbumListItem, AlbumListItemMin, Album } from "./models";
import { API_PREFIX, getApiURL } from "@/utils/api";
import { ProvidedAuthConfirmation } from "./api-auth";

const API_GROUP_PREFIX = "/albums";

/**
 * Gets list of albums
 * @returns The request parameters
 */
export function apiAlbumsGetAlbums(): RequestParams<AlbumListItem[], CommonAuthenticatedErrorHandler> {
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
 * Gets list of albums (only id and name).
 * This API is very fast, used to know the full list of albums when the vault GUI loads.
 * @returns The request parameters
 */
export function apiAlbumsGetAlbumsMin(): RequestParams<AlbumListItemMin[], CommonAuthenticatedErrorHandler> {
    return {
        method: "GET",
        url: getApiURL(`${API_PREFIX}${API_GROUP_PREFIX}?mode=min`),
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
 * Error handler for album get API
 */
export type GetAlbumErrorHandler = CommonAuthenticatedErrorHandler & {
    /**
     * Error: Album not found
     */
    notFound: () => void;
};

/**
 * Gets an album
 * @param id The album ID
 * @returns The request parameters
 */
export function apiAlbumsGetAlbum(id: number): RequestParams<Album, GetAlbumErrorHandler> {
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
 * Response for album create API
 */
export interface CreateAlbumResponse {
    /**
     * Id of the created album
     */
    album_id: number;
}

/**
 * Error handler for album create API
 */
export type CreateAlbumErrorHandler = CommonAuthenticatedErrorHandler & {
    /**
     * Error: Access denied
     */
    accessDenied: () => void;

    /**
     * Error: Invalid name
     */
    invalidName: () => void;

    /**
     * Error: Bad request
     */
    badRequest: () => void;
};

/**
 * Creates album
 * @param name Album name
 * @returns The request parameters
 */
export function apiAlbumsCreateAlbum(name: string): RequestParams<CreateAlbumResponse, CreateAlbumErrorHandler> {
    return {
        method: "POST",
        url: getApiURL(`${API_PREFIX}${API_GROUP_PREFIX}`),
        json: {
            name: name,
        },
        handleError: (err, handler) => {
            new RequestErrorHandler()
                .add(401, "*", handler.unauthorized)
                .add(400, "INVALID_NAME", handler.invalidName)
                .add(400, "*", handler.badRequest)
                .add(403, "*", handler.accessDenied)
                .add(500, "*", "serverError" in handler ? handler.serverError : handler.temporalError)
                .add("*", "*", "networkError" in handler ? handler.networkError : handler.temporalError)
                .handle(err);
        },
    };
}

/**
 * Error handler for album delete API
 */
export type DeleteAlbumErrorHandler = CommonAuthenticatedErrorHandler & {
    /**
     * Error: Access denied
     */
    accessDenied: () => void;

    /**
     * Error: Album not found
     */
    notFound: () => void;

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
 * Deletes album
 * @param id Album ID
 * @param providedAuthConfirmation Auth confirmation
 * @returns The request parameters
 */
export function apiAlbumsDeleteAlbum(
    id: number,
    providedAuthConfirmation: ProvidedAuthConfirmation,
): RequestParams<void, DeleteAlbumErrorHandler> {
    return {
        method: "POST",
        url: getApiURL(`${API_PREFIX}${API_GROUP_PREFIX}/${encodeURIComponent(id + "")}/delete`),
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
                .add(404, "*", handler.notFound)
                .add(500, "*", "serverError" in handler ? handler.serverError : handler.temporalError)
                .add("*", "*", "networkError" in handler ? handler.networkError : handler.temporalError)
                .handle(err);
        },
    };
}

/**
 * Error handler for album rename API
 */
export type RenameAlbumErrorHandler = CommonAuthenticatedErrorHandler & {
    /**
     * Error: Invalid name
     */
    invalidName: () => void;

    /**
     * Error: Bad request
     */
    badRequest: () => void;

    /**
     * Error: Access denied
     */
    accessDenied: () => void;

    /**
     * Error: Album not found
     */
    notFound: () => void;
};

/**
 * Renames album
 * @param id Album ID
 * @param name New name for the album
 * @returns The request parameters
 */
export function apiAlbumsRenameAlbum(id: number, name: string): RequestParams<void, RenameAlbumErrorHandler> {
    return {
        method: "POST",
        url: getApiURL(`${API_PREFIX}${API_GROUP_PREFIX}/${encodeURIComponent(id + "")}/rename`),
        json: {
            name: name,
        },
        handleError: (err, handler) => {
            new RequestErrorHandler()
                .add(401, "*", handler.unauthorized)
                .add(400, "INVALID_NAME", handler.invalidName)
                .add(400, "*", handler.badRequest)
                .add(403, "*", handler.accessDenied)
                .add(404, "*", handler.notFound)
                .add(500, "*", "serverError" in handler ? handler.serverError : handler.temporalError)
                .add("*", "*", "networkError" in handler ? handler.networkError : handler.temporalError)
                .handle(err);
        },
    };
}

/**
 * Error handler for album add-media API
 */
export type AddMediaAlbumErrorHandler = CommonAuthenticatedErrorHandler & {
    /**
     * Error: Max size reached
     */
    maxSizeReached: () => void;

    /**
     * Error: Bad request
     */
    badRequest: () => void;

    /**
     * Error: Access denied
     */
    accessDenied: () => void;

    /**
     * Error: Album not found
     */
    notFound: () => void;
};

/**
 * Adds media to album
 * @param id Album ID
 * @param media Media ID
 * @returns The request parameters
 */
export function apiAlbumsAddMediaToAlbum(id: number, media: number): RequestParams<void, AddMediaAlbumErrorHandler> {
    return {
        method: "POST",
        url: getApiURL(`${API_PREFIX}${API_GROUP_PREFIX}/${encodeURIComponent(id + "")}/add`),
        json: {
            media_id: media,
        },
        handleError: (err, handler) => {
            new RequestErrorHandler()
                .add(401, "*", handler.unauthorized)
                .add(400, "MAX_SIZE_REACHED", handler.maxSizeReached)
                .add(400, "*", handler.badRequest)
                .add(403, "*", handler.accessDenied)
                .add(404, "*", handler.notFound)
                .add(500, "*", "serverError" in handler ? handler.serverError : handler.temporalError)
                .add("*", "*", "networkError" in handler ? handler.networkError : handler.temporalError)
                .handle(err);
        },
    };
}

/**
 * Error handler for album remove-media API
 */
export type RemoveMediaAlbumErrorHandler = CommonAuthenticatedErrorHandler & {
    /**
     * Error: Access denied
     */
    accessDenied: () => void;

    /**
     * Error: Album not found
     */
    notFound: () => void;
};

/**
 * Removes media from album
 * @param id Album ID
 * @param media Media ID
 * @returns The request parameters
 */
export function apiAlbumsRemoveMediaFromAlbum(id: number, media: number): RequestParams<void, RemoveMediaAlbumErrorHandler> {
    return {
        method: "POST",
        url: getApiURL(`${API_PREFIX}${API_GROUP_PREFIX}/${encodeURIComponent(id + "")}/remove`),
        json: {
            media_id: media,
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
 * Error handler for album move-media API
 */
export type MoveMediaAlbumErrorHandler = CommonAuthenticatedErrorHandler & {
    /**
     * Error: Max size reached
     */
    maxSizeReached: () => void;

    /**
     * Error: Bad request
     */
    badRequest: () => void;

    /**
     * Error: Access denied
     */
    accessDenied: () => void;

    /**
     * Error: Album not found
     */
    notFound: () => void;
};

/**
 * Moves media into album
 * @param id Album ID
 * @param media Media ID
 * @param position New position for the media (index, start at 0)
 * @returns The request parameters
 */
export function apiAlbumsMoveMediaInAlbum(id: number, media: number, position: number): RequestParams<void, MoveMediaAlbumErrorHandler> {
    return {
        method: "POST",
        url: getApiURL(`${API_PREFIX}${API_GROUP_PREFIX}/${encodeURIComponent(id + "")}/move`),
        json: {
            media_id: media,
            position: position,
        },
        handleError: (err, handler) => {
            new RequestErrorHandler()
                .add(401, "*", handler.unauthorized)
                .add(400, "MAX_SIZE_REACHED", handler.maxSizeReached)
                .add(400, "*", handler.badRequest)
                .add(403, "*", handler.accessDenied)
                .add(404, "*", handler.notFound)
                .add(500, "*", "serverError" in handler ? handler.serverError : handler.temporalError)
                .add("*", "*", "networkError" in handler ? handler.networkError : handler.temporalError)
                .handle(err);
        },
    };
}

/**
 * Error handler for set thumbnail API
 */
export type SetAlbumThumbnailErrorHandler = CommonAuthenticatedErrorHandler & {
    /**
     * Error: Invalid thumbnail
     */
    invalidThumbnail: () => void;

    /**
     * Error: Bad request
     */
    badRequest: () => void;

    /**
     * Error: Access denied
     */
    accessDenied: () => void;

    /**
     * Error: Album not found
     */
    notFound: () => void;
};

/**
 * Response of the API to change the album thumbnail
 */
export interface ChangeAlbumThumbnailResponse {
    /**
     * New thumbnail URL
     */
    url: string;
}

/**
 * Sets thumbnail
 * @param id Album ID
 * @param thumbnail Thumbnail file to upload
 * @returns The request parameters
 */
export function apiAlbumsChangeAlbumThumbnail(
    id: number,
    thumbnail: File,
): RequestParams<ChangeAlbumThumbnailResponse, SetAlbumThumbnailErrorHandler> {
    const form = new FormData();
    form.append("file", thumbnail);
    return {
        method: "POST",
        url: getApiURL(`${API_PREFIX}${API_GROUP_PREFIX}/${encodeURIComponent(id + "")}/thumbnail`),
        form: form,
        handleError: (err, handler) => {
            new RequestErrorHandler()
                .add(401, "*", handler.unauthorized)
                .add(400, "INVALID_THUMBNAIL", handler.invalidThumbnail)
                .add(400, "*", handler.badRequest)
                .add(403, "*", handler.accessDenied)
                .add(404, "*", handler.notFound)
                .add(500, "*", "serverError" in handler ? handler.serverError : handler.temporalError)
                .add("*", "*", "networkError" in handler ? handler.networkError : handler.temporalError)
                .handle(err);
        },
    };
}
