// Player config utils

"use strict";

import type { MediaData } from "@/api/models";
import { MEDIA_TYPE_IMAGE } from "@/api/models";

/**
 * Renders value percent
 * @param v The value
 * @param $t The translation function
 * @returns The rendered value
 */
export function renderValuePercent(v: number, $t: (k: string) => string) {
    if (v > 1) {
        return Math.floor(v * 100) + "%";
    } else if (v < 1) {
        return Math.floor(v * 100) + "%";
    } else {
        return $t("Normal");
    }
}

/**
 * Renders playback speed
 * @param v The value
 * @param $t The translation function
 * @returns The rendered value
 */
export function renderSpeed(v: number, $t: (k: string) => string) {
    return renderValuePercent(v, $t);
}

/**
 * Renders video scale
 * @param v The value
 * @param $t The translation function
 * @returns The rendered value
 */
export function renderScale(v: number, $t: (k: string) => string) {
    return renderValuePercent(v, $t);
}

/**
 * Renders image background
 * @param b The value
 * @param $t The translation function
 * @returns The rendered value
 */
export function renderBackground(b: string, $t: (k: string) => string): string {
    switch (b) {
        case "white":
            return $t("White");
        case "black":
            return $t("Black");
        default:
            return $t("Default (Theme)");
    }
}

/**
 * Renders auto-next delay
 * @param s The value
 * @param $t The translation function
 * @returns The rendered value
 */
export function renderAutoNext(s: number, $t: (k: string) => string): string {
    if (!isNaN(s) && isFinite(s) && s > 0) {
        if (s === 1) {
            return s + " " + $t("second");
        } else {
            return s + " " + $t("seconds");
        }
    } else {
        return $t("Disabled");
    }
}

/**
 * Renders image resolution
 * @param metadata The metadata
 * @param res The resolution index
 * @param $t The translation function
 * @returns The rendered value
 */
export function renderImageResolution(metadata: MediaData, res: number, $t: (k: string) => string): string {
    if (res < 0) {
        return metadata.width + "x" + metadata.height + " (" + $t("Original") + ")" + (metadata.encoded ? "" : " (" + $t("Pending") + ")");
    } else {
        const resData = metadata.resolutions[res];

        let width = metadata.width;
        let height = metadata.height;

        if (width > height) {
            const proportionalHeight = Math.round((height * resData.width) / width);

            if (proportionalHeight > resData.height) {
                width = Math.round((width * resData.height) / height);
                height = resData.height;
            } else {
                width = resData.width;
                height = proportionalHeight;
            }
        } else {
            const proportionalWidth = Math.round((width * resData.height) / height);

            if (proportionalWidth > resData.width) {
                height = Math.round((height * resData.width) / width);
                width = resData.width;
            } else {
                width = proportionalWidth;
                height = resData.height;
            }
        }

        if (resData) {
            return width + "x" + height + "" + (resData.ready ? "" : " (" + $t("Pending") + ")");
        } else {
            return $t("Unknown");
        }
    }
}

/**
 * Renders video resolution
 * @param metadata The metadata
 * @param res The resolution index
 * @param $t The translation function
 * @returns The rendered value
 */
export function renderVideoResolution(metadata: MediaData, res: number, $t: (k: string) => string): string {
    if (res < 0) {
        return (
            metadata.width +
            "x" +
            metadata.height +
            ", " +
            metadata.fps +
            " fps (" +
            $t("Original") +
            ")" +
            (metadata.encoded ? "" : " (" + $t("Pending") + ")")
        );
    } else {
        const resData = metadata.resolutions[res];
        if (resData) {
            return resData.width + "x" + resData.height + ", " + resData.fps + " fps " + (resData.ready ? "" : " (" + $t("Pending") + ")");
        } else {
            return $t("Unknown");
        }
    }
}

/**
 * Renders resolution
 * @param metadata The metadata
 * @param res The resolution index
 * @param rTick The reload tick
 * @param $t The translation function
 * @returns The rendered value
 */
export function renderResolution(metadata: MediaData, res: number, rTick: number, $t: (k: string) => string): string {
    if (rTick < 0 || !metadata) {
        return $t("Unknown");
    }

    if (metadata.type === MEDIA_TYPE_IMAGE) {
        return renderImageResolution(metadata, res, $t);
    } else {
        return renderImageResolution(metadata, res, $t);
    }
}

/**
 * Renders animation style for audio player
 * @param s The style
 * @param $t The translation function
 * @returns The rendered value
 */
export function renderAnimStyle(s: string, $t: (k: string) => string): string {
    switch (s) {
        case "gradient":
            return $t("Gradient");
        case "rainbow":
            return $t("Rainbow");
        case "none":
            return $t("None");
        default:
            return $t("Monochrome");
    }
}

/**
 * Renders audio track
 * @param metadata The metadata
 * @param audioId The audio track ID
 * @param rTick The reload tick
 * @param $t The translation function
 * @returns The rendered value
 */
export function renderAudio(metadata: MediaData, audioId: string, rTick: number, $t: (k: string) => string): string {
    if (rTick < 0 || !metadata || !metadata.audios || !audioId) {
        return "(" + $t("From video") + ")";
    }

    for (const aud of metadata.audios) {
        if (aud.id === audioId) {
            return aud.name;
        }
    }

    return "(" + $t("From video") + ")";
}

/**
 * Renders subtitles
 * @param metadata The metadata
 * @param subId The subtitles ID
 * @param rTick The reload tick
 * @param $t The translation function
 * @returns The rendered value
 */
export function renderSubtitle(metadata: MediaData, subId: string, rTick: number, $t: (k: string) => string): string {
    if (rTick < 0 || !metadata || !metadata.subtitles || !subId) {
        return $t("No subtitles");
    }

    for (const sub of metadata.subtitles) {
        if (sub.id === subId) {
            return sub.name;
        }
    }

    if (subId && metadata.subtitles.length > 0) {
        return metadata.subtitles[0].name;
    }

    return $t("No subtitles");
}

/**
 * Renders toggle delay
 * @param d The delay
 * @param $t The translation function
 * @returns The rendered value
 */
export function renderToggleDelay(d: number, $t: (k: string) => string): string {
    switch (d) {
        case 0:
            return $t("No delay");
        case 250:
            return "0.25 s";
        case 500:
            return "0.5 s";
        case 750:
            return "0.75 s";
        case 1000:
            return "1 s";
        default:
            return "" + d;
    }
}
