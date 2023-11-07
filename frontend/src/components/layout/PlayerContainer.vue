<template>
    <div class="player-container" tabindex="-1">
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
import { MediaController } from "@/control/media";
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

import { AlbumsController } from "@/control/albums";
import { AppStatus } from "@/control/app-status";
import { AuthController } from "@/control/auth";
import { FocusTrap } from "../../utils/focus-trap";
import { closeFullscreen } from "@/utils/full-screen";

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

            hasPagePrev: AlbumsController.HasPagePrev,
            hasPageNext: AlbumsController.HasPageNext,

            minPlayer: false,

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
                AppEvents.Emit("page-media-nav-next");
            }
        },

        goPrev: function () {
            if (this.prev) {
                AppStatus.ClickOnMedia(this.prev.id, false);
            } else if (this.hasPagePrev) {
                AppEvents.Emit("page-media-nav-prev");
            }
        },

        onAlbumPosUpdate: function () {
            this.prev = AlbumsController.CurrentPrev;
            this.next = AlbumsController.CurrentNext;
            this.isInAlbum = AppStatus.CurrentAlbum >= 0;
        },

        onPagePosUpdate: function () {
            this.hasPagePrev = AlbumsController.HasPagePrev;
            this.hasPageNext = AlbumsController.HasPageNext;
        },

        updateAuthInfo: function () {
            this.canWrite = AuthController.CanWrite;
        },

        onUpdateFullScreen: function () {
            if (this.fullScreen) {
                this._handles.focusTrap.activate();
                nextTick(() => {
                    this.$el.focus();
                });
            } else {
                this._handles.focusTrap.deactivate();
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
        },

        onForceLoop: function (v: boolean) {
            this.loopForced = true;
            this.loopForcedValue = v;
        },
    },
    mounted: function () {
        this._handles = Object.create(null);
        this._handles.loadingH = this.updateLoading.bind(this);
        this._handles.updateH = this.updateMedia.bind(this);

        this._handles.focusTrap = new FocusTrap(this.$el, this.focusLost.bind(this));

        this._handles.timer = setInterval(this.checkPlayerSize.bind(this), 1000);
        this.checkPlayerSize();

        this.updateStatus();

        MediaController.AddLoadingEventListener(this._handles.loadingH);
        MediaController.AddUpdateEventListener(this._handles.updateH);

        this._handles.posUpdateH = this.onAlbumPosUpdate.bind(this);
        AppEvents.AddEventListener("album-pos-update", this._handles.posUpdateH);

        this._handles.onPagePosUpdateH = this.onPagePosUpdate.bind(this);
        AppEvents.AddEventListener("page-media-nav-update", this._handles.onPagePosUpdateH);

        this._handles.authUpdateH = this.updateAuthInfo.bind(this);

        AuthController.AddChangeEventListener(this._handles.authUpdateH);

        this._handles.albumLoadingH = this.updateAlbumsLoading.bind(this);
        AppEvents.AddEventListener("current-album-loading", this._handles.albumLoadingH);

        this._handles.goPrevH = this.goPrev.bind(this);
        AppEvents.AddEventListener("media-go-prev", this._handles.goPrevH);

        this._handles.goNextH = this.goNext.bind(this);
        AppEvents.AddEventListener("media-go-next", this._handles.goNextH);

        nextTick(() => {
            this.$el.focus();
        });
    },
    beforeUnmount: function () {
        MediaController.RemoveLoadingEventListener(this._handles.loadingH);
        MediaController.RemoveUpdateEventListener(this._handles.updateH);

        AppEvents.RemoveEventListener("album-pos-update", this._handles.posUpdateH);
        AppEvents.RemoveEventListener("page-media-nav-update", this._handles.onPagePosUpdateH);

        AuthController.RemoveChangeEventListener(this._handles.authUpdateH);

        AppEvents.RemoveEventListener("current-album-loading", this._handles.albumLoadingH);

        AppEvents.RemoveEventListener("media-go-prev", this._handles.goPrevH);
        AppEvents.RemoveEventListener("media-go-next", this._handles.goNextH);

        this._handles.focusTrap.destroy();

        clearInterval(this._handles.timer);
    },
});
</script>

<style>
@import "@/style/player/common.css";

@import "@/style/player/loader.css";

@import "@/style/player/player-top-bar.css";
@import "@/style/player/editor.css";

@import "@/style/player/player-config.css";
@import "@/style/player/context-menu.css";
@import "@/style/player/change-preview.css";

@import "@/style/player/empty-player.css";

@import "@/style/player/video-player.css";
@import "@/style/player/audio-player.css";
@import "@/style/player/image-player.css";

@import "@/style/player/timeline.css";

@import "@/style/player/volume.css";
@import "@/style/player/scale.css";

@import "@/style/player/subtitles.css";

@import "@/style/player/image-notes.css";

@import "@/style/player/ext-desc.css";
</style>
