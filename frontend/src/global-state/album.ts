// Current album global state

"use strict";

import { makeNamedApiRequest, abortNamedApiRequest } from "@asanrom/request-browser";
import { setNamedTimeout, clearNamedTimeout } from "@/utils/named-timeouts";
import { AppStatus } from "./app-status";
import { AuthController } from "./auth";
import type { Album, MediaData, MediaListItem } from "@/api/models";
import { apiAlbumsGetAlbum } from "@/api/api-albums";
import { apiMediaGetMedia } from "@/api/api-media";
import {
    addAppEventListener,
    emitAppEvent,
    EVENT_NAME_ALBUMS_CHANGED,
    EVENT_NAME_APP_STATUS_CHANGED,
    EVENT_NAME_AUTH_CHANGED,
    EVENT_NAME_CURRENT_ALBUM_LOADING,
    EVENT_NAME_CURRENT_ALBUM_MEDIA_POSITION_UPDATED,
    EVENT_NAME_CURRENT_ALBUM_UPDATED,
    EVENT_NAME_NEXT_PRE_FETCH,
    EVENT_NAME_UNAUTHORIZED,
} from "./app-events";
import { getUniqueStringId } from "@/utils/unique-id";
import { getCurrentMediaId, provideCurrentMediaData } from "./media";
import { setCachedAlbumPosition } from "./album-position-cache";
import { refreshAlbumsList } from "./albums";

/**
 * State of the current album
 */
const CurrentAlbumState = {
    /**
     * Current album ID,
     * or -1 if no album selected.
     */
    id: -1,

    /**
     * Loading status of the current album
     */
    loading: false,

    /**
     * Current album data
     */
    data: null as Album | null,
};

/**
 * Gets the current album ID
 * @returns The current album ID
 */
export function getCurrentAlbumId(): number {
    return CurrentAlbumState.id;
}

/**
 * Checks if the current album is being loaded
 * @returns True if the current album is being loaded
 */
export function isCurrentAlbumLoading(): boolean {
    return CurrentAlbumState.loading;
}

/**
 * Sets the current album loading value
 * @param loading The loading value
 */
function setCurrentAlbumLoading(loading: boolean) {
    CurrentAlbumState.loading = loading;
    emitAppEvent(EVENT_NAME_CURRENT_ALBUM_LOADING, loading);
}

/**
 * Gets the current album data
 * @returns The album data, or null
 */
export function getCurrentAlbumData(): Album | null {
    return CurrentAlbumState.data;
}

/**
 * Sets the current album data
 * @param data The album data, or null
 */
function setCurrentAlbumData(data: Album | null) {
    CurrentAlbumState.data = data;
    emitAppEvent(EVENT_NAME_CURRENT_ALBUM_UPDATED, data);
}

// Request ID for current album load
const REQUEST_ID_CURRENT_ALBUM_LOAD = getUniqueStringId();

// Delay to retry loading after an error
const LOAD_RETRY_DELAY = 1500;

/**
 * Loads the current album data
 */
function loadCurrentAlbum() {
    if (CurrentAlbumState.id < 0) {
        clearNamedTimeout(REQUEST_ID_CURRENT_ALBUM_LOAD);
        abortNamedApiRequest(REQUEST_ID_CURRENT_ALBUM_LOAD);

        setCurrentAlbumData(null);

        setCurrentAlbumLoading(false);

        updateAlbumMediaPositionStatus();

        return;
    }

    setCurrentAlbumLoading(true);

    if (AuthController.Locked) {
        return; // Vault is locked
    }

    clearNamedTimeout(REQUEST_ID_CURRENT_ALBUM_LOAD);
    makeNamedApiRequest(REQUEST_ID_CURRENT_ALBUM_LOAD, apiAlbumsGetAlbum(CurrentAlbumState.id))
        .onSuccess((album) => {
            setCurrentAlbumData(album);

            setCurrentAlbumLoading(false);

            updateAlbumMediaPositionStatus();

            AppStatus.UpdateURL();
        })
        .onRequestError((err, handleErr) => {
            handleErr(err, {
                unauthorized: () => {
                    emitAppEvent(EVENT_NAME_UNAUTHORIZED);
                },
                notFound: () => {
                    setCurrentAlbumData(null);

                    setCurrentAlbumLoading(false);

                    AppStatus.CloseAlbum();

                    refreshAlbumsList();
                },
                temporalError: () => {
                    // Retry
                    setNamedTimeout(REQUEST_ID_CURRENT_ALBUM_LOAD, LOAD_RETRY_DELAY, loadCurrentAlbum);
                },
            });
        })
        .onUnexpectedError((err) => {
            console.error(err);
            // Retry
            setNamedTimeout(REQUEST_ID_CURRENT_ALBUM_LOAD, LOAD_RETRY_DELAY, loadCurrentAlbum);
        });
}

/**
 * Refreshes current album, reloading it from the server
 */
export function refreshCurrentAlbum() {
    loadCurrentAlbum();
}

/**
 * Called when the app status changes, in order to reload if necessary
 */
