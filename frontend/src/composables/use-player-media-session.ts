// Player MediaSession Composable

"use strict";

import { MediaController } from "@/control/media";
import { addMediaSessionActionHandler, clearMediaSessionActionHandlers } from "@/utils/media-session";
import { getUniqueStringId } from "@/utils/unique-id";
import type { Ref } from "vue";
import { onBeforeUnmount, onMounted, watch } from "vue";

// Media session actions for image player
export const IMAGE_PLAYER_MEDIA_SESSION_ACTIONS: MediaSessionAction[] = ["nexttrack", "previoustrack"];

// Media session actions for video player
export const VIDEO_PLAYER_MEDIA_SESSION_ACTIONS: MediaSessionAction[] = [
    "play",
    "pause",
    "nexttrack",
    "previoustrack",
    "seekbackward",
    "seekforward",
    "seekto",
];

// Media session actions for audio player
export const AUDIO_PLAYER_MEDIA_SESSION_ACTIONS = VIDEO_PLAYER_MEDIA_SESSION_ACTIONS;

/**
 * Handles MediaSession actions and automatically updates the playing status
 * @param actions The MediaSession actions the player can handle
 * @param handler The handler function
 * @param playing The playing status (optional)
 */
export function usePlayerMediaSession(
    actions: MediaSessionAction[],
    handler: (event: MediaSessionActionDetails) => void,
    playing?: Ref<boolean>,
) {
    // Unique ID for the media session
    const mediaSessionId = getUniqueStringId();

    onMounted(() => {
        if (window.navigator && window.navigator.mediaSession) {
            MediaController.MediaSessionId = mediaSessionId;
            clearMediaSessionActionHandlers();

            addMediaSessionActionHandler(actions, handler);

            if (playing) {
                navigator.mediaSession.playbackState = playing.value ? "playing" : "paused";
            } else {
                navigator.mediaSession.playbackState = "none";
            }
        }
    });

    onBeforeUnmount(() => {
        if (window.navigator && window.navigator.mediaSession && MediaController.MediaSessionId === mediaSessionId) {
            clearMediaSessionActionHandlers();
            navigator.mediaSession.playbackState = "none";
            MediaController.MediaSessionId = "";
        }
    });

    if (playing) {
        watch(playing, () => {
            if (!window.navigator || !window.navigator.mediaSession) {
                return;
            }

            if (MediaController.MediaSessionId !== mediaSessionId) {
                return;
            }

            navigator.mediaSession.playbackState = playing.value ? "playing" : "paused";
        });
    }
}
