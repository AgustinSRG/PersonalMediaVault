// Last row padding for pages

"use strict";

import type { ComputedRef, Ref } from "vue";
import { computed, nextTick, onBeforeUnmount, onMounted, ref, watch } from "vue";

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
    // Number of items that fit in a row
    const itemsFitInRow = ref(1);

    /**
     * Updates value of itemsFitInRow
     */
    const updateItemsFitInRow = () => {
        const containerWidth = container.value?.getBoundingClientRect().width || 0;

        const itemWidth = Math.max(
            props.minItemsSize,
            Math.min(
                props.maxItemsSize,
                props.min ? containerWidth / Math.max(1, props.rowSizeMin) : containerWidth / Math.max(1, props.rowSize),
            ),
        );

        itemsFitInRow.value = Math.max(1, Math.floor(containerWidth / Math.max(1, itemWidth)));
    };

    updateItemsFitInRow();

    // Resize observer
    const resizeObserver = new ResizeObserver(() => {
        updateItemsFitInRow();
    });

    onMounted(() => {
        if (container.value) {
            resizeObserver.observe(container.value);
        }
    });

    onBeforeUnmount(() => {
        resizeObserver.disconnect();
    });

    watch([() => props.minItemsSize, () => props.rowSizeMin, () => props.rowSize], updateItemsFitInRow);

    watch([() => props.min], () => {
        nextTick(updateItemsFitInRow);
    });

    const lastRowPadding = computed(() => {
        return Math.max(0, itemsFitInRow.value - (pageItemsLength.value % itemsFitInRow.value));
    });

    return {
        lastRowPadding,
        itemsFitInRow,
    };
}
