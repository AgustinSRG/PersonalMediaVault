<template>
    <ModalDialogContainer ref="container" v-model:display="display" :lock-close="busy" :static="true" @scroll.passive="onPageScroll">
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
                <div v-if="!isUpload" class="modal-title">
                    {{ $t("Search media to add to the album") }}
                </div>
                <div v-if="isUpload" class="modal-title">
                    {{ $t("Upload media to add to the album") }}
                </div>
                <button class="modal-close-btn" :title="$t('Close')" @click="close">
                    <i class="fas fa-times"></i>
                </button>
            </div>
            <div class="modal-body no-padding">
                <div class="horizontal-filter-menu two-child modal-top-menu">
                    <a
                        href="javascript:;"
                        class="horizontal-filter-menu-item"
                        :title="$t('Upload')"
                        :class="{ selected: isUpload }"
                        @click="changeToUpload"
                        ><i class="fas fa-upload"></i> {{ $t("Upload") }}</a
                    >
                    <a
                        href="javascript:;"
                        class="horizontal-filter-menu-item"
                        :title="$t('Search')"
                        :class="{ selected: !isUpload }"
                        @click="changeToSearch"
                        ><i class="fas fa-search"></i> {{ $t("Search") }}</a
                    >
                </div>
                <PageSearch
                    v-if="!isUpload"
                    :display="true"
                    :in-modal="true"
                    :min="false"
                    :no-album="aid"
                    :page-size="pageSize"
                    :display-titles="displayTitles"
                    :row-size="rowSize"
                    :row-size-min="rowSizeMin"
                    :min-items-size="minItemSize"
                    :max-items-size="maxItemSize"
                    @select-media="selectMedia"
                ></PageSearch>
                <PageUpload v-if="isUpload" :display="true" :in-modal="true" :fixed-album="aid" @media-go="close"></PageUpload>
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
import { ref, useTemplateRef } from "vue";
import PageSearch from "@/components/pages/PageSearch.vue";
import PageUpload from "@/components/pages/PageUpload.vue";
import { makeApiRequest } from "@asanrom/request-browser";
import {
    emitAppEvent,
    EVENT_NAME_ADVANCED_SEARCH_GO_TOP,
    EVENT_NAME_ADVANCED_SEARCH_SCROLL,
    EVENT_NAME_UNAUTHORIZED,
} from "@/control/app-events";
import { AlbumsController } from "@/control/albums";
import { AuthController } from "@/control/auth";
import { PagesController } from "@/control/pages";
import { apiAlbumsAddMediaToAlbum } from "@/api/api-albums";
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
     * Album ID
     */
    aid: {
        type: Number,
        required: true,
    },
});

// Upload or search?
const isUpload = ref(true);

/**
 * Changes to upload mode
 */
const changeToUpload = () => {
    isUpload.value = true;
};

/**
 * Changes to search mode
 */
const changeToSearch = () => {
    isUpload.value = false;
};

// Busy status
const busy = ref(false);

/**
 * Called when the user selects a media element
 * @param m The media element
 * @param callback The callback
 */
const selectMedia = (m: MediaListItem, callback: () => void) => {
    if (busy.value) {
        return;
    }

    const albumId = props.aid;

    busy.value = true;

    makeApiRequest(apiAlbumsAddMediaToAlbum(albumId, m.id))
        .onSuccess(() => {
            busy.value = false;

            PagesController.ShowSnackBar($t("Successfully added to album"));

            AlbumsController.OnChangedAlbum(albumId, true);

            callback();
        })
        .onRequestError((err, handleErr) => {
            busy.value = false;

            handleErr(err, {
                unauthorized: () => {
                    emitAppEvent(EVENT_NAME_UNAUTHORIZED);
                },
                maxSizeReached: () => {
                    PagesController.ShowSnackBar(
                        $t("Error") + ": " + $t("The album reached the limit of 1024 elements. Please, consider creating another album."),
                    );
                },
                badRequest: () => {
                    PagesController.ShowSnackBar($t("Error") + ": " + $t("Bad request"));
                },
                accessDenied: () => {
                    PagesController.ShowSnackBar($t("Error") + ": " + $t("Access denied"));
                    AuthController.CheckAuthStatusSilent();
                },
                notFound: () => {
                    PagesController.ShowSnackBar($t("Error") + ": " + $t("Not found"));
                    AlbumsController.Load();
                },
                serverError: () => {
                    PagesController.ShowSnackBar($t("Error") + ": " + $t("Internal server error"));
                },
                networkError: () => {
                    PagesController.ShowSnackBar($t("Error") + ": " + $t("Could not connect to the server"));
                },
            });
        })
        .onUnexpectedError((err) => {
            busy.value = false;
            console.error(err);
        });
};

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

    if (!isUpload.value) {
        emitAppEvent(EVENT_NAME_ADVANCED_SEARCH_SCROLL, e);
    }
};

/**
 * Scrolls to the top
 */
const goTop = () => {
    if (!isUpload.value) {
        emitAppEvent(EVENT_NAME_ADVANCED_SEARCH_GO_TOP);
    }

    scrollToTop();
};
</script>
