// Player subtitles composable

"use strict";

import { EVENT_NAME_SUBTITLES_UPDATE } from "@/control/app-events";
import type { Ref } from "vue";
import { ref } from "vue";
import { onApplicationEvent } from "./on-app-event";
import { getSubtitlesLine } from "@/control/subtitles";

/**
 * Player subtitles composable
 */
export type PlayerSubtitlesComposable = {
    /**
     * Current subtitles
     */
    subtitles: Ref<string>;

    /**
     * Resets subtitles
     */
    resetSubtitles: () => void;

    /**
     * Updates subtitles based on current time
     */
    updateSubtitles: () => void;
};

/**
 * Gets the player subtitles composable
 * @param currentTime Current time ref
 * @returns The composable
 */
export function usePlayerSubtitles(currentTime: Ref<number>): PlayerSubtitlesComposable {
    // Current subtitles
    const subtitles = ref("");

    // Subtitles range
    const subtitlesStart = ref(-1);
    const subtitlesEnd = ref(-1);

    /**
     * Resets subtitles
     */
    const resetSubtitles = () => {
        subtitles.value = "";
        subtitlesStart.value = -1;
        subtitlesEnd.value = -1;
    };

    /**
     * Updates subtitles based on the current time
     */
    const updateSubtitles = () => {
        if (currentTime.value >= subtitlesStart.value && currentTime.value <= subtitlesEnd.value) {
            return;
        }
        const sub = getSubtitlesLine(currentTime.value);
        if (sub) {
            subtitles.value = sub.text;
            subtitlesStart.value = sub.start;
            subtitlesEnd.value = sub.end;
        } else {
            subtitles.value = "";
            subtitlesStart.value = 0;
            subtitlesEnd.value = 0;
        }
    };

    onApplicationEvent(EVENT_NAME_SUBTITLES_UPDATE, () => {
        resetSubtitles();
        updateSubtitles();
    });

    return {
        subtitles,
        resetSubtitles,
        updateSubtitles,
    };
}
