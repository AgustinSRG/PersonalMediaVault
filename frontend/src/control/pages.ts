// Pages common control logic

"use strict";

import { AppEvents } from "./app-events";

/**
 * Event triggered when the navigation of the current media in the current page changes
 */
export const EVENT_NAME_PAGE_MEDIA_NAV_UPDATE = "page-media-nav-update";

/**
 * Event triggered when a refresh is requested for the random page
 */
export const EVENT_NAME_RANDOM_PAGE_REFRESH = "random-page-refresh";

/**
 * Event triggered when the user requests going to the next media element in the page
 */
export const EVENT_NAME_PAGE_NAV_NEXT = "page-media-nav-next";

/**
 * Event triggered when the user requests going to the previous media element in the page
 */
export const EVENT_NAME_PAGE_NAV_PREV = "page-media-nav-prev";

/**
 * Event triggered when the user request goings to the next media element
 */
export const EVENT_NAME_GO_NEXT = "media-go-next";

/**
 * Event triggered when the user requests going to the previous media element
 */
export const EVENT_NAME_GO_PREV = "media-go-prev";

/**
 * Event triggered when the user requests going to the top of the advanced search page
 */
export const EVENT_NAME_ADVANCED_SEARCH_GO_TOP = "search-go-top";

/**
 * Event triggered when the advanced search container is scrolled
 */
export const EVENT_NAME_ADVANCED_SEARCH_SCROLL = "search-scroll";

/**
 * Event triggered when the user makes changes to the basic metadata of the current media
 * This means albums and pages should reload
 */
export const EVENT_NAME_MEDIA_METADATA_CHANGE = "media-meta-change";

/**
 * Event triggered when the user deletes the current media
 * This means albums and pages should reload
 */
export const EVENT_NAME_MEDIA_DELETE = "media-delete";

/**
 * Event triggered to show a message in the snackbar
 */
export const EVENT_NAME_SNACK_BAR = "snack";

/**
 * Management object for pages
 */
export class PagesController {
    /**
     * True if the current media has a previous element
     */
    public static HasPagePrev = false;

    /**
     * True if the current media has a next element
     */
    public static HasPageNext = false;

    /**
     * Call when a page is loaded
     * @param currentMediaIndex The index of the current media in the page
     * @param pageSize The current page size
     * @param page The page number
     * @param totalPages The total number of pages
     */
    public static OnPageLoad(currentMediaIndex: number, pageSize: number, page: number, totalPages: number) {
        if (currentMediaIndex >= 0) {
            PagesController.HasPagePrev = currentMediaIndex > 0 || page > 0;
            PagesController.HasPageNext = currentMediaIndex < pageSize - 1 || page < totalPages - 1;
        } else {
            PagesController.HasPagePrev = false;
            PagesController.HasPageNext = false;
        }
        AppEvents.Emit(EVENT_NAME_PAGE_MEDIA_NAV_UPDATE, PagesController.HasPagePrev, PagesController.HasPageNext);
    }

    /**
     * Call when a home group is loaded
     * @param hasPrev True if the group has a previous
     * @param hasNext True if the group has a next element
     */
    public static OnHomeGroupLoad(hasPrev: boolean, hasNext: boolean) {
        PagesController.HasPagePrev = hasPrev;
        PagesController.HasPageNext = hasNext;
        AppEvents.Emit(EVENT_NAME_PAGE_MEDIA_NAV_UPDATE, PagesController.HasPagePrev, PagesController.HasPageNext);
    }

    /**
     * Call when a page unloads
     */
    public static OnPageUnload() {
        PagesController.HasPagePrev = false;
        PagesController.HasPageNext = false;
        AppEvents.Emit(EVENT_NAME_PAGE_MEDIA_NAV_UPDATE, PagesController.HasPagePrev, PagesController.HasPageNext);
    }

    /**
     * Search query for the albums page
     */
    public static AlbumsPageSearch = "";

    /**
     * Emits event to show a snackbar
     * @param message The message to show
     * @param position The position of the snackbar
     */
    public static ShowSnackBar(message: string, position?: "left" | "right" | "center") {
        AppEvents.Emit(EVENT_NAME_SNACK_BAR, message, position);
    }

    /**
     * Emits event to show a snackbar (Right position)
     * @param message The message to show
     */
    public static ShowSnackBarRight(message: string) {
        PagesController.ShowSnackBar(message, "right");
    }

    /**
     * Emits event to show a snackbar (Center position)
     * @param message The message to show
     */
    public static ShowSnackBarCenter(message: string) {
        PagesController.ShowSnackBar(message, "center");
    }
}
