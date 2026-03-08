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
                <div v-if="isAlbums" class="modal-title">
                    {{ $t("Add albums") }}
                </div>
                <div v-else class="modal-title">
                    {{ $t("Add media elements") }}
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
                        :title="$t('Albums')"
                        :class="{ selected: isAlbums }"
                        @click="changeToAlbums"
                        ><i class="fas fa-list"></i> {{ $t("Albums") }}</a
                    >
                    <a
                        href="javascript:;"
                        class="horizontal-filter-menu-item"
                        :title="$t('Media')"
                        :class="{ selected: !isAlbums }"
                        @click="changeToMedia"
                        ><i class="fas fa-photo-film"></i> {{ $t("Media") }}</a
                    >
                </div>
                <PageSearch
                    v-if="!isAlbums"
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
                <PageAlbums
                    v-else
                    :in-modal="true"
                    :min="false"
                    :page-size="pageSize"
                    :display-titles="displayTitles"
                    :row-size="rowSize"
                    :row-size-min="rowSizeMin"
                    :min-items-size="minItemSize"
                    :max-items-size="maxItemSize"
                    :remove-albums-from-list="albumElements"
                    @select-album="selectAlbum"
                ></PageAlbums>
            </div>

            <div v-if="pageScroll > 0" class="modal-button-br-container">
                <button type="button" :title="$t('Go to the top')" class="modal-button-br" @click="goTop">
                    <i class="fas fa-angles-up"></i>
                </button>
            </div>
        </div>

        <ErrorMessageModal v-if="errorDisplay" v-model:display="errorDisplay" :message="error"></ErrorMessageModal>
    </ModalDialogContainer>
</template>

<script setup lang="ts">
import ModalDialogContainer from "./common/ModalDialogContainer.vue";
import type { PropType } from "vue";
import { computed, defineAsyncComponent, ref, useTemplateRef, watch } from "vue";
import PageSearch from "@/components/pages/PageSearch.vue";
import PageAlbums from "@/components/pages/PageAlbums.vue";
import { makeApiRequest } from "@asanrom/request-browser";
import { emitAppEvent, EVENT_NAME_ADVANCED_SEARCH_GO_TOP, EVENT_NAME_ADVANCED_SEARCH_SCROLL } from "@/control/app-events";
import type { HomePageElement } from "@/api/api-home";
import { apiHomeGroupAddElement, HOME_PAGE_ELEMENT_TYPE_ALBUM, HOME_PAGE_ELEMENT_TYPE_MEDIA } from "@/api/api-home";
import type { MediaListItem } from "@/api/models";
import { useI18n } from "@/composables/use-i18n";
import { useModal } from "@/composables/use-modal";
import { usePagePreferences } from "@/composables/use-page-preferences";
import { useCommonRequestErrors } from "@/composables/use-common-request-errors";
import { showSnackBar } from "@/control/snack-bar";

const ErrorMessageModal = defineAsyncComponent({
    loader: () => import("@/components/modals/ErrorMessageModal.vue"),
});

// Translation function
const { $t } = useI18n();

// Display model
const display = defineModel<boolean>("display");

// Modal container
const container = useTemplateRef("container");

// Modal composable
const { close, forceClose, scrollToTop } = useModal(display, container);

// Props
const props = defineProps({
    /**
     * Group ID
     */
    groupId: {
        type: Number,
        required: true,
    },

    /**
     * List of elements in the group
     */
    groupElements: {
        type: Array as PropType<HomePageElement[]>,
        required: true,
    },
});

// Events
const emit = defineEmits<{
    /**
     * Emitted when an element is added
     * to reload the row
     */
    (e: "added-element"): void;

    /**
     * Emitted to indicate the home page should be reloaded
     */
    (e: "must-reload"): void;
}>();

// Albums or media?
const isAlbums = ref(true);

/**
 * Changes to albums mode
 */
const changeToAlbums = () => {
    isAlbums.value = true;
};

const changeToMedia = () => {
    isAlbums.value = false;
};

// Busy status
const busy = ref(false);

