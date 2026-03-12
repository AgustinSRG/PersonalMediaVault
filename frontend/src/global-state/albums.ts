// Albums list global loader

"use strict";

import { makeNamedApiRequest } from "@asanrom/request-browser";
import { setNamedTimeout, clearNamedTimeout } from "@/utils/named-timeouts";
import { AuthController } from "./auth";
import type { AlbumListItemMin, AlbumListItemMinExt } from "@/api/models";
import { apiAlbumsGetAlbumsMin } from "@/api/api-albums";
import {
    addAppEventListener,
    emitAppEvent,
    EVENT_NAME_ALBUMS_LIST_UPDATE,
    EVENT_NAME_AUTH_CHANGED,
    EVENT_NAME_UNAUTHORIZED,
} from "./app-events";
import { getUniqueStringId } from "@/utils/unique-id";

/**
 * Global state of albums
 */
const AlbumsState = {
    /**
     * Map of albums.
     * ID -> Metadata
     */
    albumsMap: new Map<number, AlbumListItemMin>(),

    /**
     * Albums list loading status
     */
    loading: true,

    /**
     * True if the albums list was loaded at least once
     */
    initiallyLoaded: false,
};

/**
 * Gets the albums list
 * @returns The albums list
 */
export function getAlbumsList(): AlbumListItemMin[] {
    return Array.from(AlbumsState.albumsMap.values());
}

/**
 * Gets an extended version of the albums list,
 * including an extra field with the lowercased name.
 * @returns An extended version of the albums list
 */
export function getAlbumsListExt(): AlbumListItemMinExt[] {
    const result = [];

    const albums = AlbumsState.albumsMap.values();

    for (const album of albums) {
        result.push({
            id: album.id,
            name: album.name,
            nameLowerCase: album.name.toLowerCase(),
        });
    }

    return result;
}

/**
 * Gets the name of an album
 * @param id The album ID
 * @returns The album name
 */
export function getAlbumName(id: number): string {
    if (id < 0) {
        return "--";
    }

    const album = AlbumsState.albumsMap.get(id);

    if (!album) {
        return "";
    }

    return album.name;
}

/**
 * Checks if the albums list of being loaded
 * @returns True if the albums list of being loaded
 */
export function isAlbumsListLoading(): boolean {
    return AlbumsState.loading;
}

/**
 * Finds duplicated name in the albums list
 * @param name The name to find
 * @returns True if the name is found in the albums list
 */
export function albumsFindDuplicatedName(name: string): boolean {
    const nameLower = name.toLowerCase();

    const albums = AlbumsState.albumsMap.values();

    for (const album of albums) {
        if (nameLower === album.name.toLowerCase()) {
            return true;
        }
    }

    return false;
}

// Request ID for album list load
const REQUEST_ID_ALBUMS_LOAD = getUniqueStringId();

// Delay to retry loading after an error
const LOAD_RETRY_DELAY = 1500;

/**
 * Loads the albums list
 */
function loadAlbumsList() {
    AlbumsState.loading = true;

    if (AuthController.Locked) {
        return; // Vault is locked
    }

    clearNamedTimeout(REQUEST_ID_ALBUMS_LOAD);
    makeNamedApiRequest(REQUEST_ID_ALBUMS_LOAD, apiAlbumsGetAlbumsMin())
        .onSuccess((albums) => {
            AlbumsState.albumsMap.clear();

            for (const album of albums) {
                AlbumsState.albumsMap.set(album.id, album);
            }

            emitAppEvent(EVENT_NAME_ALBUMS_LIST_UPDATE, AlbumsState.albumsMap);
            AlbumsState.loading = false;
            AlbumsState.initiallyLoaded = true;
        })
        .onRequestError((err, handleErr) => {
            handleErr(err, {
                unauthorized: () => {
                    emitAppEvent(EVENT_NAME_UNAUTHORIZED);
                },
                temporalError: () => {
                    // Retry
                    setNamedTimeout(REQUEST_ID_ALBUMS_LOAD, LOAD_RETRY_DELAY, loadAlbumsList);
                },
            });
        })
        .onUnexpectedError((err) => {
            console.error(err);
            // Retry
            setNamedTimeout(REQUEST_ID_ALBUMS_LOAD, LOAD_RETRY_DELAY, loadAlbumsList);
        });
}

/**
 * Refreshes the albums list from the server
 * @param force True to force refresh, aborting the previous request
 */
export function refreshAlbumsList(force?: boolean) {
    if (!force && AlbumsState.loading) {
        return; // Already loading
    }

    loadAlbumsList();
}

/**
 * Initializes albums
 */
export function initializeAlbums() {
    addAppEventListener(EVENT_NAME_AUTH_CHANGED, loadAlbumsList);

    loadAlbumsList();
}
