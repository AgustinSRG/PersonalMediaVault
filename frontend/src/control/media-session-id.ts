// Media session ID global store

"use strict";

/**
 * Global store for media session ID
 */
const MediaSessionIdGlobalStore = {
    /**
     * Media session ID
     */
    id: "",
};

/**
 * Gets global media session ID
 * @returns The current global media session ID
 */
export function getGlobalMediaSessionId(): string {
    return MediaSessionIdGlobalStore.id;
}

/**
 * Sets global media session ID
 * @param id The media session ID
 */
export function setGlobalMediaSessionId(id: string) {
    MediaSessionIdGlobalStore.id = id;
}

/**
 * Clears global media session ID
 */
export function clearGlobalMediaSessionId() {
    MediaSessionIdGlobalStore.id = "";
}