// Key for page preferences
const pagePreferencesKey = computed(() => (isAlbums.value ? "albums" : "search"));

// Page preferences
const { pageSize, rowSize, rowSizeMin, minItemSize, maxItemSize, padding, displayTitles, roundedCorners } =
    usePagePreferences(pagePreferencesKey);

// Album elements in the group
const albumElements = ref(new Set<number>((props.groupElements || []).filter((e) => !!e.album).map((e) => e.album.id)));

// Media elements in the group
const mediaElements = ref(new Set<number>((props.groupElements || []).filter((e) => !!e.media).map((e) => e.media.id)));

watch(
    () => props.groupElements,
    () => {
        albumElements.value = new Set<number>((props.groupElements || []).filter((e) => !!e.album).map((e) => e.album.id));
        mediaElements.value = new Set<number>((props.groupElements || []).filter((e) => !!e.media).map((e) => e.media.id));
    },
);

// Request error
const { error, errorDisplay, setError, unauthorized, accessDenied, serverError, networkError } = useCommonRequestErrors();

/**
 * Selects media to be added to the group
 * @param m The media element
 * @param callback The callback
 */
const selectMedia = (m: MediaListItem, callback: () => void) => {
    if (busy.value) {
        return;
    }

    const groupId = props.groupId;

    busy.value = true;

    makeApiRequest(
        apiHomeGroupAddElement(groupId, {
            t: HOME_PAGE_ELEMENT_TYPE_MEDIA,
            i: m.id,
        }),
    )
        .onSuccess(() => {
            busy.value = false;

            showSnackBar($t("Successfully added media to row"));

            mediaElements.value.add(m.id);

            emit("added-element");

            callback();
        })
        .onRequestError((err, handleErr) => {
            busy.value = false;

            handleErr(err, {
                unauthorized,
                tooManyElements: () => {
                    setError($t("The row reached the limit of 256 elements. Please, consider using another row."));
                },
                notCustomGroup: () => {
                    emit("must-reload");
                    forceClose();
                },
                accessDenied,
                notFound: () => {
                    emit("must-reload");
                    forceClose();
                },
                serverError,
                networkError,
            });
        })
        .onUnexpectedError((err) => {
            busy.value = false;
            console.error(err);
            setError(err.message);
        });
};

/**
 * Selects album to be added to the group
 * @param albumId The album ID
 * @param callback The callback
 */
const selectAlbum = (albumId: number, callback: () => void) => {
    if (busy.value) {
        return;
    }

    busy.value = true;

    const groupId = props.groupId;

    makeApiRequest(
        apiHomeGroupAddElement(groupId, {
            t: HOME_PAGE_ELEMENT_TYPE_ALBUM,
            i: albumId,
        }),
    )
        .onSuccess(() => {
            busy.value = false;

            showSnackBar($t("Successfully added album to row"));

            albumElements.value.add(albumId);

            emit("added-element");

            callback();
        })
        .onRequestError((err, handleErr) => {
            busy.value = false;
            handleErr(err, {
                unauthorized,
                tooManyElements: () => {
                    setError($t("The row reached the limit of 256 elements. Please, consider using another row."));
                },
                notCustomGroup: () => {
                    emit("must-reload");
                    forceClose();
                },
                accessDenied,
                notFound: () => {
                    emit("must-reload");
                    forceClose();
                },
                serverError,
                networkError,
            });
        })
        .onUnexpectedError((err) => {
            busy.value = false;
            console.error(err);
            setError(err.message);
        });
};

// Page scroll
const pageScroll = ref(0);

/**
 * Event handler for 'scroll'
 * @param e The event
 */
const onPageScroll = (e: Event) => {
    pageScroll.value = (e.target as HTMLElement).scrollTop;

    if (!isAlbums.value) {
        emitAppEvent(EVENT_NAME_ADVANCED_SEARCH_SCROLL, e);
    }
};

/**
 * Scrolls to the top
 */
const goTop = () => {
    if (!isAlbums.value) {
        emitAppEvent(EVENT_NAME_ADVANCED_SEARCH_GO_TOP);
    }

    scrollToTop();
};
</script>
