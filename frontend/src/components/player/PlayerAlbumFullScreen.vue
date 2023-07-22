<template>
    <div class="player-album-container" :class="{ expanded: expanded }" tabindex="-1">
        <div v-if="!loading && albumData" class="album-header">
            <div class="album-header-title">
                <div class="album-title"><i class="fas fa-list-ol"></i> {{ albumData.name }}</div>
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
                <div class="album-post-text">{{ renderPos(currentPos) }} / {{ albumList.length }}</div>
            </div>
        </div>

        <div v-show="!loading && albumData" class="album-body">
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
                    <img v-if="item.thumbnail" :src="getThumbnail(item.thumbnail)" :alt="item.title || $t('Untitled')" loading="lazy" />
                    <div class="album-body-item-thumb-tag" v-if="item.type === 2 || item.type === 3">
                        {{ renderTime(item.duration) }}
                    </div>
                    <div class="album-body-item-thumb-pos">
                        {{ renderPos(i) }}
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
import { clone } from "@/utils/objects";
import { GetAssetURL } from "@/utils/request";
import { renderTimeSeconds } from "@/utils/time";
import { defineComponent, nextTick } from "vue";

export default defineComponent({
    name: "PlayerAlbumFullScreen",
    emits: ["close"],
    props: {
        expanded: Boolean,
    },
    data: function () {
        return {
            albumId: AlbumsController.CurrentAlbum,
            albumData: AlbumsController.CurrentAlbumData,

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
            this.albumData = AlbumsController.CurrentAlbumData;
            this.updateAlbumsList();
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

        updateAlbumsList: function () {
            const prefix = "" + Date.now() + "-";
            let i = 0;
            if (this.albumData) {
                this.albumList = this.albumData.list.map((a) => {
                    const o = clone(a);
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
            this.$emit("close");
        },

        close: function () {
            this.$emit("close");
        },

        onAlbumPosUpdate: function () {
            this.loop = AlbumsController.AlbumLoop;
            this.random = AlbumsController.AlbumRandom;
            this.currentPos = AlbumsController.CurrentAlbumPos;
            nextTick(() => {
                this.scrollToSelected();
            });
        },

        scrollToSelected: function () {
            const itemHeight = 130;
            const element = this.$el.querySelector(".album-body");

            if (!element) {
                return;
            }

            const scrollHeight = element.scrollHeight;
            const height = element.getBoundingClientRect().height;

            const itemTop = this.currentPos * itemHeight;

            const expectedTop = height / 2 - itemHeight / 2;

            const scroll = Math.max(0, Math.min(scrollHeight - height, itemTop - expectedTop));

            element.scrollTop = scroll;
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
    },
    mounted: function () {
        this.$options.albumUpdateH = this.onAlbumUpdate.bind(this);
        AppEvents.AddEventListener("current-album-update", this.$options.albumUpdateH);

        this.onAlbumPosUpdate();

        this.updateAlbumsList();

        this.$options.loadingH = this.onAlbumLoading.bind(this);
        AppEvents.AddEventListener("current-album-loading", this.$options.loadingH);

        this.$options.posUpdateH = this.onAlbumPosUpdate.bind(this);
        AppEvents.AddEventListener("album-pos-update", this.$options.posUpdateH);
    },
    beforeUnmount: function () {
        AppEvents.RemoveEventListener("current-album-update", this.$options.albumUpdateH);
        AppEvents.RemoveEventListener("current-album-loading", this.$options.loadingH);

        AppEvents.RemoveEventListener("album-pos-update", this.$options.posUpdateH);
    },
    watch: {
        expanded: function () {
            if (this.expanded) {
                nextTick(() => {
                    this.$el.focus();
                });
                this.scrollToSelected();
            }
        },
    },
});
</script>
