// Player preferences

"use strict";

import { MediaData } from "@/api/models";
import { clearLocalStorage, fetchFromLocalStorage, fetchFromLocalStorageCache, saveIntoLocalStorage } from "../utils/local-storage";

/**
 * User selected resolution (video)
 */
export interface UserSelectedResolutionVideo {
    original: boolean;
    width: number;
    height: number;
    fps: number;
}

const LS_KEY_USER_SELECTED_RESOLUTION_VIDEO = "player-pref-resolution";

/**
 * Gets user selected resolution (video)
 * @param metadata Media metadata
 * @returns The resolution index
 */
export function getUserSelectedResolutionVideo(metadata: MediaData): number {
    let r = fetchFromLocalStorage(LS_KEY_USER_SELECTED_RESOLUTION_VIDEO, {
        original: true,
        width: 0,
        height: 0,
        fps: 0,
    } as UserSelectedResolutionVideo);

    if (!r || typeof r !== "object") {
        r = {
            original: true,
            width: 0,
            height: 0,
            fps: 0,
        };
    }

    if (r.original || !metadata.resolutions || metadata.resolutions.length === 0) {
        return -1;
    }

    let currentVal = metadata.width * metadata.height * metadata.fps;

    const prefVal = r.width * r.height * r.fps;

    let currenRes = -1;

    for (let i = 0; i < metadata.resolutions.length; i++) {
        const res = metadata.resolutions[i];
        if (!res.ready) {
            continue;
        }
        const resVal = res.width * res.height * res.fps;
        if (Math.abs(resVal - prefVal) < Math.abs(currentVal - prefVal)) {
            currentVal = resVal;
            currenRes = i;
        }
    }

    return currenRes;
}

/**
 * Sets user selected resolution (Video)
 * @param metadata The media metadata
 * @param index The selected resolution index
 */
export function setUserSelectedResolutionVideo(metadata: MediaData, index: number) {
    let r: UserSelectedResolutionVideo;

    if (index < 0) {
        r = {
            original: true,
            width: 0,
            height: 0,
            fps: 0,
        };
    } else if (metadata && metadata.resolutions && metadata.resolutions[index] && metadata.resolutions[index].ready) {
        r = {
            original: false,
            width: metadata.resolutions[index].width,
            height: metadata.resolutions[index].height,
            fps: metadata.resolutions[index].fps,
        };
    }

    saveIntoLocalStorage(LS_KEY_USER_SELECTED_RESOLUTION_VIDEO, r);
}

/**
 * User selected resolution (image)
 */
export interface UserSelectedResolutionImage {
    original: boolean;
    width: number;
    height: number;
}

const LS_KEY_USER_SELECTED_RESOLUTION_IMAGE = "player-pref-resolution-img";

/**
 * Gets user selected resolution (image)
 * @param metadata Media metadata
 * @returns The resolution index
 */
export function getUserSelectedResolutionImage(metadata: MediaData): number {
    let r = fetchFromLocalStorage(LS_KEY_USER_SELECTED_RESOLUTION_IMAGE, {
        original: true,
        width: 0,
        height: 0,
    } as UserSelectedResolutionImage);

    if (!r || typeof r !== "object") {
        r = {
            original: true,
            width: 0,
            height: 0,
        };
    }

    if (r.original || !metadata.resolutions || metadata.resolutions.length === 0) {
        return -1;
    }
    let currentVal = metadata.width * metadata.height;

    const prefVal = r.width * r.height;

    let currenRes = -1;

    for (let i = 0; i < metadata.resolutions.length; i++) {
        const res = metadata.resolutions[i];
        if (!res.ready) {
            continue;
        }
        const resVal = res.width * res.height;
        if (Math.abs(resVal - prefVal) < Math.abs(currentVal - prefVal)) {
            currentVal = resVal;
            currenRes = i;
        }
    }

    return currenRes;
}

/**
 * Sets user selected resolution (image)
 * @param metadata Media metadata
 * @param index The selected resolution index
 */
export function setUserSelectedResolutionImage(metadata: MediaData, index: number) {
    let r: UserSelectedResolutionImage;
    if (index < 0) {
        r = {
            original: true,
            width: 0,
            height: 0,
        };
    } else if (metadata && metadata.resolutions && metadata.resolutions[index] && metadata.resolutions[index].ready) {
        r = {
            original: false,
            width: metadata.resolutions[index].width,
            height: metadata.resolutions[index].height,
        };
    }

    saveIntoLocalStorage(LS_KEY_USER_SELECTED_RESOLUTION_IMAGE, r);
}

