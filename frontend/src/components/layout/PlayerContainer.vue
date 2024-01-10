<template>
    <div class="player-container" :class="{ 'using-touch-device': touchDevice }" tabindex="-1">
        <EmptyPlayer
            v-if="!mediaData || mediaData.type === 0"
            :mid="mid"
            :status="status"
            :rTick="tick"
            :prev="prev"
            :pagePrev="hasPagePrev"
            :pageNext="hasPageNext"
            :next="next"
            :inAlbum="isInAlbum"
            :albumLoading="albumLoading"
            :canWrite="canWrite"
            @go-next="goNext"
            @go-prev="goPrev"
            v-model:fullscreen="fullScreen"
            @update:fullscreen="onUpdateFullScreen"
            :min="minPlayer"
            @delete="openDelete"
        ></EmptyPlayer>
        <ImagePlayer
            v-if="mediaData && mediaData.type === 1"
            :mid="mid"
            :metadata="mediaData"
            :rTick="tick"
            :prev="prev"
            :next="next"
            :pagePrev="hasPagePrev"
            :pageNext="hasPageNext"
            :inAlbum="isInAlbum"
            :canWrite="canWrite"
            @go-next="goNext"
            @go-prev="goPrev"
            v-model:fullscreen="fullScreen"
            @update:fullscreen="onUpdateFullScreen"
            v-model:showControls="showControls"
            @albums-open="openAlbums"
            @stats-open="openStats"
            v-model:display-tag-list="displayTagList"
            @ext-desc-open="openExtendedDescription"
            :min="minPlayer"
            @delete="openDelete"
        ></ImagePlayer>
        <VideoPlayer
            v-if="mediaData && mediaData.type === 2"
            :mid="mid"
            :metadata="mediaData"
            :rTick="tick"
            :prev="prev"
            :next="next"
            :pagePrev="hasPagePrev"
            :pageNext="hasPageNext"
            :inAlbum="isInAlbum"
            :canWrite="canWrite"
            @go-next="goNext"
            @go-prev="goPrev"
            v-model:fullscreen="fullScreen"
            @update:fullscreen="onUpdateFullScreen"
            v-model:userControls="showControls"
            @albums-open="openAlbums"
            @stats-open="openStats"
            v-model:display-tag-list="displayTagList"
            @ext-desc-open="openExtendedDescription"
            :min="minPlayer"
            :loopForced="loopForced"
            :loopForcedValue="loopForcedValue"
            @force-loop="onForceLoop"
            :autoPlay="!(displayAlbumList || displayExtendedDescription || displaySizeStats || displayUpload)"
            @delete="openDelete"
        ></VideoPlayer>
        <AudioPlayer
            v-if="mediaData && mediaData.type === 3"
            :mid="mid"
            :metadata="mediaData"
            :rTick="tick"
            :prev="prev"
            :next="next"
            :pagePrev="hasPagePrev"
            :pageNext="hasPageNext"
            :inAlbum="isInAlbum"
            :canWrite="canWrite"
            @go-next="goNext"
            @go-prev="goPrev"
            v-model:fullscreen="fullScreen"
            @update:fullscreen="onUpdateFullScreen"
            @albums-open="openAlbums"
            @stats-open="openStats"
            v-model:display-tag-list="displayTagList"
            @ext-desc-open="openExtendedDescription"
            :min="minPlayer"
            :loopForced="loopForced"
            :loopForcedValue="loopForcedValue"
            @force-loop="onForceLoop"
            :autoPlay="!(displayAlbumList || displayExtendedDescription || displaySizeStats || displayUpload)"
            @delete="openDelete"
        ></AudioPlayer>

        <AlbumListModal v-if="displayAlbumList" v-model:display="displayAlbumList"></AlbumListModal>

        <SizeStatsModal :mid="mid" v-if="displaySizeStats" v-model:display="displaySizeStats"></SizeStatsModal>

        <ExtendedDescriptionModal v-if="displayExtendedDescription" v-model:display="displayExtendedDescription"></ExtendedDescriptionModal>

        <MediaDeleteModal v-if="displayDelete" v-model:display="displayDelete"></MediaDeleteModal>
    </div>
</template>

