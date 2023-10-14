<template>
    <div class="album-container" tabindex="-1">
        <div v-if="!loading && loadedAlbum" class="album-header">
            <div class="album-header-title">
                <div class="album-title"><i class="fas fa-list-ol"></i> {{ albumName }}</div>
                <button type="button" :title="$t('Close')" class="album-header-btn album-close-btn" @click="closePage">
                    <i class="fas fa-times"></i>
                </button>
            </div>
            <div class="album-header-controls">
                <div class="album-buttons">
                    <button type="button" :title="$t('Loop')" class="album-header-btn" :class="{ toggled: loop }" @click="toggleLoop">
                        <i class="fas fa-repeat"></i>
                    </button>

                    <button type="button" :title="$t('Random')" class="album-header-btn" :class="{ toggled: random }" @click="toggleRandom">
                        <i class="fas fa-shuffle"></i>
                    </button>

                    <button v-if="canWrite" type="button" :title="$t('Add media')" class="album-header-btn" @click="addMediaToAlbum">
                        <i class="fas fa-plus"></i>
                    </button>

                    <button v-if="canWrite" type="button" :title="$t('Rename')" class="album-header-btn" @click="renameAlbum">
                        <i class="fas fa-pencil-alt"></i>
                    </button>

                    <button type="button" :title="$t('Favorite')" class="album-header-btn" :class="{ toggled: isFav }" @click="toggleFav">
                        <i class="fas fa-star"></i>
                    </button>

                    <button v-if="canWrite" type="button" :title="$t('Delete')" class="album-header-btn" @click="deleteAlbum">
                        <i class="fas fa-trash-alt"></i>
                    </button>
                </div>
                <div class="album-post-text">{{ renderPos(currentPos) }} / {{ albumListLength }}</div>
            </div>
        </div>
        <div v-show="!loading && loadedAlbum" class="album-body" @scroll.passive="onScroll" tabindex="-1">
            <a
                v-for="item in albumList"
                :key="item.pos"
                :href="getMediaURL(item)"
                target="_blank"
                rel="noopener noreferrer"
                class="album-body-item"
                :class="{ current: item.pos === currentPos }"
                :title="item.title || $t('Untitled')"
                @click="clickMedia(item, $event)"
            >
                <div class="album-body-item-thumbnail">
                    <div v-if="!item.thumbnail" class="no-thumb">
                        <i v-if="item.type === 1" class="fas fa-image"></i>
                        <i v-else-if="item.type === 2" class="fas fa-video"></i>
                        <i v-else-if="item.type === 3" class="fas fa-headphones"></i>
                        <i v-else class="fas fa-ban"></i>
                    </div>
                    <img v-if="item.thumbnail" :src="getThumbnail(item.thumbnail)" :alt="item.title || $t('Untitled')" loading="lazy" />
                    <div class="album-body-item-thumb-tag" v-if="item.type === 2 || item.type === 3">
                        {{ renderTime(item.duration) }}
                    </div>
                    <div class="album-body-item-thumb-pos">
                        {{ renderPos(item.pos) }}
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
                    @click="showOptions(item, item.pos, $event)"
                    @mousedown="stopPropagationEvent"
                    @touchstart="stopPropagationEvent"
                >
                    <i class="fas fa-bars"></i>
                </button>
            </a>
        </div>
        <AlbumContextMenu
            v-model:shown="contextShown"
            :mediaIndex="contextIndex"
            :albumLength="albumListLength"
            :x="contextX"
            :y="contextY"
            @move-up="moveMediaUp"
            @move-down="moveMediaDown"
            @change-pos="changeMediaPos"
            @media-remove="removeMedia"
        ></AlbumContextMenu>
        <AlbumAddMediaModal v-if="displayAlbumAddMedia" v-model:display="displayAlbumAddMedia" :aid="albumId"></AlbumAddMediaModal>
        <AlbumRenameModal v-if="displayAlbumRename" v-model:display="displayAlbumRename"></AlbumRenameModal>
        <AlbumDeleteModal v-if="displayAlbumDelete" v-model:display="displayAlbumDelete"></AlbumDeleteModal>
        <AlbumMovePosModal ref="movePosModal" v-model:display="displayAlbumMovePos"></AlbumMovePosModal>
        <LoadingOverlay v-if="loading"></LoadingOverlay>
    </div>
</template>

<script lang="ts">
import { AlbumsAPI } from "@/api/api-albums";
import { AlbumsController } from "@/control/albums";
import { AppEvents } from "@/control/app-events";
import { AppPreferences } from "@/control/app-preferences";
import { AppStatus } from "@/control/app-status";
import { AuthController } from "@/control/auth";
import { KeyboardManager } from "@/control/keyboard";
import { GenerateURIQuery, GetAssetURL, Request } from "@/utils/request";
import { renderTimeSeconds } from "@/utils/time";
import { defineAsyncComponent, defineComponent, nextTick } from "vue";