const MAX_CACHE_PLAY_TIME_SIZE = 100;
const LS_KEY_PLAY_TIME_CACHE = "player-play-time-cache";

interface PlayTimeCacheEntry {
    mid: number;
    time: number;
}

/**
 * Gets cached initial time
 * @param mid The media ID
 * @returns The initial time in seconds
 */
export function getCachedInitialTime(mid: number): number {
    let cache = fetchFromLocalStorage(LS_KEY_PLAY_TIME_CACHE, [] as PlayTimeCacheEntry[]);

    if (!cache || !Array.isArray(cache)) {
        cache = [];
    }

    for (const entry of cache) {
        if (!entry || typeof entry !== "object") {
            continue;
        }

        if (entry.mid === mid) {
            const time = entry.time;
            if (typeof time === "number" && !isNaN(time) && isFinite(time) && time >= 0) {
                return time;
            } else {
                return 0;
            }
        }
    }

    return 0;
}

/**
 * Delay to save the current time inside the cache (milliseconds)
 */
export const CURRENT_TIME_UPDATE_DELAY = 2000;

/**
 * Sets cached initial time
 * @param mid The media ID
 * @param time The cached current time
 */
export function setCachedInitialTime(mid: number, time: number) {
    let cache = fetchFromLocalStorage(LS_KEY_PLAY_TIME_CACHE, [] as PlayTimeCacheEntry[]);

    if (!cache || !Array.isArray(cache)) {
        cache = [];
    }

    // Remove elements
    cache = cache.filter((e) => {
        if (!e || typeof e !== "object") {
            return false;
        }

        return e.mid !== mid;
    });

    while (cache.length >= MAX_CACHE_PLAY_TIME_SIZE) {
        cache.shift();
    }

    // Add

    cache.push({
        mid: mid,
        time: time,
    });

    saveIntoLocalStorage(LS_KEY_PLAY_TIME_CACHE, cache);
}

/**
 * Clears cached initial times
 */
export function clearCachedTimes() {
    clearLocalStorage(LS_KEY_PLAY_TIME_CACHE);
}

const MAX_CACHE_ALBUM_POS_SIZE = 100;
const LS_KEY_ALBUM_POS_CACHE = "player-album-pos-cache";

interface AlbumPositionCacheEntry {
    id: number;
    pos: number;
}

/**
 * Gets cached current album position
 * @param id Album ID
 * @returns Current cached position
 */
export function getCachedAlbumPosition(id: number): number {
    let cache = fetchFromLocalStorage(LS_KEY_ALBUM_POS_CACHE, [] as AlbumPositionCacheEntry[]);

    if (!cache || !Array.isArray(cache)) {
        cache = [];
    }

    for (const entry of cache) {
        if (!entry || typeof entry !== "object") {
            continue;
        }

        if (entry.id === id) {
            const pos = entry.pos;
            if (typeof pos === "number" && !isNaN(pos) && isFinite(pos) && pos >= 0) {
                return pos;
            } else {
                return 0;
            }
        }
    }

    return 0;
}

/**
 * Sets cached current album position
 * @param id Album ID
 * @param pos Current cached position
 */
export function setCachedAlbumPosition(id: number, pos: number) {
    let cache = fetchFromLocalStorage(LS_KEY_ALBUM_POS_CACHE, [] as AlbumPositionCacheEntry[]);

    if (!cache || !Array.isArray(cache)) {
        cache = [];
    }

    // Remove elements
    cache = cache.filter((e) => {
        if (!e || typeof e !== "object") {
            return false;
        }

        return e.id !== id;
    });

    while (cache.length >= MAX_CACHE_ALBUM_POS_SIZE) {
        cache.shift();
    }

    cache.push({
        id: id,
        pos: pos,
    });

    saveIntoLocalStorage(LS_KEY_ALBUM_POS_CACHE, cache);
}

/**
 * Clears cached initial times
 */
export function clearCachedAlbumPositions() {
    clearLocalStorage(LS_KEY_ALBUM_POS_CACHE);
}

const LS_KEY_VOLUME = "player-pref-volume";

/**
 * Gets player volume
 * @returns The volume (0 - 1)
 */
export function getPlayerVolume(): number {
    return Number(fetchFromLocalStorageCache(LS_KEY_VOLUME, 1)) || 0;
}

