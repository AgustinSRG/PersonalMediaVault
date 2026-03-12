// Pages global state

"use strict";

import { emitAppEvent, EVENT_NAME_PAGE_MEDIA_NAV_UPDATE } from "./app-events";

/**
 * Global page state
 */
const PageGlobalState = {
    /**
     * Does current media has a previous element in the page?
     */
    hasPrev: false,

    /**
     * Does current media has a next element in the page?
     */
    hasNext: false,
};

/**
 * Gets the global state for a page context,
 * regarding if the current media has a previous element
 * @returns True if the current media has a previous element
 */
export function getPageHasPrevGlobalState(): boolean {
    return PageGlobalState.hasPrev;
}

/**
 * Gets the global state for a page context,
 * regarding if the current media has a next element
 * @returns True if the current media has a next element
 */
export function getPageHasNextGlobalState(): boolean {
    return PageGlobalState.hasNext;
}

/**
 * Call when a page is loaded
 * @param currentMediaIndex The index of the current media in the page
 * @param pageSize The current page size
 * @param page The page number
 * @param totalPages The total number of pages
 */
export function onPageLoad(currentMediaIndex: number, pageSize: number, page: number, totalPages: number) {
    if (currentMediaIndex >= 0) {
        PageGlobalState.hasPrev = currentMediaIndex > 0 || page > 0;
        PageGlobalState.hasNext = currentMediaIndex < pageSize - 1 || page < totalPages - 1;
    } else {
        PageGlobalState.hasPrev = false;
        PageGlobalState.hasNext = false;
    }
    emitAppEvent(EVENT_NAME_PAGE_MEDIA_NAV_UPDATE, PageGlobalState.hasPrev, PageGlobalState.hasNext);
}

/**
 * Call when a home group is loaded
 * @param hasPrev True if the group has a previous
 * @param hasNext True if the group has a next element
 */
export function onHomeGroupLoad(hasPrev: boolean, hasNext: boolean) {
    PageGlobalState.hasPrev = hasPrev;
    PageGlobalState.hasNext = hasNext;
    emitAppEvent(EVENT_NAME_PAGE_MEDIA_NAV_UPDATE, PageGlobalState.hasPrev, PageGlobalState.hasNext);
}

/**
 * Call when a page unloads
 */
export function onPageUnload() {
    PageGlobalState.hasPrev = false;
    PageGlobalState.hasNext = false;
    emitAppEvent(EVENT_NAME_PAGE_MEDIA_NAV_UPDATE, PageGlobalState.hasPrev, PageGlobalState.hasNext);
}
