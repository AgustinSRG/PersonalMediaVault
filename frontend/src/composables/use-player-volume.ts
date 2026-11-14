// Volume composable

"use strict";

import { getPlayerMuted, getPlayerVolume, getVolumeMultiplier, setPlayerMuted, setPlayerVolume } from "@/local-storage/player-preferences";
import type { Ref } from "vue";
import { computed, ref, watch } from "vue";
import { onApplicationEvent } from "./on-app-event";
import { EVENT_NAME_VOLUME_MULTIPLIER_UPDATED } from "@/global-state/app-events";

/**
 * Volume composable type
 */
export type ComposablePlayerVolume = {
    /**
     * Volume ref
     */
    volume: Ref<number>;

    /**
     * Volume multiplier
     */
    volumeMultiplier: Ref<number>;

    /**
     * Effective volume
     */
    effectiveVolume: Ref<number>;

    /**
     * Volume muted?
     */
    muted: Ref<boolean>;

    /**
     * Function to call when the volume is updated by the user
     */
    onUserVolumeUpdated: () => void;

    /**
     * Changes the volume
     * @param v The new volume
     */
    changeVolume: (v: number) => void;

    /**
     * Function to call when the muted setting is updated by the user
     */
    onUserMutedUpdated: () => void;

    /**
     * Toggles the muted state
     */
    toggleMuted: () => void;
};

export function usePlayerVolume(mid: Ref<number>): ComposablePlayerVolume {
    // Current volume
    const volume = ref(getPlayerVolume());

    // Volume multiplier
    const volumeMultiplier = ref(getVolumeMultiplier(mid.value));

    onApplicationEvent(EVENT_NAME_VOLUME_MULTIPLIER_UPDATED, (mediaIdChanged: number, newMultiplier: number) => {
        if (mid.value !== mediaIdChanged) {
            return;
        }

        volumeMultiplier.value = newMultiplier;
    });

    watch(mid, () => {
        volumeMultiplier.value = getVolumeMultiplier(mid.value);
    });

    // Effective volume
    const effectiveVolume = computed<number>(() => volume.value * volumeMultiplier.value);

    /**
     * Called when the volume is updated by the user
     */
    const onUserVolumeUpdated = () => {
        setPlayerVolume(volume.value);
    };

    // Muted?
    const muted = ref(getPlayerMuted());

    /**
     * Called when the muted state is changed by the user
     */
    const onUserMutedUpdated = () => {
        setPlayerMuted(muted.value);
    };

    /**
     * Changes the volume
     * @param v The new volume
     */
    const changeVolume = (v: number) => {
        volume.value = v;
        onUserVolumeUpdated();
    };

    /**
     * Toggles the muted state
     */
    const toggleMuted = () => {
        muted.value = !muted.value;
        onUserMutedUpdated();
    };

    return {
        volume,
        volumeMultiplier,
        effectiveVolume,
        muted,
        onUserVolumeUpdated,
        changeVolume,
        onUserMutedUpdated,
        toggleMuted,
    };
}
