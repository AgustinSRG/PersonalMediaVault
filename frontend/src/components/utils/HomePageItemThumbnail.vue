<template>
    <AlbumItemThumbnail v-if="item.album" :item="item.album">
        <div v-if="editing" class="home-page-element-thumb-pos">
            {{ position + 1 }}
        </div>
        <button
            v-if="editing"
            type="button"
            class="home-page-row-context-btn"
            :title="$t('Options to modify element')"
            @click="showContextMenu"
        >
            <i class="fas fa-bars"></i>
        </button>
    </AlbumItemThumbnail>
    <MediaItemThumbnail v-else-if="item.media" :item="item.media">
        <div v-if="editing" class="home-page-element-thumb-pos">
            {{ position + 1 }}
        </div>
        <button
            v-if="editing"
            type="button"
            class="home-page-row-context-btn"
            :title="$t('Options to modify element')"
            @click="showContextMenu"
        >
            <i class="fas fa-bars"></i>
        </button>
    </MediaItemThumbnail>
    <div v-else class="search-result-thumb">
        <div class="search-result-thumb-inner">
            <div class="no-thumb">
                <i class="fas fa-ban"></i>
            </div>
        </div>
    </div>
</template>
<script setup lang="ts">
import type { HomePageElement } from "@/api/api-home";
import type { PropType } from "vue";
import AlbumItemThumbnail from "./AlbumItemThumbnail.vue";
import MediaItemThumbnail from "./MediaItemThumbnail.vue";

const emit = defineEmits<{
    /**
     * Event triggered when the user clicks the menu button
     */
    (e: "open-options", position: number, event: MouseEvent): void;
}>();

const props = defineProps({
    /**
     * The album item
     */
    item: {
        type: Object as PropType<HomePageElement>,
        required: true,
    },

    /**
     * True if editing
     */
    editing: Boolean,

    /**
     * Position in the home page row
     */
    position: Number,
});

/**
 * Handler for 'click' event on the menu button
 * @param e The mouse event
 */
const showContextMenu = (e: MouseEvent) => {
    emit("open-options", props.position, e);
};
</script>
