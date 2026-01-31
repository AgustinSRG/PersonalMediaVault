<template>
    <i :class="getClass(displayLoader, icon, extraClass)"><slot></slot></i>
</template>

<script setup lang="ts">
import { useTimeout } from "@/composables/use-timeout";
import { onMounted, ref, watch } from "vue";

// Default delay (milliseconds)
const DEFAULT_DELAY = 333;

/**
 * Generates the icon class list
 * @param displayLoader True if the loader should display
 * @param icon The icon when not loading
 * @param extraClass Extra class
 */
const getClass = (displayLoader: boolean, icon: string, extraClass: string): string => {
    const classes = [];

    if (displayLoader) {
        classes.push("fa", "fa-spinner", "fa-spin");
    } else if (icon) {
        classes.push(icon);
    }

    if (extraClass) {
        classes.push(extraClass);
    }

    return classes.join(" ");
};

const props = defineProps({
    /**
     * Icon to display when not loading
     */
    icon: String,

    /**
     * Extra class to add to the icon
     */
    extraClass: String,

    /**
     * True if the icon should display the loading status
     */
    loading: Boolean,

    /**
     * Custom delay for the loading status to display,
     * in milliseconds
     */
    delay: Number,
});

// Timeout for displaying
const displayTimeout = useTimeout();

// True to display the loader icon
const displayLoader = ref(false);

/**
 * Call when 'loading' changes to true
 */
const onLoadingTrue = () => {
    displayTimeout.set(() => {
        displayLoader.value = true;
    }, props.delay || DEFAULT_DELAY);
};

/**
 * Call when 'loading' changes to false
 */
const onLoadingFalse = () => {
    displayTimeout.clear();

    displayLoader.value = false;
};

onMounted(() => {
    if (props.loading) {
        onLoadingTrue();
    }
});

watch(
    () => props.loading,
    () => {
        if (props.loading) {
            onLoadingTrue();
        } else {
            onLoadingFalse();
        }
    },
);
</script>