function onCurrentAlbumChanged() {
    if (AppStatus.CurrentAlbum !== CurrentAlbumState.id) {
        CurrentAlbumState.id = AppStatus.CurrentAlbum;
        CurrentAlbumState.data = null;
        loadCurrentAlbum();
    }

    updateAlbumMediaPositionStatus();
}

/**
 * Call when the user makes changes to an album
 * @param albumId The album ID
 * @param noUpdateList Set to true if a list reload is not necessary
 */
export function indicateAlbumMetadataChanged(albumId: number, noUpdateList?: boolean) {
    if (CurrentAlbumState.id === albumId) {
        loadCurrentAlbum();
    }

    emitAppEvent(EVENT_NAME_ALBUMS_CHANGED);

    if (!noUpdateList) {
        refreshAlbumsList(true);
    }
}

/**
 * The position context of the current media in the current album
 */
export type AlbumMediaPositionContext = {
    /**
     * The position in the list
     */
    pos: number;

    /**
     * The previous element
     */
    prev: MediaListItem | null;

    /**
     * The next element
     */
    next: MediaListItem | null;

    /**
     * True if album loop is enabled
     */
    loop: boolean;

    /**
     * True if album order randomization is enabled
     */
    random: boolean;
};

/**
 * Position status of the current media in the current album
 */
const AlbumMediaPositionStatus: AlbumMediaPositionContext = {
    /**
     * Position of the current media in the current album.
     * -1 if the media is not in the album
     */
    pos: -1,

    /**
     * Previous element for the current media in the album
     */
    prev: null as MediaListItem | null,

    /**
     * Next element for the current media in the album
     */
    next: null as MediaListItem | null,

    /**
     * True if album loop is enabled
     */
    loop: false,

    /**
     * True if album order randomization is enabled
     */
    random: false,
};

/**
 * Gets the current position of the current media
 * in the current album
 * @returns The media position
 */
export function getCurrentAlbumMediaPosition(): number {
    return AlbumMediaPositionStatus.pos;
}

/**
 * Gets the position context of the current media in the current album.
 * This includes the position, and the previous and next elements,
 * and the loop / random order options.
 * @returns The context
 */
export function getCurrentAlbumMediaPositionContext(): Readonly<AlbumMediaPositionContext> {
    return AlbumMediaPositionStatus;
}

/**
 * Updates current album media position status
 */
export function updateAlbumMediaPositionStatus() {
    const mediaId = AppStatus.CurrentMedia;

    if (!CurrentAlbumState.data || CurrentAlbumState.loading) {
        AlbumMediaPositionStatus.pos = -1;
        AlbumMediaPositionStatus.prev = null;
        AlbumMediaPositionStatus.next = null;
        emitAppEvent(EVENT_NAME_CURRENT_ALBUM_MEDIA_POSITION_UPDATED, AlbumMediaPositionStatus);
        loadAlbumNextPreFetch();
        return;
    }

    if (mediaId < 0 && CurrentAlbumState.data.list.length > 0) {
        const albumList = CurrentAlbumState.data.list.map((a) => {
            return a.id;
        });
        AppStatus.ClickOnAlbumWithList(CurrentAlbumState.data.id, albumList);
        return;
    }

    let mediaPos = -1;

    for (let i = 0; i < CurrentAlbumState.data.list.length; i++) {
        if (mediaId === CurrentAlbumState.data.list[i].id) {
            mediaPos = i;
            break;
        }
    }

    AlbumMediaPositionStatus.pos = mediaPos;

    if (mediaPos >= 0) {
        if (AlbumMediaPositionStatus.random) {
            let randomIndex = Math.floor(Math.random() * (CurrentAlbumState.data.list.length - 1));

            if (randomIndex >= mediaPos) {
                randomIndex++;
            }

            AlbumMediaPositionStatus.prev = null;
            AlbumMediaPositionStatus.next = CurrentAlbumState.data.list[randomIndex] || null;

            if (AlbumMediaPositionStatus.loop) {
                if (AlbumMediaPositionStatus.next === null) {
                    AlbumMediaPositionStatus.next = CurrentAlbumState.data.list[0] || null;
                }
            }
        } else {
            AlbumMediaPositionStatus.prev = CurrentAlbumState.data.list[mediaPos - 1] || null;
            AlbumMediaPositionStatus.next = CurrentAlbumState.data.list[mediaPos + 1] || null;

            if (AlbumMediaPositionStatus.loop) {
                if (AlbumMediaPositionStatus.prev === null) {
                    AlbumMediaPositionStatus.prev = CurrentAlbumState.data.list[CurrentAlbumState.data.list.length - 1] || null;
                }

                if (AlbumMediaPositionStatus.next === null) {
                    AlbumMediaPositionStatus.next = CurrentAlbumState.data.list[0] || null;
                }
            }
        }
    } else {
        AlbumMediaPositionStatus.prev = null;
        AlbumMediaPositionStatus.next = null;
    }

    setCachedAlbumPosition(CurrentAlbumState.data.id, AlbumMediaPositionStatus.pos);

    emitAppEvent(EVENT_NAME_CURRENT_ALBUM_MEDIA_POSITION_UPDATED, AlbumMediaPositionStatus);
    loadAlbumNextPreFetch();
}

