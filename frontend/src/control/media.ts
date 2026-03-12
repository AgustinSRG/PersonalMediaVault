// Media data controller

"use strict";

import { makeNamedApiRequest, abortNamedApiRequest } from "@asanrom/request-browser";
import { setNamedTimeout, clearNamedTimeout } from "@/utils/named-timeouts";
import {
    addAppEventListener,
    emitAppEvent,
    EVENT_NAME_APP_STATUS_CHANGED,
    EVENT_NAME_AUTH_CHANGED,
    EVENT_NAME_MEDIA_LOADING,
    EVENT_NAME_MEDIA_UPDATE,
    EVENT_NAME_UNAUTHORIZED,
} from "./app-events";
import { AppStatus } from "./app-status";
import { AuthController } from "./auth";
import type { MediaData } from "@/api/models";
import { apiMediaGetMedia } from "@/api/api-media";
import { checkAlbumNextPrefetch } from "./album";
import { getUniqueStringId } from "@/utils/unique-id";

/**
 * Min duration in seconds to use auto-next, instead of next-end
 */
export const AUTO_LOOP_MIN_DURATION = 3;

/**
 * Number of seconds to wait for next-end
 */
export const NEXT_END_WAIT_DURATION = 8;

/**
 * Current media state
 */
const CurrentMediaState = {
    /**
     * Current media ID
     */
    id: -1,

    /**
     * Loading state
     */
    loading: true,

    /**
     * Metadata
     */
    data: null as MediaData | null,
};

/**
 * Gets the current media ID
 * @returns The current media ID
 */
export function getCurrentMediaId(): number {
    return CurrentMediaState.id;
}

/**
 * Checks if the current media is being loaded
 * @returns True if loading
 */
export function isCurrentMediaLoading(): boolean {
    return CurrentMediaState.loading;
}

/**
 * Sets the current media loading status
 * @param loading True if loading
 */
function setCurrentMediaLoading(loading: boolean) {
    CurrentMediaState.loading = loading;
    emitAppEvent(EVENT_NAME_MEDIA_LOADING, loading);
}

/**
 * Gets current media metadata
 * @returns The metadata
 */
export function getCurrentMediaData(): MediaData | null {
    return CurrentMediaState.data;
}

/**
 * Sets current media metadata
 * @param data The metadata
 */
function setCurrentMediaData(data: MediaData | null) {
    CurrentMediaState.data = data;
    emitAppEvent(EVENT_NAME_MEDIA_UPDATE, data);
}

// Request ID to load the current media
const REQUEST_ID = getUniqueStringId();

// Delay to retry loading after an error
const LOAD_RETRY_DELAY = 1500;

/**
 * Loads current media
 */
export function loadCurrentMedia() {
    if (CurrentMediaState.id < 0) {
        clearNamedTimeout(REQUEST_ID);
        abortNamedApiRequest(REQUEST_ID);

        setCurrentMediaData(null);

        setCurrentMediaLoading(false);

        return;
    }

    setCurrentMediaData(null);

    setCurrentMediaLoading(true);

    if (AuthController.Locked) {
        return; // Vault is locked
    }

    clearNamedTimeout(REQUEST_ID);
    abortNamedApiRequest(REQUEST_ID);

    if (checkAlbumNextPrefetch()) {
        return; // Pre-fetch
    }

    makeNamedApiRequest(REQUEST_ID, apiMediaGetMedia(CurrentMediaState.id))
        .onSuccess((media) => {
            setCurrentMediaData(media);

            setCurrentMediaLoading(false);
        })
        .onRequestError((err, handleErr) => {
            handleErr(err, {
                unauthorized: () => {
                    emitAppEvent(EVENT_NAME_UNAUTHORIZED);
                },
                notFound: () => {
                    setCurrentMediaData(null);

                    setCurrentMediaLoading(false);
                },
                temporalError: () => {
                    // Retry
                    setNamedTimeout(REQUEST_ID, LOAD_RETRY_DELAY, loadCurrentMedia);
                },
            });
        })
        .onUnexpectedError((err) => {
            console.error(err);
            // Retry
            setNamedTimeout(REQUEST_ID, LOAD_RETRY_DELAY, loadCurrentMedia);
        });
}

/**
 * Called when the current media ID changes
 */
function onMediaChanged() {
    if (CurrentMediaState.id !== AppStatus.CurrentMedia) {
        CurrentMediaState.id = AppStatus.CurrentMedia;
        loadCurrentMedia();
    }
}

/**
 * Provides the media data externally, without loading.
 * Called by pre-fetch or cache services
 * @param media The media data
 */
export function provideCurrentMediaData(media: MediaData) {
    setCurrentMediaData(media);
    setCurrentMediaLoading(false);
}

/**
 * Initializes media loader system
 */
export function initializeMedia() {
    addAppEventListener(EVENT_NAME_AUTH_CHANGED, loadCurrentMedia);
    addAppEventListener(EVENT_NAME_APP_STATUS_CHANGED, onMediaChanged);

    CurrentMediaState.id = AppStatus.CurrentMedia;

    loadCurrentMedia();
}

/**
 * Makes changes to the current media metadata.
 * Condition: The ID must match.
 * @param id Expected media ID
 * @param callback The callback to modify the media metadata
 */
export function modifyCurrentMediaData(id: number, callback: (metadata: MediaData) => void) {
    if (!CurrentMediaState.data || CurrentMediaState.data.id !== id) {
        return;
    }

    callback(CurrentMediaState.data);
}
