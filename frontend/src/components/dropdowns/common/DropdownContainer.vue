<template>
    <div ref="container" :class="containerClass" tabindex="-1" role="dialog" @mousedown="close" @touchstart="close" @keydown="onKeyDown">
        <div
            v-if="display"
            class="modal-dialog modal-sm"
            role="document"
            @click="stopPropagationEvent"
            @mousedown="stopPropagationEvent"
            @touchstart="stopPropagationEvent"
        >
            <slot></slot>
        </div>
    </div>
</template>

<script setup lang="ts">
import { stopPropagationEvent } from "@/utils/events";
import { FocusTrap } from "@/utils/focus-trap";
import { computed, nextTick, onBeforeUnmount, onMounted, useTemplateRef, watch } from "vue";

const props = defineProps({
    /**
     * CSS class to change position of the dropdown
     */
    positionClass: String,

    /**
     * Exception class for the focus trap
     */
    focusTrapExceptionClass: String,
});

// Display model
const display = defineModel<boolean>("display");

// CSS class for the container
const containerClass = computed(() => {
    const classList = ["modal-container", "no-transition", "modal-container-corner"];

    if (props.positionClass) {
        classList.push(props.positionClass);
    }

    if (!display.value) {
        classList.push("hidden");
    }

    return classList.join(" ");
});

/**
 * Closes the dropdown
 */
const close = () => {
    display.value = false;
};

/**
 * Event handler for 'keydown'
 * @param e The keyboard event
 */
const onKeyDown = (e: KeyboardEvent) => {
    e.stopPropagation();
    if (e.key === "Escape") {
        close();
    }
};

// Ref to the container element
const container = useTemplateRef("container");

/**
 * Automatically focuses the container element
 */
const autoFocus = () => {
    nextTick(() => {
        container.value?.focus();
    });
};

// Focus trap
let focusTrap: null | FocusTrap = null;

onMounted(() => {
    focusTrap = new FocusTrap(container.value, close, props.focusTrapExceptionClass);

    if (display.value) {
        focusTrap.activate();
        autoFocus();
    }
});

watch(display, () => {
    if (display.value) {
        focusTrap?.activate();
        autoFocus();
    } else {
        focusTrap?.deactivate();
    }
});

onBeforeUnmount(() => {
    focusTrap?.destroy();
});
</script>
