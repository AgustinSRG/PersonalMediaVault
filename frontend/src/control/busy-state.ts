// Global busy state

"use strict";

// Set to store the busy statuses
const BusySet = new Set<string>();

// Function to check for a busy state
export type BusyCheckFunction = () => boolean;

// Array of busy check functions
const BusyCheckFunctions: BusyCheckFunction[] = [];

// Add listener for window close
window.addEventListener("beforeunload", function (e: BeforeUnloadEvent) {
    let busy = false;

    for (const fn of BusyCheckFunctions) {
        if (fn()) {
            busy = true;
            break;
        }
    }

    if (BusySet.size > 0 || busy) {
        // Cancel the event
        e.preventDefault(); // If you prevent default behavior in Mozilla Firefox prompt will always be shown
        // Chrome requires returnValue to be set
        e.returnValue = "";
    }
});

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

/**
 * Adds function to globally check for a busy state
 * @param fn The function
 */
export function addGlobalBusyCheck(fn: BusyCheckFunction) {
    BusyCheckFunctions.push(fn);
}

/**
 * Removes a global busy check function
 * @param fn The function
 */
export function removeGlobalBusyCheck(fn: BusyCheckFunction) {
    for (let i = 0; i < BusyCheckFunctions.length; i++) {
        if (BusyCheckFunctions[i] === fn) {
            BusyCheckFunctions.splice(i, 1);
            return;
        }
    }
}
