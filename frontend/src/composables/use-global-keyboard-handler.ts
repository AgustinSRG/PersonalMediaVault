// Composable to register a global keyboard handler

"use strict";

import { addGlobalKeyboardHandler, removeGlobalKeyboardHandler, type KeyboardEventHandler } from "@/global-state/keyboard";
import { onBeforeUnmount } from "vue";

/**
 * Registers a global keyboard handler.
 * Automatically removes it when the component unmounts
 * @param handler The handler
 * @param priority The priority
 */
export function useGlobalKeyboardHandler(handler: KeyboardEventHandler, priority?: number) {
    addGlobalKeyboardHandler(handler, priority);

    onBeforeUnmount(() => {
        removeGlobalKeyboardHandler(handler);
    });
}
