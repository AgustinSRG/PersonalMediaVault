// Player next-end composable

"use strict";

import type { Ref } from "vue";
import { ref, watch } from "vue";
import { useTimeout } from "./use-timeout";
import type { MediaListItem } from "@/api/models";
import { getAutoNextOnEnd, getAutoNextPageDelay, setAutoNextOnEnd, setAutoNextPageDelay } from "@/global-state/player-preferences";

/**
 * Required props for player next-end
 */
export type PlayerNextOnEndRequiredProps = {
    /**
     * Next element in album
     */
    next?: MediaListItem | null;

    /**
     * Has next element in page?
     */
    pageNext?: boolean;
};

/**
 * Player next-on-end composable
 */
export type PlayerNextOnEndComposable = {
    /**
     * Wait a few seconds before going to the next media?
     */
    autoNextPageDelay: Ref<boolean>;

    /**
     * Go to next media when audio ends?
     */
    nextEnd: Ref<boolean>;

    /**
     * Pending to go to the next element
     */
    pendingNextEnd: Ref<boolean>;

    /**
     * Number of seconds to automatically go to the next element
     */
    pendingNextEndSeconds: Ref<number>;

    /**
     * Displays the next-end timer
     */
    showNextEnd: () => void;

    /**
     * Hides the next-end overlay
     */
    hideNextEnd: () => void;

    /**
     * Call when the value of 'nextEnd' is updated by the user
     */
    onUpdateNextEnd: () => void;

    /**
     * Call when the value of 'autoNextPageDelay' is updated by the user
     */
    onUpdateAutoNextPageDelay: () => void;
};

/**
 * Gets the player next-on-end composable
 * @param props The player component props
 * @param goNext The function to go to the next element
 * @returns The composable
 */
export function usePlayerNextOnEnd(props: PlayerNextOnEndRequiredProps, goNext: () => void): PlayerNextOnEndComposable {
    // Wait a few seconds before going to the next media?
    const autoNextPageDelay = ref(getAutoNextPageDelay());

    // Go to next media when audio ends?
    const nextEnd = ref(getAutoNextOnEnd());

    // Timer for next-end
    const nextEndTimer = useTimeout();

    // Pending to go to the next element
    const pendingNextEnd = ref(false);

    // Number of seconds to automatically go to the next element
    const pendingNextEndSeconds = ref(0);

    // Tick delay for next-end timer (milliseconds)
    const NEXT_END_TICK_DELAY = 1000;

    // Number of seconds to wait for next-end
    const NEXT_END_WAIT_DURATION = 8;

    /**
     * Displays the next-end timer
     */
    const showNextEnd = () => {
        pendingNextEnd.value = true;
        pendingNextEndSeconds.value = NEXT_END_WAIT_DURATION;

        nextEndTimer.set(tickNextEnd, NEXT_END_TICK_DELAY);
    };

    /**
     * Decreases the next-end timer by one.
     * When 0, goes to the next media
     */
    const tickNextEnd = () => {
        pendingNextEndSeconds.value = Math.max(0, pendingNextEndSeconds.value - 1);

        if (pendingNextEndSeconds.value <= 0) {
            pendingNextEnd.value = false;
            goNext();
            return;
        }

        nextEndTimer.set(tickNextEnd, NEXT_END_TICK_DELAY);
    };

    /**
     * Hides the next-end overlay
     */
    const hideNextEnd = () => {
        pendingNextEnd.value = false;

        nextEndTimer.clear();
    };

    watch([() => props.next, () => props.pageNext], () => {
        if (!props.next && !props.pageNext) {
            hideNextEnd();
        }
    });

    const onUpdateNextEnd = () => {
        setAutoNextOnEnd(nextEnd.value);
    };

    const onUpdateAutoNextPageDelay = () => {
        setAutoNextPageDelay(autoNextPageDelay.value);
    };

    return {
        nextEnd,
        autoNextPageDelay,
        pendingNextEnd,
        pendingNextEndSeconds,
        showNextEnd,
        hideNextEnd,
        onUpdateNextEnd,
        onUpdateAutoNextPageDelay,
    };
}
