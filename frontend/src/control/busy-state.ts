// Busy state controller

"use strict";

import { UploadController } from "./upload";

/**
 * Management object to warn the user before closing when in busy state.
 */
export class BusyStateController {
    /**
     * Set to store the busy statuses
     */
    private static BusySet = new Set();

    /**
     * Initialization logic
     */
    public static Initialize() {
        window.addEventListener("beforeunload", function (e) {
            if (
                BusyStateController.BusySet.size > 0 ||
                UploadController.GetEntries().filter((a) => a.status !== "error" && a.status !== "ready").length > 0
            ) {
                // Cancel the event
                e.preventDefault(); // If you prevent default behavior in Mozilla Firefox prompt will always be shown
                // Chrome requires returnValue to be set
                e.returnValue = "";
            }
        });
    }

    /**
     * Checks if a status is busy
     * @param key Status key
     * @returns True if busy
     */
    public static IsBusy(key: string): boolean {
        return BusyStateController.BusySet.has(key);
    }

    /**
     * Set a status as busy
     * @param key Status key
     */
    public static SetBusy(key: string) {
        BusyStateController.BusySet.add(key);
    }

    /**
     * Sets a status as free (no longer busy)
     * @param key Status key
     */
    public static RemoveBusy(key: string) {
        BusyStateController.BusySet.delete(key);
    }
}
