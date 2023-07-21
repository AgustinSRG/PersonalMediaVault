<template>
  <div class="side-bar" :class="{ hidden: !display }" @click="stopPropagationEvent" tabindex="-1" :role="initialLayout ? '' : 'dialog'" :aria-hidden="!display">
    <div class="side-bar-header">
      <div class="top-bar-logo-td">
        <button type="button" class="top-bar-button" :title="$t('Main menu')" @click="close">
          <i class="fas fa-bars"></i>
        </button>
        <img class="top-bar-logo-img" src="/img/icons/favicon.png" alt="PMV" />
        <span :title="getAppTitle()" class="top-bar-title">PMV</span>
      </div>
    </div>
    <div class="side-bar-body" tabindex="-1">
      <a class="side-bar-option" :class="{ selected: album < 0 && page === 'home' }" :title="$t('Home')" @click="goToPage('home', $event)" :href="getPageURL('home')" target="_blank" rel="noopener noreferrer">
        <div class="side-bar-option-icon"><i class="fas fa-home"></i></div>
        <div class="side-bar-option-text">{{ $t("Home") }}</div>
      </a>

      <a v-if="!!search" class="side-bar-option" :class="{ selected: album < 0 && page === 'search' }" :title="$t('Search results')" @click="goToSearch($event)" :href="getPageURL('search')" target="_blank" rel="noopener noreferrer">
        <div class="side-bar-option-icon"><i class="fas fa-search"></i></div>
        <div class="side-bar-option-text">{{ $t("Search results") }}</div>
      </a>

      <a class="side-bar-option" :class="{ selected: album < 0 && page === 'albums' }" :title="$t('Albums')" @click="goToPage('albums', $event)" :href="getPageURL('albums')" target="_blank" rel="noopener noreferrer">
        <div class="side-bar-option-icon"><i class="fas fa-list"></i></div>
        <div class="side-bar-option-text">{{ $t("Albums") }}</div>
      </a>

      <a v-if="canWrite" class="side-bar-option" :class="{ selected: album < 0 && page === 'upload' }" :title="$t('Upload')" @click="goToPage('upload', $event)" :href="getPageURL('upload')" target="_blank" rel="noopener noreferrer">
        <div class="side-bar-option-icon"><i class="fas fa-upload"></i></div>
        <div class="side-bar-option-text">{{ $t("Upload") }}</div>
      </a>

      <a class="side-bar-option" :class="{ selected: album < 0 && page === 'random' }" :title="$t('Random')" @click="goToPage('random', $event)" :href="getPageURL('random')" target="_blank" rel="noopener noreferrer">
        <div class="side-bar-option-icon"><i class="fas fa-shuffle"></i></div>
        <div class="side-bar-option-text">{{ $t("Random") }}</div>
      </a>

      <a class="side-bar-option" :class="{ selected: album < 0 && page === 'adv-search' }" :title="$t('Advanced search')" @click="goToPage('adv-search', $event)" :href="getPageURL('adv-search')" target="_blank" rel="noopener noreferrer">
        <div class="side-bar-option-icon"><i class="fas fa-search"></i></div>
        <div class="side-bar-option-text">{{ $t("Advanced search") }}</div>
      </a>

      <div class="side-bar-separator" v-if="albumsFavorite.length > 0"></div>

      <a v-for="a in albumsFavorite" :key="a.id" class="side-bar-option" :class="{ selected: album == a.id }" :title="a.name" @click="goToAlbum(a, $event)" :href="getAlbumURL(a.id)" target="_blank" rel="noopener noreferrer">
        <div class="side-bar-option-icon"><i class="fas fa-star"></i></div>
        <div class="side-bar-option-text">{{ a.name }}</div>
      </a>

      <div class="side-bar-separator"></div>

      <a v-for="a in albumsRest" :key="a.id" class="side-bar-option" :class="{ selected: album == a.id }" :title="a.name" @click="goToAlbum(a, $event)" :href="getAlbumURL(a.id)" target="_blank" rel="noopener noreferrer">
        <div class="side-bar-option-icon"><i class="fas fa-list-ol"></i></div>
        <div class="side-bar-option-text">{{ a.name }}</div>
      </a>
    </div>
  </div>
</template>

<script lang="ts">
import { AlbumsController } from "@/control/albums";
import { AppEvents } from "@/control/app-events";
import { AppPreferences } from "@/control/app-preferences";
import { AppStatus } from "@/control/app-status";
import { AuthController } from "@/control/auth";
import { GenerateURIQuery } from "@/utils/request";
import { defineComponent, nextTick } from "vue";
import { useVModel } from "../../utils/v-model";
import { FocusTrap } from "../../utils/focus-trap";

const MAX_ALBUMS_LIST_LENGTH_SIDEBAR = 10;

