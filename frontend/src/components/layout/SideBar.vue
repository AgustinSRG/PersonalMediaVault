<template>
    <div
        class="side-bar"
        :class="{ hidden: !display }"
        tabindex="-1"
        :role="initialLayout ? '' : 'dialog'"
        @click="stopPropagationEvent"
        @keydown="keyDownHandle"
    >
        <div class="side-bar-header">
            <div class="top-bar-logo-td">
                <button type="button" class="top-bar-button" :title="$t('Main menu')" @click="close">
                    <i class="fas fa-bars"></i>
                </button>
                <img class="top-bar-logo-img" src="/img/icons/favicon.png" :alt="getAppLogoText(customLogo)" />
                <span :title="getAppTitle(customTitle)" class="top-bar-title">{{ getAppLogoText(customLogo) }}</span>
            </div>
        </div>
        <div class="side-bar-body" tabindex="-1">
            <a
                class="side-bar-option"
                :class="{ selected: album < 0 && page === 'home' }"
                :title="$t('Home')"
                :href="getPageURL('home')"
                target="_blank"
                rel="noopener noreferrer"
                @click="goToPage('home', $event)"
            >
                <div class="side-bar-option-icon"><i class="fas fa-home"></i></div>
                <div class="side-bar-option-text">{{ $t("Home") }}</div>
            </a>

            <a
                class="side-bar-option"
                :class="{ selected: album < 0 && page === 'albums' }"
                :title="$t('Albums')"
                :href="getPageURL('albums', cachedAlbumsSearchParams)"
                target="_blank"
                rel="noopener noreferrer"
                @click="goToPage('albums', $event, cachedAlbumsSearchParams)"
            >
                <div class="side-bar-option-icon"><i class="fas fa-list"></i></div>
                <div class="side-bar-option-text">{{ $t("Albums") }}</div>
            </a>

            <a
                class="side-bar-option"
                :class="{ selected: album < 0 && page === 'media' }"
                :title="$t('Media')"
                :href="getPageURL('media')"
                target="_blank"
                rel="noopener noreferrer"
                @click="goToPage('media', $event)"
            >
                <div class="side-bar-option-icon"><i class="fas fa-photo-film"></i></div>
                <div class="side-bar-option-text">{{ $t("Media") }}</div>
            </a>

            <a
                class="side-bar-option"
                :class="{ selected: album < 0 && page === 'adv-search' }"
                :title="$t('Find media')"
                :href="getPageURL('adv-search')"
                target="_blank"
                rel="noopener noreferrer"
                @click="goToPage('adv-search', $event)"
            >
                <div class="side-bar-option-icon"><i class="fas fa-search"></i></div>
                <div class="side-bar-option-text">{{ $t("Find media") }}</div>
            </a>

            <a
                class="side-bar-option"
                :class="{ selected: album < 0 && page === 'random' }"
                :title="$t('Random')"
                :href="getPageURL('random')"
                target="_blank"
                rel="noopener noreferrer"
                @click="goToPage('random', $event)"
            >
                <div class="side-bar-option-icon"><i class="fas fa-shuffle"></i></div>
                <div class="side-bar-option-text">{{ $t("Random") }}</div>
            </a>

            <a
                v-if="canWrite"
                class="side-bar-option"
                :class="{ selected: album < 0 && page === 'upload' }"
                :title="$t('Upload')"
                :href="getPageURL('upload')"
                target="_blank"
                rel="noopener noreferrer"
                @click="goToPage('upload', $event)"
            >
                <div class="side-bar-option-icon"><i class="fas fa-upload"></i></div>
                <div class="side-bar-option-text">{{ $t("Upload") }}</div>
            </a>

            <div v-if="albumsFavorite.length > 0" class="side-bar-separator"></div>

            <a
                v-for="a in albumsFavorite"
                :key="a.id"
                class="side-bar-option"
                :class="{ selected: album == a.id }"
                :title="a.name"
                :href="getAlbumURL(a.id)"
                target="_blank"
                rel="noopener noreferrer"
                @click="goToAlbum(a, $event)"
            >
                <div class="side-bar-option-icon"><i class="fas fa-star"></i></div>
                <div class="side-bar-option-text">{{ a.name }}</div>
            </a>

            <div class="side-bar-separator"></div>

            <a
                v-for="a in albumsRest"
                :key="a.id"
                class="side-bar-option"
                :class="{ selected: album == a.id }"
                :title="a.name"
                :href="getAlbumURL(a.id)"
                target="_blank"
                rel="noopener noreferrer"
                @click="goToAlbum(a, $event)"
            >
                <div class="side-bar-option-icon"><i class="fas fa-list-ol"></i></div>
                <div class="side-bar-option-text">{{ a.name }}</div>
            </a>
        </div>
    </div>
</template>

