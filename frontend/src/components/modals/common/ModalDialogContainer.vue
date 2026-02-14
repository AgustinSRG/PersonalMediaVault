<template>
    <div
        ref="container"
        class="modal-container modal-container-dialog"
        :class="{ hidden: !display, closing: closing }"
        tabindex="-1"
        role="dialog"
        @keydown="onKeyDown"
        @animationend="onAnimationEnd"
        @mousedown="stopPropagationEvent"
        @touchstart="stopPropagationEvent"
        @click="stopPropagationEvent"
        @mouseup="onMouseUp"
        @touchend.passive="onTouchEnd"
        @mouseleave="onMouseLeave"
    >
        <div class="modal-out-close-area" @mousedown="onMouseDown" @touchstart="onTouchOutside"></div>
        <slot @mousedown="stopPropagationEvent" @touchstart="stopPropagationEvent" @click="stopPropagationEvent"></slot>
    </div>
</template>

<script setup lang="ts">
import { useFocusTrap } from "@/composables/use-focus-trap";
import { stopPropagationEvent } from "@/utils/events";
import type { PropType } from "vue";
import { ref, useTemplateRef, watch } from "vue";

const emit = defineEmits<{
    /**
     * Event emitted when the modal closes
     */
    (e: "close"): void;

    /**
     * Keyboard event (key pressed)
     */
    (e: "key", event: KeyboardEvent): void;

    /**
     * Forward for 'mouseup' event
     */
    (e: "mouseup", event: MouseEvent): void;

    /**
     * Forward for 'mouseleave' event
     */
    (e: "mouseleave", event: MouseEvent): void;

    /**
     * Forward for 'touchend' event
     */
    (e: "touchend", event: TouchEvent): void;
}>();

const props = defineProps({
    /**
     * True for the modal to not being able to be closed
     * by clicking outside
     */
    static: Boolean,

    /**
     * True if the modal cannot be closed
     */
    lockClose: Boolean,

    /**
     * Callback to prevent closing the modal.
     * The callback will be passed the 'close' function as argument
     * and may call it to confirm closing the modal.
     */
    closeCallback: Function as PropType<(close: () => void) => void>,
});

// Display model
const display = defineModel<boolean>("display");

// True if the modal is closing
const closing = ref(false);

// Ref to the container element
const container = useTemplateRef("container");

// Called when focus on the modal is lost
const onFocusLost = () => {
    if (display.value && !closing.value) {
        container.value?.focus();
    }
};

// Focus trap
useFocusTrap(container, display, onFocusLost);

watch(display, () => {
    if (display.value) {
        closing.value = false;
    }
});

/**
 * Closes the modal
 * @param forced True if forced
 */
const close = (forced?: boolean) => {
    if (props.lockClose && forced !== true) {
        return;
    }
    if (props.closeCallback && forced !== true) {
        props.closeCallback(() => {
            emit("close");
            closing.value = true;
        });
    } else {
        emit("close");
        closing.value = true;
    }
};

// Listeners for events

/**
 * Called when the closing animation ends
 * @param e The animation event
 */
const onAnimationEnd = (e: AnimationEvent) => {
    e.stopPropagation();
    if (e.animationName === "modal-close-animation") {
        closing.value = false;
        display.value = false;
    }
};

/**
 * Called when the modal received the 'keydown' event
 * @param e The keyboard event
 */
const onKeyDown = (e: KeyboardEvent) => {
    e.stopPropagation();
    if (e.key === "Escape" && display.value && !closing.value) {
        close();
    } else {
        emit("key", e);
    }
};

/**
 * Called when a touch event is received outside the modal
 * @param e The touch event
 */
const onTouchOutside = (e: TouchEvent) => {
    e.stopPropagation();
    if (!props.static) {
        close();
    }
};

/**
 * Called when the modal received the 'mousedown' event
 * @param e The mouse event
 */
const onMouseDown = (e: MouseEvent) => {
    e.stopPropagation();
    if (e.button === 0 && !props.static) {
        close();
    }
};

/**
 * Called when the modal received the 'mouseup' event
 * @param e The mouse event
 */
const onMouseUp = (e: MouseEvent) => {
    emit("mouseup", e);
};

/**
 * Called when the modal received the 'touchend' event
 * @param e The touch event
 */
const onTouchEnd = (e: TouchEvent) => {
    emit("touchend", e);
};

/**
 * Called when the modal received the 'mouseleave' event
 * @param e The mouse event
 */
const onMouseLeave = (e: MouseEvent) => {
    emit("mouseleave", e);
};

// Name of the class to auto focus
const AUTO_FOCUS_CLASS_NAME = "auto-focus";

/**
 * Focuses the container element
 */
const focus = () => {
    const focusElement = container.value?.querySelector(`.${AUTO_FOCUS_CLASS_NAME}`) as HTMLElement;
    if (focusElement) {
        focusElement.focus();

        if (focusElement.classList.contains("auto-select") && typeof (focusElement as HTMLInputElement).select === "function") {
            (focusElement as HTMLInputElement).select();
        }
    } else {
        container.value?.focus();
    }
};

/**
 * Scrolls the container to the top
 */
const scrollToTop = () => {
    if (container.value) {
        container.value.scrollTop = 0;
        container.value.focus();
    }
};

// Expose methods
defineExpose({
    close,
    focus,
    scrollToTop,
});
</script>