<script lang="ts">
import { AppEvents } from "@/control/app-events";
import { EVENT_NAME_MEDIA_LOADING, EVENT_NAME_MEDIA_UPDATE, MediaController } from "@/control/media";
import { defineAsyncComponent, defineComponent, nextTick } from "vue";

import LoadingOverlay from "./LoadingOverlay.vue";

const EmptyPlayer = defineAsyncComponent({
    loader: () => import("@/components/player/EmptyPlayer.vue"),
    loadingComponent: LoadingOverlay,
    delay: 1000,
});

const AudioPlayer = defineAsyncComponent({
    loader: () => import("@/components/player/AudioPlayer.vue"),
    loadingComponent: LoadingOverlay,
    delay: 1000,
});

const VideoPlayer = defineAsyncComponent({
    loader: () => import("@/components/player/VideoPlayer.vue"),
    loadingComponent: LoadingOverlay,
    delay: 1000,
});

const ImagePlayer = defineAsyncComponent({
    loader: () => import("@/components/player/ImagePlayer.vue"),
    loadingComponent: LoadingOverlay,
    delay: 1000,
});

import { AlbumsController, EVENT_NAME_CURRENT_ALBUM_LOADING, EVENT_NAME_CURRENT_ALBUM_MEDIA_POSITION_UPDATED } from "@/control/albums";
import { AppStatus } from "@/control/app-status";
import { AuthController, EVENT_NAME_AUTH_CHANGED } from "@/control/auth";
import { FocusTrap } from "../../utils/focus-trap";
import { closeFullscreen } from "@/utils/full-screen";
import {
    EVENT_NAME_GO_NEXT,
    EVENT_NAME_GO_PREV,
    EVENT_NAME_PAGE_MEDIA_NAV_UPDATE,
    EVENT_NAME_PAGE_NAV_NEXT,
    EVENT_NAME_PAGE_NAV_PREV,
    PagesController,
} from "@/control/pages";
import { isTouchDevice } from "@/utils/touch";

const AlbumListModal = defineAsyncComponent({
    loader: () => import("@/components/modals/AlbumListModal.vue"),
    loadingComponent: LoadingOverlay,
    delay: 1000,
});

const SizeStatsModal = defineAsyncComponent({
    loader: () => import("@/components/modals/SizeStatsModal.vue"),
    loadingComponent: LoadingOverlay,
    delay: 1000,
});

const ExtendedDescriptionModal = defineAsyncComponent({
    loader: () => import("@/components/modals/ExtendedDescriptionModal.vue"),
    loadingComponent: LoadingOverlay,
    delay: 1000,
});

const MediaDeleteModal = defineAsyncComponent({
    loader: () => import("@/components/modals/MediaDeleteModal.vue"),
    loadingComponent: LoadingOverlay,
    delay: 1000,
});

