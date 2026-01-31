// Composable for listening to global document events

"use strict";

import { onBeforeUnmount } from "vue";

/**
 * Listens for a global document event during the lifetime of the component.
 * The listener will be automatically cleared when the component unmounts.
 * @param eventName The event name
 * @param listener The listener
 */
export function onDocumentEvent<K extends keyof DocumentEventMap>(
    eventName: K,
    listener: (this: Document, ev: DocumentEventMap[K]) => any,
) {
    document.addEventListener(eventName, listener);

    onBeforeUnmount(() => {
        document.removeEventListener(eventName, listener);
    });
}
