// Global busy state

"use strict";

import { UploadController } from "./upload";

// Set to store the busy statuses
const BusySet = new Set();

/**
 * Initializes the global busy state
 */
export function initializeGlobalBusyState() {
    // Add listener for window close
    window.addEventListener("beforeunload", function (e: BeforeUnloadEvent) {
        if (BusySet.size > 0 || UploadController.GetPendingEntries().length > 0) {
            // Cancel the event
            e.preventDefault(); // If you prevent default behavior in Mozilla Firefox prompt will always be shown
            // Chrome requires returnValue to be set
            e.returnValue = "";
        }
    });
}

/**
 * Checks a global busy state
 * @param key Status key
 * @returns True if busy
 */
export function isGlobalBusyState(key: string): boolean {
    return BusySet.has(key);
}

/**
 * Set a global busy state
 * @param key Status key
 */
export function setGlobalBusyState(key: string) {
    BusySet.add(key);
}

/**
 * Sets a global busy state as free (no longer busy)
 * @param key Status key
 */
export function removeGlobalBusyState(key: string) {
    BusySet.delete(key);
}
