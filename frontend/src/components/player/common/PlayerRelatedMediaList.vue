<template>
    <div
        ref="container"
        class="related-media-list"
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
            v-for="media in relatedMedia || []"
            :key="media.id"
            class="related-media-item"
            tabindex="0"
            :href="getMediaURL(media.id)"
            :title="media.title || $t('Untitled')"
            target="_blank"
            rel="noopener noreferrer"
            @click="clickOnRelatedMedia"
            @keydown="clickOnEnter"
        >
            <MediaItemAlbumThumbnail :item="media"></MediaItemAlbumThumbnail>
            <div class="related-media-item-title">{{ media.title || $t("Untitled") }}</div>
        </a>
    </div>
</template>

<script setup lang="ts">
import type { PropType } from "vue";
import { useTemplateRef } from "vue";
import type { MediaListItem } from "@/api/models";
import { getFrontendUrl } from "@/utils/api";
import MediaItemAlbumThumbnail from "@/components/utils/MediaItemAlbumThumbnail.vue";
import { useFocusTrap } from "@/composables/use-focus-trap";
import { stopPropagationEvent, clickOnEnter } from "@/utils/events";

// Ref to the container element
const container = useTemplateRef("container");

// Shown model
const shown = defineModel<boolean>("shown");

// Props
defineProps({
    /**
     * List of related media
     */
    relatedMedia: {
        type: Array as PropType<MediaListItem[]>,
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
 * Gets the URL of a media element
 * @param mid The media ID
 * @returns The URL
 */
const getMediaURL = (mid: number): string => {
    return getFrontendUrl({
        media: mid,
    });
};

/**
 * The user clicked on a related media link
 * @param event The click event
 */
const clickOnRelatedMedia = (event: Event) => {
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