/**
 * Sets player volume
 * @param volume The volume (0 - 1)
 */
export function setPlayerVolume(volume: number) {
    saveIntoLocalStorage(LS_KEY_VOLUME, volume);
}

const LS_KEY_MUTED = "player-pref-muted";

/**
 * Get player muted flag
 * @returns The muted flag
 */
export function getPlayerMuted(): boolean {
    return !!fetchFromLocalStorageCache(LS_KEY_MUTED, false);
}

/**
 * Sets player muted flag
 * @param muted The muted flag
 */
export function setPlayerMuted(muted: boolean) {
    saveIntoLocalStorage(LS_KEY_MUTED, muted);
}

const LS_KEY_SCALE = "player-pref-scale";

/**
 * Gets image scale
 * @returns The image scale
 */
export function getImageScale(): number {
    return Number(fetchFromLocalStorageCache(LS_KEY_SCALE, 0)) || 0;
}

/**
 * Sets image scale
 * @param scale The image scale
 */
export function setImageScale(scale: number) {
    saveIntoLocalStorage(LS_KEY_SCALE, scale);
}

const LS_KEY_IMAGE_FIT = "player-pref-fit";

/**
 * Gets image fit flag
 * @returns The image fit flag
 */
export function getImageFit(): boolean {
    return !!fetchFromLocalStorageCache(LS_KEY_IMAGE_FIT, true);
}

/**
 * Sets image fit flag
 * @param fit The image fit flag
 */
export function setImageFit(fit: boolean) {
    saveIntoLocalStorage(LS_KEY_IMAGE_FIT, fit);
}

const LS_KEY_AUDIO_ANIMATION_STYLE = "player-pref-audio-anim";

/**
 * Gets selected audio animation style
 * @returns The animation style name
 */
export function getAudioAnimationStyle(): string {
    return fetchFromLocalStorageCache(LS_KEY_AUDIO_ANIMATION_STYLE, "gradient") + "";
}

/**
 * Sets selected audio animation style
 * @param style The animation style name
 */
export function setAudioAnimationStyle(style: string) {
    saveIntoLocalStorage(LS_KEY_AUDIO_ANIMATION_STYLE, style);
}

const LS_KEY_IMAGE_BACKGROUND = "player-pref-img-bg";

/**
 * Get selected image background style
 * @returns The background style name
 */
export function getImageBackgroundStyle(): string {
    return fetchFromLocalStorageCache(LS_KEY_IMAGE_BACKGROUND, "default") + "";
}

/**
 * Set selected image background style
 * @param style The background style name
 */
export function setImageBackgroundStyle(style: string) {
    saveIntoLocalStorage(LS_KEY_IMAGE_BACKGROUND, style);
}

const LS_KEY_AUTO_NEXT_ON_END = "player-pref-next-end";

/**
 * Gets auto next option on media ending
 * @returns Auto next flag
 */
export function getAutoNextOnEnd(): boolean {
    return !!fetchFromLocalStorageCache(LS_KEY_AUTO_NEXT_ON_END, true);
}

/**
 * Sets auto next option on media ending
 * @param autoNext Auto next flag
 */
export function setAutoNextOnEnd(autoNext: boolean) {
    saveIntoLocalStorage(LS_KEY_AUTO_NEXT_ON_END, autoNext);
}

const LS_KEY_AUTO_NEXT_PAGE_DELAY = "player-pref-next-end-delay";

/**
 * Gets option to wait several seconds before playing the next media on pages
 * @returns True for delay on auto next for pages
 */
export function getAutoNextPageDelay(): boolean {
    return !!fetchFromLocalStorageCache(LS_KEY_AUTO_NEXT_PAGE_DELAY, true);
}

/**
 * Sets option to wait several seconds before playing the next media on pages
 * @param autoNextDelay True for delay on auto next for pages
 */
export function setAutoNextPageDelay(autoNextDelay: boolean) {
    saveIntoLocalStorage(LS_KEY_AUTO_NEXT_PAGE_DELAY, autoNextDelay);
}

const LS_KEY_AUTO_NEXT_TIME = "player-pref-img-auto-next";

/**
 * Gets auto next option for images or short videos
 * @returns Number of seconds to wait for auto next, 0 = disabled
 */
export function getAutoNextTime(): number {
    return Number(fetchFromLocalStorageCache(LS_KEY_AUTO_NEXT_TIME, 0)) || 0;
}

