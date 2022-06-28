<template>
  <div class="album-container">
    <div v-if="!loading && albumData" class="album-header">
      <div class="album-header-title">
        <div class="album-title">
          <i class="fas fa-list-ol"></i> {{ albumData.name }}
        </div>
        <button
          type="button"
          :title="$t('Close')"
          class="album-header-btn album-close-btn"
          @click="closePage"
        >
          <i class="fas fa-times"></i>
        </button>
      </div>
      <div class="album-header-controls">
        <div class="album-buttons">
          <button
            type="button"
            :title="$t('Loop')"
            class="album-header-btn"
            :class="{ toggled: loop }"
            @click="toggleLoop"
          >
            <i class="fas fa-repeat"></i>
          </button>

          <button
            type="button"
            :title="$t('Random')"
            class="album-header-btn"
            :class="{ toggled: random }"
            @click="toggleRandom"
          >
            <i class="fas fa-shuffle"></i>
          </button>

          <button
            type="button"
            :title="$t('Rename')"
            class="page-header-btn"
            @click="renameAlbum"
          >
            <i class="fas fa-pencil-alt"></i>
          </button>

          <button
            type="button"
            :title="$t('Delete')"
            class="page-header-btn"
            @click="deleteAlbum"
          >
            <i class="fas fa-trash-alt"></i>
          </button>
        </div>
        <div class="album-post-text">
          {{ renderPos(currentPos) }} / {{ albumLength }}
        </div>
      </div>
    </div>
    <div v-if="!loading && albumData" class="album-body">
      <div
        v-for="(item, i) in albumData.list"
        :key="i"
        class="album-body-item"
        :class="{ current: i === currentPos }"
        :title="item.title || $t('Untitled')"
      >
        <div class="album-body-item-thumbnail">
          <div v-if="!item.thumbnail" class="no-thumb">
            <i v-if="item.type === 1" class="fas fa-image"></i>
            <i v-else-if="item.type === 2" class="fas fa-video"></i>
            <i v-else-if="item.type === 3" class="fas fa-headphones"></i>
            <i v-else class="fas fa-ban"></i>
          </div>
          <img
            v-if="item.thumbnail"
            :src="getThumbnail(item.thumbnail)"
            :alt="item.title || $t('Untitled')"
          />
          <div
            class="album-body-item-thumb-tag"
            v-if="item.type === 2 || item.type === 3"
          >
            {{ renderTime(item.duration) }}
          </div>
          <div class="album-body-item-thumb-pos">
            {{ renderPos(i) }}
          </div>
        </div>

        <div class="album-body-item-title">{{item.title || $t('Untitled')}}</div>
      </div>
    </div>
    <div v-if="loading" class="album-loader">
      <div class="loading-overlay-loader">
        <div></div>
        <div></div>
        <div></div>
        <div></div>
      </div>
    </div>
  </div>
</template>

<script lang="ts">
import { AlbumsController } from "@/control/albums";
import { AppEvents } from "@/control/app-events";
import { AppStatus } from "@/control/app-status";
import { GetAssetURL } from "@/utils/request";
import { renderTimeSeconds } from "@/utils/time-utils";
import { defineComponent } from "vue";

export default defineComponent({
  name: "AlbumContainer",
  data: function () {
    return {
      albumId: AlbumsController.CurrentAlbum,
      albumData: AlbumsController.CurrentAlbumData,

      loading: AlbumsController.CurrentAlbumLoading,

      currentPos: -1,
      albumLength: 0,

      loop: false,
      random: false,
    };
  },
  methods: {
    onAlbumUpdate: function () {
      this.albumId = AlbumsController.CurrentAlbum;
      this.albumData = AlbumsController.CurrentAlbumData;
    },

    onAlbumLoading: function (l) {
      if (l) {
        if (this.albumId !== AlbumsController.CurrentAlbum) {
          this.loading = true;
        }
      } else {
        this.loading = false;
      }
    },

    closePage: function () {
      AppStatus.CloseAbum();
    },

    toggleLoop: function () {
      this.loop = !this.loop;
    },

    toggleRandom: function () {
      this.random = !this.random;
    },

    renameAlbum: function () {},

    renderPos: function (p) {
      if (p < 0) {
        return "?";
      } else {
        return "" + (p + 1);
      }
    },

    deleteAlbum: function () {},

    renderTime: function (s: number): string {
      return renderTimeSeconds(s);
    },

    getThumbnail(thumb: string) {
      return GetAssetURL(thumb);
    },
  },
  mounted: function () {
    this.$options.albumUpdateH = this.onAlbumUpdate.bind(this);
    AppEvents.AddEventListener(
      "current-album-update",
      this.$options.albumUpdateH
    );

    this.$options.loadingH = this.onAlbumLoading.bind(this);
    AppEvents.AddEventListener("current-album-loading", this.$options.loadingH);
  },
  beforeUnmount: function () {
    AppEvents.RemoveEventListener(
      "current-album-update",
      this.$options.albumUpdateH
    );
    AppEvents.RemoveEventListener(
      "current-album-loading",
      this.$options.loadingH
    );
  },
});
</script>

