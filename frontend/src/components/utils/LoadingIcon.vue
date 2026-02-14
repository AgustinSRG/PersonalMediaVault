<template>
    <i :class="computedClass"><slot></slot></i>
</template>

<script setup lang="ts">
import { useTimeout } from "@/composables/use-timeout";
import { computed, onMounted, ref, watch } from "vue";

// Default delay (milliseconds)
const DEFAULT_DELAY = 333;

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

// True to display the loader icon
const displayLoader = ref(false);

// Computed class for the icon
const computedClass = computed(() => {
    const classes = [];

    if (displayLoader.value) {
        classes.push("fa", "fa-spinner", "fa-spin");
    } else if (props.icon) {
        classes.push(props.icon);
    }

    if (props.extraClass) {
        classes.push(props.extraClass);
    }

    return classes.join(" ");
});

// Timeout for displaying
const displayTimeout = useTimeout();

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
