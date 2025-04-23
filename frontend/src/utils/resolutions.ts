// Video and image standard resolutions

"use strict";

/**
 * Standard video resolution
 */
export type VideoResolutionStandard = {
    // Name of the resolution
    name: string;

    // Witch (px)
    width: number;

    // Height (px)
    height: number;

    // Frames per second
    fps: number;
};

/**
 * Standard video resolutions
 */
export const STANDARD_VIDEO_RESOLUTIONS: VideoResolutionStandard[] = [
    {
        name: "144p",
        width: 256,
        height: 144,
        fps: 30,
    },
    {
        name: "240p",
        width: 352,
        height: 240,
        fps: 30,
    },
    {
        name: "360p",
        width: 480,
        height: 360,
        fps: 30,
    },
    {
        name: "480p",
        width: 858,
        height: 480,
        fps: 30,
    },
    {
        name: "720p",
        width: 1280,
        height: 720,
        fps: 30,
    },
    {
        name: "720p60",
        width: 1280,
        height: 720,
        fps: 60,
    },
    {
        name: "1080p",
        width: 1920,
        height: 1080,
        fps: 30,
    },
    {
        name: "1080p60",
        width: 1920,
        height: 1080,
        fps: 60,
    },
    {
        name: "2k",
        width: 2048,
        height: 1152,
        fps: 30,
    },
    {
        name: "2k60",
        width: 2048,
        height: 1152,
        fps: 60,
    },
    {
        name: "4k",
        width: 3860,
        height: 2160,
        fps: 30,
    },
    {
        name: "4k60",
        width: 3860,
        height: 2160,
        fps: 60,
    },
];

/**
 * Toggleable video resolution
 */
export type VideoResolutionStandardToggleable = VideoResolutionStandard & {
    // True if the resolution is enabled
    enabled: boolean;
};

/**
 * Standard video resolution
 */
export type ImageResolutionStandard = {
    // Name of the resolution
    name: string;

    // Witch (px)
    width: number;

    // Height (px)
    height: number;
};

export const STANDARD_IMAGE_RESOLUTIONS: ImageResolutionStandard[] = STANDARD_VIDEO_RESOLUTIONS.filter((r) => {
    return r.fps === 30;
});

/**
 * Toggleable image resolution
 */
export type ImageResolutionStandardToggleable = ImageResolutionStandard & {
    // True if the resolution is enabled
    enabled: boolean;
};
