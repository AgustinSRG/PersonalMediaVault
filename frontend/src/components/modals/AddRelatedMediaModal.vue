<template>
    <ModalDialogContainer ref="container" v-model:display="display" :static="true" @scroll.passive="onPageScroll">
        <div
            class="modal-dialog modal-xl modal-height-100"
            role="document"
            :class="{ 'rounded-corners-cells': roundedCorners }"
            :style="{
                '--row-size': rowSize,
                '--row-size-min': rowSizeMin,
                '--min-cell-size': minItemSize + 'px',
                '--max-cell-size': maxItemSize + 'px',
                '--cell-padding': padding + 'px',
            }"
        >
            <div class="modal-header">
                <div class="modal-title">
                    {{ $t("Add related media") }}
                </div>
                <button class="modal-close-btn" :title="$t('Close')" @click="close">
                    <i class="fas fa-times"></i>
                </button>
            </div>
            <div class="modal-body no-padding">
                <PageSearch
                    :in-modal="true"
                    :min="false"
                    :no-album="-1"
                    :page-size="pageSize"
                    :display-titles="displayTitles"
                    :row-size="rowSize"
                    :row-size-min="rowSizeMin"
                    :min-items-size="minItemSize"
                    :max-items-size="maxItemSize"
                    :remove-media-from-list="mediaElements"
                    @select-media="selectMedia"
                ></PageSearch>
            </div>

            <div v-if="pageScroll > 0" class="modal-button-br-container">
                <button type="button" :title="$t('Go to the top')" class="modal-button-br" @click="goTop">
                    <i class="fas fa-angles-up"></i>
                </button>
            </div>
        </div>
    </ModalDialogContainer>
</template>

<script setup lang="ts">
import ModalDialogContainer from "./common/ModalDialogContainer.vue";
import type { PropType } from "vue";
import { computed, ref, useTemplateRef } from "vue";
import PageSearch from "@/components/pages/PageSearch.vue";
import { emitAppEvent, EVENT_NAME_ADVANCED_SEARCH_GO_TOP, EVENT_NAME_ADVANCED_SEARCH_SCROLL } from "@/control/app-events";
import type { MediaListItem } from "@/api/models";
import { useI18n } from "@/composables/use-i18n";
import { useModal } from "@/composables/use-modal";
import { usePagePreferences } from "@/composables/use-page-preferences";

// Translation function
const { $t } = useI18n();

// Display model
const display = defineModel<boolean>("display");

// Modal container
const container = useTemplateRef("container");

// Modal composable
const { close, scrollToTop } = useModal(display, container);

// Props
const props = defineProps({
    /**
     * Media ID
     */
    mid: {
        type: Number,
        required: true,
    },

    /**
     * Related media array
     */
    relatedMedia: {
        type: Array as PropType<MediaListItem[]>,
        required: true,
    },
});

// Events
const emit = defineEmits<{
    /**
     * Emitted when media is added
     * @param media The media element
     * @param callback The callback to call in order to indicate the media was added
     */
    (e: "add-media", media: MediaListItem, callback: () => void): void;
}>();

/**
 * Call when the user selects a media element
 * @param m Media element
 * @param callback Callback
 */
const selectMedia = (m: MediaListItem, callback: () => void) => {
    emit("add-media", m, callback);
};

// Set of media elements not to show
const mediaElements = computed(() => new Set<number>((props.relatedMedia || []).map((e) => e.id).concat(props.mid)));

// Page preferences
const { pageSize, rowSize, rowSizeMin, minItemSize, maxItemSize, padding, displayTitles, roundedCorners } = usePagePreferences("search");

// Page scroll
const pageScroll = ref(0);

/**
 * Event handler for 'scroll'
 * @param e The event
 */
const onPageScroll = (e: Event) => {
    pageScroll.value = (e.target as HTMLElement).scrollTop;

    emitAppEvent(EVENT_NAME_ADVANCED_SEARCH_SCROLL, e);
};

/**
 * Scrolls to the top
 */
const goTop = () => {
    emitAppEvent(EVENT_NAME_ADVANCED_SEARCH_GO_TOP);

    scrollToTop();
};
</script>
