<template>
  <div
    class="side-bar"
    :class="{ hidden: !display }"
    @click="stopPropagationEvent"
  >
    <div class="side-bar-header">
      <div class="top-bar-logo-td">
        <button
          type="button"
          class="top-bar-button"
          :title="$t('Main menu')"
          @click="close"
        >
          <i class="fas fa-bars"></i>
        </button>
        <span :title="$t('Personal Media Vault')" class="top-bar-title"
          ><i class="fab fa-youtube"></i> PMV</span
        >
        <span :title="$t('Personal Media Vault')" class="top-bar-title-min"
          ><i class="fab fa-youtube"></i> PMV</span
        >
      </div>
    </div>
    <div class="side-bar-body">
      <div
        class="side-bar-option"
        :class="{ selected: album < 0 && page === 'home' }"
        tabindex="0"
        :title="$t('Home')"
        @click="goToPage('home')"
      >
        <div class="side-bar-option-icon"><i class="fas fa-home"></i></div>
        <div class="side-bar-option-text">{{ $t("Home") }}</div>
      </div>

      <div
        v-if="!!search"
        class="side-bar-option"
        :class="{ selected: album < 0 && page === 'search' }"
        tabindex="0"
        :title="$t('Search results')"
        @click="goToSearch"
      >
        <div class="side-bar-option-icon"><i class="fas fa-search"></i></div>
        <div class="side-bar-option-text">{{ $t("Search results") }}</div>
      </div>

      <div
        class="side-bar-option"
        :class="{ selected: album < 0 && page === 'albums' }"
        tabindex="0"
        :title="$t('Albums')"
        @click="goToPage('albums')"
      >
        <div class="side-bar-option-icon"><i class="fas fa-list"></i></div>
        <div class="side-bar-option-text">{{ $t("Albums") }}</div>
      </div>

      <div
        class="side-bar-option"
        :class="{ selected: album < 0 && page === 'upload' }"
        tabindex="0"
        :title="$t('Upload')"
        @click="goToPage('upload')"
      >
        <div class="side-bar-option-icon"><i class="fas fa-upload"></i></div>
        <div class="side-bar-option-text">{{ $t("Upload") }}</div>
      </div>

      <div
        class="side-bar-option"
        :class="{ selected: album < 0 && page === 'random' }"
        tabindex="0"
        :title="$t('Random')"
        @click="goToPage('random')"
      >
        <div class="side-bar-option-icon"><i class="fas fa-shuffle"></i></div>
        <div class="side-bar-option-text">{{ $t("Random") }}</div>
      </div>

      <div class="side-bar-separator"></div>

      <div
        v-for="a in albums"
        :key="a.id"
        class="side-bar-option"
        :class="{ selected: album == a.id }"
        tabindex="0"
        :title="a.name"
        @click="goToAlbum(a)"
      >
        <div class="side-bar-option-icon"><i class="fas fa-list-ol"></i></div>
        <div class="side-bar-option-text">{{ a.name }}</div>
      </div>
    </div>
  </div>
</template>

<script lang="ts">
import { AlbumsController } from "@/control/albums";
import { AppEvents } from "@/control/app-events";
import { AppStatus } from "@/control/app-status";
import { defineComponent } from "vue";
import { useVModel } from "../../utils/vmodel";

export default defineComponent({
  name: "SideBar",
  emits: ["update:display"],
  props: {
    display: Boolean,
  },
  setup(props) {
    return {
      displayStatus: useVModel(props, "display"),
    };
  },
  data: function () {
    return {
      page: AppStatus.CurrentPage,
      album: AppStatus.CurrentAlbum,
      layout: AppStatus.CurrentLayout,
      search: AppStatus.CurrentSearch,

      albums: [],
    };
  },
  methods: {
    close: function () {
      this.displayStatus = false;
    },

    updateStatus: function () {
      if (AppStatus.CurrentLayout !== "initial") {
        this.displayStatus = false;
      } else if (this.layout !== "initial") {
        this.displayStatus = true;
      }

      this.layout = AppStatus.CurrentLayout;

      this.page = AppStatus.CurrentPage;
      this.album = AppStatus.CurrentAlbum;

      this.search = AppStatus.CurrentSearch;
    },

    goToPage: function (p) {
      AppStatus.GoToPage(p);
    },

    goToSearch: function () {
      AppStatus.GoToSearch(this.search);
    },

    goToAlbum: function (a) {
      AppStatus.ClickOnAlbum(a.id, a.list.length > 0 ? a.list[0] : -1);
    },

    stopPropagationEvent: function (e) {
      e.stopPropagation();
    },

    updateAlbums: function () {
      this.albums = AlbumsController.GetAlbumsListCopy().sort((a, b) => {
        if (a.nameLowerCase < b.nameLowerCase) {
          return -1;
        } else if (a.nameLowerCase > b.nameLowerCase) {
          return 1;
        } else {
          return 1;
        }
      });
    },
  },
  mounted: function () {
    this.$options.statusUpdater = this.updateStatus.bind(this);

    AppEvents.AddEventListener(
      "app-status-update",
      this.$options.statusUpdater
    );

    this.$options.albumsUpdater = this.updateAlbums.bind(this);

    AppEvents.AddEventListener("albums-update", this.$options.albumsUpdater);

    this.updateStatus();
    this.updateAlbums();
  },
  beforeUnmount: function () {
    AppEvents.RemoveEventListener(
      "app-status-update",
      this.$options.statusUpdater
    );

    AppEvents.RemoveEventListener("albums-update", this.$options.albumsUpdater);
  },
});
</script>

<style>
.side-bar {
  position: absolute;
  top: 0;
  left: 0;
  width: 240px;
  max-width: 100%;
  height: 100%;
  background: #212121;
  display: flex;
  flex-direction: column;
  transition: left 0.2s;
  z-index: 10;
}

.side-bar.hidden {
  left: -300px;
}

.side-bar-header {
  width: 100%;
  height: 56px;
  display: flex;
  white-space: nowrap;
  flex-direction: row;
  align-items: center;
}

.side-bar-body {
  height: calc(100% - 56px);
  width: 100%;
  display: flex;
  flex-direction: column;
  overflow: auto;
}

.side-bar-option {
  width: 100%;
  cursor: pointer;

  display: flex;
  flex-direction: row;
  align-items: center;
}

.side-bar-option:hover {
  background: rgba(255, 255, 255, 0.1);
}

.side-bar-option.selected,
.side-bar-option.selected:hover {
  background: rgba(255, 255, 255, 0.2);
}

.side-bar-option-icon {
  width: 72px;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 18px;
  padding: 1rem;
}

.side-bar-option-text {
  width: calc(100% - 72px);
  padding-top: 1rem;
  padding-bottom: 1rem;
  padding-right: 1rem;
  font-size: 1.1rem;
  text-overflow: ellipsis;
  overflow: hidden;
}

.side-bar-separator {
  border-top: solid 1px rgba(255, 255, 255, 0.1);
  width: 100%;
  height: 1px;
  margin-top: 0.25rem;
  margin-bottom: 0.25rem;
}
</style>
