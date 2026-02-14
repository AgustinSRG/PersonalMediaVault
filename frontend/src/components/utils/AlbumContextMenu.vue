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
        <div v-if="mediaIndex > 0" tabindex="0" class="album-body-item-options-menu-btn" @click="moveMediaUp" @keydown="clickOnEnter">
            <i class="fas fa-arrow-up"></i> {{ $t("Move up") }}
        </div>
        <div
            v-if="mediaIndex < albumLength - 1"
            tabindex="0"
            class="album-body-item-options-menu-btn"
            @keydown="clickOnEnter"
            @click="moveMediaDown"
        >
            <i class="fas fa-arrow-down"></i> {{ $t("Move down") }}
        </div>
        <div tabindex="0" class="album-body-item-options-menu-btn" @keydown="clickOnEnter" @click="changePosition">
            <i class="fas fa-arrows-up-down-left-right"></i>
            {{ $t("Change position") }}
        </div>
        <div tabindex="0" class="album-body-item-options-menu-btn" @keydown="clickOnEnter" @click="removeMedia">
            <i class="fas fa-trash-alt"></i> {{ $t("Remove from the album") }}
        </div>
    </div>
</template>

<script setup lang="ts">
import { reactive, useTemplateRef, watch } from "vue";
import { stopPropagationEvent, clickOnEnter } from "@/utils/events";
import { useFocusTrap } from "@/composables/use-focus-trap";
import { onDocumentEvent } from "@/composables/on-document-event";

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
     * Index of the media in the album list
     */
    mediaIndex: {
        type: Number,
        required: true,
    },

    /**
     * Length of the album list
     */
    albumLength: {
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
     * The user want to move the media element up in the list
     */
    (e: "move-up", index: number): void;

    /**
     * The user want to move the media element down in the list
     */
    (e: "move-down", index: number): void;

    /**
     * The user want to move the media element to another position
     */
    (e: "change-pos", index: number): void;

    /**
     * The user want to remove the media element from the list
     */
    (e: "media-remove", index: number): void;
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
 * Computes the dimensions of the context menu
 */
const computeDimensions = () => {
    const pageWidth = window.innerWidth;
    const pageHeight = window.innerHeight;

    const x = props.x;
    const y = props.y;

    if (y > pageHeight / 2) {
        const bottom = pageHeight - y;
        const right = pageWidth - x;

        const maxWidth = pageWidth - right;

        const maxHeight = pageHeight - bottom;

        dimensions.top = "auto";
        dimensions.left = "auto";
        dimensions.right = right + "px";
        dimensions.bottom = bottom + "px";
        dimensions.width = "auto";
        dimensions.maxWidth = maxWidth + "px";
        dimensions.maxHeight = maxHeight + "px";
    } else {
        const top = y;
        const right = pageWidth - x;

        const maxWidth = pageWidth - right;

        const maxHeight = pageHeight - top;

        dimensions.top = top + "px";
        dimensions.left = "auto";
        dimensions.right = right + "px";
        dimensions.bottom = "auto";
        dimensions.width = "auto";
        dimensions.maxWidth = maxWidth + "px";
        dimensions.maxHeight = maxHeight + "px";
    }
};

computeDimensions();
watch(() => props.x, computeDimensions);
watch(() => props.y, computeDimensions);

// Ref to the container element
const container = useTemplateRef("container");

// Focus trap
useFocusTrap(container, shown, hide, "album-body-btn", true);

/**
 * The user wants to move the media element up
 */
const moveMediaUp = () => {
    emit("move-up", props.mediaIndex);
    hide();
};

/**
 * The user wants to move the media element down
 */
const moveMediaDown = () => {
    emit("move-down", props.mediaIndex);
    hide();
};

/**
 * The user want to change the position in the album
 */
const changePosition = () => {
    emit("change-pos", props.mediaIndex);
    hide();
};

/**
 * The user want to remove the mdia
 */
const removeMedia = () => {
    emit("media-remove", props.mediaIndex);
    hide();
};
</script>