export default defineComponent({
    name: "SideBar",
    emits: ["update:display", "skip-to-content"],
    props: {
        display: Boolean,
        initialLayout: Boolean,
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

            canWrite: AuthController.CanWrite,

            albums: [],
            albumsFavorite: [],
            albumsRest: [],
        };
    },
    methods: {
        close: function () {
            this.displayStatus = false;
        },

        getAppTitle: function () {
            return AuthController.Title || this.$t("Personal Media Vault");
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

        goToPage: function (p, e) {
            if (e) {
                e.preventDefault();
            }
            AppStatus.GoToPageNoSplit(p);
            nextTick(() => {
                this.$emit("skip-to-content");
            });
        },

        goToSearch: function (e) {
            if (e) {
                e.preventDefault();
            }
            AppStatus.GoToSearch(this.search);
            nextTick(() => {
                this.$emit("skip-to-content");
            });
        },

        goToAlbum: function (a, e) {
            if (e) {
                e.preventDefault();
            }
            AppStatus.ClickOnAlbum(a.id);
            nextTick(() => {
                this.$emit("skip-to-content");
            });
        },

        getPageURL: function (page: string): string {
            return (
                window.location.protocol +
        "//" +
        window.location.host +
        window.location.pathname +
        GenerateURIQuery({
            page: page,
        })
            );
        },

        getAlbumURL: function (albumId: number): string {
            return (
                window.location.protocol +
        "//" +
        window.location.host +
        window.location.pathname +
        GenerateURIQuery({
            album: albumId + "",
        })
            );
        },

        stopPropagationEvent: function (e) {
            e.stopPropagation();
        },

        updateAlbums: function () {
            this.albums = AlbumsController.GetAlbumsListCopy().sort((a, b) => {
                const lruA = AppPreferences.AlbumPositionMap[a.id + ""] || 0;
                const lruB = AppPreferences.AlbumPositionMap[b.id + ""] || 0;
                if (lruA > lruB) {
                    return -1;
                } else if (lruA < lruB) {
                    return 1;
                } else if (a.nameLowerCase < b.nameLowerCase) {
                    return -1;
                } else if (a.nameLowerCase > b.nameLowerCase) {
                    return 1;
                } else {
                    return 0;
                }
            });
            const favIdList = AppPreferences.FavAlbums;
            const albumsFavorite = [];
            const albumsRest = [];
            this.albums.forEach((album) => {
                if (favIdList.includes(album.id + "")) {
                    albumsFavorite.push(album);
                } else {
                    albumsRest.push(album);
                }
            });
            this.albumsFavorite = albumsFavorite;
            this.albumsRest = albumsRest.slice(0, MAX_ALBUMS_LIST_LENGTH_SIDEBAR);
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
        },

        putAlbumFirst: function (albumId: number) {
            for (let i = 0; i < this.albumsFavorite.length; i++) {
                if (this.albumsFavorite[i].id === albumId) {
                    const albumEntry = this.albumsFavorite.splice(i, 1)[0];
                    this.albumsFavorite.unshift(albumEntry);
                    return;
                }
            }
            for (let i = 0; i < this.albumsRest.length; i++) {
                if (this.albumsRest[i].id === albumId) {
                    const albumEntry = this.albumsRest.splice(i, 1)[0];
                    this.albumsRest.unshift(albumEntry);
                    return;
                }
            }
            for (let i = 0; i < this.albums.length; i++) {
                if (this.albums[i].id === albumId) {
                    const albumEntry = this.albums[i];
                    this.albumsRest.unshift(albumEntry);
                    if (this.albumsRest.length > MAX_ALBUMS_LIST_LENGTH_SIDEBAR) {
                        this.albumsRest.pop();
                    }
                    return;
                }
            }
        },

        undoScroll: function () {
            const e = this.$el.querySelector(".side-bar-body");
            if (e) {
                e.scrollTop = 0;
            }
        },

        lostFocus: function () {
            if (!this.initialLayout) {
                this.close();
            }
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
        AppEvents.AddEventListener(
            "albums-fav-updated",
            this.$options.albumsUpdater
        );

        this.$options.albumGoTop = this.putAlbumFirst.bind(this);

        AppEvents.AddEventListener("album-sidebar-top", this.$options.albumGoTop);

        this.$options.authUpdateH = this.updateAuthInfo.bind(this);

        AppEvents.AddEventListener(
            "auth-status-changed",
            this.$options.authUpdateH
        );

        this.$options.focusTrap = new FocusTrap(
            this.$el,
            this.lostFocus.bind(this)
        );

        if (this.display) {
            this.$options.focusTrap.activate();
        }

        this.updateStatus();
        this.updateAlbums();
    },
    beforeUnmount: function () {
        AppEvents.RemoveEventListener(
            "app-status-update",
            this.$options.statusUpdater
        );

        AppEvents.RemoveEventListener("albums-update", this.$options.albumsUpdater);
        AppEvents.RemoveEventListener(
            "albums-fav-updated",
            this.$options.albumsUpdater
        );

        AppEvents.RemoveEventListener(
            "album-sidebar-top",
            this.$options.albumGoTop
        );

        AppEvents.RemoveEventListener(
            "auth-status-changed",
            this.$options.authUpdateH
        );

        if (this.$options.focusTrap) {
            this.$options.focusTrap.destroy();
        }
    },
    watch: {
        display: function () {
            if (this.display) {
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
    },
});
</script>
