// Subtitles controller

"use strict";

import { GetAssetURL, Request } from "@/utils/request";
import { findSubtitlesEntry, parseSRT, SubtitlesEntry } from "@/utils/srt";
import { Timeouts } from "@/utils/timeout";
import { AppEvents } from "./app-events";
import { AppStatus } from "./app-status";
import { MediaController } from "./media";
import { getSelectedSubtitles } from "./player-preferences";
import { AuthController, EVENT_NAME_UNAUTHORIZED } from "./auth";

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
        AuthController.AddChangeEventListener(SubtitlesController.Load);
        AppStatus.AddEventListener(SubtitlesController.OnMediaChanged);
        MediaController.AddUpdateEventListener(SubtitlesController.Load);

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
            Timeouts.Abort(REQUEST_ID_SUBTITLES_LOAD);
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
                SubtitlesController.SubtitlesFileURL = GetAssetURL(sub.url);
                break;
            }
        }

        if (!SubtitlesController.SubtitlesFileURL) {
            Request.Abort(REQUEST_ID_SUBTITLES_LOAD);
            Timeouts.Abort(REQUEST_ID_SUBTITLES_LOAD);
            AppEvents.Emit(EVENT_NAME_SUBTITLES_UPDATE);
            return;
        }

        Timeouts.Abort(REQUEST_ID_SUBTITLES_LOAD);
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
                        Timeouts.Set(REQUEST_ID_SUBTITLES_LOAD, 1500, SubtitlesController.Load);
                    })
                    .handle(err);
            })
            .onUnexpectedError((err) => {
                console.error(err);
                // Retry
                Timeouts.Set(REQUEST_ID_SUBTITLES_LOAD, 1500, SubtitlesController.Load);
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
