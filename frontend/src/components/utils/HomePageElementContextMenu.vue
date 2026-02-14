<template>
    <div
        ref="container"
        class="album-body-item-options-menu"
        :class="{
            hidden: !shown,
        }"
        :style="{
            top: dimensions.top,
            left: dimensions.left,
            right: dimensions.right,
            bottom: dimensions.bottom,
            width: dimensions.width,
            'max-width': dimensions.maxWidth,
            'max-height': dimensions.maxHeight,
        }"
        tabindex="-1"
        @mousedown="stopPropagationEvent"
        @touchstart="stopPropagationEvent"
        @click="stopPropagationEvent"
        @dblclick="stopPropagationEvent"
    >
        <div tabindex="0" class="album-body-item-options-menu-btn" @keydown="clickOnEnter" @click="changePosition">
            <i class="fas fa-arrows-up-down-left-right"></i>
            {{ $t("Change position") }}
        </div>
        <div tabindex="0" class="album-body-item-options-menu-btn" @keydown="clickOnEnter" @click="removeElement">
            <i class="fas fa-trash-alt"></i> {{ $t("Remove from row") }}
        </div>
    </div>
</template>

<script setup lang="ts">
import { reactive, useTemplateRef, watch } from "vue";
import { stopPropagationEvent, clickOnEnter } from "@/utils/events";
import { onDocumentEvent } from "@/composables/on-document-event";
import { useFocusTrap } from "@/composables/use-focus-trap";
import { useI18n } from "@/composables/use-i18n";

// Translation function
const { $t } = useI18n();

// Shown model
const shown = defineModel<boolean>("shown");

/**
 * Hides the context menu
 */
const hide = () => {
    shown.value = false;
};

onDocumentEvent("mousedown", hide);
onDocumentEvent("touchstart", hide);

// Props
const props = defineProps({
    /**
     * Index of the element in the list
     */
    elementIndex: {
        type: Number,
        required: true,
    },

    /**
     * X coordinate of the context menu
     */
    x: {
        type: Number,
        required: true,
    },

    /**
     * Y coordinate of the context menu
     */
    y: {
        type: Number,
        required: true,
    },
});

const emit = defineEmits<{
    /**
     * The user want to move the element
     */
    (e: "change-pos", index: number): void;

    /**
     * The user want to remove element
     */
    (e: "element-remove", index: number): void;
}>();

// Dimensions
const dimensions = reactive({
    top: "",
    left: "",
    right: "",
    bottom: "",

    width: "",

    maxWidth: "",
    maxHeight: "",
});

/**
 * Computes the context menu dimensions
 */
const computeDimensions = () => {
    const pageWidth = window.innerWidth;
    const pageHeight = window.innerHeight;

    const x = props.x;
    const y = props.y;

    if (y > pageHeight / 2) {
        const bottom = pageHeight - y;

        const maxHeight = pageHeight - bottom;

        dimensions.top = "auto";
        dimensions.bottom = bottom + "px";

        dimensions.maxHeight = maxHeight + "px";
    } else {
        const top = y;

        const maxHeight = pageHeight - top;

        dimensions.top = top + "px";
        dimensions.bottom = "auto";

        dimensions.maxHeight = maxHeight + "px";
    }

    if (x > pageWidth / 2) {
        const right = pageWidth - x;
        const maxWidth = pageWidth - right;

        dimensions.left = "auto";
        dimensions.right = right + "px";

        dimensions.width = "auto";
        dimensions.maxWidth = maxWidth + "px";
    } else {
        const maxWidth = pageWidth - x;

        dimensions.left = x + "px";
        dimensions.right = "auto";

        dimensions.width = "auto";
        dimensions.maxWidth = maxWidth + "px";
    }
};

computeDimensions();
watch(() => props.x, computeDimensions);
watch(() => props.y, computeDimensions);

// Ref to the container element
const container = useTemplateRef("container");

// Focus trap
useFocusTrap(container, shown, hide, "home-page-row-context-btn", true);

/**
 * The user wants to change the position of the element
 */
const changePosition = () => {
    emit("change-pos", props.elementIndex);
    hide();
};

/**
 * The user wants to remove the element
 */
const removeElement = () => {
    emit("element-remove", props.elementIndex);
    hide();
};
</script>
