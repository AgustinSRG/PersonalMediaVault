<template>
  <div class="player-container">
    <EmptyPlayer v-if="!mdata || mdata.type === 0" :mid="mid" :status="status" :rtick="tick" :prev="prev" :next="next" @gonext="goNext" @goprev="goPrev" v-model:fullscreen="fullScreen"></EmptyPlayer>
    <ImagePlayer v-if="mdata && mdata.type === 1" :mid="mid" :metadata="mdata" :rtick="tick" :prev="prev" :next="next" @gonext="goNext" @goprev="goPrev" v-model:fullscreen="fullScreen" v-model:showcontrols="showControls" @albums-open="openAlbums"></ImagePlayer>
    <VideoPlayer v-if="mdata && mdata.type === 2" :mid="mid" :metadata="mdata" :rtick="tick" :prev="prev" :next="next" @gonext="goNext" @goprev="goPrev" v-model:fullscreen="fullScreen" @albums-open="openAlbums"></VideoPlayer>
    <AudioPlayer v-if="mdata && mdata.type === 3" :mid="mid" :metadata="mdata" :rtick="tick" :prev="prev" :next="next" @gonext="goNext" @goprev="goPrev" v-model:fullscreen="fullScreen" @albums-open="openAlbums"></AudioPlayer>
  </div>
</template>

<script lang="ts">
import { AppEvents } from "@/control/app-events";
import { MediaController } from "@/control/media";
import { defineComponent } from "vue";

import EmptyPlayer from "@/components/player/EmptyPlayer.vue";
import AudioPlayer from "@/components/player/AudioPlayer.vue";
import VideoPlayer from "@/components/player/VideoPlayer.vue";
import ImagePlayer from "@/components/player/ImagePlayer.vue";

export default defineComponent({
  name: "PlayerContainer",
  emits: ["albums-open"],
  components: {
    EmptyPlayer,
    AudioPlayer,
    VideoPlayer,
    ImagePlayer,
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

      prev: null,
      next: null,
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
      this.$emit("albums-open");
    },

    goNext: function () {

    },

    goPrev: function () {

    },
  },
  mounted: function () {
    this.$options.loadingH = this.updateLoading.bind(this);
    this.$options.updateH = this.updateMedia.bind(this);

    this.updateStatus();

    AppEvents.AddEventListener("current-media-loading", this.$options.loadingH);
    AppEvents.AddEventListener("current-media-update", this.$options.updateH);
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
  },
});
</script>

<style>
.player-container {
  position: absolute;
  top: 57px;
  height: calc(100% - 57px);
  left: 0;
  width: 100%;
  background: blue;
  overflow: auto;
}

.layout-media-split .player-container {
  width: calc(100% - 500px);
}

.layout-album .player-container {
  width: calc(100% - 500px);
}

.layout-initial .player-container {
  display: none;
}
</style>
