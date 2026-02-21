<template>
    <div
        ref="container"
        class="player-context-menu"
        :class="{
            hidden: !shown,
            closing: closing,
        }"
        :style="{
            top: top,
            left: left,
            right: right,
            bottom: bottom,
            width: width,
            'max-width': maxWidth,
            'max-height': maxHeight,
        }"
        tabindex="-1"
        @mousedown="stopPropagationEvent"
        @touchstart="stopPropagationEvent"
        @click="stopPropagationEvent"
        @dblclick="stopPropagationEvent"
        @contextmenu="stopPropagationEvent"
        @animationend="onAnimationEnd"
    >
        <table class="player-context-menu-table">
            <tbody>
                <tr
                    v-if="(type === 'video' || type === 'audio') && !isShort"
                    class="tr-button"
                    tabindex="0"
                    @click="toggleLoop"
                    @keydown="clickOnEnter"
                >
                    <td>
                        <i class="fas fa-repeat icon-config"></i>
                        <span class="context-entry-title">{{ $t("Loop") }}</span>
                    </td>
                    <td class="td-right">
                        <i class="fas fa-check" :class="{ 'check-uncheck': !loop }"></i>
                    </td>
                </tr>

                <tr
                    v-if="(type === 'video' || type === 'audio') && hasSlices"
                    class="tr-button"
                    tabindex="0"
                    @click="toggleSliceLoop"
                    @keydown="clickOnEnter"
                >
                    <td>
                        <i class="fas fa-repeat icon-config"></i>
                        <span class="context-entry-title">{{ $t("Time slice loop") }}</span>
                    </td>
                    <td class="td-right">
                        <i class="fas fa-check" :class="{ 'check-uncheck': !sliceLoop }"></i>
                    </td>
                </tr>

                <tr v-if="type === 'image'" class="tr-button" tabindex="0" @click="toggleFit" @keydown="clickOnEnter">
                    <td>
                        <i class="fas fa-magnifying-glass icon-config"></i>
                        <span class="context-entry-title">{{ $t("Fit image") }}</span>
                    </td>
                    <td class="td-right">
                        <i class="fas fa-check" :class="{ 'check-uncheck': !fit }"></i>
                    </td>
                </tr>

                <tr
                    v-if="type === 'image' || type === 'video'"
                    class="tr-button"
                    tabindex="0"
                    @keydown="clickOnEnter"
                    @click="toggleControls"
                >
                    <td>
                        <i class="fas fa-eye-slash icon-config"></i>
                        <span class="context-entry-title">{{ $t("Hide controls") }}</span>
                    </td>
                    <td class="td-right">
                        <i class="fas fa-check" :class="{ 'check-uncheck': controls }"></i>
                    </td>
                </tr>

                <tr class="tr-button" tabindex="0" @click="showTags" @keydown="clickOnEnter">
                    <td>
                        <i class="fas fa-tag icon-config"></i>
                        <span class="context-entry-title">{{ $t("Tags") }}</span>
                    </td>
                    <td class="td-right"></td>
                </tr>

                <tr v-if="type === 'image' && canWrite" class="tr-button" tabindex="0" @click="toggleNotes" @keydown="clickOnEnter">
                    <td>
                        <i class="fas fa-pencil-alt icon-config"></i>
                        <span class="context-entry-title">{{ $t("Edit image notes") }}</span>
                    </td>
                    <td class="td-right">
                        <i class="fas fa-check" :class="{ 'check-uncheck': !notesEdit }"></i>
                    </td>
                </tr>

                <tr
                    v-if="(type === 'video' || type === 'audio') && canWrite"
                    class="tr-button"
                    tabindex="0"
                    @click="toggleTimeSlices"
                    @keydown="clickOnEnter"
                >
                    <td>
                        <i class="fas fa-pencil-alt icon-config"></i>
                        <span class="context-entry-title">{{ $t("Edit time slices") }}</span>
                    </td>
                    <td class="td-right">
                        <i class="fas fa-check" :class="{ 'check-uncheck': !timeSlicesEdit }"></i>
                    </td>
                </tr>

                <tr v-if="hasDescription || canWrite" class="tr-button" tabindex="0" @click="showDescription" @keydown="clickOnEnter">
                    <td>
                        <i class="fas fa-file-lines icon-config"></i>
                        <span class="context-entry-title">{{ $t("Description") }}</span>
                    </td>
                    <td class="td-right"></td>
                </tr>

                <tr v-if="url" class="tr-button" tabindex="0" @keydown="clickOnEnter" @click="download">
                    <td>
                        <i class="fas fa-download icon-config"></i>
                        <span class="context-entry-title">{{ $t("Download") }}</span>
                    </td>
                    <td class="td-right"></td>
                </tr>

                <tr v-if="url" class="tr-button" tabindex="0" @keydown="clickOnEnter" @click="openStats">
                    <td>
                        <i class="fas fa-bars-progress icon-config"></i>
                        <span class="context-entry-title">{{ $t("Size Statistics") }}</span>
                    </td>
                    <td class="td-right"></td>
                </tr>

                <tr v-if="url" class="tr-button" tabindex="0" @keydown="clickOnEnter" @click="refreshMedia">
                    <td>
                        <i class="fas fa-sync-alt icon-config"></i>
                        <span class="context-entry-title">{{ $t("Refresh") }}</span>
                    </td>
                    <td class="td-right"></td>
                </tr>
            </tbody>
        </table>
    </div>
</template>

<script setup lang="ts">
import { MediaController } from "@/control/media";
import type { PropType } from "vue";
import { onMounted, ref, useTemplateRef, watch } from "vue";
import { useI18n } from "@/composables/use-i18n";
import { useUserPermissions } from "@/composables/use-user-permissions";
import { useFocusTrap } from "@/composables/use-focus-trap";
import { clickOnEnter, stopPropagationEvent } from "@/utils/events";

