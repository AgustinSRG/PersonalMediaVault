<template>
    <div class="player-album-container" tabindex="-1">
        <div v-if="!loading && loadedAlbum" class="album-header">
            <div class="album-header-title">
                <div class="album-title"><i class="fas fa-list-ol"></i> {{ albumName }}</div>
                <button type="button" :title="$t('Close')" class="album-header-btn album-close-btn" @click="close">
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
                </div>
                <div class="album-post-text">{{ renderPos(currentPos) }} / {{ albumListLength }}</div>
            </div>
        </div>

        <div v-show="!loading && loadedAlbum" class="album-body" @scroll.passive="onScroll" tabindex="-1">
            <div
                v-for="item in albumList"
                :key="item.pos"
                class="album-body-item"
                :class="{ current: item.pos === currentPos }"
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
                    <img v-if="item.thumbnail" :src="getThumbnail(item.thumbnail)" :alt="item.title || $t('Untitled')" loading="lazy" />
                    <div class="album-body-item-thumb-tag" v-if="item.type === 2 || item.type === 3">
                        {{ renderTime(item.duration) }}
                    </div>
                    <div class="album-body-item-thumb-pos">
                        {{ renderPos(item.pos) }}
                    </div>
                </div>

                <div class="album-body-item-title no-btn">
                    {{ item.title || $t("Untitled") }}
                </div>
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
import { BigListScroller } from "@/utils/big-list-scroller";
import { GetAssetURL } from "@/utils/request";
import { renderTimeSeconds } from "@/utils/time";
import { defineComponent, nextTick } from "vue";

const INITIAL_WINDOW_SIZE = 100;

export default defineComponent({
    name: "PlayerAlbumFullScreen",
    emits: ["close"],
    data: function () {
        return {
            albumId: AlbumsController.CurrentAlbum,
            albumName: AlbumsController.CurrentAlbumData ? AlbumsController.CurrentAlbumData.name : "",
            albumListLength: AlbumsController.CurrentAlbumData ? AlbumsController.CurrentAlbumData.list.length : 0,
            loadedAlbum: !!AlbumsController.CurrentAlbumData,

            albumList: [],

            loading: AlbumsController.CurrentAlbumLoading,

            currentPos: AlbumsController.CurrentAlbumPos,

            loop: false,
            random: false,
        };
    },
    methods: {
        onAlbumUpdate: function () {
            this.albumId = AlbumsController.CurrentAlbum;
            this.loadedAlbum = !!AlbumsController.CurrentAlbumData;
            this.albumName = AlbumsController.CurrentAlbumData ? AlbumsController.CurrentAlbumData.name : "";
            this.albumListLength = AlbumsController.CurrentAlbumData ? AlbumsController.CurrentAlbumData.list.length : 0;
            this.updateAlbumList();
        },

        updateAlbumList: function () {
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

        toggleLoop: function () {
            AlbumsController.ToggleLoop();
        },

        toggleRandom: function () {
            AlbumsController.ToggleRandom();
        },

        renderPos: function (p) {
            if (p < 0) {
                return "?";
            } else {
                return "" + (p + 1);
            }
        },

        renderTime: function (s: number): string {
            return renderTimeSeconds(s);
        },

        getThumbnail(thumb: string) {
            return GetAssetURL(thumb);
        },

        clickMedia: function (item) {
            AppStatus.ClickOnMedia(item.id, false);
            this.$emit("close");
        },

        close: function () {
            this.$emit("close");
        },

        onAlbumPosUpdate: function () {
            this.loop = AlbumsController.AlbumLoop;
            this.random = AlbumsController.AlbumRandom;
            this.currentPos = AlbumsController.CurrentAlbumPos;
            this._handles.listScroller.moveWindowToElement(this.currentPos);
            nextTick(() => {
                this.scrollToSelected();
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

        stopPropagationEvent: function (e) {
            e.stopPropagation();
        },

        clickOnEnter: function (event) {
            if (event.key === "Enter") {
                event.preventDefault();
                event.stopPropagation();
                event.target.click();
            }
        },

        onScroll: function (e) {
            this._handles.listScroller.checkElementScroll(e.target);
        },

        checkContainerHeight: function () {
            const cont = this.$el.querySelector(".album-body");

            if (!cont) {
                return;
            }

            const el = this.$el.querySelector(".album-body-item");

            if (!el) {
                return;
            }

            const changed = this._handles.listScroller.checkScrollContainerHeight(cont, el);

            if (changed) {
                this.onAlbumPosUpdate();
            }
        },

        autoFocus: function () {
            nextTick(() => {
                const focusEl = this.$el.querySelector(".album-body");
                if (focusEl) {
                    focusEl.focus();
                    this.checkContainerHeight();
                }
            });
        },
    },
    mounted: function () {
        this._handles = Object.create(null);
        this._handles.albumUpdateH = this.onAlbumUpdate.bind(this);
        AppEvents.AddEventListener("current-album-update", this._handles.albumUpdateH);

        this._handles.listScroller = new BigListScroller(INITIAL_WINDOW_SIZE, {
            get: () => {
                return this.albumList;
            },
            set: (l) => {
                this.albumList = l;
            },
        });

        this.updateAlbumList();

        this.onAlbumPosUpdate();

        this._handles.loadingH = this.onAlbumLoading.bind(this);
        AppEvents.AddEventListener("current-album-loading", this._handles.loadingH);

        this._handles.posUpdateH = this.onAlbumPosUpdate.bind(this);
        AppEvents.AddEventListener("album-pos-update", this._handles.posUpdateH);

        this._handles.checkContainerTimer = setInterval(this.checkContainerHeight.bind(this), 1000);

        this.autoFocus();
    },
    beforeUnmount: function () {
        AppEvents.RemoveEventListener("current-album-update", this._handles.albumUpdateH);
        AppEvents.RemoveEventListener("current-album-loading", this._handles.loadingH);

        AppEvents.RemoveEventListener("album-pos-update", this._handles.posUpdateH);

        clearInterval(this._handles.checkContainerTimer);
    },
});
</script>
