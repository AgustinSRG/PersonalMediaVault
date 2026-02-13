// Pages common control logic

"use strict";

import { emitAppEvent, EVENT_NAME_PAGE_MEDIA_NAV_UPDATE, EVENT_NAME_SNACK_BAR } from "./app-events";

/**
 * Snackbar position
 */
export type SnackBarPosition = "left" | "right" | "center";

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
        emitAppEvent(EVENT_NAME_PAGE_MEDIA_NAV_UPDATE, PagesController.HasPagePrev, PagesController.HasPageNext);
    }

    /**
     * Call when a home group is loaded
     * @param hasPrev True if the group has a previous
     * @param hasNext True if the group has a next element
     */
    public static OnHomeGroupLoad(hasPrev: boolean, hasNext: boolean) {
        PagesController.HasPagePrev = hasPrev;
        PagesController.HasPageNext = hasNext;
        emitAppEvent(EVENT_NAME_PAGE_MEDIA_NAV_UPDATE, PagesController.HasPagePrev, PagesController.HasPageNext);
    }

    /**
     * Call when a page unloads
     */
    public static OnPageUnload() {
        PagesController.HasPagePrev = false;
        PagesController.HasPageNext = false;
        emitAppEvent(EVENT_NAME_PAGE_MEDIA_NAV_UPDATE, PagesController.HasPagePrev, PagesController.HasPageNext);
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
    public static ShowSnackBar(message: string, position?: SnackBarPosition) {
        emitAppEvent(EVENT_NAME_SNACK_BAR, message, position);
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
