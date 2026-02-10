// Composable for setting intervals

"use strict";

import { onBeforeUnmount } from "vue";

/**
 * Composable interval
 */
export type ComposableInterval = {
    /**
     * Sets the interval.
     * If the interval was already set, it clears the previous one.
     * @param callback  The interval callback
     * @param time The interval time (milliseconds)
     */
    set: (callback: () => void, time: number) => void;

    /**
     * Clears the interval.
     */
    clear: () => void;
};

/**
 * Creates a composable interval.
 * The interval will be automatically cleared when the component is unmounted.
 * @returns A composable that allows to set and clear the interval
 */
export function useInterval(): ComposableInterval {
    let interval: ReturnType<typeof setInterval> | null = null;

    /**
     * Function to set the interval
     * @param callback The interval callback
     * @param time The time (milliseconds)
     */
    const set = (callback: () => void, time: number) => {
        if (interval !== null) {
            clearInterval(interval);
        }

        interval = setInterval(callback, time);
    };

    /**
     * Function to clear the interval
     */
    const clear = () => {
        if (interval !== null) {
            clearInterval(interval);
        }

        interval = null;
    };

    // Ensure the interval is cleared when the component is unmounted
    onBeforeUnmount(clear);

    return { set, clear };
}
