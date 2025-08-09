// Plugin for global events and other custom methods

"use strict";

import type { App } from "vue";
import { AppEvents } from "./control/app-events";
import type { KeyboardEventHandler } from "./control/keyboard";
import { KeyboardManager } from "./control/keyboard";

type CallbackFunctionVariadic = (...args: any[]) => void;

declare module "vue" {
    interface ComponentCustomProperties {
        /**
         * Override refs
         */
        $refs: never;

        /**
         * Mapping of app event handlers
         */
        $appEventHandlers: Map<string, CallbackFunctionVariadic>;

        /**
         * Mapping of document event handlers
         */
        $documentEventHandlers: Map<string, CallbackFunctionVariadic>;

        /**
         * List of keyboard handlers
         */
        $keyboardHandlers: KeyboardEventHandler[];

        /**
         * Listens to a global custom App event, and removes the listener after the component has been removed
         * @param eventName Event name
         * @param handler Event handler function
         */
        $listenOnAppEvent: (eventName: string, handler: CallbackFunctionVariadic) => void;

        /**
         * Listens to a document
         * @param eventName Event name
         * @param handler Event handler function
         */
        $listenOnDocumentEvent: <K extends keyof DocumentEventMap>(
            eventName: K,
            listener: (this: Document, ev: DocumentEventMap[K]) => any,
        ) => void;

        /**
         * Adds a keyboard handler
         * @param handler The handler function
         * @param priority The priority
         */
        $addKeyboardHandler: (handler: KeyboardEventHandler, priority?: number) => void;

        /**
         * Clicks element whenever the enter key is pressed
         * @param event The keyboard event
         */
        clickOnEnter: (event: KeyboardEvent) => void;

        /**
         * Prevents the propagation of an event to the parent elements
         * @param event The DOM event
         */
        stopPropagationEvent: (event: Event) => void;
    }
}

export const appEventsPlugin = {
    install: (app: App) => {
        app.mixin({
            beforeUnmount() {
                if (this.$appEventHandlers) {
                    this.$appEventHandlers.forEach((handler, eventName) => {
                        AppEvents.RemoveEventListener(eventName, handler);
                    });
                }

                if (this.$documentEventHandlers) {
                    this.$documentEventHandlers.forEach((listener, eventName) => {
                        document.removeEventListener(eventName, listener);
                    });
                }

                if (this.$keyboardHandlers) {
                    this.$keyboardHandlers.forEach(KeyboardManager.RemoveHandler);
                }
            },
            methods: {
                $listenOnAppEvent: function (eventName: string, handler: CallbackFunctionVariadic) {
                    if (!this.$appEventHandlers) {
                        this.$appEventHandlers = new Map();
                    }
                    if (this.$appEventHandlers.has(eventName)) {
                        throw new Error("Already listening for app event '" + eventName + "' on this component");
                    }
                    this.$appEventHandlers.set(eventName, handler);
                    AppEvents.AddEventListener(eventName, handler);
                },
                $listenOnDocumentEvent: function <K extends keyof DocumentEventMap>(
                    eventName: K,
                    listener: (this: Document, ev: DocumentEventMap[K]) => any,
                ) {
                    if (!this.$documentEventHandlers) {
                        this.$documentEventHandlers = new Map();
                    }
                    if (this.$documentEventHandlers.has(eventName)) {
                        throw new Error("Already listening for document event '" + eventName + "' on this component");
                    }
                    this.$documentEventHandlers.set(eventName, listener);
                    document.addEventListener(eventName, listener);
                },
                $addKeyboardHandler: function (handler: KeyboardEventHandler, priority?: number) {
                    if (!this.$keyboardHandlers) {
                        this.$keyboardHandlers = [];
                    }
                    this.$keyboardHandlers.push(handler);
                    KeyboardManager.AddHandler(handler, priority);
                },
                clickOnEnter: function (event: KeyboardEvent) {
                    if (event.key === "Enter") {
                        event.preventDefault();
                        event.stopPropagation();
                        (event.target as HTMLElement).click();
                    }
                },
                stopPropagationEvent: function (e: Event) {
                    e.stopPropagation();
                },
            },
        });
    },
};
