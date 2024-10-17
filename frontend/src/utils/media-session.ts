// MediaSession Utils

"use strict";

// Global status of MediaSession
const MediaSessionStatus = {
    // List of actions with handlers set
    actions: new Set<MediaSessionAction>(),
};

/**
 * Adds an action handler for MediaSession
 * @param action The action or list of actions
 * @param handler The handler
 */
export function addMediaSessionActionHandler(
    action: MediaSessionAction | MediaSessionAction[],
    handler: (details: MediaSessionActionDetails) => void,
) {
    if (typeof action === "string") {
        navigator.mediaSession.setActionHandler(action, handler);
        MediaSessionStatus.actions.add(action);
    } else {
        for (const a of action) {
            navigator.mediaSession.setActionHandler(a, handler);
            MediaSessionStatus.actions.add(a);
        }
    }
}

/**
 * Clears all action handlers for MediaSession
 */
export function clearMediaSessionActionHandlers() {
    for (const action of MediaSessionStatus.actions) {
        navigator.mediaSession.setActionHandler(action, null);
    }

    MediaSessionStatus.actions.clear();
}
