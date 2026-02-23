<template>
    <div
        ref="container"
        class="player-attachments-list"
        :class="{ hidden: !shown }"
        tabindex="-1"
        role="dialog"
        @click="stopPropagationEvent"
        @dblclick="stopPropagationEvent"
        @mousedown="stopPropagationEvent"
        @touchstart="stopPropagationEvent"
        @contextmenu="stopPropagationEvent"
        @mouseenter="enter"
        @mouseleave="leave"
        @keydown="keyDownHandle"
    >
        <a
            v-for="att in attachments || []"
            :key="att.id"
            class="player-attachment-link"
            tabindex="0"
            :href="getAssetURL(att.url)"
            target="_blank"
            rel="noopener noreferrer"
            @click="clickAttachmentLink"
            @keydown="clickOnEnter"
        >
            <div class="attachment-icon-link">
                <i class="fas fa-paperclip"></i>
            </div>
            <div class="attachment-name">
                {{ att.name }}
            </div>
        </a>
    </div>
</template>

<script setup lang="ts">
import type { PropType } from "vue";
import { useTemplateRef } from "vue";
import type { MediaAttachment } from "@/api/models";
import { getAssetURL } from "@/utils/api";
import { stopPropagationEvent, clickOnEnter } from "@/utils/events";
import { useFocusTrap } from "@/composables/use-focus-trap";

// Ref to the container element
const container = useTemplateRef("container");

// Shown model
const shown = defineModel<boolean>("shown");

// Props
defineProps({
    /**
     * List of attachments
     */
    attachments: {
        type: Array as PropType<MediaAttachment[]>,
        required: true,
    },
});

// Emits
const emit = defineEmits<{
    /**
     * The user enters the control
     */
    (e: "enter"): void;

    /**
     * The user leaves the control
     */
    (e: "leave"): void;
}>();

/**
 * Closes the control
 */
const close = () => {
    shown.value = false;
};

/**
 * The user enters the config
 */
const enter = () => {
    emit("enter");
};

/**
 * The user leaves the config
 */
const leave = () => {
    emit("leave");
};

/**
 * The user clicked on a link
 * @param event The click event
 */
const clickAttachmentLink = (event: Event) => {
    event.stopPropagation();
    close();
};

/**
 * Event handler for 'keydown'
 * @param e The keyboard event
 */
const keyDownHandle = (e: KeyboardEvent) => {
    if (e.ctrlKey) {
        return;
    }
    if (e.key === "Escape") {
        close();
        e.stopPropagation();
    }
};

// Focus trap
useFocusTrap(container, shown, close, "player-settings-no-trap", true);
</script>
