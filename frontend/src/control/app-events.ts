// Global events manager

"use strict";

type CallbackFunctionVariadic = (...args: any[]) => void;

/**
 * Event triggered to show a message in the snackbar
 */
export const EVENT_NAME_SNACK_BAR = "snack";

/**
 * Global events manager
 */
export class AppEvents {
    /**
     * Event handlers
     */
    private static events: Map<string, CallbackFunctionVariadic[]> = new Map();

    /**
     * Adds event handler
     * @param eventName Event name
     * @param handler Event handler function
     */
    public static AddEventListener(eventName: string, handler: CallbackFunctionVariadic) {
        if (!AppEvents.events.has(eventName)) {
            AppEvents.events.set(eventName, []);
        }
        AppEvents.events.get(eventName).push(handler);
    }

    /**
     * Emits event
     * @param eventName Event name
     * @param args Event arguments
     */
    public static Emit(eventName: string, ...args: any[]) {
        if (AppEvents.events.has(eventName)) {
            for (const handler of AppEvents.events.get(eventName)) {
                try {
                    handler(...args);
                } catch (ex) {
                    console.error(ex);
                }
            }
        }
    }

    /**
     * Removes event handler
     * @param eventName Event name
     * @param handler Event handler function
     */
    public static RemoveEventListener(eventName: string, handler: CallbackFunctionVariadic) {
        if (!AppEvents.events.has(eventName)) {
            return;
        }
        const arr = AppEvents.events.get(eventName);
        const i = arr.indexOf(handler);
        if (i >= 0) {
            arr.splice(i, 1);
            if (arr.length === 0) {
                AppEvents.events.delete(eventName);
            }
        }
    }

    /**
     * Emits event to show a snackbar
     * @param message The message to show
     */
    public static ShowSnackBar(message: string) {
        AppEvents.Emit(EVENT_NAME_SNACK_BAR, message);
    }
}
