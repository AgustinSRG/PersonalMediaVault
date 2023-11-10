// Keyboard events manager

"use strict";

/**
 * Handler for keyboard events
 */
export type KeyboardEventHandler = (event: KeyboardEvent) => boolean;

/**
 * Keyboard manager.
 * Handles global key presses.
 */
export class KeyboardManager {
    /**
     * Keyboard handlers
     */
    private static handlers: { priority: number; fn: KeyboardEventHandler }[] = [];

    /**
     * Initialization logic
     */
    public static Initialize() {
        document.addEventListener("keydown", KeyboardManager.Handle);
    }

    /**
     * Handles keydown event
     * @param event The event
     */
    private static Handle(event: KeyboardEvent) {
        if (!event.ctrlKey && event.target && ["input", "select"].includes(((<Node>event.target).nodeName + "").toLowerCase())) {
            return;
        }
        if (event.key === "Enter" && event.target && ["button", "a"].includes(((<Node>event.target).nodeName + "").toLowerCase())) {
            return;
        }
        for (const handler of KeyboardManager.handlers) {
            if (handler.fn(event)) {
                event.preventDefault();
                return;
            }
        }
    }

    /**
     * Sorts event handlers by priority
     */
    private static Sort() {
        KeyboardManager.handlers = KeyboardManager.handlers.sort((a, b) => {
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
     * Adds keyboard handler
     * @param handler The handler function
     * @param priority The priority
     */
    public static AddHandler(handler: KeyboardEventHandler, priority?: number) {
        KeyboardManager.handlers.push({
            priority: priority || 0,
            fn: handler,
        });
        KeyboardManager.Sort();
    }

    /**
     * Removes keyboard handler
     * @param handler The handler function
     */
    public static RemoveHandler(handler: KeyboardEventHandler) {
        for (let i = 0; i < KeyboardManager.handlers.length; i++) {
            if (KeyboardManager.handlers[i].fn === handler) {
                KeyboardManager.handlers.splice(i, 1);
                return;
            }
        }
    }
}
