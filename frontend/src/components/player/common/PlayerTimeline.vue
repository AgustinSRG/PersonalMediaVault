<template>
    <div
        class="player-timeline"
        :class="{ hidden: hidden }"
        @mouseenter="enter"
        @mouseleave="leave"
        @mousemove="mouseMoveTimeline"
        @dblclick="stopPropagationEvent"
        @click="clickTimeline"
        @mousedown="grabTimelineByMouse"
        @touchstart.passive="grabTimelineByTouch"
    >
        <div class="player-timeline-back"></div>
        <div class="player-timeline-buffer" :style="{ width: bufferedWidth }"></div>
        <div class="player-timeline-current" :style="{ width: currentWidth }"></div>

        <div
            v-for="(ts, tsi) in timeSlices"
            :key="tsi"
            class="player-timeline-split"
            :class="{ 'start-split': ts.start <= 0 }"
            :style="{ left: getTimelineBarWidth(ts.start, duration) }"
        ></div>

        <div class="player-timeline-thumb" :style="{ left: thumbLeft }"></div>
    </div>
</template>
<script setup lang="ts">
import { stopPropagationEvent } from "@/utils/events";
import type { NormalizedTimeSlice } from "@/utils/time-slices";
import { isTouchDevice } from "@/utils/touch";
import type { PropType } from "vue";
import { computed } from "vue";

// Props
const props = defineProps({
    /**
     * True to hide timeline
     */
    hidden: {
        type: Boolean,
        required: true,
    },

    /**
     * Media duration
     */
    duration: {
        type: Number,
        required: true,
    },

    /**
     * Buffered time
     */
    bufferedTime: {
        type: Number,
        required: true,
    },

    /**
     * Current time
     */
    currentTime: {
        type: Number,
        required: true,
    },

    /**
     * List of time slices
     */
    timeSlices: {
        type: Array as PropType<NormalizedTimeSlice[]>,
        required: true,
    },
});

// Emits
const emit = defineEmits<{
    /**
     * Mouse enters the timeline
     */
    (e: "enter"): void;

    /**
     * Mouse leaves the timeline
     */
    (e: "leave"): void;

    /**
     * Updates the tooltip based on the position
     * of the mouse / touch
     */
    (e: "update-tooltip", x: number): void;

    /**
     * Click
     */
    (e: "click", event: Event): void;

    /**
     * Grab the timeline
     */
    (e: "grab", x: number): void;
}>();

/**
 * Gets timeline bar width
 * @param time The time
 * @param duration The duration
 */
const getTimelineBarWidth = (time: number, duration: number) => {
    if (duration > 0) {
        return Math.min((time / duration) * 100, 100) + "%";
    } else {
        return "0";
    }
};

// Buffered bar width
const bufferedWidth = computed(() => getTimelineBarWidth(props.bufferedTime, props.duration));

// Current time bar width
const currentWidth = computed(() => getTimelineBarWidth(props.currentTime, props.duration));

// Thumb left position
const thumbLeft = computed(() => {
    if (props.duration > 0) {
        return "calc(" + Math.min((props.currentTime / props.duration) * 100, 100) + "% - 7px)";
    } else {
        return "-7px";
    }
});

/**
 * Mouse enters the timeline
 */
const enter = () => {
    emit("enter");
};

/**
 * Mouse leaves the timeline
 */
const leave = () => {
    emit("leave");
};

/**
 * Event handler for 'mousemove'
 * @param e The mouse event
 */
const mouseMoveTimeline = (e: MouseEvent) => {
    if (isTouchDevice()) {
        return;
    }

    emit("update-tooltip", e.pageX);
};

/**
 * Event handler for 'click'
 * @param e The click event
 */
const clickTimeline = (e: Event) => {
    emit("click", e);
};

/**
 * Grabs timeline using the mouse
 * @param e The mouse event
 */
const grabTimelineByMouse = (e: MouseEvent) => {
    e.stopPropagation();
    if (e.button === 0) {
        emit("grab", e.pageX);
    }
};

/**
 * Grabs timeline using the touch screen
 * @param e The touch event
 */
const grabTimelineByTouch = (e: TouchEvent) => {
    e.stopPropagation();

    if (!e.touches[0]) {
        return;
    }

    const x = e.touches[0].pageX;

    emit("grab", x);

    emit("update-tooltip", x);
};
</script>
