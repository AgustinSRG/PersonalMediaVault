// Subtitles controller

"use strict";

import { RequestErrorHandler, abortNamedApiRequest, makeNamedApiRequest } from "@asanrom/request-browser";
import type { SubtitlesEntry } from "@/utils/srt";
import { findSubtitlesEntry, parseSRT } from "@/utils/srt";
import { setNamedTimeout, clearNamedTimeout } from "@/utils/named-timeouts";
import { AppStatus } from "./app-status";
import { MediaController } from "./media";
import { getSelectedSubtitles } from "./player-preferences";
import { getAssetURL } from "@/utils/api";
import {
    addAppEventListener,
    emitAppEvent,
    EVENT_NAME_APP_STATUS_CHANGED,
    EVENT_NAME_AUTH_CHANGED,
    EVENT_NAME_MEDIA_UPDATE,
    EVENT_NAME_SUBTITLES_UPDATE,
    EVENT_NAME_UNAUTHORIZED,
} from "./app-events";

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
        addAppEventListener(EVENT_NAME_AUTH_CHANGED, SubtitlesController.Load);
        addAppEventListener(EVENT_NAME_APP_STATUS_CHANGED, SubtitlesController.OnMediaChanged);
        addAppEventListener(EVENT_NAME_MEDIA_UPDATE, SubtitlesController.Load);

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
            abortNamedApiRequest(REQUEST_ID_SUBTITLES_LOAD);
            clearNamedTimeout(REQUEST_ID_SUBTITLES_LOAD);
            SubtitlesController.SelectedSubtitles = "";
            SubtitlesController.SubtitlesFileURL = "";
            SubtitlesController.Subtitles = [];
            emitAppEvent(EVENT_NAME_SUBTITLES_UPDATE);
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

        if (!SubtitlesController.SelectedSubtitles && prefSubtitles && subtitles.length > 0) {
            SubtitlesController.SelectedSubtitles = subtitles[0].id;
            SubtitlesController.SubtitlesFileURL = getAssetURL(subtitles[0].url);
        }

        if (!SubtitlesController.SubtitlesFileURL) {
            abortNamedApiRequest(REQUEST_ID_SUBTITLES_LOAD);
            clearNamedTimeout(REQUEST_ID_SUBTITLES_LOAD);
            emitAppEvent(EVENT_NAME_SUBTITLES_UPDATE);
            return;
        }

        clearNamedTimeout(REQUEST_ID_SUBTITLES_LOAD);
        makeNamedApiRequest(REQUEST_ID_SUBTITLES_LOAD, {
            method: "GET",
            url: SubtitlesController.SubtitlesFileURL,
        })
            .onSuccess((srtText) => {
                SubtitlesController.Subtitles = parseSRT(srtText);
                emitAppEvent(EVENT_NAME_SUBTITLES_UPDATE);
            })
            .onRequestError((err) => {
                new RequestErrorHandler()
                    .add(401, "*", () => {
                        emitAppEvent(EVENT_NAME_UNAUTHORIZED);
                    })
                    .add(404, "*", () => {
                        SubtitlesController.Subtitles = [];
                        emitAppEvent(EVENT_NAME_SUBTITLES_UPDATE);
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
