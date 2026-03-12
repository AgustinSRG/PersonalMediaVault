// Composable for listening to application events

"use strict";

import type { AppEventsMap } from "@/global-state/app-events";
import { addAppEventListener, removeAppEventListener } from "@/global-state/app-events";
import { onBeforeUnmount } from "vue";

/**
 * Listens for an application event during the lifetime of the component.
 * The listener will be automatically cleared when the component unmounts.
 * @param eventName The event name
 * @param listener The listener
 */
export function onApplicationEvent<K extends keyof AppEventsMap>(eventName: K, listener: AppEventsMap[K]) {
    addAppEventListener(eventName, listener);

    onBeforeUnmount(() => {
        removeAppEventListener(eventName, listener);
    });
}
