<template>
  <div class="player-container" tabindex="-1">
    <EmptyPlayer
      v-if="!mdata || mdata.type === 0"
      :mid="mid"
      :status="status"
      :rtick="tick"
      :prev="prev"
      :next="next"
      :inalbum="isInAlbum"
      :albumloading="albumLoading"
      :canwrite="canWrite"
      @gonext="goNext"
      @goprev="goPrev"
      v-model:fullscreen="fullScreen"
      @update:fullscreen="onUpdateFullScreen"
    ></EmptyPlayer>
    <ImagePlayer
      v-if="mdata && mdata.type === 1"
      :mid="mid"
      :metadata="mdata"
      :rtick="tick"
      :prev="prev"
      :next="next"
      :inalbum="isInAlbum"
      :canwrite="canWrite"
      @gonext="goNext"
      @goprev="goPrev"
      v-model:fullscreen="fullScreen"
      @update:fullscreen="onUpdateFullScreen"
      v-model:showcontrols="showControls"
      @albums-open="openAlbums"
      @stats-open="openStats"
    ></ImagePlayer>
    <VideoPlayer
      v-if="mdata && mdata.type === 2"
      :mid="mid"
      :metadata="mdata"
      :rtick="tick"
      :prev="prev"
      :next="next"
      :inalbum="isInAlbum"
      :canwrite="canWrite"
      @gonext="goNext"
      @goprev="goPrev"
      v-model:fullscreen="fullScreen"
      @update:fullscreen="onUpdateFullScreen"
      @albums-open="openAlbums"
      @stats-open="openStats"
    ></VideoPlayer>
    <AudioPlayer
      v-if="mdata && mdata.type === 3"
      :mid="mid"
      :metadata="mdata"
      :rtick="tick"
      :prev="prev"
      :next="next"
      :inalbum="isInAlbum"
      :canwrite="canWrite"
      @gonext="goNext"
      @goprev="goPrev"
      v-model:fullscreen="fullScreen"
      @update:fullscreen="onUpdateFullScreen"
      @albums-open="openAlbums"
      @stats-open="openStats"
    ></AudioPlayer>

    <AlbumListModal v-model:display="displayAlbumList"></AlbumListModal>

    <SizeStatsModal
      :mid="mid"
      v-model:display="displaySizeStats"
    ></SizeStatsModal>
  </div>
</template>

<script lang="ts">
import { AppEvents } from "@/control/app-events";
import { MediaController } from "@/control/media";
import { defineComponent, nextTick } from "vue";

import EmptyPlayer from "@/components/player/EmptyPlayer.vue";
import AudioPlayer from "@/components/player/AudioPlayer.vue";
import VideoPlayer from "@/components/player/VideoPlayer.vue";
import ImagePlayer from "@/components/player/ImagePlayer.vue";
import { AlbumsController } from "@/control/albums";
import { AppStatus } from "@/control/app-status";
import { AuthController } from "@/control/auth";
import { FocusTrap } from "../../utils/focus-trap";
import { closeFullscreen } from "@/utils/full-screen";

import AlbumListModal from "../modals/AlbumListModal.vue";
import SizeStatsModal from "../modals/SizeStatsModal.vue";

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
  },
  data: function () {
    return {
      tick: 0,
      status: "loading",
      loading: MediaController.Loading,
      mid: MediaController.MediaId,
      mdata: MediaController.MediaData,

      fullScreen: false,
      showControls: true,

      prev: AlbumsController.CurrentPrev,
      next: AlbumsController.CurrentNext,
      isInAlbum: AppStatus.CurrentAlbum >= 0,
      albumLoading: AlbumsController.CurrentAlbumLoading,

      canWrite: AuthController.CanWrite,

      displayAlbumList: false,
      displaySizeStats: false,
    };
  },
  methods: {
    updateMedia: function () {
      this.mid = MediaController.MediaId;
      if (MediaController.MediaData !== this.mdata) {
        this.mdata = MediaController.MediaData;
        this.tick++;
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
      } else if (this.mdata) {
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

    goNext: function () {
      if (this.next) {
        AppStatus.ClickOnMedia(this.next.id, false);
      }
    },

    goPrev: function () {
      if (this.prev) {
        AppStatus.ClickOnMedia(this.prev.id, false);
      }
    },

    onAlbumPosUpdate: function () {
      this.prev = AlbumsController.CurrentPrev;
      this.next = AlbumsController.CurrentNext;
      this.isInAlbum = AppStatus.CurrentAlbum >= 0;
    },

    updateAuthInfo: function () {
      this.canWrite = AuthController.CanWrite;
    },

    onUpdateFullScreen: function () {
      if (this.fullScreen) {
        if (this.$options.focusTrap) {
          this.$options.focusTrap.activate();
        }
        nextTick(() => {
          this.$el.focus();
        });
      } else {
        if (this.$options.focusTrap) {
          this.$options.focusTrap.deactivate();
        }
      }
    },

    focusLost: function () {
      closeFullscreen();
    },
  },
  mounted: function () {
    this.$options.loadingH = this.updateLoading.bind(this);
    this.$options.updateH = this.updateMedia.bind(this);

    this.$options.focusTrap = new FocusTrap(
      this.$el,
      this.focusLost.bind(this)
    );

    this.updateStatus();

    AppEvents.AddEventListener("current-media-loading", this.$options.loadingH);
    AppEvents.AddEventListener("current-media-update", this.$options.updateH);

    this.$options.posUpdateH = this.onAlbumPosUpdate.bind(this);
    AppEvents.AddEventListener("album-pos-update", this.$options.posUpdateH);

    this.$options.authUpdateH = this.updateAuthInfo.bind(this);

    AppEvents.AddEventListener(
      "auth-status-changed",
      this.$options.authUpdateH
    );

    this.$options.albumLoadingH = this.updateAlbumsLoading.bind(this);
    AppEvents.AddEventListener(
      "current-album-loading",
      this.$options.albumLoadingH
    );
  },
  beforeUnmount: function () {
    AppEvents.RemoveEventListener(
      "current-media-loading",
      this.$options.loadingH
    );
    AppEvents.RemoveEventListener(
      "current-media-update",
      this.$options.updateH
    );

    AppEvents.RemoveEventListener("album-pos-update", this.$options.posUpdateH);

    AppEvents.RemoveEventListener(
      "auth-status-changed",
      this.$options.authUpdateH
    );

    AppEvents.RemoveEventListener(
      "current-album-loading",
      this.$options.albumLoadingH
    );

    if (this.$options.focusTrap) {
      this.$options.focusTrap.destroy();
    }
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
</style>