/**
 * Toggles album loop option
 * @returns The current album loop value
 */
export function albumToggleLoop(): boolean {
    AlbumMediaPositionStatus.loop = !AlbumMediaPositionStatus.loop;
    if (AlbumMediaPositionStatus.loop) {
        AlbumMediaPositionStatus.random = false;
    }
    updateAlbumMediaPositionStatus();

    return AlbumMediaPositionStatus.loop;
}

/**
 * Toggles album order randomize option
 * @returns The current album order randomize option value
 */
export function albumToggleRandom(): boolean {
    AlbumMediaPositionStatus.random = !AlbumMediaPositionStatus.random;
    if (AlbumMediaPositionStatus.random) {
        AlbumMediaPositionStatus.loop = false;
    }
    updateAlbumMediaPositionStatus();

    return AlbumMediaPositionStatus.random;
}

/**
 * State for pre-fetching the next media of the album
 */
const AlbumNextPrefetchState = {
    /**
     * Is the next element available for pre-fetch?
     */
    available: false,

    /**
     * Is the pre-fetch element being loaded?
     */
    loading: false,

    /**
     * Pre-fetched data
     */
    data: null as MediaData | null,
};

/**
 * Gets the media data of the next element in the album (pe-fetched)
 * @returns The metadata or null
 */
export function getAlbumNextPrefetchData(): MediaData | null {
    return AlbumNextPrefetchState.data;
}

// Request ID for pre-fetch
const REQUEST_ID_NEXT_PRE_FETCH = getUniqueStringId();

/**
 * Pre-fetches the next media element for faster transition
 */
export function loadAlbumNextPreFetch() {
    if (AlbumMediaPositionStatus.next === null || AlbumMediaPositionStatus.next.id === getCurrentMediaId()) {
        clearNamedTimeout(REQUEST_ID_NEXT_PRE_FETCH);
        abortNamedApiRequest(REQUEST_ID_NEXT_PRE_FETCH);

        AlbumNextPrefetchState.data = null;
        AlbumNextPrefetchState.loading = false;
        AlbumNextPrefetchState.available = false;
        emitAppEvent(EVENT_NAME_NEXT_PRE_FETCH);
        return;
    }

    AlbumNextPrefetchState.data = null;
    AlbumNextPrefetchState.loading = true;
    AlbumNextPrefetchState.available = false;

    if (AuthController.Locked) {
        return; // Vault is locked
    }

    const mediaId = AlbumMediaPositionStatus.next.id;

    clearNamedTimeout(REQUEST_ID_NEXT_PRE_FETCH);
    makeNamedApiRequest(REQUEST_ID_NEXT_PRE_FETCH, apiMediaGetMedia(mediaId))
        .onSuccess((media) => {
            AlbumNextPrefetchState.data = media;
            AlbumNextPrefetchState.loading = false;
            AlbumNextPrefetchState.available = true;
            emitAppEvent(EVENT_NAME_NEXT_PRE_FETCH);
        })
        .onRequestError((err, handleErr) => {
            handleErr(err, {
                unauthorized: () => {
                    emitAppEvent(EVENT_NAME_UNAUTHORIZED);
                },
                notFound: () => {
                    AlbumNextPrefetchState.data = null;
                    AlbumNextPrefetchState.loading = false;
                    AlbumNextPrefetchState.available = true;
                    emitAppEvent(EVENT_NAME_NEXT_PRE_FETCH);
                },
                temporalError: () => {
                    // Retry
                    setNamedTimeout(REQUEST_ID_NEXT_PRE_FETCH, LOAD_RETRY_DELAY, loadAlbumNextPreFetch);
                },
            });
        })
        .onUnexpectedError((err) => {
            console.error(err);
            // Retry
            setNamedTimeout(REQUEST_ID_NEXT_PRE_FETCH, LOAD_RETRY_DELAY, loadAlbumNextPreFetch);
        });
}

/**
 * Checks if pre-fetch is available
 * @returns True if the pre-fetched data was set, so no load is needed
 */
export function checkAlbumNextPrefetch(): boolean {
    if (AlbumMediaPositionStatus.next === null) {
        return false;
    }

    if (AlbumMediaPositionStatus.next.id !== getCurrentMediaId()) {
        return false;
    }

    if (AlbumNextPrefetchState.available) {
        provideCurrentMediaData(AlbumNextPrefetchState.data);
        return true;
    } else {
        return false;
    }
}

/**
 * Initializes current album state
 */
export function initializeAlbum() {
    addAppEventListener(EVENT_NAME_AUTH_CHANGED, loadCurrentAlbum);
    addAppEventListener(EVENT_NAME_APP_STATUS_CHANGED, onCurrentAlbumChanged);

    CurrentAlbumState.id = AppStatus.CurrentAlbum;

    loadCurrentAlbum();
}
