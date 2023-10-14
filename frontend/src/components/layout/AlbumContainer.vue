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
        <div
            v-show="!loading && loadedAlbum"
            class="album-body"
            :class="{ 'is-dragging': dragging }"
            @scroll.passive="onScroll"
            tabindex="-1"
        >
            <a
                v-for="item in albumList"
                :key="item.pos"
                :href="getMediaURL(item)"
                target="_blank"
                rel="noopener noreferrer"
                class="album-body-item"
                :class="{
                    current: item.pos === currentPos,
                    dragging: item.pos === draggingPosition,
                    'dragging-over': item.pos === draggingOverPosition,
                }"
                :title="item.title || $t('Untitled')"
                @click="clickMedia(item, $event)"
                @dragstart="onDrag(item, $event)"
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
            <div
                v-if="dragging && albumList.length > 0 && draggingOverPosition > albumList[albumList.length - 1].pos"
                class="dragging-padding-bottom"
            ></div>
        </div>
        <div
            v-if="dragging && draggingItem"
            class="album-dragging-helper"
            :style="{ top: mouseY - 65 + 'px', left: mouseX - 65 + 'px' }"
            @click="stopPropagationEvent"
        >
            <div
                :href="getMediaURL(draggingItem)"
                target="_blank"
                rel="noopener noreferrer"
                class="album-body-item"
                :class="{ current: draggingItem.pos === currentPos }"
                :title="draggingItem.title || $t('Untitled')"
            >
                <div class="album-body-item-thumbnail">
                    <div v-if="!draggingItem.thumbnail" class="no-thumb">
                        <i v-if="draggingItem.type === 1" class="fas fa-image"></i>
                        <i v-else-if="draggingItem.type === 2" class="fas fa-video"></i>
                        <i v-else-if="draggingItem.type === 3" class="fas fa-headphones"></i>
                        <i v-else class="fas fa-ban"></i>
                    </div>
                    <img
                        v-if="draggingItem.thumbnail"
                        :src="getThumbnail(draggingItem.thumbnail)"
                        :alt="draggingItem.title || $t('Untitled')"
                        loading="lazy"
                    />
                    <div class="album-body-item-thumb-tag" v-if="draggingItem.type === 2 || draggingItem.type === 3">
                        {{ renderTime(draggingItem.duration) }}
                    </div>
                    <div class="album-body-item-thumb-pos">
                        {{ renderPos(draggingItem.pos) }}
                    </div>
                </div>

                <div class="album-body-item-title">
                    {{ draggingItem.title || $t("Untitled") }}
                </div>

                <button v-if="canWrite" type="button" :title="$t('Options')" class="album-body-btn" disabled>
                    <i class="fas fa-bars"></i>
                </button>
            </div>
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

const INITIAL_WINDOW_SIZE = 100;

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

interface AlbumListItem {
    pos: number;
    id: number;
    type: number;
    title: string;
    thumbnail: string;
    duration: number;
}

