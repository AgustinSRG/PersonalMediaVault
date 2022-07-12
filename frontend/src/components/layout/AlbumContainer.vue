<template>
  <div class="album-container" tabindex="-1">
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
            v-if="canWrite"
            type="button"
            :title="$t('Rename')"
            class="album-header-btn"
            @click="renameAlbum"
          >
            <i class="fas fa-pencil-alt"></i>
          </button>

          <button
            v-if="canWrite"
            type="button"
            :title="$t('Delete')"
            class="album-header-btn"
            @click="deleteAlbum"
          >
            <i class="fas fa-trash-alt"></i>
          </button>
        </div>
        <div class="album-post-text">
          {{ renderPos(currentPos) }} / {{ albumList.length }}
        </div>
      </div>
    </div>
    <div
      v-show="!loading && albumData"
      class="album-body"
      @scroll.passive="closeOptionsMenu"
    >
      <div
        v-for="(item, i) in albumList"
        :key="item.list_id"
        class="album-body-item"
        :class="{ current: i === currentPos }"
        :title="item.title || $t('Untitled')"
        tabindex="0"
        @click="clickMedia(item)"
        @keydown="clickOnEnter"
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

        <div class="album-body-item-title">
          {{ item.title || $t("Untitled") }}
        </div>

        <button
          v-if="canWrite"
          type="button"
          :title="$t('Options')"
          class="album-body-btn"
          @click="showOptions(item, i, $event)"
          @mousedown="stopPropagationEvent"
          @toutchstart.passive="stopPropagationEvent"
        >
          <i class="fas fa-bars"></i>
        </button>
      </div>
    </div>
    <AlbumContextMenu
      v-model:shown="contextShown"
      :mindex="contextIndex"
      :mlength="albumList.length"
      :x="contextX"
      :y="contextY"
      @move-up="moveMediaUp"
      @move-down="moveMediaDown"
      @media-remove="removeMedia"
    ></AlbumContextMenu>
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
import { AmbumsAPI } from "@/api/api-albums";
import { AlbumsController } from "@/control/albums";
import { AppEvents } from "@/control/app-events";
import { AppStatus } from "@/control/app-status";
import { AuthController } from "@/control/auth";
import { copyObject } from "@/utils/objects";
import { GetAssetURL, Request } from "@/utils/request";
import { renderTimeSeconds } from "@/utils/time-utils";
import { isTouchDevice } from "@/utils/touch";
import { defineComponent, nextTick } from "vue";

import AlbumContextMenu from "./AlbumContextMenu.vue";

declare var Sortable;