export default defineComponent({
    name: "PlayerContainer",
    emits: [],
    components: {
        EmptyPlayer,
        AudioPlayer,
        VideoPlayer,
        ImagePlayer,
        AlbumListModal,
        SizeStatsModal,
        ExtendedDescriptionModal,
        MediaDeleteModal,
    },
    props: {
        displayUpload: Boolean,
    },
    setup() {
        return {
            focusTrap: null as FocusTrap,
            timer: null,
        };
    },
    data: function () {
        return {
            tick: 0,
            status: "loading",
            loading: MediaController.Loading,
            mid: MediaController.MediaId,
            mediaData: MediaController.MediaData,

            fullScreen: false,
            showControls: true,

            prev: AlbumsController.CurrentPrev,
            next: AlbumsController.CurrentNext,
            isInAlbum: AppStatus.CurrentAlbum >= 0,
            albumLoading: AlbumsController.CurrentAlbumLoading,

            canWrite: AuthController.CanWrite,

            displayAlbumList: false,
            displaySizeStats: false,
            displayTagList: false,
            displayExtendedDescription: false,
            displayDelete: false,

            hasPagePrev: PagesController.HasPagePrev,
            hasPageNext: PagesController.HasPageNext,

            minPlayer: false,
            touchDevice: isTouchDevice(),

            loopForced: false,
            loopForcedValue: false,
        };
    },
    methods: {
        updateMedia: function () {
            this.displayDelete = false;
            this.mid = MediaController.MediaId;
            if (MediaController.MediaData !== this.mediaData) {
                this.mediaData = MediaController.MediaData;
                this.tick++;
                if (this.mid) {
                    this.$el.focus();
                }
            }
            this.updateStatus();
        },

        updateLoading: function (l) {
            this.loading = l;
            this.updateStatus();
        },

        updateAlbumsLoading: function (l) {
            this.albumLoading = l;
        },

        updateStatus: function () {
            if (this.loading) {
                this.status = "loading";
            } else if (this.mediaData) {
                this.status = "200";
            } else if (this.mid === -1) {
                this.status = "none";
            } else {
                this.status = "404";
            }
        },

        openAlbums: function () {
            this.displayAlbumList = true;
        },

        openStats: function () {
            this.displaySizeStats = true;
        },

        openExtendedDescription: function () {
            this.displayExtendedDescription = true;
        },

        openDelete: function () {
            this.displayDelete = true;
        },

        goNext: function () {
            if (this.next) {
                AppStatus.ClickOnMedia(this.next.id, false);
            } else if (this.hasPageNext) {
                AppEvents.Emit(EVENT_NAME_PAGE_NAV_NEXT);
            }
        },

        goPrev: function () {
            if (this.prev) {
                AppStatus.ClickOnMedia(this.prev.id, false);
            } else if (this.hasPagePrev) {
                AppEvents.Emit(EVENT_NAME_PAGE_NAV_PREV);
            }
        },

        onAlbumPosUpdate: function () {
            this.prev = AlbumsController.CurrentPrev;
            this.next = AlbumsController.CurrentNext;
            this.isInAlbum = AppStatus.CurrentAlbum >= 0;
        },

        onPagePosUpdate: function (prev: boolean, next: boolean) {
            this.hasPagePrev = prev;
            this.hasPageNext = next;
        },

        updateAuthInfo: function () {
            this.canWrite = AuthController.CanWrite;
        },

        onUpdateFullScreen: function () {
            if (this.fullScreen) {
                this.focusTrap.activate();
                nextTick(() => {
                    this.$el.focus();
                });
            } else {
                this.focusTrap.deactivate();
            }
        },

        focusLost: function () {
            closeFullscreen();
        },

        checkPlayerSize() {
            const rect = this.$el.getBoundingClientRect();
            const width = rect.width;
            const height = rect.height;

            if (width < 480 || height < 360) {
                this.minPlayer = true;
            } else {
                this.minPlayer = false;
            }

            this.touchDevice = isTouchDevice();
        },

        onForceLoop: function (v: boolean) {
            this.loopForced = true;
            this.loopForcedValue = v;
        },
    },
    mounted: function () {
        this.focusTrap = new FocusTrap(this.$el, this.focusLost.bind(this));

        this.timer = setInterval(this.checkPlayerSize.bind(this), 1000);
        this.checkPlayerSize();

        this.updateStatus();

        this.$listenOnAppEvent(EVENT_NAME_MEDIA_LOADING, this.updateLoading.bind(this));

        this.$listenOnAppEvent(EVENT_NAME_MEDIA_UPDATE, this.updateMedia.bind(this));

        this.$listenOnAppEvent(EVENT_NAME_CURRENT_ALBUM_MEDIA_POSITION_UPDATED, this.onAlbumPosUpdate.bind(this));

        this.$listenOnAppEvent(EVENT_NAME_PAGE_MEDIA_NAV_UPDATE, this.onPagePosUpdate.bind(this));

        this.$listenOnAppEvent(EVENT_NAME_AUTH_CHANGED, this.updateAuthInfo.bind(this));

        this.$listenOnAppEvent(EVENT_NAME_CURRENT_ALBUM_LOADING, this.updateAlbumsLoading.bind(this));

        this.$listenOnAppEvent(EVENT_NAME_GO_PREV, this.goPrev.bind(this));

        this.$listenOnAppEvent(EVENT_NAME_GO_NEXT, this.goNext.bind(this));

        nextTick(() => {
            this.$el.focus();
        });
    },
    beforeUnmount: function () {
        this.focusTrap.destroy();
        clearInterval(this.timer);
    },
});
</script>
