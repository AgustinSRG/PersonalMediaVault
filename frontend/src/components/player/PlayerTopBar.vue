<template>
    <div
        class="player-top-bar"
        :class="{
            hidden: !shown,
            'with-album': inAlbum,
            'album-expand': albumExpanded,
            expanded: expanded && !albumExpanded,
            expanding: expanding,
            contracting: clickedContract,
        }"
        tabindex="-1"
        @click="clickTopBar"
        @dblclick="stopPropagationEvent"
        @mousedown="stopPropagationEvent"
        @touchstart="stopPropagationEvent"
        @contextmenu="stopPropagationEvent"
        @keydown="onKeyDown"
        @animationstart="onAnimationStart"
        @animationend="onAnimationEnd"
    >
        <div v-if="!albumExpanded" class="player-title-container">
            <div class="player-title-left">
                <button type="button" :title="$t('View Album')" class="player-btn" @click="expandAlbum">
                    <i class="fas fa-list-ol"></i>
                </button>
            </div>
            <div class="player-title">
                <div v-if="metadata">{{ metadata.title }}</div>
            </div>
            <div class="player-title-right">
                <button v-if="metadata && !expanded" type="button" :title="$t('Expand')" class="player-btn" @click="expandTitle">
                    <i class="fas fa-chevron-down"></i>
                </button>

                <button v-if="metadata && expanded" type="button" :title="$t('Close')" class="player-btn" @click="closeTitle">
                    <i class="fas fa-chevron-up"></i>
                </button>
            </div>
        </div>

        <PlayerAlbumFullScreen v-if="albumExpanded" @close="closeAlbum"></PlayerAlbumFullScreen>
        <PlayerMediaEditor v-if="expanded" @changed="onEditDone" @open-description="openDescription"></PlayerMediaEditor>
    </div>
</template>

<script lang="ts">
import { MediaController } from "@/control/media";
import type { PropType } from "vue";
import { defineAsyncComponent, defineComponent, nextTick } from "vue";
import { useVModel } from "../../utils/v-model";
import PlayerAlbumFullScreen from "./PlayerAlbumFullScreen.vue";
import { AuthController } from "@/control/auth";
import type { MediaData } from "@/api/models";
import { ExitPreventer } from "@/control/exit-prevent";

const PlayerMediaEditor = defineAsyncComponent({
    loader: () => import("@/components/player/editor/PlayerMediaEditor.vue"),
});

export default defineComponent({
    name: "PlayerTopBar",
    components: {
        PlayerAlbumFullScreen,
        PlayerMediaEditor,
    },
    props: {
        mid: Number,
        metadata: Object as PropType<MediaData>,

        inAlbum: Boolean,

        shown: Boolean,
        fullscreen: Boolean,
        expanded: Boolean,
        albumExpanded: Boolean,
    },
    emits: ["update:expanded", "update:albumExpanded", "click-player", "open-description"],
    setup(props) {
        return {
            expandedState: useVModel(props, "expanded"),
            albumExpandedState: useVModel(props, "albumExpanded"),
        };
    },
    data: function () {
        return {
            dirty: false,
            expanding: false,

            clickedContract: false,
        };
    },
    watch: {
        fullscreen: function () {
            this.albumExpandedState = false;
        },

        expanded: function () {
            this.clickedContract = !this.expanded;
            if (this.expanded) {
                nextTick(() => {
                    const el = this.$el.querySelector(".player-media-editor");
                    if (el) {
                        el.focus();
                    }
                });
            }
            if (this.dirty) {
                this.dirty = false;
                setTimeout(() => {
                    MediaController.Load();
                }, 100);
            }
        },

        albumExpanded: function () {
            this.clickedContract = false;
            if (this.albumExpanded) {
                nextTick(() => {
                    const el = this.$el.querySelector(".player-album-container");
                    if (el) {
                        el.focus();
                    }
                });
            }
        },
    },
    mounted: function () {
        this.$addKeyboardHandler(this.handleGlobalKey.bind(this));
    },
    methods: {
        clickTopBar: function (e) {
            e.stopPropagation();
            this.$emit("click-player");
        },

        expandTitle: function () {
            this.albumExpandedState = false;
            this.expandedState = true;
        },

        onEditDone: function () {
            this.dirty = true;
        },

        closeTitle: function () {
            ExitPreventer.TryExit(() => {
                this.expandedState = false;
            });
        },

        expandAlbum: function () {
            this.albumExpandedState = true;
            this.expandedState = false;
        },

        closeAlbum: function () {
            this.albumExpandedState = false;
        },

        close: function () {
            this.closeTitle();
            this.closeAlbum();
        },

        onKeyDown: function (e: KeyboardEvent) {
            if (!this.expanded && !this.albumExpanded) {
                return;
            }
            e.stopPropagation();
            if (e.key === "Escape") {
                e.preventDefault();
                this.close();
            }
        },

        handleGlobalKey: function (event: KeyboardEvent): boolean {
            if (AuthController.Locked || !event.key || event.ctrlKey) {
                return false;
            }

            if (event.key.toUpperCase() === "E") {
                this.expandTitle();
                return true;
            }

            return false;
        },

        onAnimationStart: function (event: AnimationEvent) {
            if (event.animationName == "player-top-bar-expand" && this.expanded) {
                this.expanding = true;
            }
        },

        onAnimationEnd: function (event: AnimationEvent) {
            if (event.animationName == "player-top-bar-expand" && this.expanded) {
                this.expanding = false;
                const autoFocus = this.$el.querySelector(".auto-focus");
                if (autoFocus) {
                    autoFocus.focus();
                }
            }
        },

        openDescription: function () {
            this.$emit("open-description");
        },
    },
});
</script>