import AlbumContextMenu from "./AlbumContextMenu.vue";
import LoadingOverlay from "./LoadingOverlay.vue";

import AlbumMovePosModal from "@/components/modals/AlbumMovePosModal.vue";

const INITIAL_WINDOW_SIZE = 20;

const AlbumRenameModal = defineAsyncComponent({
    loader: () => import("@/components/modals/AlbumRenameModal.vue"),
    loadingComponent: LoadingOverlay,
    delay: 1000,
});

const AlbumDeleteModal = defineAsyncComponent({
    loader: () => import("@/components/modals/AlbumDeleteModal.vue"),
    loadingComponent: LoadingOverlay,
    delay: 1000,
});

const AlbumAddMediaModal = defineAsyncComponent({
    loader: () => import("@/components/modals/AlbumAddMediaModal.vue"),
    loadingComponent: LoadingOverlay,
    delay: 1000,
});

import { useVModel } from "@/utils/v-model";
import { BigListScroller } from "@/utils/big-list-scroller";

export default defineComponent({
    name: "AlbumContainer",
    emits: ["update:displayUpload"],
    components: {
        AlbumContextMenu,
        LoadingOverlay,
        AlbumRenameModal,
        AlbumDeleteModal,
        AlbumMovePosModal,
        AlbumAddMediaModal,
    },
    props: {
        displayUpload: Boolean,
    },
    setup(props) {
        return {
            displayAlbumAddMedia: useVModel(props, "displayUpload"),
        };
    },
    data: function () {
        return {
            albumId: AlbumsController.CurrentAlbum,
            albumName: AlbumsController.CurrentAlbumData ? AlbumsController.CurrentAlbumData.name : "",
            albumListLength: AlbumsController.CurrentAlbumData ? AlbumsController.CurrentAlbumData.list.length : 0,
            loadedAlbum: !!AlbumsController.CurrentAlbumData,

            albumList: [],

            isFav: AppPreferences.FavAlbums.includes(AlbumsController.CurrentAlbum + ""),

            loading: AlbumsController.CurrentAlbumLoading,

            currentPos: AlbumsController.CurrentAlbumPos,
            currentPosMedia: AlbumsController.CurrentAlbumPos >= 0 ? AppStatus.CurrentMedia : -1,
            mustScroll: true,

            currentMenuOpen: -1,
            contextShown: false,
            contextIndex: -1,
            contextX: 0,
            contextY: 0,

            loop: false,
            random: false,

            canWrite: AuthController.CanWrite,

            displayAlbumRename: false,
            displayAlbumDelete: false,
            displayAlbumMovePos: false,
        };
    },
    methods: {
        onAlbumUpdate: function () {
            if (this.albumId !== AlbumsController.CurrentAlbum) {
                this.mustScroll = true;
            }

            this.albumId = AlbumsController.CurrentAlbum;
            this.loadedAlbum = !!AlbumsController.CurrentAlbumData;

            this.albumName = AlbumsController.CurrentAlbumData ? AlbumsController.CurrentAlbumData.name : "";
            this.albumListLength = AlbumsController.CurrentAlbumData ? AlbumsController.CurrentAlbumData.list.length : 0;

            this.isFav = AppPreferences.FavAlbums.includes(AlbumsController.CurrentAlbum + "");

            this.updateAlbumList(!this.mustScroll);
        },

        updateAlbumList: function (preserveWindowScroll?: boolean) {
            this.currentMenuOpen = -1;

            const centerPosition = this._handles.listScroller.getCenterPosition();

            this._handles.listScroller.reset();

            if (this.loadedAlbum) {
                let i = 0;
                this._handles.listScroller.addElements(
                    AlbumsController.CurrentAlbumData.list.map((m) => {
                        return {
                            pos: i++,
                            id: m.id,
                            type: m.type,
                            title: m.title,
                            thumbnail: m.thumbnail,
                            duration: m.duration,
                        };
                    }),
                );

                if (preserveWindowScroll) {
                    this._handles.listScroller.moveWindowToElement(centerPosition);
                }
            }

            nextTick(() => {
                this.checkContainerHeight();
            });
        },

        onScroll: function (e) {
            this.closeOptionsMenu();
            this._handles.listScroller.checkElementScroll(e.target);
        },

        checkContainerHeight: function () {
            const cont = this.$el.querySelector(".album-body");

            if (!cont) {
                return;
            }

            this._handles.listScroller.checkElementScroll(cont);

            const el = this.$el.querySelector(".album-body-item");

            if (!el) {
                return;
            }

            const changed = this._handles.listScroller.checkScrollContainerHeight(cont, el);

            if (changed) {
                this.mustScroll = true;
                this.onAlbumPosUpdate();
            }
        },

        closeOptionsMenu: function () {
            if (this.contextShown) {
                this.contextShown = false;
            }
        },

        onAlbumLoading: function (l: boolean) {
            if (l) {
                if (this.albumId !== AlbumsController.CurrentAlbum) {
                    this.loading = true;
                }
            } else {
                this.loading = false;
            }
        },

        closePage: function () {
            AppStatus.CloseAlbum();
        },

        toggleLoop: function () {
            AlbumsController.ToggleLoop();
            if (AlbumsController.AlbumLoop) {
                AppEvents.Emit("snack", this.$t("Album loop enabled"));
            } else {
                AppEvents.Emit("snack", this.$t("Album loop disabled"));
            }
        },

        toggleRandom: function () {
            AlbumsController.ToggleRandom();
            if (AlbumsController.AlbumRandom) {
                AppEvents.Emit("snack", this.$t("Album shuffle enabled"));
            } else {
                AppEvents.Emit("snack", this.$t("Album shuffle disabled"));
            }
        },

        toggleFav: function () {
            if (this.isFav) {
                AppPreferences.albumRemoveFav(AlbumsController.CurrentAlbum + "");
                AppEvents.Emit("snack", this.$t("Album removed from favorites"));
            } else {
                AppPreferences.albumAddFav(AlbumsController.CurrentAlbum + "");
                AppEvents.Emit("snack", this.$t("Album added to favorites"));
            }
        },

        addMediaToAlbum: function () {
            this.displayAlbumAddMedia = true;
        },

        renameAlbum: function () {
            this.displayAlbumRename = true;
        },

        renderPos: function (p) {
            if (p < 0) {
                return "?";
            } else {
                return "" + (p + 1);
            }
        },

        deleteAlbum: function () {
            this.displayAlbumDelete = true;
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

        clickMedia: function (item, e) {
            if (e) {
                e.preventDefault();
            }
            AppStatus.ClickOnMedia(item.id, false);
        },

        getMediaURL: function (item) {
            return (
                window.location.protocol +
                "//" +
                window.location.host +
                window.location.pathname +
                GenerateURIQuery({
                    media: item.id + "",
                    album: this.albumId + "",
                })
            );
        },

        showOptions: function (item: { pos: number }, i: number, event) {
            event.preventDefault();
            event.stopPropagation();

            if (this.contextShown && this.currentMenuOpen === item.pos) {
                this.currentMenuOpen = "";
                this.contextShown = false;
            } else {
                this.currentMenuOpen = item.pos;
                this.contextShown = true;
                this.contextIndex = i;

                const targetRect = event.target.getBoundingClientRect();

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

            const newPosMedia = AlbumsController.CurrentAlbumPos >= 0 ? AppStatus.CurrentMedia : -1;

            let mustScroll = false;

            if (this.mustScroll || this.currentPosMedia !== newPosMedia) {
                this.mustScroll = false;
                mustScroll = true;
            }

            this.currentPos = AlbumsController.CurrentAlbumPos;
            this.currentPosMedia = newPosMedia;

            if (mustScroll) {
                this._handles.listScroller.moveWindowToElement(this.currentPos);
                nextTick(() => {
                    this.scrollToSelected();
                });
            }
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
            if (i < this.albumListLength - 1) {
                AlbumsController.MoveCurrentAlbumOrder(i, i + 1);
            }
        },

        changeMediaPos: function (i: number) {
            this.$refs.movePosModal.show({
                pos: i,
                callback: (newPos: number) => {
                    if (isNaN(newPos) || !isFinite(newPos)) {
                        return;
                    }
                    newPos = Math.min(newPos, this.albumListLength - 1);
                    newPos = Math.max(0, newPos);
                    if (newPos === i) {
                        return;
                    }
                    AlbumsController.MoveCurrentAlbumOrder(i, newPos);
                },
            });
        },

        removeMedia: function (i: number) {
            const completeAlbumList = this._handles.listScroller.list;
            const media = completeAlbumList[i];
            if (!media) {
                return;
            }
            const albumId = this.albumId;
            Request.Do(AlbumsAPI.RemoveMediaFromAlbum(albumId, media.id))
                .onSuccess(() => {
                    AppEvents.Emit("snack", this.$t("Successfully removed from album"));
                    AlbumsController.OnChangedAlbum(albumId, true);
                    AppEvents.Emit("albums-list-change");
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
            const cont = this.$el.querySelector(".album-body");

            if (!cont) {
                return;
            }

            const contBounds = cont.getBoundingClientRect();

            const el = this.$el.querySelector(".album-body-item.current");

            if (!el) {
                return;
            }

            const elBounds = el.getBoundingClientRect();
            const elTopRel = elBounds.top - contBounds.top + cont.scrollTop;

            const expectedTop = contBounds.height / 2 - elBounds.height / 2;

            const scroll = Math.max(0, Math.min(cont.scrollHeight - contBounds.height, elTopRel - expectedTop));

            cont.scrollTop = scroll;
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
            if (this._handles.sortable) {
                this._handles.sortable.option("disabled", !this.canWrite);
            }
        },

        updateFav: function () {
            this.isFav = AppPreferences.FavAlbums.includes(AlbumsController.CurrentAlbum + "");
        },

        handleGlobalKey: function (event: KeyboardEvent): boolean {
            if (AuthController.Locked || AppStatus.CurrentLayout !== "album" || !event.key || event.ctrlKey) {
                return false;
            }

            const completeAlbumList = this._handles.listScroller.list;

            switch (event.key) {
                case "l":
                case "l":
                    this.toggleLoop();
                    return true;
                case "r":
                case "R":
                    this.toggleRandom();
                    return true;

                case "q":
                case "Q":
                    this.closePage();
                    return true;
                case "f":
                case "F":
                    this.toggleFav();
                    return true;
                case "Home":
                    if (completeAlbumList.length > 0) {
                        this.clickMedia(completeAlbumList[0]);
                    }
                    return true;
                case "End":
                    if (completeAlbumList.length > 0) {
                        this.clickMedia(completeAlbumList[completeAlbumList.length - 1]);
                    }
                    return true;
                default:
                    return false;
            }
        },
    },
    mounted: function () {
        if (this.displayAlbumAddMedia) {
            this.displayAlbumAddMedia = false;
        }
        this._handles = Object.create(null);
        this._handles.albumUpdateH = this.onAlbumUpdate.bind(this);
        AppEvents.AddEventListener("current-album-update", this._handles.albumUpdateH);

        this._handles.handleGlobalKeyH = this.handleGlobalKey.bind(this);
        KeyboardManager.AddHandler(this._handles.handleGlobalKeyH, 10);

        this._handles.listScroller = new BigListScroller(INITIAL_WINDOW_SIZE, {
            get: () => {
                return this.albumList;
            },
            set: (l) => {
                this.albumList = l;
            },
        });

        this.updateAlbumList();

        this.mustScroll = true;

        this.onAlbumPosUpdate();

        this._handles.loadingH = this.onAlbumLoading.bind(this);
        AppEvents.AddEventListener("current-album-loading", this._handles.loadingH);

        this._handles.posUpdateH = this.onAlbumPosUpdate.bind(this);
        AppEvents.AddEventListener("album-pos-update", this._handles.posUpdateH);

        this._handles.authUpdateH = this.updateAuthInfo.bind(this);

        AppEvents.AddEventListener("auth-status-changed", this._handles.authUpdateH);

        this._handles.favUpdateH = this.updateFav.bind(this);
        AppEvents.AddEventListener("albums-fav-updated", this._handles.favUpdateH);

        // Sortable
        /*if (!isTouchDevice()) {
            const element = this.$el.querySelector(".album-body");
            this._handles.sortable = Sortable.create(element, {
                onUpdate: this.onUpdateSortable.bind(this),
                disabled: !this.canWrite,
                scroll: element,
                forceAutoscrollFallback: true,
                scrollSensitivity: 100,
            });
        }*/

        this._handles.checkContainerTimer = setInterval(this.checkContainerHeight.bind(this), 1000);
    },
    beforeUnmount: function () {
        AppEvents.RemoveEventListener("current-album-update", this._handles.albumUpdateH);
        AppEvents.RemoveEventListener("current-album-loading", this._handles.loadingH);

        AppEvents.RemoveEventListener("album-pos-update", this._handles.posUpdateH);

        AppEvents.RemoveEventListener("auth-status-changed", this._handles.authUpdateH);

        AppEvents.RemoveEventListener("albums-fav-updated", this._handles.favUpdateH);

        KeyboardManager.RemoveHandler(this._handles.handleGlobalKeyH);

        // Sortable
        /*if (this._handles.sortable) {
            this._handles.sortable.destroy();
            this._handles.sortable = null;
        }*/

        clearInterval(this._handles.checkContainerTimer);
    },
});
</script>
