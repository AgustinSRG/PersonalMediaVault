// Player auto next composable

"use strict";

import { getAutoNextTime, setAutoNextTime } from "@/control/player-preferences";
import type { Ref } from "vue";
import { watch } from "vue";
import { useTimeout } from "./use-timeout";
import type { MediaListItem } from "@/api/models";
import { useI18n } from "./use-i18n";
import { PagesController } from "@/control/pages";
import { renderAutoNext } from "@/utils/player-config";

/**
 * Required props for player auto-next
 */
export type PlayerAutoNextRequiredProps = {
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
 * Player auto-next composable
 */
export type PlayerAutoNextComposable = {
    /**
     * Clears auto-next timer
     */
    clearAutoNextTimer: () => void;

    /**
     * Starts auto-next timer
     */
    setupAutoNextTimer: () => void;

    /**
     * Toggles auto next
     */
    toggleAutoNext: () => void;
};

/**
 * Gets the player auto-next composable
 * @param props Player component props
 * @param isShort Reference to indicate if the media is short enough for auto-next
 * @param refs Required references
 * @param goNext Function to go to the next element
 * @returns The composable
 */
export function usePlayerAutoNext(
    props: PlayerAutoNextRequiredProps,
    isShort: Ref<boolean>,
    refs: { displayConfig: Ref<boolean>; expandedTitle: Ref<boolean>; displayAttachments: Ref<boolean>; displayRelatedMedia: Ref<boolean> },
    goNext: () => void,
): PlayerAutoNextComposable {
    // Translation
    const { $t } = useI18n();

    // Interval to check for auto-next
    const autoNextTimer = useTimeout();

    /**
     * Clears auto-next timer
     */
    const clearAutoNextTimer = () => {
        autoNextTimer.clear();
    };

    /**
     * Sets up auto-next timer
     */
    const setupAutoNextTimer = () => {
        autoNextTimer.clear();

        if (!isShort.value) {
            return;
        }

        const timerS = getAutoNextTime();

        if (isNaN(timerS) || !isFinite(timerS) || timerS <= 0) {
            return;
        }

        if (!props.next && !props.pageNext) {
            return;
        }

        const ms = timerS * 1000;

        autoNextTimer.set(() => {
            if (refs.displayConfig.value || refs.expandedTitle.value || refs.displayAttachments.value || refs.displayRelatedMedia.value) {
                setupAutoNextTimer();
            } else {
                goNext();
            }
        }, ms);
    };

    // Reset auto-next timer if next state changes
    watch([() => props.next, () => props.pageNext], setupAutoNextTimer);

    // Default value for active auto next
    const DEFAULT_ACTIVE_AUTO_NEXT = 10;

    /**
     * Toggles auto next
     */
    const toggleAutoNext = () => {
        let oldTime = getAutoNextTime(true);

        if (oldTime <= 0) {
            oldTime = DEFAULT_ACTIVE_AUTO_NEXT;
        }

        let newTime = getAutoNextTime();

        if (newTime <= 0) {
            // Enable
            newTime = oldTime;
        } else {
            setAutoNextTime(newTime, true); // Store to re-enable later
            newTime = 0;
        }

        setAutoNextTime(newTime);

        PagesController.ShowSnackBar($t("Auto next") + ": " + renderAutoNext(newTime, $t));

        setupAutoNextTimer();
    };

    return {
        clearAutoNextTimer,
        setupAutoNextTimer,
        toggleAutoNext,
    };
}
