// Subtitles controller

"use strict";

import { RequestErrorHandler, abortNamedApiRequest, makeNamedApiRequest } from "@asanrom/request-browser";
import type { SubtitlesEntry } from "@/utils/srt";
import { findSubtitlesEntry, parseSRT } from "@/utils/srt";
import { setNamedTimeout, clearNamedTimeout } from "@/utils/named-timeouts";
import { AppStatus } from "./app-status";
import { getSelectedSubtitles } from "./player-preferences";
import { getAssetURL } from "@/utils/api";
import {
    addAppEventListener,
    emitAppEvent,
    EVENT_NAME_APP_STATUS_CHANGED,
    EVENT_NAME_AUTH_CHANGED,
    EVENT_NAME_MEDIA_UPDATE,
    EVENT_NAME_SUBTITLES_CHANGED,
    EVENT_NAME_SUBTITLES_UPDATE,
    EVENT_NAME_UNAUTHORIZED,
} from "./app-events";
import { getUniqueStringId } from "@/utils/unique-id";
import { getCurrentMediaData } from "./media";
import { LOAD_RETRY_DELAY } from "@/constants";

// Request ID for loading
const REQUEST_ID_SUBTITLES_LOAD = getUniqueStringId();

/**
 * Current status of subtitles
 */
const SubtitlesStatus = {
    /**
     * ID of the current media
     */
    mediaId: AppStatus.CurrentMedia,

    /**
     * ID of the selected subtitles
     */
    selectedSubtitles: "",

    /**
     * URL of the SubRip file to download
     */
    url: "",

    /**List of subtitles entries, after being parsed
     *
     */
    subtitles: [] as SubtitlesEntry[],
};

/**
 * Gets subtitles line by time
 * @param time The current time
 * @returns The subtitles entry
 */
export function getSubtitlesLine(time: number): SubtitlesEntry {
    return findSubtitlesEntry(SubtitlesStatus.subtitles, time);
}

/**
 * Loads current subtitles file
 */
function loadSubtitles() {
    const mediaData = getCurrentMediaData();

    if (!mediaData) {
        abortNamedApiRequest(REQUEST_ID_SUBTITLES_LOAD);
        clearNamedTimeout(REQUEST_ID_SUBTITLES_LOAD);
        SubtitlesStatus.selectedSubtitles = "";
        SubtitlesStatus.url = "";
        SubtitlesStatus.subtitles = [];
        emitAppEvent(EVENT_NAME_SUBTITLES_UPDATE);
        return;
    }

    const subtitles = mediaData.subtitles || [];
    const prefSubtitles = getSelectedSubtitles();

    SubtitlesStatus.selectedSubtitles = "";
    SubtitlesStatus.url = "";
    SubtitlesStatus.subtitles = [];

    for (const sub of subtitles) {
        if (sub.id === prefSubtitles) {
            SubtitlesStatus.selectedSubtitles = sub.id;
            SubtitlesStatus.url = getAssetURL(sub.url);
            break;
        }
    }

    if (!SubtitlesStatus.selectedSubtitles && prefSubtitles && subtitles.length > 0) {
        SubtitlesStatus.selectedSubtitles = subtitles[0].id;
        SubtitlesStatus.url = getAssetURL(subtitles[0].url);
    }

    if (!SubtitlesStatus.url) {
        abortNamedApiRequest(REQUEST_ID_SUBTITLES_LOAD);
        clearNamedTimeout(REQUEST_ID_SUBTITLES_LOAD);
        emitAppEvent(EVENT_NAME_SUBTITLES_UPDATE);
        return;
    }

    clearNamedTimeout(REQUEST_ID_SUBTITLES_LOAD);
    makeNamedApiRequest(REQUEST_ID_SUBTITLES_LOAD, {
        method: "GET",
        url: SubtitlesStatus.url,
    })
        .onSuccess((srtText) => {
            SubtitlesStatus.subtitles = parseSRT(srtText);
            emitAppEvent(EVENT_NAME_SUBTITLES_UPDATE);
        })
        .onRequestError((err) => {
            new RequestErrorHandler()
                .add(401, "*", () => {
                    emitAppEvent(EVENT_NAME_UNAUTHORIZED);
                })
                .add(404, "*", () => {
                    SubtitlesStatus.subtitles = [];
                    emitAppEvent(EVENT_NAME_SUBTITLES_UPDATE);
                })
                .add("*", "*", () => {
                    // Retry
                    setNamedTimeout(REQUEST_ID_SUBTITLES_LOAD, LOAD_RETRY_DELAY, loadSubtitles);
                })
                .handle(err);
        })
        .onUnexpectedError((err) => {
            console.error(err);
            // Retry
            setNamedTimeout(REQUEST_ID_SUBTITLES_LOAD, LOAD_RETRY_DELAY, loadSubtitles);
        });
}

// Load when auth status changes
addAppEventListener(EVENT_NAME_AUTH_CHANGED, loadSubtitles);

// Load when current media changes
addAppEventListener(EVENT_NAME_MEDIA_UPDATE, loadSubtitles);

// Load when the app status changes if current media ID changes
addAppEventListener(EVENT_NAME_APP_STATUS_CHANGED, () => {
    if (SubtitlesStatus.mediaId !== AppStatus.CurrentMedia) {
        SubtitlesStatus.mediaId = AppStatus.CurrentMedia;
        SubtitlesStatus.selectedSubtitles = "";
        SubtitlesStatus.url = "";
        SubtitlesStatus.subtitles = [];

        loadSubtitles();
    }
});

addAppEventListener(EVENT_NAME_SUBTITLES_CHANGED, (sub) => {
    if (SubtitlesStatus.selectedSubtitles !== sub) {
        loadSubtitles();
    }
});

// Initially load the subtitles
loadSubtitles();
