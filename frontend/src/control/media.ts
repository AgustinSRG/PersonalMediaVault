// Media data controller

"use strict";

import { MediaAPI } from "@/api/api-media";
import { Request } from "@/utils/request";
import { Timeouts } from "@/utils/timeout";
import { AlbumsController } from "./albums";
import { AppEvents } from "./app-events";
import { AppStatus } from "./app-status";
import { AuthController, EVENT_NAME_UNAUTHORIZED } from "./auth";
import { MediaData } from "@/api/models";

/**
 * Type value for images
 */
export const MEDIA_TYPE_IMAGE = 1;

/**
 * Type value for videos
 */
export const MEDIA_TYPE_VIDEO = 2;

/**
 * Type value for audios
 */
export const MEDIA_TYPE_AUDIO = 3;

/**
 * Min duration in seconds to use auto-next, instead of next-end
 */
export const AUTO_LOOP_MIN_DURATION = 3;

/**
 * Number of seconds to wait for next-end
 */
export const NEXT_END_WAIT_DURATION = 8;

const REQUEST_ID = "media-current-load";

const EVENT_NAME_LOADING = "current-media-loading";
const EVENT_NAME_UPDATE = "current-media-update";

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
     * Initialization logic
     */
    public static Initialize() {
        AuthController.AddChangeEventListener(MediaController.Load);
        AppStatus.AddEventListener(MediaController.OnMediaChanged);

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
            Timeouts.Abort(REQUEST_ID);
            Request.Abort(REQUEST_ID);

            MediaController.MediaData = null;
            AppEvents.Emit(EVENT_NAME_UPDATE, null);
            MediaController.Loading = false;
            AppEvents.Emit(EVENT_NAME_LOADING, false);

            return;
        }

        MediaController.MediaData = null;
        AppEvents.Emit(EVENT_NAME_UPDATE, null);

        MediaController.Loading = true;
        AppEvents.Emit(EVENT_NAME_LOADING, true);

        if (AuthController.Locked) {
            return; // Vault is locked
        }

        Timeouts.Abort(REQUEST_ID);
        Request.Abort(REQUEST_ID);

        if (AlbumsController.CheckAlbumNextPrefetch()) {
            return; // Pre-fetch
        }

        Request.Pending(REQUEST_ID, MediaAPI.GetMedia(MediaController.MediaId))
            .onSuccess((media) => {
                MediaController.MediaData = media;
                AppEvents.Emit(EVENT_NAME_UPDATE, MediaController.MediaData);

                MediaController.Loading = false;
                AppEvents.Emit(EVENT_NAME_LOADING, false);
            })
            .onRequestError((err) => {
                Request.ErrorHandler()
                    .add(401, "*", () => {
                        AppEvents.Emit(EVENT_NAME_UNAUTHORIZED);
                    })
                    .add(404, "*", () => {
                        MediaController.MediaData = null;
                        AppEvents.Emit(EVENT_NAME_UPDATE, MediaController.MediaData);

                        MediaController.Loading = false;
                        AppEvents.Emit(EVENT_NAME_LOADING, false);
                    })
                    .add("*", "*", () => {
                        // Retry
                        Timeouts.Set(REQUEST_ID, 1500, MediaController.Load);
                    })
                    .handle(err);
            })
            .onUnexpectedError((err) => {
                console.error(err);
                // Retry
                Timeouts.Set(REQUEST_ID, 1500, MediaController.Load);
            });
    }

    /**
     * Sets the media data externally, without loading
     * @param media The media data
     */
    public static SetMediaData(media: MediaData) {
        MediaController.MediaData = media;
        AppEvents.Emit(EVENT_NAME_UPDATE, MediaController.MediaData);

        MediaController.Loading = false;
        AppEvents.Emit(EVENT_NAME_LOADING, false);
    }

    /**
     * Counter to make unique pending Ids
     */
    private static NextPendingId = 0;

    /**
     * Gets an unique ID for a request to check the media
     */
    public static GetPendingId() {
        MediaController.NextPendingId++;
        return "pending-check-" + Date.now() + "-" + MediaController.NextPendingId;
    }

    /**
     * Adds event listener to check for metadata updates
     * @param handler Event handler
     */
    public static AddUpdateEventListener(handler: (m: MediaData) => void) {
        AppEvents.AddEventListener(EVENT_NAME_UPDATE, handler);
    }

    /**
     * Removes event listener (Updated event)
     * @param handler Event handler
     */
    public static RemoveUpdateEventListener(handler: (m: MediaData) => void) {
        AppEvents.RemoveEventListener(EVENT_NAME_UPDATE, handler);
    }

    /**
     * Adds event listener to check for request loading status updates
     * @param handler Event handler
     */
    public static AddLoadingEventListener(handler: (loading: boolean) => void) {
        AppEvents.AddEventListener(EVENT_NAME_LOADING, handler);
    }

    /**
     * Removes event listener (Loading event)
     * @param handler Event handler
     */
    public static RemoveLoadingEventListener(handler: (loading: boolean) => void) {
        AppEvents.RemoveEventListener(EVENT_NAME_LOADING, handler);
    }
}
