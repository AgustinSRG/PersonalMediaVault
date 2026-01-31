// Media data controller

"use strict";

import { makeNamedApiRequest, abortNamedApiRequest } from "@asanrom/request-browser";
import { setNamedTimeout, clearNamedTimeout } from "@/utils/named-timeouts";
import { AlbumsController } from "./albums";
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

/**
 * Min duration in seconds to use auto-next, instead of next-end
 */
export const AUTO_LOOP_MIN_DURATION = 3;

/**
 * Number of seconds to wait for next-end
 */
export const NEXT_END_WAIT_DURATION = 8;

const REQUEST_ID = "media-current-load";

/**
 * Management object to fetch media metadata
 */
export class MediaController {
    /**
     * True if the metadata is being loaded from the server.
     */
    public static Loading = true;

    /**
     * ID of the media being fetched
     */
    public static MediaId = -1;

    /**
     * Media metadata fetched from the server
     */
    public static MediaData: MediaData = null;

    /**
     * Media session ID
     */
    public static MediaSessionId = "";

    /**
     * Initialization logic
     */
    public static Initialize() {
        addAppEventListener(EVENT_NAME_AUTH_CHANGED, MediaController.Load);
        addAppEventListener(EVENT_NAME_APP_STATUS_CHANGED, MediaController.OnMediaChanged);

        MediaController.MediaId = AppStatus.CurrentMedia;

        MediaController.Load();
    }

    /**
     * Called when the current media ID changes
     */
    private static OnMediaChanged() {
        if (MediaController.MediaId !== AppStatus.CurrentMedia) {
            MediaController.MediaId = AppStatus.CurrentMedia;
            MediaController.Load();
        }
    }

    /**
     * Loads the current media metadata.
     */
    public static Load() {
        if (MediaController.MediaId < 0) {
            clearNamedTimeout(REQUEST_ID);
            abortNamedApiRequest(REQUEST_ID);

            MediaController.MediaData = null;
            emitAppEvent(EVENT_NAME_MEDIA_UPDATE, null);
            MediaController.Loading = false;
            emitAppEvent(EVENT_NAME_MEDIA_LOADING, false);

            return;
        }

        MediaController.MediaData = null;
        emitAppEvent(EVENT_NAME_MEDIA_UPDATE, null);

        MediaController.Loading = true;
        emitAppEvent(EVENT_NAME_MEDIA_LOADING, true);

        if (AuthController.Locked) {
            return; // Vault is locked
        }

        clearNamedTimeout(REQUEST_ID);
        abortNamedApiRequest(REQUEST_ID);

        if (AlbumsController.CheckAlbumNextPrefetch()) {
            return; // Pre-fetch
        }

        makeNamedApiRequest(REQUEST_ID, apiMediaGetMedia(MediaController.MediaId))
            .onSuccess((media) => {
                MediaController.MediaData = media;
                emitAppEvent(EVENT_NAME_MEDIA_UPDATE, MediaController.MediaData);

                MediaController.Loading = false;
                emitAppEvent(EVENT_NAME_MEDIA_LOADING, false);
            })
            .onRequestError((err, handleErr) => {
                handleErr(err, {
                    unauthorized: () => {
                        emitAppEvent(EVENT_NAME_UNAUTHORIZED);
                    },
                    notFound: () => {
                        MediaController.MediaData = null;
                        emitAppEvent(EVENT_NAME_MEDIA_UPDATE, MediaController.MediaData);

                        MediaController.Loading = false;
                        emitAppEvent(EVENT_NAME_MEDIA_LOADING, false);
                    },
                    temporalError: () => {
                        // Retry
                        setNamedTimeout(REQUEST_ID, 1500, MediaController.Load);
                    },
                });
            })
            .onUnexpectedError((err) => {
                console.error(err);
                // Retry
                setNamedTimeout(REQUEST_ID, 1500, MediaController.Load);
            });
    }

    /**
     * Sets the media data externally, without loading
     * @param media The media data
     */
    public static SetMediaData(media: MediaData) {
        MediaController.MediaData = media;
        emitAppEvent(EVENT_NAME_MEDIA_UPDATE, MediaController.MediaData);

        MediaController.Loading = false;
        emitAppEvent(EVENT_NAME_MEDIA_LOADING, false);
    }
}
