// Last row padding for pages

"use strict";

import type { ComputedRef, Ref } from "vue";
import { computed, onBeforeUnmount, onMounted, ref } from "vue";

export type LastRowPaddingComposable = {
    /**
     * Padding for the last row
     */
    lastRowPadding: Ref<number>;

    /**
     * Number of items that fit in a row
     */
    itemsFitInRow: Ref<number>;
};

/**
 * Gets the number of elements for the last row padding in pages.
 * @param props The page component properties
 * @param container The page container
 * @param pageItems The page items
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
    pageItemsLength: ComputedRef<number>,
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

    const itemsFitInRow = computed(() => {
        const itemWidth = Math.max(
            props.minItemsSize,
            Math.min(
                props.maxItemsSize,
                props.min ? containerWidth.value / Math.max(1, props.rowSizeMin) : containerWidth.value / Math.max(1, props.rowSize),
            ),
        );

        return Math.max(1, Math.floor(containerWidth.value / Math.max(1, itemWidth)));
    });

    const lastRowPadding = computed(() => {
        return Math.max(0, itemsFitInRow.value - (pageItemsLength.value % itemsFitInRow.value));
    });

    return {
        lastRowPadding,
        itemsFitInRow,
    };
}
