// Keyboard events manager

"use strict";

/**
 * Handler for keyboard events
 */
export type KeyboardEventHandler = (event: KeyboardEvent) => boolean;

/**
 * Keyboard handlers global status
 */
const KeyboardHandlersStatus = {
    /**
     * List of keyboard handlers
     */
    handlers: [] as { priority: number; fn: KeyboardEventHandler }[],
};

// Listen for document 'keydown' global event
document.addEventListener("keydown", (event: KeyboardEvent) => {
    if (!event.ctrlKey && event.target && ["input", "select"].includes(((<Node>event.target).nodeName + "").toLowerCase())) {
        return;
    }
    if (event.key === "Enter" && event.target && ["button", "a"].includes(((<Node>event.target).nodeName + "").toLowerCase())) {
        return;
    }
    for (const handler of KeyboardHandlersStatus.handlers) {
        if (handler.fn(event)) {
            event.preventDefault();
            return;
        }
    }
});

/**
 * Sorts the keyboard handlers by priority
 */
function sortHandlers() {
    KeyboardHandlersStatus.handlers = KeyboardHandlersStatus.handlers.sort((a, b) => {
        if (a.priority > b.priority) {
            return -1;
        } else if (a.priority < b.priority) {
            return 1;
        } else {
            return 0;
        }
    });
}

/**
 * Adds global keyboard handler
 * @param handler The handler function
 * @param priority The priority
 */
export function addGlobalKeyboardHandler(handler: KeyboardEventHandler, priority?: number) {
    KeyboardHandlersStatus.handlers.push({
        priority: priority || 0,
        fn: handler,
    });
    sortHandlers();
}

/**
 * Removes global keyboard handler
 * @param handler The handler function
 */
export function removeGlobalKeyboardHandler(handler: KeyboardEventHandler) {
    for (let i = 0; i < KeyboardHandlersStatus.handlers.length; i++) {
        if (KeyboardHandlersStatus.handlers[i].fn === handler) {
            KeyboardHandlersStatus.handlers.splice(i, 1);
            return;
        }
    }
}
