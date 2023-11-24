// Plugin for global events and other custom methods

"use strict";

import { App } from "vue";
import { AppEvents } from "./control/app-events";
import { KeyboardEventHandler, KeyboardManager } from "./control/keyboard";

type CallbackFunctionVariadic = (...args: any[]) => void;

declare module "vue" {
    interface ComponentCustomProperties {
        /**
         * Override refs
         */
        $refs: {
            modalContainer?: {
                /**
                 * Closes the modal
                 * @param forced True to force (no wait, no condition)
                 */
                close: (forced?: boolean) => void;
            };

            // Position change modal. TODO: Use event instead
            movePosModal?: {
                show: (options: { pos: number; callback: (newPos: number) => void }) => void;
            };

            // Delete account modal. TODO: Use event instead
            deleteModal?: {
                show: (options: { name: string; callback: () => void }) => void;
            };

            // Delete attachment modal. TODO: Use event instead
            attachmentDeleteModal?: {
                show: (options: { name: string; callback: () => void }) => void;
            };

            // Delete audio track modal. TODO: Use event instead
            audioTrackDeleteModal?: {
                show: (options: { name: string; callback: () => void }) => void;
            };

            // Delete subtitles modal. TODO: Use event instead
            subtitlesDeleteModal?: {
                show: (options: { name: string; callback: () => void }) => void;
            };

            // Resolution confirmation modal. TODO: Use event instead
            resolutionConfirmationModal?: {
                show: (options: {
                    type: number;
                    deleting: boolean;
                    name: string;
                    width: number;
                    height: number;
                    fps: number;
                    callback: () => void;
                }) => void;
            };

            advSearch?: {
                onScroll: (e: Event) => void;
                goTop: () => void;
            };
        };

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
    }
}

export const appEventsPlugin = {
    install: (app: App) => {
        app.mixin({
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
            },
            beforeUnmount() {
                this.$appEventHandlers && this.$appEventHandlers.forEach((handler, eventName) => {
                    AppEvents.RemoveEventListener(eventName, handler);
                });
                this.$documentEventHandlers && this.$documentEventHandlers.forEach((listener, eventName) => {
                    document.removeEventListener(eventName, listener);
                });
                this.$keyboardHandlers && this.$keyboardHandlers.forEach(KeyboardManager.RemoveHandler);
            },
        });
    },
};
