// Page preferences composable

"use strict";

import { getPagePreferences } from "@/control/app-preferences";
import type { AppStatusPage } from "@/control/app-status";
import type { Ref } from "vue";
import { ref } from "vue";
import { onApplicationEvent } from "./on-app-event";
import { EVENT_NAME_PAGE_PREFERENCES_UPDATED } from "@/control/app-events";

/**
 * Page preferences composable
 */
export type PagePreferencesComposable = {
    /**
     * Max elements per page
     */
    pageSize: Ref<number>;

    /**
     * Elements per row (preferred)
     */
    rowSize: Ref<number>;

    /**
     * Elements per row (min)
     */
    rowSizeMin: Ref<number>;

    /**
     * Min item size (pixels)
     */
    minItemSize: Ref<number>;

    /**
     * Max item size (pixels)
     */
    maxItemSize: Ref<number>;

    /**
     * Padding for items (pixels)
     */
    padding: Ref<number>;

    /**
     * True to display titles
     */
    displayTitles: Ref<boolean>;

    /**
     * True to use rounded corners
     */
    roundedCorners: Ref<boolean>;
};

/**
 * Gets the page preferences composable
 * @param page The page name
 * @returns The page preferences composable
 */
export function usePagePreferences(page: AppStatusPage): PagePreferencesComposable {
    // Initial preferences
    const initialPagePreferences = getPagePreferences(page);

    // Max elements per page
    const pageSize = ref(initialPagePreferences.pageSize);

    // Elements per row (preferred)
    const rowSize = ref(initialPagePreferences.rowSize);

    // Elements per row (min)
    const rowSizeMin = ref(initialPagePreferences.rowSizeMin);

    // Min item size (pixels)
    const minItemSize = ref(initialPagePreferences.minItemSize);

    // Max item size (pixels)
    const maxItemSize = ref(initialPagePreferences.maxItemSize);

    // Padding for items (pixels)
    const padding = ref(initialPagePreferences.padding);

    // True to display titles
    const displayTitles = ref(initialPagePreferences.displayTitles);

    // True to use rounded corners
    const roundedCorners = ref(initialPagePreferences.roundedCorners);

    onApplicationEvent(EVENT_NAME_PAGE_PREFERENCES_UPDATED, () => {
        const pagePreferences = getPagePreferences(page);

        pageSize.value = pagePreferences.pageSize;

        rowSize.value = pagePreferences.rowSize;

        rowSizeMin.value = pagePreferences.rowSizeMin;

        minItemSize.value = pagePreferences.minItemSize;

        maxItemSize.value = pagePreferences.maxItemSize;

        padding.value = pagePreferences.padding;

        displayTitles.value = pagePreferences.displayTitles;

        roundedCorners.value = pagePreferences.roundedCorners;
    });

    return {
        pageSize,
        rowSize,
        rowSizeMin,
        minItemSize,
        maxItemSize,
        padding,
        displayTitles,
        roundedCorners,
    };
}
