// Player common features composable

"use strict";

import type { MediaListItem } from "@/api/models";
import { closeFullscreen, openFullscreen } from "@/utils/full-screen";
import type { HelpTooltipType } from "@/utils/player-tooltip";
import { isTouchDevice } from "@/utils/touch";
import type { Ref } from "vue";
import { ref, watch } from "vue";
import { useGlobalKeyboardHandler } from "./use-global-keyboard-handler";
import { AuthController } from "@/global-state/auth";
import { AppStatus } from "@/global-state/app-status";

/**
 * Common player props
 */
interface PlayerCommonProps {
    /**
     * Player reload rick
     */
    rTick: number;

    /**
     * Next element in album
     */
    next?: MediaListItem | null;

    /**
     * Previous element in album
     */
    prev?: MediaListItem | null;

    /**
     * Has next element in page
     */
    pageNext?: boolean;

    /**
     * Has prev element in page
     */
    pagePrev?: boolean;
}

/**
 * Player common features composable
 */
export type PlayerCommonComposable = {
    /**
     * True if title is expanded
     */
    expandedTitle: Ref<boolean>;

    /**
     * True if expanded album in full screen
     */
    expandedAlbum: Ref<boolean>;

    /**
     * Type of displayed tooltip
     */
    helpTooltip: Ref<HelpTooltipType>;

    /**
     * Enters a tooltip
     * @param t The tooltip
     */
    enterTooltip: (t: HelpTooltipType) => void;

    /**
     * Leaves a tooltip
     * @param t The tooltip
     */
    leaveTooltip: (t: HelpTooltipType) => void;

    /**
     * Clears the tooltip
     */
    clearTooltip: () => void;

    /**
     * Goes to the next element
     */
    goNext: () => void;

    /**
     * Goes to the previous element
     */
    goPrev: () => void;

    /**
     * Toggles the full screen status
     */
    toggleFullScreen: () => void;
};

/**
 * Player common emits
 */
export type PlayerCommonEmits = {
    /**
     * Go to the next media
     */
    (e: "go-next"): void;

    /**
     * Go to the previous media
     */
    (e: "go-prev"): void;

    /**
     * The user wants to delete the media
     */
    (e: "delete"): void;
};

// Priority for player keyboard handlers
export const PLAYER_KEYBOARD_HANDLER_PRIORITY = 100;

/**
 * Gets the player common features for the composition API
 * @param props The player props
 * @param emit The player emit function
 * @param fullScreen The full screen model
 * @returns The composable
 */
export function usePlayerCommon(props: PlayerCommonProps, emit: PlayerCommonEmits, fullScreen: Ref<boolean>): PlayerCommonComposable {
    // True if expanded title
    const expandedTitle = ref(false);

    // Close the title when the media reloads
    watch(
        () => props.rTick,
        () => {
            expandedTitle.value = false;
        },
    );

    // True if expanded album in full screen
    const expandedAlbum = ref(false);

    // Type of displayed tooltip
    const helpTooltip = ref<HelpTooltipType>("");

    /**
     * Enters a tooltip
     * @param t The tooltip
     */
    const enterTooltip = (t: HelpTooltipType) => {
        if (isTouchDevice()) {
            helpTooltip.value = "";
            return;
        }
        helpTooltip.value = t;
    };

    /**
     * Leaves a tooltip
     * @param t The tooltip
     */
    const leaveTooltip = (t: HelpTooltipType) => {
        if (t === helpTooltip.value) {
            helpTooltip.value = "";
        }
    };

    /**
     * Clears the tooltip
     */
    const clearTooltip = () => {
        helpTooltip.value = "";
    };

    /**
     * Goes to the next element
     */
    const goNext = () => {
        if (props.next || props.pageNext) {
            emit("go-next");
        }
    };

    /**
     * Goes to the previous element
     */
    const goPrev = () => {
        if (props.prev || props.pagePrev) {
            emit("go-prev");
        }
    };

    /**
     * Toggles the full screen status
     */
    const toggleFullScreen = () => {
        if (!fullScreen.value) {
            openFullscreen();
        } else {
            closeFullscreen();
        }
        fullScreen.value = !fullScreen.value;
    };

    // Global keyboard handler
    useGlobalKeyboardHandler((event: KeyboardEvent): boolean => {
        if (AuthController.Locked || !AppStatus.IsPlayerVisible() || !event.key || event.ctrlKey) {
            return false;
        }
        let caught = true;

        switch (event.key) {
            case "F":
            case "f":
                if (event.altKey || event.shiftKey) {
                    caught = false;
                } else {
                    toggleFullScreen();
                }
                break;
            case "PageUp":
                if (event.altKey || event.shiftKey) {
                    caught = false;
                } else if (props.prev || props.pagePrev) {
                    goPrev();
                } else {
                    caught = false;
                }
                break;
            case "PageDown":
                if (event.altKey || event.shiftKey) {
                    caught = false;
                } else if (props.next || props.pageNext) {
                    goNext();
                } else {
                    caught = false;
                }
                break;
            case "Delete":
                emit("delete");
                break;
            default:
                caught = false;
        }

        return caught;
    }, PLAYER_KEYBOARD_HANDLER_PRIORITY);

    return {
        expandedTitle,
        expandedAlbum,
        helpTooltip,
        enterTooltip,
        leaveTooltip,
        clearTooltip,
        goNext,
        goPrev,
        toggleFullScreen,
    };
}
