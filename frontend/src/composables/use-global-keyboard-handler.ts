// Composable to register a global keyboard handler

"use strict";

import type { KeyboardEventHandler } from "@/control/keyboard";
import { KeyboardManager } from "@/control/keyboard";
import { onBeforeUnmount } from "vue";

/**
 * Registers a global keyboard handler.
 * Automatically removes it when the component unmounts
 * @param handler The handler
 * @param priority The priority
 */
export function useGlobalKeyboardHandler(handler: KeyboardEventHandler, priority?: number) {
    KeyboardManager.AddHandler(handler, priority);

    onBeforeUnmount(() => {
        KeyboardManager.RemoveHandler(handler);
    });
}
