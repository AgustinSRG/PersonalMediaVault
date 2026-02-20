<template>
    <div class="player-media-change-preview">
        <MediaItemAlbumThumbnail :item="media"></MediaItemAlbumThumbnail>
        <div class="player-media-change-preview-details">
            <div class="player-media-change-preview-title">
                {{ next ? $t("Next") : $t("Previous") }}:<br />
                {{ media.title || $t("Untitled") }}
            </div>
        </div>
    </div>
</template>

<script setup lang="ts">
import type { PropType } from "vue";
import { type MediaListItem } from "@/api/models";
import MediaItemAlbumThumbnail from "@/components/utils/MediaItemAlbumThumbnail.vue";
import { useI18n } from "@/composables/use-i18n";

// Translation function
const { $t } = useI18n();

defineProps({
    /**
     * The media item
     */
    media: {
        type: Object as PropType<MediaListItem>,
        required: true,
    },

    /**
     * True if it is the next element,
     * false if it is the previous element
     */
    next: Boolean,
});
</script>
