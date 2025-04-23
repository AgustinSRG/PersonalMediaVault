<template>
    <div class="album-container" tabindex="-1" @keydown="onHeaderKeyDown">
        <div v-if="!loading && loadedAlbum" class="album-header">
            <div class="album-header-title">
                <div class="album-title" :title="albumName"><i class="fas fa-list-ol"></i> {{ albumName }}</div>
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

                    <button type="button" :title="$t('Favorite')" class="album-header-btn" :class="{ toggled: isFav }" @click="toggleFav">
                        <i class="fas fa-star"></i>
                    </button>

                    <button v-if="canWrite" type="button" :title="$t('Add media')" class="album-header-btn" @click="addMediaToAlbum">
                        <i class="fas fa-plus"></i>
                    </button>

                    <button v-if="canWrite" type="button" :title="$t('Rename')" class="album-header-btn" @click="renameAlbum">
                        <i class="fas fa-pencil-alt"></i>
                    </button>

                    <button v-if="canWrite" type="button" :title="$t('Thumbnail')" class="album-header-btn" @click="changeAlbumThumbnail">
                        <i class="fas fa-image"></i>
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
            tabindex="-1"
            @scroll.passive="onScroll"
            @keydown="onBodyKeyDown"
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
                    <ThumbImage v-if="item.thumbnail" :src="getThumbnail(item.thumbnail)"></ThumbImage>
                    <DurationIndicator
                        v-if="item.type === 2 || item.type === 3"
                        :type="item.type"
                        :duration="item.duration"
                        :small="true"
                    ></DurationIndicator>
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
                    <ThumbImage v-if="draggingItem.thumbnail" :src="getThumbnail(draggingItem.thumbnail)"></ThumbImage>
                    <DurationIndicator
                        v-if="draggingItem.type === 2 || draggingItem.type === 3"
                        :type="draggingItem.type"
                        :duration="draggingItem.duration"
                    ></DurationIndicator>
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
            :media-index="contextIndex"
            :album-length="albumListLength"
            :x="contextX"
            :y="contextY"
            @move-up="moveMediaUp"
            @move-down="moveMediaDown"
            @change-pos="changeMediaPos"
            @media-remove="removeMedia"
        ></AlbumContextMenu>
        <AlbumAddMediaModal v-if="displayAlbumAddMedia" v-model:display="displayAlbumAddMedia" :aid="albumId"></AlbumAddMediaModal>
        <AlbumRenameModal v-if="displayAlbumRename" v-model:display="displayAlbumRename"></AlbumRenameModal>
        <AlbumChangeThumbnailModal
            v-if="displayAlbumChangeThumbnail"
            v-model:display="displayAlbumChangeThumbnail"
        ></AlbumChangeThumbnailModal>
        <AlbumDeleteModal v-if="displayAlbumDelete" v-model:display="displayAlbumDelete"></AlbumDeleteModal>
        <AlbumGoToPosModal v-if="displayAlbumGoPos" v-model:display="displayAlbumGoPos"></AlbumGoToPosModal>
        <AlbumMovePosModal
            v-model:display="displayAlbumMovePos"
            :position-to-move="positionToMove"
            :album-list-length="albumListLength"
        ></AlbumMovePosModal>
        <LoadingOverlay v-if="loading"></LoadingOverlay>
    </div>
</template>