export default defineComponent({
  name: "AlbumContainer",
  emits: ["album-rename", "album-delete"],
  components: {
    AlbumContextMenu,
  },
  data: function () {
    return {
      albumId: AlbumsController.CurrentAlbum,
      albumData: AlbumsController.CurrentAlbumData,

      albumList: [],

      loading: AlbumsController.CurrentAlbumLoading,

      currentPos: AlbumsController.CurrentAlbumPos,

      currentMenuOpen: "",
      contextShown: false,
      contextIndex: -1,
      contextX: 0,
      contextY: 0,

      loop: false,
      random: false,

      canWrite: AuthController.CanWrite,
    };
  },
  methods: {
    onAlbumUpdate: function () {
      this.albumId = AlbumsController.CurrentAlbum;
      this.albumData = AlbumsController.CurrentAlbumData;
      this.updateAlbumsList();
    },

    closeOptionsMenu: function () {
      if (this.contextShown) {
        this.contextShown = false;
      }
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
      AlbumsController.ToggleLoop();
    },

    toggleRandom: function () {
      AlbumsController.ToggleRandom();
    },

    renameAlbum: function () {
      this.$emit("album-rename");
    },

    renderPos: function (p) {
      if (p < 0) {
        return "?";
      } else {
        return "" + (p + 1);
      }
    },

    deleteAlbum: function () {
      this.$emit("album-delete");
    },

    renderTime: function (s: number): string {
      return renderTimeSeconds(s);
    },

    getThumbnail(thumb: string) {
      return GetAssetURL(thumb);
    },

    onUpdateSortable: function (event) {
      AlbumsController.MoveCurrentAlbumOrder(event.oldIndex, event.newIndex);
    },

    updateAlbumsList: function () {
      this.currentMenuOpen = "";
      const prefix = "" + Date.now() + "-";
      let i = 0;
      if (this.albumData) {
        this.albumList = this.albumData.list.map((a) => {
          const o = copyObject(a);
          o.list_id = prefix + i;
          i++;
          return o;
        });
      } else {
        this.albumList = [];
      }
    },

    clickMedia: function (item) {
      AppStatus.ClickOnMedia(item.id, false);
    },

    showOptions: function (item, i, event) {
      event.stopPropagation();

      if (this.contextShown && this.currentMenuOpen === item.list_id) {
        this.currentMenuOpen = "";
        this.contextShown = false;
      } else {
        this.currentMenuOpen = item.list_id;
        this.contextShown = true;
        this.contextIndex = i;

        var targetRect = event.target.getBoundingClientRect();

        this.contextX = targetRect.left + targetRect.width;

        if (targetRect.top > window.innerHeight / 2) {
          this.contextY = targetRect.top;
        } else {
          this.contextY = targetRect.top + targetRect.height;
        }
      }
    },

    onAlbumPosUpdate: function () {
      this.loop = AlbumsController.AlbumLoop;
      this.random = AlbumsController.AlbumRandom;
      this.currentPos = AlbumsController.CurrentAlbumPos;
      nextTick(() => {
        this.scrollToSelected();
      });
    },

    stopPropagationEvent: function (e) {
      e.stopPropagation();
    },

    moveMediaUp: function (i: number) {
      if (i > 0) {
        AlbumsController.MoveCurrentAlbumOrder(i, i - 1);
      }
    },

    moveMediaDown: function (i: number) {
      if (i < this.albumList.length - 1) {
        AlbumsController.MoveCurrentAlbumOrder(i, i + 1);
      }
    },

    removeMedia: function (i: number) {
      const media = this.albumList[i];
      if (!media) {
        return;
      }
      const albumId = this.albumId;
      Request.Do(AmbumsAPI.RemoveMediaFromAlbum(albumId, media.id))
        .onSuccess(() => {
          AppEvents.Emit("snack", this.$t("Successfully removed from album"));
          AlbumsController.OnChangedAlbum(albumId);
        })
        .onRequestError((err) => {
          Request.ErrorHandler()
            .add(401, "*", () => {
              AppEvents.Emit("unauthorized");
            })
            .handle(err);
        })
        .onUnexpectedError((err) => {
          console.error(err);
        });
    },

    scrollToSelected: function () {
      const itemHeight = 128;
      const element = this.$el.querySelector(".album-body");

      if (!element) {
        return;
      }

      const scrollHeight = element.scrollHeight;
      const height = element.getBoundingClientRect().height;

      const itemTop = this.currentPos * itemHeight;

      const expectedTop = height / 2 - itemHeight / 2;

      const scroll = Math.max(
        0,
        Math.min(scrollHeight - height, itemTop - expectedTop)
      );

      element.scrollTop = scroll;
    },

    clickOnEnter: function (event) {
      if (event.key === "Enter") {
        event.preventDefault();
        event.stopPropagation();
        event.target.click();
      }
    },

    updateAuthInfo: function () {
      this.canWrite = AuthController.CanWrite;
      if (this.$options.sortable) {
        this.$options.sortable.option("disabled", !this.canWrite);
      }
    },
  },
  mounted: function () {
    this.$options.albumUpdateH = this.onAlbumUpdate.bind(this);
    AppEvents.AddEventListener(
      "current-album-update",
      this.$options.albumUpdateH
    );

    this.onAlbumPosUpdate();

    this.updateAlbumsList();

    this.$options.loadingH = this.onAlbumLoading.bind(this);
    AppEvents.AddEventListener("current-album-loading", this.$options.loadingH);

    this.$options.posUpdateH = this.onAlbumPosUpdate.bind(this);
    AppEvents.AddEventListener("album-pos-update", this.$options.posUpdateH);

    this.$options.authUpdateH = this.updateAuthInfo.bind(this);

    AppEvents.AddEventListener(
      "auth-status-changed",
      this.$options.authUpdateH
    );

    // Sortable
    if (!isTouchDevice()) {
      var element = this.$el.querySelector(".album-body");
      this.$options.sortable = Sortable.create(element, {
        onUpdate: this.onUpdateSortable.bind(this),
        disabled: !this.canWrite,
      });
    }
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

    AppEvents.RemoveEventListener("album-pos-update", this.$options.posUpdateH);

    AppEvents.RemoveEventListener(
      "auth-status-changed",
      this.$options.authUpdateH
    );

    // Sortable
    if (this.$options.sortable) {
      this.$options.sortable.destroy();
      this.$options.sortable = null;
    }
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
  border-left: solid 1px var(--theme-border-color);
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
  border-bottom: solid 1px var(--theme-border-color);

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
  color: var(--theme-btn-color);
  background: transparent;
  border-radius: 100vw;
}

.album-body-btn {
  display: block;
  width: 32px;
  height: 32px;
  box-shadow: none;
  border: none;
  cursor: pointer;
  font-size: 20px;
  color: var(--theme-btn-color);
  background: transparent;
  border-radius: 100vw;
}

.light-theme .album-header-btn.toggled {
  background: rgba(0, 0, 0, 0.2);
}

.dark-theme .album-header-btn.toggled {
  background: rgba(255, 255, 255, 0.2);
}

.album-header-btn:disabled,
.album-body-btn:disabled {
  opacity: 0.7;
  cursor: default;
}

.album-header-btn:not(:disabled):hover,
.album-body-btn:not(:disabled):hover {
  color: var(--theme-btn-hover-color);
}

.album-body-item {
  position: relative;
  display: flex;
  padding: 8px;
  align-items: center;
  cursor: pointer;
}

.light-theme .album-body-item:hover {
  background: rgba(0, 0, 0, 0.1);
}

.light-theme .album-body-item.current,
.light-theme .album-body-item.current:hover {
  background: rgba(0, 0, 0, 0.2);
}

.dark-theme .album-body-item:hover {
  background: rgba(255, 255, 255, 0.1);
}

.dark-theme .album-body-item.current,
.dark-theme .album-body-item.current:hover {
  background: rgba(255, 255, 255, 0.2);
}

.album-body-item-thumbnail {
  position: relative;
  width: 114px;
  height: 114px;
  display: flex;
  justify-content: center;
  align-items: center;
  border-radius: 4px;
  overflow: hidden;
}

.light-theme .album-body-item-thumbnail {
  background: rgba(0, 0, 0, 0.1);
}

.dark-theme .album-body-item-thumbnail {
  background: rgba(255, 255, 255, 0.1);
}

.album-body-item-thumbnail img {
  width: 100%;
  height: 100%;
}

.album-body-item-thumb-tag {
  position: absolute;
  background: rgba(0, 0, 0, 0.75);
  color: white;
  bottom: 0.25rem;
  right: 0.25rem;
  font-size: small;
  padding: 0.25rem;
}

.album-body-item-thumb-pos {
  position: absolute;
  background: rgba(0, 0, 0, 0.75);
  color: white;
  top: 0.25rem;
  left: 0.25rem;
  font-size: small;
  padding: 0.25rem;
}

.album-body-item-title {
  padding: 0.5rem;
  font-size: 1.1rem;
  width: calc(100% - 114px - 32px);
  height: 114px;
  overflow: hidden;
  text-overflow: ellipsis;
}
</style>
