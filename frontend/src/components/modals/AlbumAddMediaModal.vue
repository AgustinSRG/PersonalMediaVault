<template>
    <ModalDialogContainer v-model:display="displayStatus" :close-signal="closeSignal" :static="true" @scroll.passive="onPageScroll">
        <div
            v-if="display"
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
                    <a href="javascript:;" class="horizontal-filter-menu-item" :class="{ selected: isUpload }" @click="changeToUpload"
                        ><i class="fas fa-upload"></i> {{ $t("Upload") }}</a
                    >
                    <a href="javascript:;" class="horizontal-filter-menu-item" :class="{ selected: !isUpload }" @click="changeToSearch"
                        ><i class="fas fa-search"></i> {{ $t("Search") }}</a
                    >
                </div>
                <PageAdvancedSearch
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
                ></PageAdvancedSearch>
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

<script lang="ts">
import { defineComponent, nextTick } from "vue";
import { useVModel } from "../../utils/v-model";

import PageAdvancedSearch from "@/components/pages/PageAdvancedSearch.vue";
import PageUpload from "@/components/pages/PageUpload.vue";
import { makeApiRequest } from "@asanrom/request-browser";
import { AppEvents } from "@/control/app-events";
import { AlbumsController } from "@/control/albums";
import { AuthController, EVENT_NAME_UNAUTHORIZED } from "@/control/auth";
import { EVENT_NAME_ADVANCED_SEARCH_GO_TOP, EVENT_NAME_ADVANCED_SEARCH_SCROLL, PagesController } from "@/control/pages";
import { apiAlbumsAddMediaToAlbum } from "@/api/api-albums";
import { EVENT_NAME_PAGE_PREFERENCES_UPDATED, getPagePreferences } from "@/control/app-preferences";
import type { MediaListItem } from "@/api/models";

export default defineComponent({
    name: "AlbumAddMediaModal",
    components: {
        PageAdvancedSearch,
        PageUpload,
    },
    props: {
        display: Boolean,
        aid: Number,
    },
    emits: ["update:display"],
    setup(props) {
        return {
            displayStatus: useVModel(props, "display"),
        };
    },
    data: function () {
        const pagePreferences = getPagePreferences("adv-search");
        return {
            busy: false,

            isUpload: true,

            pageSize: pagePreferences.pageSize,

            rowSize: pagePreferences.rowSize,
            rowSizeMin: pagePreferences.rowSizeMin,

            minItemSize: pagePreferences.minItemSize,
            maxItemSize: pagePreferences.maxItemSize,

            padding: pagePreferences.padding,

            displayTitles: pagePreferences.displayTitles,
            roundedCorners: pagePreferences.roundedCorners,

            pageScroll: 0,

            closeSignal: 0,
        };
    },
    watch: {
        display: function () {
            if (this.display) {
                nextTick(() => {
                    this.$el.focus();
                });
            }
        },
    },
    mounted: function () {
        this.$listenOnAppEvent(EVENT_NAME_PAGE_PREFERENCES_UPDATED, this.updatePagePreferences.bind(this));

        if (this.display) {
            nextTick(() => {
                this.$el.focus();
            });
        }
    },
    methods: {
        close: function () {
            this.closeSignal++;
        },

        changeToUpload: function () {
            this.isUpload = true;
        },

        changeToSearch: function () {
            this.isUpload = false;
        },

        selectMedia: function (m: MediaListItem, callback: () => void) {
            if (this.busy) {
                return;
            }
            const albumId = this.aid;
            this.busy = true;
            // Add
            makeApiRequest(apiAlbumsAddMediaToAlbum(albumId, m.id))
                .onSuccess(() => {
                    this.busy = false;
                    PagesController.ShowSnackBar(this.$t("Successfully added to album"));
                    AlbumsController.OnChangedAlbum(albumId, true);
                    callback();
                })
                .onRequestError((err, handleErr) => {
                    this.busy = false;
                    handleErr(err, {
                        unauthorized: () => {
                            AppEvents.Emit(EVENT_NAME_UNAUTHORIZED);
                        },
                        maxSizeReached: () => {
                            PagesController.ShowSnackBar(
                                this.$t("Error") +
                                    ": " +
                                    this.$t("The album reached the limit of 1024 elements. Please, consider creating another album."),
                            );
                        },
                        badRequest: () => {
                            PagesController.ShowSnackBar(this.$t("Error") + ": " + this.$t("Bad request"));
                        },
                        accessDenied: () => {
                            PagesController.ShowSnackBar(this.$t("Error") + ": " + this.$t("Access denied"));
                            AuthController.CheckAuthStatusSilent();
                        },
                        notFound: () => {
                            PagesController.ShowSnackBar(this.$t("Error") + ": " + this.$t("Not found"));
                            AlbumsController.Load();
                        },
                        serverError: () => {
                            PagesController.ShowSnackBar(this.$t("Error") + ": " + this.$t("Internal server error"));
                        },
                        networkError: () => {
                            PagesController.ShowSnackBar(this.$t("Error") + ": " + this.$t("Could not connect to the server"));
                        },
                    });
                })
                .onUnexpectedError((err) => {
                    this.busy = false;
                    console.error(err);
                });
        },

        updatePagePreferences: function () {
            const pagePreferences = getPagePreferences("adv-search");

            this.pageSize = pagePreferences.pageSize;

            this.rowSize = pagePreferences.rowSize;
            this.rowSizeMin = pagePreferences.rowSizeMin;

            this.minItemSize = pagePreferences.minItemSize;
            this.maxItemSize = pagePreferences.maxItemSize;

            this.padding = pagePreferences.padding;

            this.displayTitles = pagePreferences.displayTitles;
            this.roundedCorners = pagePreferences.roundedCorners;
        },

        onPageScroll: function (e: Event) {
            this.pageScroll = (e.target as HTMLElement).scrollTop;

            if (!this.isUpload) {
                AppEvents.Emit(EVENT_NAME_ADVANCED_SEARCH_SCROLL, e);
            }
        },

        goTop: function () {
            if (!this.isUpload) {
                AppEvents.Emit(EVENT_NAME_ADVANCED_SEARCH_GO_TOP);

                nextTick(() => {
                    this.$el.scrollTop = 0;
                    this.$el.focus();
                });
            } else {
                this.$el.scrollTop = 0;
                this.$el.focus();
            }
        },
    },
});
</script>
