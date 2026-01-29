<template>
    <div :class="'thumb' + (className ? ' ' + className : '')">
        <img v-if="!isError" ref="image" :src="src" loading="lazy" @load="onSuccessfulImageLoad" @error="onError" />
        <div v-if="displayLoader" class="thumb-load">
            <i class="fa fa-spinner fa-spin"></i>
        </div>
    </div>
</template>

<script setup lang="ts">
import { useTimeout } from "@/composables/use-timeout";
import { onMounted, ref, useTemplateRef } from "vue";

// Delay to show the loader (milliseconds)
const LOADER_VISIBILITY_DELAY = 333;

// Delay to reload after error (milliseconds)
const RELOAD_DELAY = 1500;

defineProps({
    /**
     * Source URL of the thumbnail image
     */
    src: {
        type: String,
        required: true,
    },

    /**
     * A CSS class name for the image container
     */
    className: String,
});

// Timeout to display the loader after some time passed while loading
const displayLoaderTimeout = useTimeout();

// Timeout for reloading the image
const reloadTimeout = useTimeout();

// True if the loader is visible, false otherwise
const displayLoader = ref(false);

// True if the image loading process ended with an error, false otherwise
const isError = ref(false);

/**
 * Function called when the image starts loading.
 * This can occur:
 *  - On component mount
 *  - When the source URL change
 */
const startLoading = () => {
    isError.value = false;

    reloadTimeout.clear();

    displayLoaderTimeout.set(() => {
        displayLoader.value = true;
    }, LOADER_VISIBILITY_DELAY);
};

/**
 * Function called when the image is successfully loaded.
 */
const onSuccessfulImageLoad = () => {
    displayLoaderTimeout.clear();
    reloadTimeout.clear();

    displayLoader.value = false;
};

/**
 * Function called when the image loading results in an error.
 * It will retry to load the image after certain delay.
 */
const onError = () => {
    isError.value = true;

    reloadTimeout.set(() => {
        startLoading();
    }, RELOAD_DELAY);
};

// Reference to the image element
const imageElementRef = useTemplateRef("image");

// When counted, if the image is not instantly loaded,
// call 'startLoading' in order to initiate the loading process
onMounted(() => {
    const elem = imageElementRef.value;
    if (!elem || !elem.complete) {
        startLoading();
    }
});
</script>
