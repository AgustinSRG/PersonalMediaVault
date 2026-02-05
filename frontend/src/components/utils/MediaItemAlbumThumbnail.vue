<template>
    <div class="album-body-item-thumbnail" :title="item.title || $t('Untitled')">
        <ThumbImage v-if="item.thumbnail" :src="getAssetURL(item.thumbnail)"></ThumbImage>
        <MediaNoThumbnail v-else :type="item.type"></MediaNoThumbnail>

        <DurationIndicator
            v-if="item.type === MEDIA_TYPE_VIDEO || item.type === MEDIA_TYPE_AUDIO"
            :type="item.type"
            :duration="item.duration"
            :small="true"
        ></DurationIndicator>

        <div v-if="displayPosition" class="album-body-item-thumb-pos">
            {{ renderPosition(item.pos) }}
        </div>
    </div>
</template>

<script setup lang="ts">
import { MEDIA_TYPE_AUDIO, MEDIA_TYPE_VIDEO, type MediaListItem } from "@/api/models";
import type { PropType } from "vue";
import MediaNoThumbnail from "./MediaNoThumbnail.vue";
import ThumbImage from "./ThumbImage.vue";
import DurationIndicator from "./DurationIndicator.vue";
import { useI18n } from "@/composables/use-i18n";
import { getAssetURL } from "@/utils/api";

const { $t } = useI18n();

type OptionalPositionedMediaListItem = MediaListItem & {
    /**
     * Position in the list
     */
    pos?: number;
};

defineProps({
    /**
     * The media item
     */
    item: {
        type: Object as PropType<OptionalPositionedMediaListItem>,
        required: true,
    },

    /**
     * True to display the position
     */
    displayPosition: Boolean,
});

/**
 * Renders the position
 * @param p The position
 * @returns The rendered position
 */
const renderPosition = (p: number | undefined): string => {
    if (typeof p !== "number" || p < 0) {
        return "?";
    } else {
        return "" + (p + 1);
    }
};
</script>