// Ref to the container element
const container = useTemplateRef("container");

// Translation
const { $t } = useI18n();

// User permissions
const { canWrite } = useUserPermissions();

// Shown model
const shown = defineModel<boolean>("shown");

// Types of player
type PlayerType = "video" | "audio" | "image";

// Props
const props = defineProps({
    /**
     * Type of player
     */
    type: {
        type: String as PropType<PlayerType>,
        required: true,
    },

    /**
     * X coordinate
     */
    x: {
        type: Number,
        required: true,
    },

    y: {
        type: Number,
        required: true,
    },

    /**
     * Download URL
     */
    url: String,

    /**
     * Media title
     */
    title: String,

    /**
     * Has slices?
     */
    hasSlices: Boolean,

    /**
     * Is short duration?
     */
    isShort: Boolean,

    /**
     * Has description?
     */
    hasDescription: Boolean,
});

// Loop
const loop = defineModel<boolean>("loop");

// Fit image
const fit = defineModel<boolean>("fit");

// Show controls
const controls = defineModel<boolean>("controls");

// Slice loop
const sliceLoop = defineModel<boolean>("sliceLoop");

// Image notes edit mode?
const notesEdit = defineModel<boolean>("notesEdit");

// Time slices edit mode?
const timeSlicesEdit = defineModel<boolean>("timeSlicesEdit");

// Emits
const emit = defineEmits<{
    /**
     * Emitted when closed
     */
    (e: "close"): void;

    /**
     * Opens the tags widget
     */
    (e: "open-tags"): void;

    /**
     * Opens the description widget
     */
    (e: "open-desc"): void;

    /**
     * Opens the stats modal
     */
    (e: "stats"): void;
}>();

// True if the modal is closing
const closing = ref(false);

/**
 * Show the menu
 */
const show = () => {
    const shouldFocus = shown.value;

    shown.value = true;
    closing.value = false;

    if (shouldFocus) {
        container.value?.focus();
    }
};

/**
 * Hides the menu
 */
const hide = () => {
    emit("close");
    closing.value = true;
};

// Reset 'closing' when shown changes to true
watch(shown, () => {
    if (shown.value) {
        closing.value = false;
    }
});

// Coordinates
const top = ref("");
const left = ref("");
const right = ref("");
const bottom = ref("");
const width = ref("");
const maxWidth = ref("");
const maxHeight = ref("");

/**
 * Computes the dimensions of the context menu
 */
const computeDimensions = () => {
    if (!container.value) {
        return;
    }

    const pageWidth = window.innerWidth;
    const pageHeight = window.innerHeight || 1;

    const elementBounds = container.value.getBoundingClientRect();

    const x = props.x;
    const y = props.y;

    if (y + elementBounds.height > pageHeight) {
        bottom.value = "0";
        top.value = "auto";
    } else {
        top.value = y + "px";
        bottom.value = "auto";
    }

    const mh = pageHeight;

    maxHeight.value = mh + "px";

    const l = x;
    const mw = pageWidth - l;

    left.value = l + "px";
    right.value = "auto";
    width.value = "auto";
    maxWidth.value = mw + "px";
};

onMounted(() => {
    if (shown.value) {
        computeDimensions();
    }
});

watch([() => props.x, () => props.y, shown], computeDimensions);

/**
 * Toggles loop value
 */
const toggleLoop = () => {
    loop.value = !loop.value;
    hide();
};

/**
 * Toggles slice loop value
 */
const toggleSliceLoop = () => {
    sliceLoop.value = !sliceLoop.value;
    hide();
};

/**
 * Toggles the image fit value
 */
const toggleFit = () => {
    fit.value = !fit.value;
    hide();
};

/**
 * Toggles the notes edit mode
 */
const toggleNotes = () => {
    notesEdit.value = !notesEdit.value;
    hide();
};

/**
 * Toggles the time slices edit mode
 */
const toggleTimeSlices = () => {
    timeSlicesEdit.value = !timeSlicesEdit.value;
    hide();
};

/**
 * Toggles controls visibility
 */
const toggleControls = () => {
    controls.value = !controls.value;
    hide();
};

/**
 * Shows the tags widget
 */
const showTags = () => {
    emit("open-tags");
    hide();
};

/**
 * Shows the description widget
 */
const showDescription = () => {
    emit("open-desc");
    hide();
};

/**
 * Opens the size stats modal
 */
const openStats = () => {
    emit("stats");
    hide();
};

/**
 * Refreshes the media
 */
const refreshMedia = () => {
    MediaController.Load();
    hide();
};

/**
 * Downloads the media
 */
const download = () => {
    const link = document.createElement("a");
    link.target = "_blank";
    link.rel = "noopener noreferrer";

    const titlePart = props.title ? "&filename=" + encodeURIComponent(props.title) : "";

    if ((props.url + "").includes("?")) {
        link.href = props.url + "&download=force" + titlePart;
    } else {
        link.href = props.url + "?download=force" + titlePart;
    }

    link.click();

    hide();
};

/**
 * Called when the closing animation ends
 * @param e The animation event
 */
const onAnimationEnd = (e: AnimationEvent) => {
    e.stopPropagation();
    if (closing.value && e.animationName === "player-context-menu-close-animation") {
        closing.value = false;
        shown.value = false;
    }
};

// Focus trap
useFocusTrap(
    container,
    shown,
    () => {
        hide();
    },
    null,
    true,
);

// Exposed methods
defineExpose({
    hide,
    show,
});
</script>
