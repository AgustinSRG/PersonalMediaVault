<template>
    <div
        class="empty-player"
        :class="{
            'player-min': min,
            'full-screen': fullscreen,
        }"
        @dblclick="toggleFullScreen"
    >
        <div v-if="status === 'loading' || (status === 'none' && albumLoading)" class="player-loader">
            <div class="player-lds-ring">
                <div></div>
                <div></div>
                <div></div>
                <div></div>
            </div>
        </div>

        <div v-if="status === '404'" class="player-error-container">
            <div class="player-info-icon"><i class="fas fa-ban"></i></div>
            <div class="player-error">{{ $t("Media asset does not exist or was removed from the vault") }}</div>
        </div>

        <div v-if="status === 'none' && !albumLoading" class="player-error-container">
            <div class="player-info-icon"><i class="fas fa-list-ol"></i></div>
            <div class="player-info">{{ $t("The album is empty") }}</div>
            <div class="player-info">{{ $t("Browse the vault in order to add media to it") }}</div>
        </div>

        <div class="player-controls" @dblclick="stopPropagationEvent" @mouseleave="leaveControls" @contextmenu="stopPropagationEvent">
            <div class="player-controls-left">
                <button
                    v-if="!!next || !!prev || pagePrev || pageNext"
                    :disabled="!prev && !pagePrev"
                    type="button"
                    :title="$t('Previous')"
                    class="player-btn player-btn-action-prev"
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
                    class="player-btn player-btn-action-next"
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

        <div v-if="!prev && pagePrev && helpTooltip === 'prev'" class="player-tooltip player-help-tip-left">
            {{ $t("Previous") }}
        </div>

        <div v-if="!next && pageNext && helpTooltip === 'next'" class="player-tooltip player-help-tip-left">
            {{ $t("Next") }}
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
            v-model:expanded="expandedTitle"
            v-model:album-expanded="expandedAlbum"
            :mid="mid"
            :metadata="null"
            :shown="true"
            :fullscreen="fullscreen"
            :in-album="inAlbum"
        ></PlayerTopBar>
    </div>
</template>

<script lang="ts">
import type { PropType } from "vue";
import { defineComponent } from "vue";

import PlayerMediaChangePreview from "./PlayerMediaChangePreview.vue";
import PlayerTopBar from "./PlayerTopBar.vue";

import { openFullscreen, closeFullscreen } from "../../utils/full-screen";
import { useVModel } from "../../utils/v-model";
import { AppStatus } from "@/control/app-status";
import { AuthController } from "@/control/auth";
import { isTouchDevice } from "@/utils/touch";
import type { MediaListItem } from "@/api/models";

export default defineComponent({
    name: "EmptyPlayer",
    components: {
        PlayerMediaChangePreview,
        PlayerTopBar,
    },
    props: {
        mid: Number,
        status: String,

        albumLoading: Boolean,

        fullscreen: Boolean,

        rTick: Number,

        next: Object as PropType<MediaListItem | null>,
        prev: Object as PropType<MediaListItem | null>,
        inAlbum: Boolean,

        pageNext: Boolean,
        pagePrev: Boolean,

        canWrite: Boolean,

        min: Boolean,
    },
    emits: ["go-next", "go-prev", "update:fullscreen", "delete"],
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
    watch: {
        rTick: function () {
            this.expandedTitle = false;
        },
    },
    mounted: function () {
        // Load player preferences
        this.$addKeyboardHandler(this.onKeyPress.bind(this), 100);
        this.$listenOnAppEvent("fullscreenchange", this.onExitFullScreen.bind(this));
    },
    methods: {
        enterTooltip: function (t: string) {
            if (isTouchDevice()) {
                this.helpTooltip = "";
                return;
            }
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
                case "PageUp":
                    if (event.altKey || event.shiftKey) {
                        caught = false;
                    } else if (this.prev || this.pagePrev) {
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
                case "PageDown":
                    if (event.altKey || event.shiftKey) {
                        caught = false;
                    } else if (this.next || this.pageNext) {
                        this.goNext();
                    } else {
                        caught = false;
                    }
                    break;
                case "Delete":
                    this.$emit("delete");
                    break;
                default:
                    caught = false;
            }

            return caught;
        },
    },
});
</script>