<style>
.album-container {
  position: absolute;
  top: 57px;
  height: calc(100% - 57px);
  right: 0;
  width: 500px;
  border-left: solid 1px rgba(255, 255, 255, 0.1);
  display: none;
}

.layout-album .album-container {
  display: block;
}

@media (max-width: 1000px) {
  .album-container {
    width: calc(100%);
    height: calc(100% - 57px - 40px);
  }

  .layout-media-split.focus-left .album-container {
    display: none;
  }
}

.album-loader {
  position: absolute;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
  display: flex;
  justify-content: center;
  align-items: center;
}

.album-header {
  position: absolute;
  top: 0;
  left: 0;
  width: 100%;
  height: 114px;
  border-bottom: solid 1px rgba(255, 255, 255, 0.1);

  display: flex;
  flex-direction: column;
}

.album-header-title {
  width: 100%;
  height: 57px;
  display: flex;
  flex-direction: row;
  justify-content: center;
  align-items: center;
}

.album-header-controls {
  width: 100%;
  height: 57px;
  display: flex;
  flex-direction: row;
  justify-content: space-between;
  align-items: center;
}

.album-post-text {
  padding: 0 1rem;
}

.album-buttons {
  display: flex;
  flex-direction: row;
  padding-left: calc(1rem - 12px);
}

.album-buttons button {
  margin-right: 0.25rem;
}

.album-title {
  width: calc(100% - 48px);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  font-size: 1.2rem;
  padding: 0 1rem;
}

.album-title i {
  margin-right: 1rem;
}

.album-body {
  position: absolute;
  top: 114px;
  left: 0;
  width: 100%;
  height: calc(100% - 114px);
  overflow-y: auto;

  display: flex;
  flex-direction: column;
}

/* Buttons */

.album-header-btn {
  display: block;
  width: 48px;
  height: 48px;
  box-shadow: none;
  border: none;
  cursor: pointer;
  font-size: 24px;
  color: rgba(255, 255, 255, 0.75);
  background: transparent;
  outline: none;
  border-radius: 100vw;
}

.album-header-btn.toggled {
  background: rgba(255, 255, 255, 0.2);
}

.album-header-btn:disabled {
  opacity: 0.7;
  cursor: default;
}

.album-header-btn:hover {
  color: white;
}

.album-header-btn:disabled:hover {
  color: rgba(255, 255, 255, 0.75);
}

.album-header-btn:focus {
  outline: none;
}

.album-body-item {
  position: relative;
  display: flex;
  padding: 0.5rem;
  align-items: center;
  cursor: pointer;
}

.album-body-item:hover {
  background: rgba(255, 255, 255, 0.1);
}

.album-body-item.current,
.album-body-item.current:hover {
  background: rgba(255, 255, 255, 0.2);
}

.album-body-item-thumbnail {
  position: relative;
  width: 114px;
  height: 114px;
}

.album-body-item-thumbnail img {
  width: 100%;
  height: 100%;
}

.album-body-item-thumb-tag {
  position: absolute;
  background: rgba(0, 0, 0, 0.3);
  bottom: 0.25rem;
  right: 0.25rem;
  font-size: small;
  padding: 0.25rem;
}

.album-body-item-thumb-pos {
  position: absolute;
  background: rgba(0, 0, 0, 0.3);
  top: 0.25rem;
  left: 0.25rem;
  font-size: small;
  padding: 0.25rem;
}

.album-body-item-title {
  padding: 0.5rem;
  font-size: 1.1rem;
  width: calc(100% - 114px);
  height: 114px;
  overflow: hidden;
  text-overflow: ellipsis;
}
</style>
