<template>
    <ModalDialogContainer ref="modalContainer" v-model:display="displayStatus" :static="true" @scroll.passive="onPageScroll">
        <div
            v-if="display"
            class="modal-dialog modal-xl modal-height-100"
            role="document"
            :class="{
                'items-fit-many': pageItemsFit <= 0,
                'items-size-small': pageItemsSize === 'small',
                'items-size-big': pageItemsSize === 'big',
                'items-size-normal': pageItemsSize !== 'small' && pageItemsSize !== 'big',
            }"
            :style="{ '--page-items-fit': pageItemsFit }"
        >
            <div class="modal-header">
                <div class="modal-title" v-if="!isUpload">
                    {{ $t("Search media to add to the album") }}
                </div>
                <div class="modal-title" v-if="isUpload">
                    {{ $t("Upload media to add to the album") }}
                </div>
                <button class="modal-close-btn" :title="$t('Close')" @click="close">
                    <i class="fas fa-times"></i>
                </button>
            </div>
            <div class="modal-body no-padding">
                <div class="horizontal-filter-menu two-child modal-top-menu">
                    <a href="javascript:;" @click="changeToUpload" class="horizontal-filter-menu-item" :class="{ selected: isUpload }"
                        ><i class="fas fa-upload"></i> {{ $t("Upload") }}</a
                    >
                    <a href="javascript:;" @click="changeToSearch" class="horizontal-filter-menu-item" :class="{ selected: !isUpload }"
                        ><i class="fas fa-search"></i> {{ $t("Search") }}</a
                    >
                </div>
                <PageAdvancedSearch
                    v-if="!isUpload"
                    ref="advSearch"
                    :display="true"
                    :inModal="true"
                    :noAlbum="aid"
                    @select-media="selectMedia"
                ></PageAdvancedSearch>
                <PageUpload v-if="isUpload" :display="true" :inModal="true" :fixedAlbum="aid" @media-go="close"></PageUpload>
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
import { Request } from "@asanrom/request-browser";
import { AppEvents } from "@/control/app-events";
import { AlbumsController } from "@/control/albums";
import { EVENT_NAME_PAGE_ITEMS_UPDATED, getPageItemsFit, getPageItemsSize } from "@/control/app-preferences";
import { AuthController, EVENT_NAME_UNAUTHORIZED } from "@/control/auth";
import { PagesController } from "@/control/pages";
import { apiAlbumsAddMediaToAlbum } from "@/api/api-albums";

export default defineComponent({
    components: {
        PageAdvancedSearch,
        PageUpload,
    },
    name: "AlbumAddMediaModal",
    emits: ["update:display"],
    props: {
        display: Boolean,
        aid: Number,
    },
    setup(props) {
        return {
            displayStatus: useVModel(props, "display"),
        };
    },
    data: function () {
        return {
            busy: false,

            isUpload: true,

            pageItemsFit: getPageItemsFit(),
            pageItemsSize: getPageItemsSize(),

            pageScroll: 0,
        };
    },
    methods: {
        close: function () {
            this.$refs.modalContainer.close();
        },

        changeToUpload: function () {
            this.isUpload = true;
        },

        changeToSearch: function () {
            this.isUpload = false;
        },

        selectMedia: function (mid, callback) {
            if (this.busy) {
                return;
            }
            const albumId = this.aid;
            this.busy = true;
            // Add
            Request.Do(apiAlbumsAddMediaToAlbum(albumId, mid))
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

        updatePageItemsPreferences: function () {
            this.pageItemsFit = getPageItemsFit();
            this.pageItemsSize = getPageItemsSize();
        },

        onPageScroll: function (e: Event) {
            this.pageScroll = (e.target as HTMLElement).scrollTop;

            if (!this.isUpload && this.$refs.advSearch) {
                this.$refs.advSearch.onScroll(e);
            }
        },

        goTop: function () {
            if (!this.isUpload && this.$refs.advSearch) {
                this.$refs.advSearch.goTop();

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
    mounted: function () {
        this.$listenOnAppEvent(EVENT_NAME_PAGE_ITEMS_UPDATED, this.updatePageItemsPreferences.bind(this));

        if (this.display) {
            nextTick(() => {
                this.$el.focus();
            });
        }
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
});
</script>