<script lang="ts">
import {
    AlbumsController,
    EVENT_NAME_ALBUMS_CHANGED,
    EVENT_NAME_CURRENT_ALBUM_LOADING,
    EVENT_NAME_CURRENT_ALBUM_MEDIA_POSITION_UPDATED,
    EVENT_NAME_CURRENT_ALBUM_UPDATED,
} from "@/control/albums";
import { AppEvents } from "@/control/app-events";
import { EVENT_NAME_FAVORITE_ALBUMS_UPDATED, albumAddFav, albumIsFavorite, albumRemoveFav } from "@/control/app-preferences";
import { AppStatus } from "@/control/app-status";
import { AuthController, EVENT_NAME_AUTH_CHANGED, EVENT_NAME_UNAUTHORIZED } from "@/control/auth";
import { generateURIQuery, getAssetURL } from "@/utils/api";
import { makeApiRequest } from "@asanrom/request-browser";
import { defineAsyncComponent, defineComponent, nextTick } from "vue";
import AlbumContextMenu from "./AlbumContextMenu.vue";
import LoadingOverlay from "./LoadingOverlay.vue";
import AlbumMovePosModal from "@/components/modals/AlbumMovePosModal.vue";
import { useVModel } from "@/utils/v-model";
import { BigListScroller } from "@/utils/big-list-scroller";
import { isTouchDevice } from "@/utils/touch";
import { PagesController } from "@/control/pages";
import { apiAlbumsRemoveMediaFromAlbum } from "@/api/api-albums";
import ThumbImage from "../utils/ThumbImage.vue";
import DurationIndicator from "../utils/DurationIndicator.vue";

const INITIAL_WINDOW_SIZE = 100;

const AlbumGoToPosModal = defineAsyncComponent({
    loader: () => import("@/components/modals/AlbumGoToPosModal.vue"),
    loadingComponent: LoadingOverlay,
    delay: 1000,
});

const AlbumRenameModal = defineAsyncComponent({
    loader: () => import("@/components/modals/AlbumRenameModal.vue"),
    loadingComponent: LoadingOverlay,
    delay: 1000,
});

