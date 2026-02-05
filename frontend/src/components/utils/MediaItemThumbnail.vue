<template>
    <div class="search-result-thumb" :title="hintTitle">
        <div class="search-result-thumb-inner">
            <ThumbImage v-if="item.thumbnail" :src="getAssetURL(item.thumbnail)"></ThumbImage>
            <MediaNoThumbnail v-else :type="item.type"></MediaNoThumbnail>

            <DurationIndicator
                v-if="item.type === MEDIA_TYPE_AUDIO || item.type === MEDIA_TYPE_VIDEO"
                :type="item.type"
                :duration="item.duration"
            ></DurationIndicator>

            <slot></slot>
        </div>
    </div>
</template>

<script setup lang="ts">
import { MEDIA_TYPE_AUDIO, MEDIA_TYPE_VIDEO, type MediaListItem } from "@/api/models";
import type { PropType } from "vue";
import { computed } from "vue";
import MediaNoThumbnail from "./MediaNoThumbnail.vue";
import ThumbImage from "./ThumbImage.vue";
import DurationIndicator from "./DurationIndicator.vue";
import { getAssetURL } from "@/utils/api";
import { useTagNames } from "@/composables/use-tags-names";
import { useI18n } from "@/composables/use-i18n";

const { $t } = useI18n();

const props = defineProps({
    /**
     * The media item
     */
    item: {
        type: Object as PropType<MediaListItem>,
        required: true,
    },
});

// Tags
const { getTagName } = useTagNames();

// The hint title for the element
const hintTitle = computed(() => {
    const parts = [props.item.title || $t("Untitled")];

    if (props.item.tags.length > 0) {
        const tagNames = [];

        for (const tag of props.item.tags) {
            tagNames.push(getTagName(tag));
        }

        parts.push($t("Tags") + ": " + tagNames.join(", "));
    }

    return parts.join("\n");
});
</script>
