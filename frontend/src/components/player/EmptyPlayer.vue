<template>
    <div
        class="empty-player"
        :class="{
            'player-min': min,
            'full-screen': fullscreen,
        }"
        @dblclick="toggleFullScreen"
    >
        <div class="player-loader" v-if="status === 'loading' || (status === 'none' && albumLoading)">
            <div class="player-lds-ring">
                <div></div>
                <div></div>
                <div></div>
                <div></div>
            </div>
        </div>

        <div class="player-error-container" v-if="status === '404'">
            <div class="player-info-icon"><i class="fas fa-ban"></i></div>
            <div class="player-error">{{ $t("Media asset does not exist or was removed from the vault") }}</div>
        </div>

        <div class="player-error-container" v-if="status === 'none' && !albumLoading">
            <div class="player-info-icon"><i class="fas fa-list-ol"></i></div>
            <div class="player-info">{{ $t("The album is empty") }}</div>
            <div class="player-info">{{ $t("Browse the vault in order to add media to it") }}</div>
        </div>

        <div class="player-controls" @dblclick="stopPropagationEvent" @mouseleave="leaveControls">
            <div class="player-controls-left">
                <button
                    v-if="!!next || !!prev || pagePrev || pageNext"
                    :disabled="!prev && !pagePrev"
                    type="button"
                    :title="$t('Previous')"
                    class="player-btn"
                    @click="goPrev"
                    @mouseenter="enterTooltip('prev')"
                    @mouseleave="leaveTooltip('prev')"
                >
                    <i class="fas fa-backward-step"></i>
                </button>

                <button disabled type="button" :title="$t('Play')" class="player-btn">
                    <i class="fas fa-play"></i>
                </button>

                <button
                    v-if="!!next || !!prev || pagePrev || pageNext"
                    :disabled="!next && !pageNext"
                    type="button"
                    :title="$t('Next')"
                    class="player-btn"
                    @click="goNext"
                    @mouseenter="enterTooltip('next')"
                    @mouseleave="leaveTooltip('next')"
                >
                    <i class="fas fa-forward-step"></i>
                </button>
            </div>

            <div class="player-controls-right">
                <button
                    v-if="!fullscreen"
                    type="button"
                    :title="$t('Full screen')"
                    class="player-btn player-expand-btn"
                    @click="toggleFullScreen"
                    @mouseenter="enterTooltip('full-screen')"
                    @mouseleave="leaveTooltip('full-screen')"
                >
                    <i class="fas fa-expand"></i>
                </button>
                <button
                    v-if="fullscreen"
                    type="button"
                    :title="$t('Exit full screen')"
                    class="player-btn player-expand-btn"
                    @click="toggleFullScreen"
                    @mouseenter="enterTooltip('full-screen-exit')"
                    @mouseleave="leaveTooltip('full-screen-exit')"
                >
                    <i class="fas fa-compress"></i>
                </button>
            </div>
        </div>

        <div v-if="prev && helpTooltip === 'prev'" class="player-tooltip player-help-tip-left">
            <PlayerMediaChangePreview :media="prev" :next="false"></PlayerMediaChangePreview>
        </div>

        <div v-if="next && helpTooltip === 'next'" class="player-tooltip player-help-tip-left">
            <PlayerMediaChangePreview :media="next" :next="true"></PlayerMediaChangePreview>
        </div>

        <div v-if="helpTooltip === 'full-screen'" class="player-tooltip player-help-tip-right">
            {{ $t("Full screen") }}
        </div>
        <div v-if="helpTooltip === 'full-screen-exit'" class="player-tooltip player-help-tip-right">
            {{ $t("Exit full screen") }}
        </div>

        <PlayerTopBar
            :mid="mid"
            :metadata="null"
            :shown="true"
            :fullscreen="fullscreen"
            v-model:expanded="expandedTitle"
            v-model:albumExpanded="expandedAlbum"
            :inAlbum="inAlbum"
        ></PlayerTopBar>
    </div>
</template>

<script lang="ts">
import { defineComponent } from "vue";

import PlayerMediaChangePreview from "./PlayerMediaChangePreview.vue";
import PlayerTopBar from "./PlayerTopBar.vue";

