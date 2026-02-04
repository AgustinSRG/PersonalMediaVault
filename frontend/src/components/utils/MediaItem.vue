<template>
    <div class="search-result-item" :class="{ current: current, 'last-current': lastCurrent }">
        <a class="clickable" :href="mediaUrl" target="_blank" rel="noopener noreferrer" @click="onClick">
            <MediaItemThumbnail :item="item"></MediaItemThumbnail>
            <div v-if="displayTitles" class="search-result-title">
                {{ item.title || $t("Untitled") }}
            </div>
        </a>
    </div>
</template>

<script setup lang="ts">
import type { MediaListItem } from "@/api/models";
import { useI18n } from "@/composables/use-i18n";
import { getFrontendUrl } from "@/utils/api";
import type { PropType } from "vue";
import { computed } from "vue";
import MediaItemThumbnail from "./MediaItemThumbnail.vue";

const { $t } = useI18n();

const emit = defineEmits<{
    /**
     * The user clicked in the element in order to go
     * to the media or select it
     */
    (e: "go"): void;
}>();

const props = defineProps({
    /**
     * The media item
     */
    item: {
        type: Object as PropType<MediaListItem>,
        required: true,
    },

    /**
     * True if the item is selected
     */
    current: {
        type: Boolean,
        default: false,
    },

    /**
     * True if the media is the last selected one
     */
    lastCurrent: {
        type: Boolean,
        default: false,
    },

    /**
     * True to display titles
     */
    displayTitles: {
        type: Boolean,
        default: false,
    },
});

/**
 * Handler for 'click' event
 * @param e The pointer event
 */
const onClick = (e: PointerEvent) => {
    e.preventDefault();

    emit("go");
};

/**
 * Media URL
 */
const mediaUrl = computed(() => getFrontendUrl({ media: props.item?.id }));
</script>
