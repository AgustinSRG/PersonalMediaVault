// Player time slices composable

"use strict";

import type { MediaData } from "@/api/models";
import { getCurrentMediaData } from "@/global-state/media";
import { renderTimeSeconds } from "@/utils/time";
import type { NormalizedTimeSlice } from "@/utils/time-slices";
import { findTimeSlice, normalizeTimeSlices } from "@/utils/time-slices";
import type { Ref } from "vue";
import { ref } from "vue";

/**
 * Required props for the time slices functionality
 */
export type PlayerTimeSlicesRequiredProps = {
    /**
     * Media metadata
     */
    metadata?: MediaData;
};

/**
 * Player time slices composable
 */
export type PlayerTimeSlicesComposable = {
    /**
     * List of time slices
     */
    timeSlices: Ref<NormalizedTimeSlice[]>;

    /**
     * Current time slice
     */
    currentTimeSlice: Ref<NormalizedTimeSlice | null>;

    /**
     * Current time slice name
     */
    currentTimeSliceName: Ref<string>;

    /**
     * True if time slices are being edited
     */
    timeSlicesEdit: Ref<boolean>;

    /**
     * True to enable loop for the current time slice
     */
    sliceLoop: Ref<boolean>;

    /**
     * Initializes the time slices from the metadata
     */
    initializeTimeSlices: () => void;

    /**
     * Finds time slice name
     * @param time The time
     * @returns The time slice name, or en empty string
     */
    findTimeSliceName: (time: number) => string;

    /**
     * Updates current time slice
     */
    updateCurrentTimeSlice: () => void;

    /**
     * Refreshes the time slices
     */
    refreshTimeSlices: () => void;
};

/**
 * Gets the player time slices composable
 * @param props The player component props
 * @param currentTime The current time ref
 * @param setTime A function to set the time
 * @returns The composable
 */
export function usePlayerTimeSlices(
    props: PlayerTimeSlicesRequiredProps,
    currentTime: Ref<number>,
    setTime: (time: number, save?: boolean) => void,
): PlayerTimeSlicesComposable {
    // List of time slices
    const timeSlices = ref<NormalizedTimeSlice[]>([]);

    // Current time slice (based on current time)
    const currentTimeSlice = ref<NormalizedTimeSlice | null>(null);
    const currentTimeSliceName = ref("");
    const currentTimeSliceStart = ref(0);
    const currentTimeSliceEnd = ref(0);

    // True if edit mode for time slices is enabled
    const timeSlicesEdit = ref(false);

    // True to enable loop for the current time slice
    const sliceLoop = ref(false);

    /**
     * Initializes the time slices from the metadata
     */
    const initializeTimeSlices = () => {
        timeSlices.value = normalizeTimeSlices(
            (props.metadata.time_slices || []).sort((a, b) => {
                if (a.time < b.time) {
                    return -1;
                } else if (a.time > b.time) {
                    return 1;
                } else {
                    return 0;
                }
            }),
            props.metadata.duration,
        );

        currentTimeSlice.value = null;
        currentTimeSliceName.value = "";
        currentTimeSliceStart.value = 0;
        currentTimeSliceEnd.value = 0;

        sliceLoop.value = false;
    };

    /**
     * Find the time slice by the current time
     * @param time The current time
     */
    const findTimeSliceName = (time: number) => {
        const slice = findTimeSlice(timeSlices.value, time);
        if (slice) {
            return slice.name + " (" + renderTimeSeconds(slice.end - slice.start) + ")";
        } else {
            return "";
        }
    };

    /**
     * Updates the current time slice based on the current time
     */
    const updateCurrentTimeSlice = () => {
        if (
            currentTimeSlice.value &&
            sliceLoop.value &&
            currentTime.value >= currentTimeSlice.value.end &&
            currentTimeSlice.value.end > currentTimeSlice.value.start
        ) {
            setTime(currentTimeSlice.value.start, false);
            return;
        }

        const slice = findTimeSlice(timeSlices.value, currentTime.value);
        if (slice) {
            currentTimeSlice.value = slice;
            currentTimeSliceName.value = slice.name;
            currentTimeSliceStart.value = slice.start;
            currentTimeSliceEnd.value = slice.end;
        } else {
            currentTimeSlice.value = null;
            currentTimeSliceName.value = "";
            currentTimeSliceStart.value = 0;
            currentTimeSliceEnd.value = 0;
        }
    };

    /**
     * Called when time slices need to be refreshed
     */
    const refreshTimeSlices = () => {
        const metadata = getCurrentMediaData();

        if (!metadata) {
            return;
        }

        timeSlices.value = normalizeTimeSlices(
            (metadata.time_slices || []).sort((a, b) => {
                if (a.time < b.time) {
                    return -1;
                } else if (a.time > b.time) {
                    return 1;
                } else {
                    return 0;
                }
            }),
            metadata.duration,
        );

        currentTimeSlice.value = null;
        currentTimeSliceName.value = "";
        currentTimeSliceStart.value = 0;
        currentTimeSliceEnd.value = 0;

        updateCurrentTimeSlice();
    };

    return {
        timeSlices,
        currentTimeSlice,
        currentTimeSliceName,
        timeSlicesEdit,
        sliceLoop,
        initializeTimeSlices,
        findTimeSliceName,
        updateCurrentTimeSlice,
        refreshTimeSlices,
    };
}
