// Last row padding for pages

"use strict";

import type { Ref } from "vue";
import { computed, onBeforeUnmount, onMounted, ref } from "vue";

export type LastRowPaddingComposable = {
    /**
     * Padding for the last row
     */
    lastRowPadding: Ref<number>;

    /**
     * Columns to preserve for row padding
     */
    rowPaddingPreserveCols?: Ref<number>;
};

/**
 * Gets the number of elements for the last row padding in pages.
 * @param props The page component properties
 * @param container The page container
 * @param pageSize The actual page size
 * @param windowPosition Optional ref to the scroller window position
 * @returns A computed reference to the number of element for row padding
 */
export function usePageLastRowPadding(
    props: {
        minItemsSize: number;
        maxItemsSize: number;
        min: boolean;
        rowSize: number;
        rowSizeMin: number;
    },
    container: Readonly<Ref<HTMLElement>>,
    pageSize: Ref<number>,
    windowPosition?: Ref<number>,
): LastRowPaddingComposable {
    // Window width
    const containerWidth = ref(0);

    const resizeObserver = new ResizeObserver(() => {
        containerWidth.value = container.value?.getBoundingClientRect().width || 0;
    });

    onMounted(() => {
        containerWidth.value = container.value?.getBoundingClientRect().width || 0;

        if (container.value) {
            resizeObserver.observe(container.value);
        }
    });

    onBeforeUnmount(() => {
        resizeObserver.disconnect();
    });

    const lastRowPadding = computed(() => {
        const itemWidth = Math.max(
            props.minItemsSize,
            Math.min(
                props.maxItemsSize,
                props.min ? containerWidth.value / Math.max(1, props.rowSizeMin) : containerWidth.value / Math.max(1, props.rowSize),
            ),
        );

        const elementsFitInRow = Math.max(1, Math.floor(containerWidth.value / Math.max(1, itemWidth)));

        if (windowPosition) {
            const lastWindowElement = windowPosition.value + pageSize.value - 1;

            return Math.max(0, elementsFitInRow - 1 - (lastWindowElement % elementsFitInRow));
        } else {
            return Math.max(0, elementsFitInRow - (pageSize.value % elementsFitInRow));
        }
    });

    const rowPaddingPreserveCols = computed(() => {
        const itemWidth = Math.max(
            props.minItemsSize,
            Math.min(
                props.maxItemsSize,
                props.min ? containerWidth.value / Math.max(1, props.rowSizeMin) : containerWidth.value / Math.max(1, props.rowSize),
            ),
        );

        const itemsFitInRow = Math.max(1, Math.floor(containerWidth.value / Math.max(1, itemWidth)));

        if (windowPosition) {
            return windowPosition.value % itemsFitInRow;
        } else {
            return 0;
        }
    });

    return {
        lastRowPadding,
        rowPaddingPreserveCols,
    };
}
