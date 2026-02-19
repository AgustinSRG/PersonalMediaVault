// Last row padding for pages

"use strict";

import type { ComputedRef, Ref, ShallowRef } from "vue";
import { computed, onBeforeUnmount, onMounted, ref } from "vue";

/**
 * Gets the number of elements for the last row padding in pages.
 * @param props The page component properties
 * @param container The page container
 * @param pageSize The actual page size
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
    container: Readonly<ShallowRef<HTMLElement>>,
    pageSize: Ref<number>,
): ComputedRef<number> {
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

    return computed(() => {
        const itemWidth = Math.max(
            props.minItemsSize,
            Math.min(
                props.maxItemsSize,
                props.min ? containerWidth.value / Math.max(1, props.rowSizeMin) : containerWidth.value / Math.max(1, props.rowSize),
            ),
        );

        const elementsFitInRow = Math.max(1, Math.floor(containerWidth.value / Math.max(1, itemWidth)));

        return Math.max(0, elementsFitInRow - (pageSize.value % elementsFitInRow));
    });
}