/**
 * Sets auto next option for images or short videos
 * @param autoNextSeconds Number of seconds to wait for auto next, 0 = disabled
 */
export function setAutoNextTime(autoNextSeconds: number) {
    saveIntoLocalStorage(LS_KEY_AUTO_NEXT_TIME, autoNextSeconds);
}

const LS_KEY_IMAGE_NOTES_VISIBLE = "player-pref-img-notes-v";

/**
 * Gets image notes visibility
 * @returns Image notes visibility
 */
export function getImageNotesVisible(): boolean {
    return !!fetchFromLocalStorageCache(LS_KEY_IMAGE_NOTES_VISIBLE, true);
}

/**
 * Sets image notes visibility
 * @param visible Image notes visibility
 */
export function setImageNotesVisible(visible: boolean) {
    saveIntoLocalStorage(LS_KEY_IMAGE_NOTES_VISIBLE, visible);
}

const LS_KEY_SUBTITLES_SELECTED = "player-pref-subtitles";

/**
 * Gets selected subtitles
 * @returns Selected subtitles ID
 */
export function getSelectedSubtitles(): string {
    return fetchFromLocalStorageCache(LS_KEY_SUBTITLES_SELECTED, "") + "";
}

/**
 * Sets selected subtitles
 * @param sub Selected subtitles ID
 */
export function setSelectedSubtitles(sub: string) {
    saveIntoLocalStorage(LS_KEY_SUBTITLES_SELECTED, sub);
}

/**
 * Event triggered when the subtitles options change
 */
export const EVENT_NAME_SUBTITLES_OPTIONS_CHANGED = "subtitles-options-changed";

/**
 * Subtitles options
 */
export interface SubtitlesOptions {
    /**
     * Labeled size (s, m, l, xl, xll) - Default l
     */
    size: string;

    /**
     * Custom subtitles size (percent, 100 = 100%)
     */
    customSize: number;

    /**
     * Allow colors from SRT?
     */
    allowColors: boolean;

    /**
     * Allow line breaks?
     */
    allowLineBreaks: boolean;

    /**
     * Background opacity style (100, 75, 50, 25, 0)
     */
    bg: string;

    /**
     * Subtitles position (top, bottom)
     */
    pos: string;

    /**
     * Margin (pixels)
     */
    margin: number;
}

const LS_KEY_SUBTITLES_OPTIONS = "player-pref-subtitles-options";

/**
 * Gets subtitles style options
 * @returns The configured subtitles options
 */
export function getSubtitlesOptions(): SubtitlesOptions {
    const result: SubtitlesOptions = {
        size: "l",
        customSize: 150,
        allowColors: true,
        allowLineBreaks: true,
        bg: "0",
        pos: "bottom",
        margin: 0,
    };

    const parsedOptions: unknown = fetchFromLocalStorageCache(LS_KEY_SUBTITLES_OPTIONS, null);

    if (!parsedOptions || typeof parsedOptions !== "object") {
        return result;
    }

    if ("size" in parsedOptions && typeof parsedOptions.size === "string") {
        result.size = parsedOptions.size;
    }

    if ("customSize" in parsedOptions && typeof parsedOptions.customSize === "number" && !isNaN(parsedOptions.customSize)) {
        result.customSize = parsedOptions.customSize;
    }

    if ("allowColors" in parsedOptions && typeof parsedOptions.allowColors === "boolean") {
        result.allowColors = parsedOptions.allowColors;
    }

    if ("allowLineBreaks" in parsedOptions && typeof parsedOptions.allowLineBreaks === "boolean") {
        result.allowLineBreaks = parsedOptions.allowLineBreaks;
    }

    if ("bg" in parsedOptions && typeof parsedOptions.bg === "string") {
        result.bg = parsedOptions.bg;
    }

    if ("pos" in parsedOptions && typeof parsedOptions.pos === "string") {
        result.pos = parsedOptions.pos;
    }

    if ("margin" in parsedOptions && typeof parsedOptions.margin === "number") {
        result.margin = parsedOptions.margin;
    }

    return result;
}

/**
 * Sets the subtitles style options
 * @param options The subtitles options
 */
export function setSubtitlesOptions(options: SubtitlesOptions) {
    saveIntoLocalStorage(LS_KEY_SUBTITLES_OPTIONS, options);
}

const LS_KEY_AUDIO_TRACK = "player-pref-audio-track";

/**
 * Gets selected audio track
 * @returns The selected audio track
 */
