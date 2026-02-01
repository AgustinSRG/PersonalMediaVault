// Composable for setting timeouts

"use strict";

import { onBeforeUnmount } from "vue";

/**
 * Composable timeout
 */
export type ComposableTimeout = {
    /**
     * Checks if the timeout is set
     * @returns True if the timeout is set
     */
    isSet: () => boolean;

    /**
     * Sets the timeout.
     * If the timeout was already set, it clears the previous one.
     * @param callback  The timeout callback
     * @param time The timeout time (milliseconds)
     */
    set: (callback: () => void, time: number) => void;

    /**
     * Clears the timeout.
     */
    clear: () => void;
};

/**
 * Creates a composable timeout.
 * The timeout will be automatically cleared when the component is unmounted.
 * @returns A composable that allows to set and clear the timeout
 */
export function useTimeout(): ComposableTimeout {
    let timeout: ReturnType<typeof setTimeout> | null = null;

    /**
     * Checks if the timeout is set
     * @returns True if the timeout is set
     */
    const isSet = () => {
        return timeout !== null;
    };

    /**
     * Function to set the timeout
     * @param callback The timeout callback
     * @param time The time (milliseconds)
     */
    const set = (callback: () => void, time: number) => {
        if (timeout !== null) {
            clearTimeout(timeout);
        }

        timeout = setTimeout(() => {
            timeout = null;
            callback();
        }, time);
    };

    /**
     * Function to clear the timeout
     */
    const clear = () => {
        if (timeout !== null) {
            clearTimeout(timeout);
        }

        timeout = null;
    };

    // Ensure the timeout is cleared when the component is unmounted
    onBeforeUnmount(clear);

    return { isSet, set, clear };
}
