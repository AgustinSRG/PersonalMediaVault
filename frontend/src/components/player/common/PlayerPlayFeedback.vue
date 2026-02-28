<template>
    <div class="player-feedback-container">
        <div v-if="feedback === 'play'" key="play" class="player-feedback player-feedback-play" @animationend="onFeedBackAnimationEnd">
            <div><i class="fas fa-play"></i></div>
        </div>
        <div v-if="feedback === 'pause'" key="pause" class="player-feedback player-feedback-pause" @animationend="onFeedBackAnimationEnd">
            <div><i class="fas fa-pause"></i></div>
        </div>
    </div>
</template>

<script setup lang="ts">
import type { PlayerPlayFeedbackType } from "@/utils/player";

// Feedback model
const feedback = defineModel<PlayerPlayFeedbackType>("feedback");

/**
 * Called when feedback animation ends
 */
const onFeedBackAnimationEnd = () => {
    feedback.value = "";
};
</script>
