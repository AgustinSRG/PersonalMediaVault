// Home page utils

"use strict";

import type { HomePageElement } from "@/api/api-home";
import { getUniqueStringId } from "./unique-id";

/**
 * Types of groups in the home page
 */
export const HomePageGroupTypes = {
    /**
     * Custom group with custom elements
     */
    Custom: 0,

    /**
     * Recently uploaded media
     */
    RecentMedia: 1,

    /**
     * Recently modified albums
     */
    RecentAlbums: 2,
};

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

const HomePageSilentSaveActionsStatus = {
    actions: new Set<string>(),
    waitCallbacks: [] as (() => void)[],
};

/**
 * Waits for all silent save actions to finish in the home page
 * @param callback The callback
 */
export function waitForHomePageSilentSaveActions(callback: () => void) {
    if (HomePageSilentSaveActionsStatus.actions.size === 0) {
        callback();
    } else {
        HomePageSilentSaveActionsStatus.waitCallbacks.push(callback);
    }
}

/**
 * Performs a silent save action
 * @param fn The function, it receives a parameter (callback). It must be called after the request is done.
 */
export function doHomePageSilentSaveAction(fn: (callback: () => void) => void) {
    const id = getUniqueStringId();

    HomePageSilentSaveActionsStatus.actions.add(id);

    const callback = () => {
        HomePageSilentSaveActionsStatus.actions.delete(id);

        if (HomePageSilentSaveActionsStatus.actions.size === 0) {
            for (const cb of HomePageSilentSaveActionsStatus.waitCallbacks) {
                try {
                    cb();
                } catch (ex) {
                    console.error(ex);
                }
            }
            HomePageSilentSaveActionsStatus.waitCallbacks = [];
        }
    };

    try {
        fn(callback);
    } catch (ex) {
        console.error(ex);
        callback();
    }
}
