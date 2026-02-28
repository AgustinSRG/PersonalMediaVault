<template>
    <div class="player-tooltip" :style="{ left: x + 'px' }">
        <div v-if="image && !imageInvalid" class="player-tooltip-image">
            <img :src="image" @error="onImageError" @load="onImageLoaded" />
            <div v-if="imageLoading" class="player-tooltip-image-loading">
                <div class="player-tooltip-image-loader">
                    <div></div>
                    <div></div>
                    <div></div>
                    <div></div>
                </div>
            </div>
        </div>
        <div class="player-tooltip-text">{{ text }}</div>
        <div v-if="timeSlice" class="player-tooltip-text">
            {{ timeSlice }}
        </div>
    </div>
</template>

<script setup lang="ts">
import { ref, watch } from "vue";

// Props
const props = defineProps({
    /**
     * Tooltip X position
     */
    x: {
        type: Number,
        required: true,
    },

    /**
     * Text to display in the tooltip
     */
    text: {
        type: String,
        required: true,
    },

    /**
     * Name of the time slice
     */
    timeSlice: String,

    /**
     * Image
     */
    image: String,
});

// Loading image?
const imageLoading = ref(true);

// Invalid image?
const imageInvalid = ref(false);

watch(
    () => props.image,
    (newVal: string, oldVal: string) => {
        if (newVal === oldVal) {
            return;
        }

        imageLoading.value = true;
        imageInvalid.value = false;
    },
);

/**
 * Called when image is fully loaded
 */
const onImageLoaded = () => {
    imageLoading.value = false;
};

/**
 * Called when the image fails to load
 */
const onImageError = () => {
    imageInvalid.value = true;
};
</script>