export function getSelectedAudioTrack(): string {
    return fetchFromLocalStorageCache(LS_KEY_AUDIO_TRACK, "") + "";
}

/**
 * Sets selected audio track
 * @returns The selected audio track
 */
export function setSelectedAudioTrack(track: string) {
    saveIntoLocalStorage(LS_KEY_AUDIO_TRACK, track);
}

const LS_KEY_TOGGLE_PLAY_DELAY = "player-pref-toggle-delay";

/**
 * Gets toggle play delay
 * @returns Toggle play delay (ms)
 */
export function getTogglePlayDelay(): number {
    return Number(fetchFromLocalStorageCache(LS_KEY_TOGGLE_PLAY_DELAY, 250)) || 0;
}

/**
 * Sets toggle play delay
 * @param delay Toggle play delay (ms)
 */
export function setTogglePlayDelay(delay: number) {
    saveIntoLocalStorage(LS_KEY_TOGGLE_PLAY_DELAY, delay);
}

const LS_KEY_EXTENDED_DESCRIPTION_SIZE = "player-pref-ext-desc-font-size";

/**
 * Gets extended description size
 * @returns The extended description font size
 */
export function getExtendedDescriptionSize(): number {
    return Number(fetchFromLocalStorageCache(LS_KEY_EXTENDED_DESCRIPTION_SIZE, 18)) || 18;
}

/**
 * Sets extended description size
 * @param size The extended description font size
 */
export function setExtendedDescriptionSize(size: number) {
    saveIntoLocalStorage(LS_KEY_EXTENDED_DESCRIPTION_SIZE, size);
}

const LS_KEY_SHOW_AUDIO_TITLE = "player-pref-audio-title";

/**
 * Gets the option to show audio title
 * @returns True of false
 */
export function getShowAudioTitle(): boolean {
    return !!fetchFromLocalStorageCache(LS_KEY_SHOW_AUDIO_TITLE, true);
}

/**
 * Sets option to show audio title
 * @param show True or false
 */
export function setShowAudioTitle(show: boolean) {
    saveIntoLocalStorage(LS_KEY_SHOW_AUDIO_TITLE, show);
}

const LS_KEY_SHOW_AUDIO_THUMBNAIL = "player-pref-audio-thumbnail";

/**
 * Gets the option to show audio thumbnail
 * @returns True of false
 */
export function getShowAudioThumbnail(): boolean {
    return !!fetchFromLocalStorageCache(LS_KEY_SHOW_AUDIO_THUMBNAIL, true);
}

/**
 * Sets option to show audio thumbnail
 * @param show True or false
 */
export function setShowAudioThumbnail(show: boolean) {
    saveIntoLocalStorage(LS_KEY_SHOW_AUDIO_THUMBNAIL, show);
}

/**
 * Clears player preferences
 */
export function clearPlayerPreferences() {
    clearLocalStorage(LS_KEY_USER_SELECTED_RESOLUTION_VIDEO);
    clearLocalStorage(LS_KEY_USER_SELECTED_RESOLUTION_IMAGE);

    clearLocalStorage(LS_KEY_VOLUME);
    clearLocalStorage(LS_KEY_MUTED);

    clearLocalStorage(LS_KEY_SCALE);
    clearLocalStorage(LS_KEY_IMAGE_FIT);

    clearLocalStorage(LS_KEY_AUDIO_ANIMATION_STYLE);

    clearLocalStorage(LS_KEY_IMAGE_BACKGROUND);

    clearLocalStorage(LS_KEY_AUTO_NEXT_ON_END);
    clearLocalStorage(LS_KEY_AUTO_NEXT_PAGE_DELAY);
    clearLocalStorage(LS_KEY_AUTO_NEXT_TIME);

    clearLocalStorage(LS_KEY_IMAGE_NOTES_VISIBLE);

    clearLocalStorage(LS_KEY_SUBTITLES_SELECTED);
    clearLocalStorage(LS_KEY_SUBTITLES_OPTIONS);

    clearLocalStorage(LS_KEY_AUDIO_TRACK);

    clearLocalStorage(LS_KEY_TOGGLE_PLAY_DELAY);

    clearLocalStorage(LS_KEY_EXTENDED_DESCRIPTION_SIZE);

    clearLocalStorage(LS_KEY_SHOW_AUDIO_TITLE);
    clearLocalStorage(LS_KEY_SHOW_AUDIO_THUMBNAIL);
}
