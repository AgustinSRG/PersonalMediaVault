<template>
    <div
        ref="container"
        class="player-volume-control"
        :class="{ expanded: expanded, 'player-min': min }"
        :style="{ width: fulWidth }"
        @mouseenter="onEnter"
        @mouseleave="onLeave"
    >
        <button class="player-volume-btn" :title="$t('Volume')" @click="toggleMuted">
            <i v-if="!muted && volume > 0.5" class="fas fa-volume-up"></i>
            <i v-if="!muted && volume <= 0" class="fas fa-volume-off"></i>
            <i v-if="!muted && volume > 0 && volume <= 0.5" class="fas fa-volume-down"></i>
            <i v-if="muted" class="fas fa-volume-mute"></i>
        </button>
        <div
            class="player-volume-btn-expand"
            :class="{ hidden: !expanded }"
            :style="{ width: barContainerWidth }"
            @mousedown="grabMouse"
            @touchstart.passive="grabTouch"
        >
            <div class="player-volume-bar-container" :style="{ width: barContainerInnerWidth }">
                <div class="player-volume-bar" :style="{ width: barWidth }"></div>
                <div class="player-volume-current" :style="{ width: barCurrentWidth }"></div>
                <div class="player-volume-thumb" :style="{ left: thumbLeft }"></div>
            </div>
        </div>
    </div>
</template>

<script setup lang="ts">
import { useTemplateRef } from "vue";
import { useI18n } from "@/composables/use-i18n";
import { usePlayerSlider } from "@/composables/use-player-slider";

// Ref to the container element
const container = useTemplateRef("container");

// Translation
const { $t } = useI18n();

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

// Volume
const volume = defineModel<number>("volume");

// Muted
const muted = defineModel<boolean>("muted");

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
} = usePlayerSlider(container, props, volume, muted, expanded);

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
 * Toggles the muted value
 */
const toggleMuted = () => {
    muted.value = !muted.value;
};
</script>
