<template>
    <div class="search-result-thumb" :title="hintTitle">
        <div class="search-result-thumb-inner">
            <ThumbImage v-if="item.thumbnail" :src="getAssetURL(item.thumbnail)"></ThumbImage>
            <MediaNoThumbnail v-else :type="item.type"></MediaNoThumbnail>

            <VideoItemPreview
                v-if="displayPreview && item.type === MEDIA_TYPE_VIDEO"
                :mid="item.id"
                :hidden="displayPreviewHidden"
            ></VideoItemPreview>

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
import { type MediaListItem } from "@/api/models";
import type { PropType } from "vue";
import { computed, onMounted, ref, watch } from "vue";
import MediaNoThumbnail from "./MediaNoThumbnail.vue";
import ThumbImage from "./ThumbImage.vue";
import DurationIndicator from "./DurationIndicator.vue";
import VideoItemPreview from "./VideoItemPreview.vue";
import { getAssetURL } from "@/utils/api";
import { useTagNames } from "@/composables/use-tags-names";
import { useI18n } from "@/composables/use-i18n";
import { MEDIA_TYPE_VIDEO, MEDIA_TYPE_AUDIO } from "@/constants";
import { useTimeout } from "@/composables/use-timeout";

// Translation
const { $t } = useI18n();

// Props
const props = defineProps({
    /**
     * The media item
     */
    item: {
        type: Object as PropType<MediaListItem>,
        required: true,
    },

    /**
     * True to disable previews
     */
    disablePreviews: Boolean,

    /**
     * True if the element is being hovered with the mouse
     */
    hovering: Boolean,
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

// True to display the preview
const displayPreview = ref(false);

// True if preview is hiding
const displayPreviewHiding = ref(false);

// True if the preview is hidden
const displayPreviewHidden = ref(false);

// Timeout to display the preview
const displayPreviewTimer = useTimeout();

// Delay to display the preview
const DISPLAY_VIDEO_PREVIEW_SHOW_DELAY = 500;

// Delay to start hiding the preview
const DISPLAY_VIDEO_PREVIEW_HIDING_DELAY = 200;

// Delay to hide the preview (wait for animation)
const DISPLAY_VIDEO_PREVIEW_HIDE_DELAY = 300;

/**
 * Updates hover state
 */
const updateHover = () => {
    if (props.disablePreviews) {
        displayPreview.value = false;
        displayPreviewHiding.value = false;
        displayPreviewHidden.value = false;
        displayPreviewTimer.clear();
        return;
    }

    if (props.hovering) {
        if (!displayPreview.value && !displayPreviewTimer.isSet()) {
            displayPreviewHidden.value = false;
            displayPreviewHiding.value = false;
            displayPreviewTimer.set(() => {
                displayPreview.value = true;
            }, DISPLAY_VIDEO_PREVIEW_SHOW_DELAY);
        } else if (displayPreviewHiding.value || displayPreviewHidden.value) {
            displayPreviewHiding.value = false;
            displayPreviewHidden.value = false;
            displayPreviewTimer.clear();
        }
    } else {
        if (displayPreview.value) {
            if (!displayPreviewHiding.value && !displayPreviewHidden.value) {
                displayPreviewHiding.value = true;
                displayPreviewTimer.set(() => {
                    displayPreviewHidden.value = true;
                    displayPreviewTimer.set(() => {
                        displayPreview.value = false;
                    }, DISPLAY_VIDEO_PREVIEW_HIDE_DELAY);
                }, DISPLAY_VIDEO_PREVIEW_HIDING_DELAY);
            }
        } else {
            displayPreviewHiding.value = false;
            displayPreviewHidden.value = false;
            displayPreviewTimer.clear();
        }
    }
};

onMounted(updateHover);
watch([() => props.disablePreviews, () => props.hovering], updateHover);
</script>
