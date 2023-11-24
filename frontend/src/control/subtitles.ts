// Subtitles controller

"use strict";

import { Request } from "@asanrom/request-browser";
import { findSubtitlesEntry, parseSRT, SubtitlesEntry } from "@/utils/srt";
import { setNamedTimeout, clearNamedTimeout } from "@/utils/named-timeouts";
import { AppEvents } from "./app-events";
import { AppStatus, EVENT_NAME_APP_STATUS_CHANGED } from "./app-status";
import { EVENT_NAME_MEDIA_UPDATE, MediaController } from "./media";
import { getSelectedSubtitles } from "./player-preferences";
import { EVENT_NAME_AUTH_CHANGED, EVENT_NAME_UNAUTHORIZED } from "./auth";
import { getAssetURL } from "@/utils/api";

/**
 * Event triggered when a subtitles file is loaded
 */
export const EVENT_NAME_SUBTITLES_UPDATE = "subtitles-update";

const REQUEST_ID_SUBTITLES_LOAD = "subtitles-load";

/**
 * Management object to load subtitles.
 */
export class SubtitlesController {
    /**
     * Media ID owner of the subtitles being loaded
     */
    public static MediaId = -1;

    /**
     * ID of the selected subtitles
     */
    public static SelectedSubtitles = "";

    /**
     * URL of the SubRip file to download
     */
    public static SubtitlesFileURL = "";

    /**
     * List of subtitles entries, after being parsed
     */
    public static Subtitles: SubtitlesEntry[] = [];

    /**
     * Initialization logic
     */
    public static Initialize() {
        AppEvents.AddEventListener(EVENT_NAME_AUTH_CHANGED, SubtitlesController.Load);
        AppEvents.AddEventListener(EVENT_NAME_APP_STATUS_CHANGED, SubtitlesController.OnMediaChanged);
        AppEvents.AddEventListener(EVENT_NAME_MEDIA_UPDATE, SubtitlesController.Load);

        SubtitlesController.MediaId = AppStatus.CurrentMedia;

        SubtitlesController.Load();
    }

    /**
     * Called when the app status changed, in order to check if the current media changed
     */
    private static OnMediaChanged() {
        if (SubtitlesController.MediaId !== AppStatus.CurrentMedia) {
            SubtitlesController.MediaId = AppStatus.CurrentMedia;
            SubtitlesController.SelectedSubtitles = "";
            SubtitlesController.SubtitlesFileURL = "";
            SubtitlesController.Subtitles = [];
            SubtitlesController.Load();
        }
    }

    /**
     * Loads the subtitles
     */
    public static Load() {
        if (!MediaController.MediaData) {
            Request.Abort(REQUEST_ID_SUBTITLES_LOAD);
            clearNamedTimeout(REQUEST_ID_SUBTITLES_LOAD);
            SubtitlesController.SelectedSubtitles = "";
            SubtitlesController.SubtitlesFileURL = "";
            SubtitlesController.Subtitles = [];
            AppEvents.Emit(EVENT_NAME_SUBTITLES_UPDATE);
            return;
        }

        const subtitles = MediaController.MediaData.subtitles || [];
        const prefSubtitles = getSelectedSubtitles();

        SubtitlesController.SelectedSubtitles = "";
        SubtitlesController.SubtitlesFileURL = "";
        SubtitlesController.Subtitles = [];

        for (const sub of subtitles) {
            if (sub.id === prefSubtitles) {
                SubtitlesController.SelectedSubtitles = sub.id;
                SubtitlesController.SubtitlesFileURL = getAssetURL(sub.url);
                break;
            }
        }

        if (!SubtitlesController.SubtitlesFileURL) {
            Request.Abort(REQUEST_ID_SUBTITLES_LOAD);
            clearNamedTimeout(REQUEST_ID_SUBTITLES_LOAD);
            AppEvents.Emit(EVENT_NAME_SUBTITLES_UPDATE);
            return;
        }

        clearNamedTimeout(REQUEST_ID_SUBTITLES_LOAD);
        Request.Pending(REQUEST_ID_SUBTITLES_LOAD, {
            method: "GET",
            url: SubtitlesController.SubtitlesFileURL,
        })
            .onSuccess((srtText) => {
                SubtitlesController.Subtitles = parseSRT(srtText);
                AppEvents.Emit(EVENT_NAME_SUBTITLES_UPDATE);
            })
            .onRequestError((err) => {
                Request.ErrorHandler()
                    .add(401, "*", () => {
                        AppEvents.Emit(EVENT_NAME_UNAUTHORIZED);
                    })
                    .add(404, "*", () => {
                        SubtitlesController.Subtitles = [];
                        AppEvents.Emit(EVENT_NAME_SUBTITLES_UPDATE);
                    })
                    .add("*", "*", () => {
                        // Retry
                        setNamedTimeout(REQUEST_ID_SUBTITLES_LOAD, 1500, SubtitlesController.Load);
                    })
                    .handle(err);
            })
            .onUnexpectedError((err) => {
                console.error(err);
                // Retry
                setNamedTimeout(REQUEST_ID_SUBTITLES_LOAD, 1500, SubtitlesController.Load);
            });
    }

    /**
     * Changes current subtitles
     * @param sub The current subtitles ID
     */
    public static OnSubtitlesChanged(sub: string) {
        if (SubtitlesController.SelectedSubtitles !== sub) {
            SubtitlesController.Load();
        }
    }

    /**
     * Gets subtitles line by time
     * @param time The current time
     * @returns The subtitles entry
     */
    public static GetSubtitlesLine(time: number): SubtitlesEntry {
        return findSubtitlesEntry(SubtitlesController.Subtitles, time);
    }
}

SubtitlesController.Initialize();
