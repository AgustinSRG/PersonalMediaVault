// Touch devices utils

"use strict";

/**
 * Checks if the device is a touch device
 * @returns True if the browser supports the touch events and touches are enabled
 */
export function isTouchDevice() {
    return "ontouchstart" in window && navigator.maxTouchPoints > 0;
}
