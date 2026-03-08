// Snack bar control logic

"use strict";

import { emitAppEvent, EVENT_NAME_SNACK_BAR } from "./app-events";

/**
 * Snackbar position
 */
export type SnackBarPosition = "left" | "right" | "center";

/**
 * Emits event to show a snackbar
 * @param message The message to show
 * @param position The position of the snackbar
 */
export function showSnackBar(message: string, position?: SnackBarPosition) {
    emitAppEvent(EVENT_NAME_SNACK_BAR, message, position);
}

/**
 * Emits event to show a snackbar (Right position)
 * @param message The message to show
 */
export function showSnackBarRight(message: string) {
    showSnackBar(message, "right");
}

/**
 * Emits event to show a snackbar (Center position)
 * @param message The message to show
 */
export function ShowSnackBarCenter(message: string) {
    showSnackBar(message, "center");
}
