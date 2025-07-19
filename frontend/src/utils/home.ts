// Home page utils

"use strict";

import type { HomePageElement } from "@/api/api-home";
import { HomePageGroupTypes } from "@/api/api-home";

/**
 * Gets the default group name
 * @param groupType The group type
 * @param $t The translation function
 * @returns The default group name
 */
export function getDefaultGroupName(groupType: number, $t: (key: string) => string): string {
    switch (groupType) {
        case HomePageGroupTypes.RecentMedia:
            return $t("Media (Recently uploaded)");
        case HomePageGroupTypes.RecentAlbums:
            return $t("Albums (Recently modified)");
        default:
            return $t("Custom row");
    }
}

/**
 * Event triggered when the scroll in the home page changed
 */
export const EVENT_NAME_HOME_SCROLL_CHANGED = "home-page-scroll-changed";

// Data to pass when start moving a group
export type HomePageGroupStartMovingData = {
    startX: number; // The start X position
    startY: number; // The start Y position

    offsetX: number; // The offset in X
    offsetY: number; // The offset in Y

    width: number; // The width
    height: number; // The height

    initialElements: HomePageElement[] | null;
    initialScroll: number;
};