import { openFullscreen, closeFullscreen } from "../../utils/full-screen";
import { useVModel } from "../../utils/v-model";
import { KeyboardManager } from "@/control/keyboard";
import { AppStatus } from "@/control/app-status";
import { AuthController } from "@/control/auth";

export default defineComponent({
    components: {
        PlayerMediaChangePreview,
        PlayerTopBar,
    },
    name: "EmptyPlayer",
    emits: ["go-next", "go-prev", "update:fullscreen"],
    props: {
        mid: Number,
        status: String,

        albumLoading: Boolean,

        fullscreen: Boolean,

        rTick: Number,

        next: Object,
        prev: Object,
        inAlbum: Boolean,

        pageNext: Boolean,
        pagePrev: Boolean,

        canWrite: Boolean,

        min: Boolean,
    },
    setup(props) {
        return {
            fullScreenState: useVModel(props, "fullscreen"),
        };
    },
    data: function () {
        return {
            helpTooltip: "",

            expandedTitle: false,
            expandedAlbum: false,
        };
    },
    methods: {
        enterTooltip: function (t: string) {
            this.helpTooltip = t;
        },

        leaveTooltip: function (t: string) {
            if (t === this.helpTooltip) {
                this.helpTooltip = "";
            }
        },

        goNext: function () {
            if (this.next || this.pageNext) {
                this.$emit("go-next");
            }
        },

        goPrev: function () {
            if (this.prev || this.pagePrev) {
                this.$emit("go-prev");
            }
        },

        mouseLeavePlayer: function () {
            this.helpTooltip = "";
        },

        leaveControls: function () {
            this.helpTooltip = "";
        },

        toggleFullScreen: function () {
            if (!this.fullscreen) {
                openFullscreen();
            } else {
                closeFullscreen();
            }
            this.fullScreenState = !this.fullScreenState;
        },
        onExitFullScreen: function () {
            if (!document.fullscreenElement) {
                this.fullScreenState = false;
            }
        },
        stopPropagationEvent: function (e) {
            e.stopPropagation();
        },

        onKeyPress: function (event: KeyboardEvent): boolean {
            if (AuthController.Locked || !AppStatus.IsPlayerVisible() || !event.key || event.ctrlKey) {
                return false;
            }
            let caught = true;
            switch (event.key) {
            case "F":
            case "f":
                if (event.altKey || event.shiftKey) {
                    caught = false;
                } else {
                    this.toggleFullScreen();
                }
                break;
            case "ArrowLeft":
                if (this.prev || this.pagePrev) {
                    this.goPrev();
                } else {
                    caught = false;
                }
                break;
            case "ArrowRight":
                if (this.next || this.pageNext) {
                    this.goNext();
                } else {
                    caught = false;
                }
                break;
            default:
                caught = false;
            }

            return caught;
        },
    },
    mounted: function () {
        this._handles = Object.create(null);
        // Load player preferences

        this._handles.keyHandler = this.onKeyPress.bind(this);
        KeyboardManager.AddHandler(this._handles.keyHandler, 100);

        this._handles.exitFullScreenListener = this.onExitFullScreen.bind(this);
        document.addEventListener("fullscreenchange", this._handles.exitFullScreenListener);
        document.addEventListener("webkitfullscreenchange", this._handles.exitFullScreenListener);
        document.addEventListener("mozfullscreenchange", this._handles.exitFullScreenListener);
        document.addEventListener("MSFullscreenChange", this._handles.exitFullScreenListener);
    },
    beforeUnmount: function () {
        document.removeEventListener("fullscreenchange", this._handles.exitFullScreenListener);
        document.removeEventListener("webkitfullscreenchange", this._handles.exitFullScreenListener);
        document.removeEventListener("mozfullscreenchange", this._handles.exitFullScreenListener);
        document.removeEventListener("MSFullscreenChange", this._handles.exitFullScreenListener);
        KeyboardManager.RemoveHandler(this._handles.keyHandler);
    },
    watch: {
        rTick: function () {
            this.expandedTitle = false;
        },
    },
});
</script>
