<template>
    <div
        ref="container"
        class="player-scale-control"
        :class="{ expanded: expanded, 'player-min': min }"
        :style="{ width: fulWidth }"
        @mouseenter="onEnter"
        @mouseleave="onLeave"
    >
        <button class="player-scale-btn" @click="toggleFit">
            <i class="fas fa-magnifying-glass"></i>
        </button>
        <div
            class="player-scale-btn-expand"
            :class="{ hidden: !expanded }"
            :style="{ width: barContainerWidth }"
            @mousedown="grabMouse"
            @touchstart.passive="grabTouch"
        >
            <div class="player-scale-bar-container" :style="{ width: barContainerInnerWidth }">
                <div class="player-scale-bar" :style="{ width: barWidth }"></div>
                <div class="player-scale-current" :style="{ width: barCurrentWidth }"></div>
                <div class="player-scale-thumb" :style="{ left: thumbLeft }"></div>
            </div>
        </div>
    </div>
</template>

<script setup lang="ts">
import { useTemplateRef } from "vue";
import { usePlayerSlider } from "@/composables/use-player-slider";

// Ref to the container element
const container = useTemplateRef("container");

// Props
const props = defineProps({
    /**
     * Width of the control (px)
     */
    width: {
        type: Number,
        required: true,
    },

    /**
     * Miniature mode?
     */
    min: Boolean,
});

// Emits
const emit = defineEmits<{
    /**
     * The user enters the control
     */
    (e: "enter"): void;

    /**
     * The user leaves the control
     */
    (e: "leave"): void;
}>();

// Scale
const scale = defineModel<number>("scale");

// Fit image?
const fit = defineModel<boolean>("fit");

// Expanded?
const expanded = defineModel<boolean>("expanded");

// Player slider common
const {
    fullWidth: fulWidth,
    barContainerWidth,
    barContainerInnerWidth,
    barWidth,
    barCurrentWidth,
    thumbLeft,
    expand,
    grabMouse,
    grabTouch,
} = usePlayerSlider(container, props, scale, fit, expanded);

/**
 * The user enters the control
 */
const onEnter = () => {
    emit("enter");
    expand();
};

/**
 * The user leaves the control
 */
const onLeave = () => {
    emit("leave");
};

/**
 * Toggles the fit value
 */
const toggleFit = () => {
    fit.value = !fit.value;
};
</script>