<script lang="ts">
import type { AlbumListItemMinExt } from "@/control/albums";
import { AlbumsController, EVENT_NAME_ALBUMS_LIST_UPDATE } from "@/control/albums";
import {
    EVENT_NAME_ALBUM_SIDEBAR_TOP,
    EVENT_NAME_FAVORITE_ALBUMS_UPDATED,
    getAlbumFavoriteList,
    getAlbumsOrderMap,
} from "@/control/app-preferences";
import type { AppStatusPage } from "@/control/app-status";
import { AppStatus, EVENT_NAME_APP_STATUS_CHANGED } from "@/control/app-status";
import { AuthController, EVENT_NAME_AUTH_CHANGED } from "@/control/auth";
import { getFrontendUrl } from "@/utils/api";
import { defineComponent, nextTick } from "vue";
import { useVModel } from "../../utils/v-model";
import { FocusTrap } from "../../utils/focus-trap";

const MAX_ALBUMS_LIST_LENGTH_SIDEBAR = 10;

export default defineComponent({
    name: "SideBar",
    props: {
        display: Boolean,
        initialLayout: Boolean,
    },
    emits: ["update:display", "skip-to-content"],
    setup(props) {
        return {
            focusTrap: null as FocusTrap,
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

            albums: [] as AlbumListItemMinExt[],
            albumsFavorite: [] as AlbumListItemMinExt[],
            albumsRest: [] as AlbumListItemMinExt[],

            cachedAlbumsSearchParams: AppStatus.CurrentPage === "albums" ? AppStatus.SearchParams : "",

            customTitle: AuthController.Title,
            customLogo: AuthController.Logo,
        };
    },
    watch: {
        display: function () {
            if (this.display) {
                this.focusTrap.activate();
                if (!this.initialLayout) {
                    nextTick(() => {
                        this.$el.focus();
                    });
                }
            } else {
                this.focusTrap.deactivate();
            }
        },
    },
    mounted: function () {
        this.$listenOnAppEvent(EVENT_NAME_APP_STATUS_CHANGED, this.updateStatus.bind(this));

        this.$listenOnAppEvent(EVENT_NAME_ALBUMS_LIST_UPDATE, this.updateAlbums.bind(this));
        this.$listenOnAppEvent(EVENT_NAME_FAVORITE_ALBUMS_UPDATED, this.updateAlbums.bind(this));

        this.$listenOnAppEvent(EVENT_NAME_ALBUM_SIDEBAR_TOP, this.putAlbumFirst.bind(this));

        this.$listenOnAppEvent(EVENT_NAME_AUTH_CHANGED, this.updateAuthInfo.bind(this));

        this.focusTrap = new FocusTrap(this.$el, this.lostFocus.bind(this));

        if (this.display) {
            this.focusTrap.activate();
        }

        this.updateStatus();
        this.updateAlbums();
    },
    beforeUnmount: function () {
        this.focusTrap.destroy();
    },
    methods: {
        close: function () {
            this.displayStatus = false;
        },

        getAppTitle: function (customTitle: string) {
            return customTitle || this.$t("Personal Media Vault");
        },

        getAppLogoText: function (customLogo: string) {
            return customLogo || "PMV";
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

            if (AppStatus.CurrentPage === "albums") {
                this.cachedAlbumsSearchParams = AppStatus.SearchParams;
            }
        },

        goToPage: function (p: AppStatusPage, e: Event, searchParams?: string) {
            if (e) {
                e.preventDefault();
            }
            AppStatus.GoToPageConditionalSplit(p, searchParams);
            nextTick(() => {
                this.$emit("skip-to-content");
            });
            if (window.innerWidth < 1000) {
                this.close();
            }
        },

        goToAlbum: function (a, e) {
            if (e) {
                e.preventDefault();
            }
            AppStatus.ClickOnAlbum(a.id);
            nextTick(() => {
                this.$emit("skip-to-content");
            });
            if (window.innerWidth < 1000) {
                this.close();
            }
        },

        getPageURL: function (page: AppStatusPage, searchParams?: string): string {
            return getFrontendUrl({
                page: page,
                sp: searchParams || null,
            });
        },

        getAlbumURL: function (albumId: number): string {
            return getFrontendUrl({
                album: albumId,
            });
        },

        updateAlbums: function () {
            const albumsOrderMap = getAlbumsOrderMap();
            this.albums = AlbumsController.GetAlbumsListMin().sort((a, b) => {
                const lruA = albumsOrderMap[a.id + ""] || 0;
                const lruB = albumsOrderMap[b.id + ""] || 0;
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
            const favIdList = getAlbumFavoriteList();
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

        updateAuthInfo: function () {
            this.canWrite = AuthController.CanWrite;
            this.customTitle = AuthController.Title;
            this.customLogo = AuthController.Logo;
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

        keyDownHandle: function (e: KeyboardEvent) {
            if (!this.initialLayout && e.key === "Escape") {
                e.stopPropagation();
                this.close();
            }
        },
    },
});
</script>
