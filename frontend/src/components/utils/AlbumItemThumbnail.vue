<template>
    <div class="search-result-thumb" :title="title">
        <div class="search-result-thumb-inner">
            <ThumbImage v-if="item.thumbnail" :src="getAssetURL(item.thumbnail)"></ThumbImage>
            <div v-else class="no-thumb">
                <i class="fas fa-list-ol"></i>
            </div>

            <div class="thumb-bottom-right-tag" :title="sizeTitle"><i class="fas fa-list-ol"></i> {{ sizeText }}</div>

            <slot></slot>
        </div>
    </div>
</template>

<script setup lang="ts">
import { useI18n } from "@/composables/use-i18n";
import ThumbImage from "./ThumbImage.vue";
import type { PropType } from "vue";
import { computed } from "vue";
import type { AlbumListItem } from "@/api/models";
import { renderDate } from "@/utils/time";
import { getAssetURL } from "@/utils/api";

// Translation
const { $t, locale } = useI18n();

const props = defineProps({
    /**
     * The album item
     */
    item: {
        type: Object as PropType<AlbumListItem>,
        required: true,
    },
});

/**
 * The title for the item
 */
const title = computed(() => {
    return (
        (props.item.name || $t("Untitled album")) +
        (props.item.lm ? "\n" + $t("Last modified") + ": " + renderDate(props.item.lm, locale.value) : "")
    );
});

/**
 * Text to display the size of the album
 */
const sizeText = computed(() => {
    if (props.item.size > 0) {
        return props.item.size + " " + (props.item.size === 1 ? $t("item") : $t("items"));
    } else {
        return $t("Empty");
    }
});

/**
 * Title for the size tag
 */
const sizeTitle = computed(() => {
    return $t("Album") + " - " + sizeText.value;
});
</script>
