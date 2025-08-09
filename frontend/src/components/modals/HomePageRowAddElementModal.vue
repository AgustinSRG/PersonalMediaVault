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
                    <a href="javascript:;" class="horizontal-filter-menu-item" :class="{ selected: isAlbums }" @click="changeToAlbums"
                        ><i class="fas fa-list"></i> {{ $t("Albums") }}</a
                    >
                    <a href="javascript:;" class="horizontal-filter-menu-item" :class="{ selected: !isAlbums }" @click="changeToMedia"
                        ><i class="fas fa-photo-film"></i> {{ $t("Media") }}</a
                    >
                </div>
                <PageAdvancedSearch
                    v-if="!isAlbums"
                    :display="true"
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
                ></PageAdvancedSearch>
                <PageAlbums
                    v-else
                    :display="true"
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
    </ModalDialogContainer>
</template>

<script lang="ts">
import type { PropType } from "vue";
import { defineComponent, nextTick } from "vue";
import { useVModel } from "../../utils/v-model";

import PageAdvancedSearch from "@/components/pages/PageAdvancedSearch.vue";
import PageAlbums from "@/components/pages/PageAlbums.vue";
import { makeApiRequest } from "@asanrom/request-browser";
import { AppEvents } from "@/control/app-events";
import { AuthController, EVENT_NAME_UNAUTHORIZED } from "@/control/auth";
import { EVENT_NAME_ADVANCED_SEARCH_GO_TOP, EVENT_NAME_ADVANCED_SEARCH_SCROLL, PagesController } from "@/control/pages";
import { EVENT_NAME_PAGE_PREFERENCES_UPDATED, getPagePreferences } from "@/control/app-preferences";
import type { HomePageElement } from "@/api/api-home";
import { apiHomeGroupAddElement, HOME_PAGE_ELEMENT_TYPE_ALBUM, HOME_PAGE_ELEMENT_TYPE_MEDIA } from "@/api/api-home";
import type { MediaListItem } from "@/api/models";

export default defineComponent({
    name: "HomePageRowAddElementModal",
    components: {
        PageAdvancedSearch,
        PageAlbums,
    },
    props: {
        display: Boolean,
        groupId: Number,
        groupElements: Array as PropType<HomePageElement[]>,
    },
    emits: ["update:display", "must-reload", "added-element"],
    setup(props) {
        return {
            displayStatus: useVModel(props, "display"),
        };
    },
    data: function () {
        const pagePreferences = getPagePreferences("adv-search");
        return {
            busy: false,

            isAlbums: true,

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

            albumElements: new Set<number>((this.groupElements || []).filter((e) => !!e.album).map((e) => e.album.id)),
            mediaElements: new Set<number>((this.groupElements || []).filter((e) => !!e.media).map((e) => e.media.id)),
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
        groupElements: function () {
            this.albumElements = new Set<number>((this.groupElements || []).filter((e) => !!e.album).map((e) => e.album.id));
            this.mediaElements = new Set<number>((this.groupElements || []).filter((e) => !!e.media).map((e) => e.media.id));
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

        changeToAlbums: function () {
            this.isAlbums = true;
            this.updatePagePreferences();
        },

        changeToMedia: function () {
            this.isAlbums = false;
            this.updatePagePreferences();
        },

        selectMedia: function (m: MediaListItem, callback: () => void) {
            if (this.busy) {
                return;
            }
            const groupId = this.groupId;
            this.busy = true;
            // Add
            makeApiRequest(
                apiHomeGroupAddElement(groupId, {
                    t: HOME_PAGE_ELEMENT_TYPE_MEDIA,
                    i: m.id,
                }),
            )
                .onSuccess(() => {
                    this.busy = false;
                    PagesController.ShowSnackBar(this.$t("Successfully added media to row"));
                    this.mediaElements.add(m.id);
                    this.$emit("added-element");
                    callback();
                })
                .onRequestError((err, handleErr) => {
                    this.busy = false;
                    handleErr(err, {
                        unauthorized: () => {
                            AppEvents.Emit(EVENT_NAME_UNAUTHORIZED);
                        },
                        tooManyElements: () => {
                            PagesController.ShowSnackBar(
                                this.$t("Error") +
                                    ": " +
                                    this.$t("The row reached the limit of 256 elements. Please, consider using another row."),
                            );
                        },
                        notCustomGroup: () => {
                            this.$emit("must-reload");
                            this.close();
                        },
                        accessDenied: () => {
                            PagesController.ShowSnackBar(this.$t("Error") + ": " + this.$t("Access denied"));
                            AuthController.CheckAuthStatusSilent();
                        },
                        notFound: () => {
                            PagesController.ShowSnackBar(this.$t("Error") + ": " + this.$t("Not found"));
                            this.$emit("must-reload");
                            this.close();
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

        selectAlbum: function (albumId: number, callback: () => void) {
            if (this.busy) {
                return;
            }
            const groupId = this.groupId;
            this.busy = true;
            // Add
            makeApiRequest(
                apiHomeGroupAddElement(groupId, {
                    t: HOME_PAGE_ELEMENT_TYPE_ALBUM,
                    i: albumId,
                }),
            )
                .onSuccess(() => {
                    this.busy = false;
                    PagesController.ShowSnackBar(this.$t("Successfully added album to row"));
                    this.albumElements.add(albumId);
                    this.$emit("added-element");
                    callback();
                })
                .onRequestError((err, handleErr) => {
                    this.busy = false;
                    handleErr(err, {
                        unauthorized: () => {
                            AppEvents.Emit(EVENT_NAME_UNAUTHORIZED);
                        },
                        tooManyElements: () => {
                            PagesController.ShowSnackBar(
                                this.$t("Error") +
                                    ": " +
                                    this.$t("The row reached the limit of 256 elements. Please, consider using another row."),
                            );
                        },
                        notCustomGroup: () => {
                            this.$emit("must-reload");
                        },
                        accessDenied: () => {
                            PagesController.ShowSnackBar(this.$t("Error") + ": " + this.$t("Access denied"));
                            AuthController.CheckAuthStatusSilent();
                        },
                        notFound: () => {
                            PagesController.ShowSnackBar(this.$t("Error") + ": " + this.$t("Not found"));
                            this.$emit("must-reload");
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
            const pagePreferences = getPagePreferences(this.isAlbums ? "albums" : "adv-search");

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

            if (!this.isAlbums) {
                AppEvents.Emit(EVENT_NAME_ADVANCED_SEARCH_SCROLL, e);
            }
        },

        goTop: function () {
            if (!this.isAlbums) {
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
