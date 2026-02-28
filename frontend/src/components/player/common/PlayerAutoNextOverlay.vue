<template>
    <div
        class="player-auto-next-overlay"
        @click="stopPropagationEvent"
        @mousedown="stopPropagationEvent"
        @touchstart="stopPropagationEvent"
    >
        <div class="next-end-container">
            <div class="next-end-wait-msg">
                {{ $t("Next media will play in") }} <b>{{ pendingNextEndSeconds }}</b>
            </div>
            <div class="next-end-wait-buttons">
                <button type="button" class="btn btn-primary" @click="cancel"><i class="fas fa-times"></i> {{ $t("Cancel") }}</button>
                <button type="button" class="btn btn-primary" @click="next"><i class="fas fa-forward-step"></i> {{ $t("Next") }}</button>
            </div>
            <div class="next-end-wait-buttons">
                <button type="button" class="btn btn-primary" @click="play"><i class="fas fa-repeat"></i> {{ $t("Play again") }}</button>
                <button type="button" class="btn btn-primary" @click="loop"><i class="fas fa-repeat"></i> {{ $t("Loop") }}</button>
            </div>
        </div>
    </div>
</template>

<script setup lang="ts">
import { useI18n } from "@/composables/use-i18n";
import { stopPropagationEvent } from "@/utils/events";

// Translation
const { $t } = useI18n();

// Props
defineProps({
    /**
     * Number of seconds to go next automatically
     */
    pendingNextEndSeconds: {
        type: Number,
        required: true,
    },
});

// Emits
const emit = defineEmits<{
    /**
     * The user wants to stop going to the next element
     */
    (e: "cancel"): void;

    /**
     * The user wants to go next
     */
    (e: "next"): void;

    /**
     * The user wants to play again
     */
    (e: "play"): void;

    /**
     * The user wants to enable loop and play again
     */
    (e: "loop"): void;
}>();

/**
 * The user wants to stop going to the next element
 */
const cancel = () => {
    emit("cancel");
};

/**
 * The user wants to go next
 */
const next = () => {
    emit("next");
};

/**
 * The user want to play again
 */
const play = () => {
    emit("play");
};

/**
 * The user wants to enable loop and play again
 */
const loop = () => {
    emit("loop");
};
</script>