import { useVModel } from "@/utils/v-model";
import { BigListScroller } from "@/utils/big-list-scroller";
import { isTouchDevice } from "@/utils/touch";

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

            albumList: [] as AlbumListItem[],

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

            dragging: false,
            draggingPosition: -1,
            draggingItem: null as AlbumListItem,
            mouseX: 0,
            mouseY: 0,
            draggingOverPosition: -1,
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

            let currentScroll = 0;

            const conEl = this.$el.querySelector(".album-body");

            if (conEl) {
                currentScroll = conEl.scrollTop;
            }

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
                    nextTick(() => {
                        const conEl = this.$el.querySelector(".album-body");

                        if (conEl) {
                            conEl.scrollTop = currentScroll;
                        }
                    });
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

        renderPos: function (p: number) {
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

        clickMedia: function (item: AlbumListItem, e: MouseEvent) {
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

        showOptions: function (item: { pos: number }, i: number, event: MouseEvent) {
            event.preventDefault();
            event.stopPropagation();

            if (this.contextShown && this.currentMenuOpen === item.pos) {
                this.currentMenuOpen = "";
                this.contextShown = false;
            } else {
                this.currentMenuOpen = item.pos;
                this.contextShown = true;
                this.contextIndex = i;

                const targetRect = (event.target as HTMLElement).getBoundingClientRect();

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

        stopPropagationEvent: function (e: Event) {
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

        clickOnEnter: function (event: KeyboardEvent) {
            if (event.key === "Enter") {
                event.preventDefault();
                event.stopPropagation();
                (event.target as HTMLElement).click();
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
                case "L":
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

        onDrag: function (item: AlbumListItem, e: DragEvent) {
            e.preventDefault();

            if (isTouchDevice()) {
                return; // Disabled for touch devices
            }

            if (!this.canWrite) {
                return; // Cannot alter the album
            }

            this.dragging = true;
            this.draggingItem = item;
            this.draggingPosition = item.pos;
            this.draggingOverPosition = item.pos;
            if (this._handles.dragCheckInterval) {
                clearInterval(this._handles.dragCheckInterval);
                this._handles.dragCheckInterval = null;
            }
            this._handles.dragCheckInterval = setInterval(this.onDragCheck.bind(this), 40);
        },

        onDocumentMouseMove: function (e: MouseEvent) {
            this.mouseX = e.pageX;
            this.mouseY = e.pageY;
        },

        onDocumentMouseUp: function () {
            if (!this.dragging) {
                return;
            }

            if (this._handles.dragCheckInterval) {
                clearInterval(this._handles.dragCheckInterval);
                this._handles.dragCheckInterval = null;
            }

            // Check drop position

            this.checkDropPosition();

            // Move element if needed

            if (this.draggingOverPosition >= 0 && this.draggingPosition >= 0) {
                const initialPosition = this.draggingPosition;
                const finalPosition =
                    this.draggingOverPosition > this.draggingPosition ? this.draggingOverPosition - 1 : this.draggingOverPosition;

                if (initialPosition !== finalPosition) {
                    AlbumsController.MoveCurrentAlbumOrder(initialPosition, finalPosition);
                }
            }

            // Reset dragging
            this.dragging = false;
            this.draggingPosition = -1;
            this.draggingOverPosition = -1;
        },

        onDragCheck: function () {
            const con = this.$el.querySelector(".album-body");

            if (!con) {
                return;
            }

            const conBounds = con.getBoundingClientRect();

            if (this.mouseX >= conBounds.left - 420) {
                // Auto scroll

                const relTop = (this.mouseY - conBounds.top) / (conBounds.height || 1);
                const scrollStep = Math.floor(conBounds.height / 10);

                if (relTop <= 0.1) {
                    con.scrollTop = Math.max(0, con.scrollTop - scrollStep);
                } else if (relTop >= 0.9) {
                    con.scrollTop = Math.min(con.scrollHeight - conBounds.height, con.scrollTop + scrollStep);
                }
            }

            // Check drop position
            this.checkDropPosition();
        },

        checkDropPosition: function () {
            const con = this.$el.querySelector(".album-body");

            if (!con) {
                return;
            }

            const conBounds = con.getBoundingClientRect();

            if (this.albumList.length === 0) {
                return;
            }

            const item = con.querySelector(".album-body-item:not(.dragging)");

            if (!item) {
                return;
            }

            if (this.mouseX < conBounds.left - 420) {
                this.draggingOverPosition = -1;
                return;
            }

            const itemHeight = item.getBoundingClientRect().height;

            const firstPos = this.albumList[0].pos;
            const lastPos = this.albumList[this.albumList.length - 1].pos;
            const containerScrollTop = con.scrollTop;

            this.draggingOverPosition = Math.min(
                lastPos + 1,
                Math.max(firstPos, firstPos + Math.round((this.mouseY - conBounds.top + containerScrollTop) / itemHeight)),
            );
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

        this._handles.checkContainerTimer = setInterval(this.checkContainerHeight.bind(this), 1000);

        this._handles.onDocumentMouseMoveH = this.onDocumentMouseMove.bind(this);
        document.addEventListener("mousemove", this._handles.onDocumentMouseMoveH);

        this._handles.onDocumentMouseUpH = this.onDocumentMouseUp.bind(this);
        document.addEventListener("mouseup", this._handles.onDocumentMouseUpH);
    },
    beforeUnmount: function () {
        AppEvents.RemoveEventListener("current-album-update", this._handles.albumUpdateH);
        AppEvents.RemoveEventListener("current-album-loading", this._handles.loadingH);

        AppEvents.RemoveEventListener("album-pos-update", this._handles.posUpdateH);

        AppEvents.RemoveEventListener("auth-status-changed", this._handles.authUpdateH);

        AppEvents.RemoveEventListener("albums-fav-updated", this._handles.favUpdateH);

        KeyboardManager.RemoveHandler(this._handles.handleGlobalKeyH);

        clearInterval(this._handles.checkContainerTimer);

        document.removeEventListener("mousemove", this._handles.onDocumentMouseMoveH);
        document.removeEventListener("mouseup", this._handles.onDocumentMouseUpH);

        if (this._handles.dragCheckInterval) {
            clearInterval(this._handles.dragCheckInterval);
            this._handles.dragCheckInterval = null;
        }
    },
});
</script>