const AlbumChangeThumbnailModal = defineAsyncComponent({
    loader: () => import("@/components/modals/AlbumChangeThumbnailModal.vue"),
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

export default defineComponent({
    name: "AlbumContainer",
    components: {
        AlbumContextMenu,
        LoadingOverlay,
        AlbumRenameModal,
        AlbumDeleteModal,
        AlbumMovePosModal,
        AlbumAddMediaModal,
        AlbumGoToPosModal,
        AlbumChangeThumbnailModal,
        ThumbImage,
        DurationIndicator,
    },
    props: {
        displayUpload: Boolean,
    },
    emits: ["update:displayUpload"],
    setup(props) {
        return {
            listScroller: null as BigListScroller,
            dragCheckInterval: null,
            checkContainerTimer: null,
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

            isFav: albumIsFavorite(AlbumsController.CurrentAlbum),

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
            displayAlbumChangeThumbnail: false,
            displayAlbumDelete: false,
            displayAlbumMovePos: false,
            displayAlbumGoPos: false,

            dragging: false,
            draggingPosition: -1,
            draggingItem: null as AlbumListItem,
            mouseX: 0,
            mouseY: 0,
            draggingOverPosition: -1,
            positionToMove: 0,
        };
    },
    mounted: function () {
        if (this.displayAlbumAddMedia) {
            this.displayAlbumAddMedia = false;
        }

        this.$listenOnAppEvent(EVENT_NAME_CURRENT_ALBUM_UPDATED, this.onAlbumUpdate.bind(this));

        this.$addKeyboardHandler(this.handleGlobalKey.bind(this), 10);

        this.listScroller = new BigListScroller(INITIAL_WINDOW_SIZE, {
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

        this.$listenOnAppEvent(EVENT_NAME_CURRENT_ALBUM_LOADING, this.onAlbumLoading.bind(this));

        this.$listenOnAppEvent(EVENT_NAME_CURRENT_ALBUM_MEDIA_POSITION_UPDATED, this.onAlbumPosUpdate.bind(this));

        this.$listenOnAppEvent(EVENT_NAME_AUTH_CHANGED, this.updateAuthInfo.bind(this));

        this.$listenOnAppEvent(EVENT_NAME_FAVORITE_ALBUMS_UPDATED, this.updateFav.bind(this));

        this.checkContainerTimer = setInterval(this.checkContainerHeight.bind(this), 1000);

        this.$listenOnDocumentEvent("mousemove", this.onDocumentMouseMove.bind(this));

        this.$listenOnDocumentEvent("mouseup", this.onDocumentMouseUp.bind(this));
    },
    beforeUnmount: function () {
        clearInterval(this.checkContainerTimer);

        if (this.dragCheckInterval) {
            clearInterval(this.dragCheckInterval);
            this.dragCheckInterval = null;
        }
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

            this.isFav = albumIsFavorite(AlbumsController.CurrentAlbum);

            this.updateAlbumList(!this.mustScroll);
        },

        updateAlbumList: function (preserveWindowScroll?: boolean) {
            this.currentMenuOpen = -1;

            const currentWindowPosition = this.listScroller.windowPosition;

            let currentScroll = 0;

            const conEl = this.$el.querySelector(".album-body");

            if (conEl) {
                currentScroll = conEl.scrollTop;
            }

            this.listScroller.reset();

            if (this.loadedAlbum) {
                let i = 0;
                this.listScroller.addElements(
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
                    this.listScroller.setWindowPosition(currentWindowPosition);
                    nextTick(() => {
                        const conEl = this.$el.querySelector(".album-body") as HTMLElement;

                        if (conEl) {
                            conEl.scrollTop = Math.min(currentScroll, conEl.scrollHeight - conEl.getBoundingClientRect().height);
                        }
                    });
                }
            }

            nextTick(() => {
                this.checkContainerHeight();
            });
        },

        onScroll: function (e: Event) {
            this.closeOptionsMenu();
            this.listScroller.checkElementScroll(e.target as HTMLElement);
        },

        onHeaderKeyDown: function (e: KeyboardEvent) {
            if (e.key === "ArrowUp" || e.key === "ArrowDown") {
                e.stopPropagation();

                const elem = this.$el.querySelector(".album-body");

                if (elem) {
                    elem.focus();
                }
            }
        },

        onBodyKeyDown: function (e: KeyboardEvent) {
            if (e.key === "ArrowUp" || e.key === "ArrowDown") {
                e.stopPropagation();
            }
        },

        checkContainerHeight: function () {
            const cont = this.$el.querySelector(".album-body");

            if (!cont) {
                return;
            }

            this.listScroller.checkElementScroll(cont);

            const el = this.$el.querySelector(".album-body-item");

            if (!el) {
                return;
            }

            const changed = this.listScroller.checkScrollContainerHeight(cont, el);

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
                PagesController.ShowSnackBar(this.$t("Album loop enabled"));
            } else {
                PagesController.ShowSnackBar(this.$t("Album loop disabled"));
            }
        },

        toggleRandom: function () {
            AlbumsController.ToggleRandom();
            if (AlbumsController.AlbumRandom) {
                PagesController.ShowSnackBar(this.$t("Album shuffle enabled"));
            } else {
                PagesController.ShowSnackBar(this.$t("Album shuffle disabled"));
            }
        },

        toggleFav: function () {
            if (this.isFav) {
                this.isFav = false;
                albumRemoveFav(AlbumsController.CurrentAlbum);
                PagesController.ShowSnackBar(this.$t("Album removed from favorites"));
            } else {
                this.isFav = true;
                albumAddFav(AlbumsController.CurrentAlbum);
                PagesController.ShowSnackBar(this.$t("Album added to favorites"));
            }
        },

        addMediaToAlbum: function () {
            this.displayAlbumAddMedia = true;
        },

        renameAlbum: function () {
            this.displayAlbumRename = true;
        },

        changeAlbumThumbnail: function () {
            this.displayAlbumChangeThumbnail = true;
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

        getThumbnail(thumb: string) {
            return getAssetURL(thumb);
        },

        clickMedia: function (item: AlbumListItem, e?: MouseEvent) {
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
                generateURIQuery({
                    media: item.id + "",
                    album: this.albumId + "",
                })
            );
        },

        showOptions: function (item: { pos: number }, i: number, event: MouseEvent) {
            event.preventDefault();
            event.stopPropagation();

            if (this.contextShown && this.currentMenuOpen === item.pos) {
                this.currentMenuOpen = -1;
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

            if (mustScroll && this.currentPos >= 0) {
                nextTick(() => {
                    this.listScroller.moveWindowToElement(this.currentPos);
                    nextTick(() => {
                        this.scrollToSelected();
                    });
                });
            }
        },

        stopPropagationEvent: function (e: Event) {
            e.stopPropagation();
        },

        moveMediaUp: function (i: number) {
            if (i > 0) {
                AlbumsController.MoveCurrentAlbumOrder(i, i - 1, this.$t);
            }
        },

        moveMediaDown: function (i: number) {
            if (i < this.albumListLength - 1) {
                AlbumsController.MoveCurrentAlbumOrder(i, i + 1, this.$t);
            }
        },

        changeMediaPos: function (i: number) {
            this.positionToMove = i;
            this.displayAlbumMovePos = true;
        },

        removeMedia: function (i: number) {
            const completeAlbumList = this.listScroller.list;
            const media = completeAlbumList[i];
            if (!media) {
                return;
            }
            const albumId = this.albumId;
            makeApiRequest(apiAlbumsRemoveMediaFromAlbum(albumId, media.id))
                .onSuccess(() => {
                    PagesController.ShowSnackBar(this.$t("Successfully removed from album"));
                    AlbumsController.OnChangedAlbum(albumId, true);
                    AppEvents.Emit(EVENT_NAME_ALBUMS_CHANGED);
                })
                .onRequestError((err, handleErr) => {
                    handleErr(err, {
                        unauthorized: () => {
                            AppEvents.Emit(EVENT_NAME_UNAUTHORIZED);
                        },
                        accessDenied: () => {
                            PagesController.ShowSnackBar(this.$t("Error") + ": " + this.$t("Access denied"));
                            AuthController.CheckAuthStatusSilent();
                        },
                        notFound: () => {
                            PagesController.ShowSnackBar(this.$t("Error") + ": " + this.$t("Not found"));
                            AlbumsController.OnChangedAlbum(albumId);
                        },
                        serverError: () => {
                            PagesController.ShowSnackBar(this.$t("Error") + ": " + this.$t("Internal server error"));
                        },
                        networkError: () => {
                            PagesController.ShowSnackBar(this.$t("Error") + ": " + this.$t("Could not connect to the server"));
                        },
                    });
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
        },

        updateFav: function () {
            this.isFav = albumIsFavorite(AlbumsController.CurrentAlbum);
        },

        handleGlobalKey: function (event: KeyboardEvent): boolean {
            if (AuthController.Locked || AppStatus.CurrentLayout !== "album" || !event.key || event.ctrlKey) {
                return false;
            }

            const completeAlbumList = this.listScroller.list;

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
                case "g":
                case "G":
                    this.displayAlbumGoPos = true;
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
            if (this.dragCheckInterval) {
                clearInterval(this.dragCheckInterval);
                this.dragCheckInterval = null;
            }
            this.dragCheckInterval = setInterval(this.onDragCheck.bind(this), 40);
        },

        onDocumentMouseMove: function (e: MouseEvent) {
            this.mouseX = e.pageX;
            this.mouseY = e.pageY;
        },

        onDocumentMouseUp: function () {
            if (!this.dragging) {
                return;
            }

            if (this.dragCheckInterval) {
                clearInterval(this.dragCheckInterval);
                this.dragCheckInterval = null;
            }

            // Check drop position

            this.checkDropPosition();

            // Move element if needed

            if (this.draggingOverPosition >= 0 && this.draggingPosition >= 0) {
                const initialPosition = this.draggingPosition;
                const finalPosition =
                    this.draggingOverPosition > this.draggingPosition ? this.draggingOverPosition - 1 : this.draggingOverPosition;

                if (initialPosition !== finalPosition) {
                    AlbumsController.MoveCurrentAlbumOrder(initialPosition, finalPosition, this.$t);
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
});
</script>
