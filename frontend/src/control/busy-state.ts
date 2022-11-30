// Busy state controller

"use strict";

import { UploadController } from "./upload";

export class BusyStateController {
    public static Busy = false;

    private static BusySet = new Set();

    public static Initialize() {
        window.addEventListener('beforeunload', function (e) {
            if (BusyStateController.Busy || UploadController.GetEntries().filter(a => a.status !== "error" && a.status !== "ready").length > 0) {
                // Cancel the event
                e.preventDefault(); // If you prevent default behavior in Mozilla Firefox prompt will always be shown
                // Chrome requires returnValue to be set
                e.returnValue = '';
            }
        });
    }

    public static IsBusy(key: string): boolean {
        return BusyStateController.BusySet.has(key);
    }

    public static SetBusy(key: string) {
        BusyStateController.BusySet.add(key);
        BusyStateController.Busy = true;
    }

    public static RemoveBusy(key: string) {
        BusyStateController.BusySet.delete(key);
        BusyStateController.Busy = BusyStateController.BusySet.size > 0;
    }
}
