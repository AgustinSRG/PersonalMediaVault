<template>
    <ModalDialogContainer ref="container" v-model:display="display" :static="true" @scroll.passive="onPageScroll">
        <div
            class="modal-dialog modal-xl modal-height-100"
            role="document"
            :class="{ 'rounded-corners-cells': roundedCorners }"
            :style="{
                '--row-size': rowSize,
                '--row-size-min': rowSize,
                '--min-cell-size': minItemSize + 'px',
                '--max-cell-size': maxItemSize + 'px',
                '--cell-padding': padding + 'px',
            }"
        >
            <div class="modal-header">
                <div class="modal-title">
                    {{ $t("Select an image to find similar results") }}
                </div>
                <button type="button" class="modal-close-btn" :title="$t('Close')" @click="close">
                    <i class="fas fa-times"></i>
                </button>
            </div>
            <div class="modal-body no-padding">
                <div class="horizontal-filter-menu two-child modal-top-menu">
                    <a href="javascript:;" class="horizontal-filter-menu-item selected" :title="$t('Search')"
                        ><i class="fas fa-search"></i> {{ $t("Search") }}</a
                    >
                    <a href="javascript:;" class="horizontal-filter-menu-item" :title="$t('Select file')" @click="openFileSelectDialog"
                        ><i class="fas fa-upload"></i> {{ $t("Select file") }}</a
                    >
                </div>
                <input ref="hiddenFileInput" type="file" class="file-hidden" name="image-select" @change="onImageFileChanged" />
                <PageSearch
                    :in-modal="true"
                    :no-search-by-image="true"
                    :skip-media-with-no-thumbnail="true"
                    :min="false"
                    :page-size="pageSize"
                    :display-titles="displayTitles"
                    :row-size="rowSize"
                    :row-size-min="rowSizeMin"
                    :min-items-size="minItemSize"
                    :max-items-size="maxItemSize"
                    :scroll-key="scrollKey"
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
import { ref, useTemplateRef } from "vue";
import PageSearch from "@/components/pages/PageSearch.vue";
import { emitAppEvent, EVENT_NAME_ADVANCED_SEARCH_GO_TOP, EVENT_NAME_ADVANCED_SEARCH_SCROLL } from "@/global-state/app-events";
import type { MediaListItem } from "@/api/models";
import { useI18n } from "@/composables/use-i18n";
import { useModal } from "@/composables/use-modal";
import { usePagePreferences } from "@/composables/use-page-preferences";
import { getUniqueStringId } from "@/utils/unique-id.ts";

const emit = defineEmits<{
    /**
     * Event to select a media
     */
    (e: "select-media", mid: number, thumbnail: string): void;

    /**
     * Event to select a file
     */
    (e: "select-file", file: File): void;
}>();

// Translation function
const { $t } = useI18n();

// Display model
const display = defineModel<boolean>("display");

// Modal container
const container = useTemplateRef("container");

// Modal composable
const { close, scrollToTop } = useModal(display, container);

/**
 * Called when the user selects a media element
 * @param m The media element
 * @param callback The callback
 */
const selectMedia = (m: MediaListItem, callback: () => void) => {
    emit("select-media", m.id, m.thumbnail || "");

    callback();

    close();
};

// Page preferences
const { pageSize, rowSize, rowSizeMin, minItemSize, maxItemSize, padding, displayTitles, roundedCorners } = usePagePreferences("search");

// Page scroll
const pageScroll = ref(0);

// Scroll key
const scrollKey = getUniqueStringId();

/**
 * Event handler for 'scroll'
 * @param e The event
 */
const onPageScroll = (e: Event) => {
    e.stopPropagation();

    pageScroll.value = (e.target as HTMLElement).scrollTop;

    emitAppEvent(EVENT_NAME_ADVANCED_SEARCH_SCROLL, e, scrollKey);
};

/**
 * Scrolls to the top
 */
const goTop = () => {
    emitAppEvent(EVENT_NAME_ADVANCED_SEARCH_GO_TOP, scrollKey);

    scrollToTop();
};

// Hidden file input element
const hiddenFileInput = useTemplateRef("hiddenFileInput");

/**
 * Opens the dialog to select an image file
 */
const openFileSelectDialog = () => {
    if (hiddenFileInput.value) {
        hiddenFileInput.value.value = null;
        hiddenFileInput.value.click();
    }
};

/**
 * Event handler for 'change' on the file input
 * @param e The event
 */
const onImageFileChanged = (e: InputEvent) => {
    const data = (e.target as HTMLInputElement).files;
    if (data && data.length > 0) {
        emit("select-file", data[0]);
        close();
    }
};
</script>
