<template>
    <div
        class="image-player player-settings-no-trap"
        :class="{
            'player-min': min,
            'no-controls': !showControls,
            'full-screen': fullscreen,
            'bg-black': background === 'black',
            'bg-white': background === 'white',
        }"
        @mousemove="playerMouseMove"
        @click="clickPlayer"
        @dblclick="toggleFullScreen"
        @mouseleave="mouseLeavePlayer"
        @touchmove="playerMouseMove"
        @contextmenu="onContextMenu"
        @wheel="onMouseWheel"
    >
        <div class="image-prefetch-container">
            <img v-if="prefetchURL" :src="prefetchURL" />
        </div>
        <div class="image-scroller" :class="{ 'cursor-hidden': !cursorShown }" @mousedown="grabScroll">
            <img
                v-if="imageURL"
                :src="imageURL"
                :key="rTick"
                @load="onImageLoaded"
                @error="onMediaError"
                :style="{
                    width: imageWidth,
                    height: imageHeight,
                    top: imageTop,
                    left: imageLeft,
                }"
            />

            <ImageNotes
                :editing="notesEditMode"
                :contextOpen="contextMenuShown"
                :width="imageWidth"
                :height="imageHeight"
                :top="imageTop"
                :left="imageLeft"
            ></ImageNotes>
        </div>

        <div class="player-loader" v-if="loading && !mediaError">
            <div class="player-lds-ring">
                <div></div>
                <div></div>
                <div></div>
                <div></div>
            </div>
        </div>

        <PlayerEncodingPending
            v-if="(!loading && !imageURL && imagePending) || mediaError"
            :mid="mid"
            :tid="imagePendingTask"
            :res="currentResolution"
            :error="mediaError"
        ></PlayerEncodingPending>

        <div
            class="player-controls"
            :class="{ hidden: !showControls }"
            @click="clickControls"
            @dblclick="stopPropagationEvent"
            @mouseenter="enterControls"
            @mouseleave="leaveControls"
        >
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

                <ScaleControl
                    ref="scaleControl"
                    :min="min"
                    :width="min ? 70 : 100"
                    v-model:fit="fit"
                    v-model:scale="scale"
                    v-model:expanded="scaleShown"
                    @update:scale="onUserScaleUpdated"
                    @update:fit="onUserFitUpdated"
                    @enter="enterTooltip('scale')"
                    @leave="leaveTooltip('scale')"
                ></ScaleControl>
            </div>

            <div class="player-controls-right">
                <button
                    type="button"
                    :title="$t('Manage albums')"
                    class="player-btn"
                    @click="manageAlbums"
                    @mouseenter="enterTooltip('albums')"
                    @mouseleave="leaveTooltip('albums')"
                >
                    <i class="fas fa-list-ol"></i>
                </button>

                <button
                    type="button"
                    :title="$t('Player Configuration')"
                    class="player-btn player-settings-no-trap"
                    @click="showConfig"
                    @mouseenter="enterTooltip('config')"
                    @mouseleave="leaveTooltip('config')"
                >
                    <i class="fas fa-cog"></i>
                </button>

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

        <div v-if="helpTooltip === 'scale'" class="player-tooltip player-help-tip-left">
            {{ $t("Scale") }} ({{ fit ? $t("Fit") : renderScale(scale) }})
        </div>

        <div v-if="!displayConfig && helpTooltip === 'config'" class="player-tooltip player-help-tip-right">
            {{ $t("Player Configuration") }}
        </div>

        <div v-if="!displayConfig && helpTooltip === 'albums'" class="player-tooltip player-help-tip-right">
            {{ $t("Manage albums") }}
        </div>

        <div v-if="helpTooltip === 'full-screen'" class="player-tooltip player-help-tip-right">
            {{ $t("Full screen") }}
        </div>
        <div v-if="helpTooltip === 'full-screen-exit'" class="player-tooltip player-help-tip-right">
            {{ $t("Exit full screen") }}
        </div>

        <ImagePlayerConfig
            v-model:shown="displayConfig"
            v-model:resolution="currentResolution"
            v-model:background="background"
            @update:resolution="onResolutionUpdated"
            @update:background="onBackgroundChanged"
            @update-auto-next="setupAutoNextTimer"
            :rTick="internalTick"
            :metadata="metadata"
            @enter="enterControls"
            @leave="leaveControls"
        ></ImagePlayerConfig>

        <PlayerTopBar
            v-if="metadata"
            :mid="mid"
            :metadata="metadata"
            :shown="showControls"
            :fullscreen="fullscreen"
            v-model:expanded="expandedTitle"
            v-model:albumExpanded="expandedAlbum"
            :inAlbum="inAlbum"
            @click-player="clickControls"
        ></PlayerTopBar>

        <PlayerContextMenu
            type="image"
            v-model:shown="contextMenuShown"
            :x="contextMenuX"
            :y="contextMenuY"
            v-model:fit="fit"
            @update:fit="onUserFitUpdated"
            :url="imageURL"
            v-model:controls="showControlsState"
            :canWrite="canWrite"
            :hasExtendedDescription="hasExtendedDescription"
            v-model:notesEdit="notesEditMode"
            @stats="openStats"
            @open-tags="openTags"
            @open-ext-desc="openExtendedDescription"
        ></PlayerContextMenu>
    </div>
</template>

<script lang="ts">
import { PlayerPreferences } from "@/control/player-preferences";
import { defineComponent, nextTick } from "vue";

import ScaleControl from "./ScaleControl.vue";
import PlayerMediaChangePreview from "./PlayerMediaChangePreview.vue";
import PlayerTopBar from "./PlayerTopBar.vue";
import PlayerEncodingPending from "./PlayerEncodingPending.vue";
import ImageNotes from "./ImageNotes.vue";

import { openFullscreen, closeFullscreen } from "../../utils/full-screen";
import { isTouchDevice } from "@/utils/touch";
import ImagePlayerConfig from "./ImagePlayerConfig.vue";
import PlayerContextMenu from "./PlayerContextMenu.vue";
import { GetAssetURL } from "@/utils/request";
import { useVModel } from "../../utils/v-model";
import { AuthController } from "@/control/auth";
import { AppStatus } from "@/control/app-status";
import { KeyboardManager } from "@/control/keyboard";
import { AlbumsController } from "@/control/albums";
import { AppEvents } from "@/control/app-events";
import { MEDIA_TYPE_IMAGE } from "@/utils/constants";
import { MediaController } from "@/control/media";

const SCALE_RANGE = 2;
const SCALE_RANGE_PERCENT = SCALE_RANGE * 100;
const SCALE_STEP = 0.1 / SCALE_RANGE;
const SCALE_STEP_MIN = 0.01 / SCALE_RANGE;

export default defineComponent({
    components: {
        ScaleControl,
        ImagePlayerConfig,
        PlayerMediaChangePreview,
        PlayerTopBar,
        PlayerContextMenu,
        PlayerEncodingPending,
        ImageNotes,
    },
    name: "ImagePlayer",
    emits: ["go-next", "go-prev", "update:fullscreen", "update:showControls", "albums-open", "stats-open", "tags-open", "ext-desc-open"],
    props: {
        mid: Number,
        metadata: Object,
        rTick: Number,

        showControls: Boolean,

        fullscreen: Boolean,

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
            showControlsState: useVModel(props, "showControls"),
        };
    },
    data: function () {
        return {
            loading: true,

            imageURL: "",
            imagePending: false,
            imagePendingTask: 0,

            currentResolution: -1,
            width: 0,
            height: 0,

            displayConfig: false,

            imageTop: "0",
            imageLeft: "0",
            imageWidth: "auto",
            imageHeight: "auto",

            lastControlsInteraction: Date.now(),
            mouseInControls: false,

            scale: 0,
            fit: true,
            scaleShown: isTouchDevice(),

            background: "default",

            internalTick: 0,

            helpTooltip: "",

            expandedTitle: false,
            expandedAlbum: false,

            contextMenuX: 0,
            contextMenuY: 0,
            contextMenuShown: false,

            scrollGrabbed: false,
            scrollGrabX: 0,
            scrollGrabY: 0,
            scrollGrabTop: 0,
            scrollGrabLeft: 0,

            cursorShown: false,

            prefetchURL: "",

            notesEditMode: false,
            hasExtendedDescription: false,

            mediaError: false,
        };
    },
    methods: {
        onContextMenu: function (e) {
            this.contextMenuX = e.pageX;
            this.contextMenuY = e.pageY;
            this.contextMenuShown = true;
            e.preventDefault();
        },

        manageAlbums: function () {
            this.$emit("albums-open");
        },

        openStats: function () {
            this.$emit("stats-open");
        },

        openTags: function () {
            this.$emit("tags-open");
        },

        openExtendedDescription: function () {
            if (!this.hasExtendedDescription && !this.canWrite) {
                return;
            }
            this.$emit("ext-desc-open");
        },

        grabScroll: function (e) {
            const scroller = this.$el.querySelector(".image-scroller");

            if (!scroller) {
                return;
            }

            this.scrollGrabTop = scroller.scrollTop;
            this.scrollGrabLeft = scroller.scrollLeft;

            this.scrollGrabbed = true;
            if (e.touches && e.touches.length > 0) {
                this.scrollGrabX = e.touches[0].pageX;
                this.scrollGrabY = e.touches[0].pageY;
            } else {
                this.scrollGrabX = e.pageX;
                this.scrollGrabY = e.pageY;
            }
        },

        moveScrollByMouse: function (x: number, y: number) {
            const scroller = this.$el.querySelector(".image-scroller");

            if (!scroller) {
                return;
            }

            const rect = scroller.getBoundingClientRect();

            const maxScrollLeft = scroller.scrollWidth - rect.width;
            const maxScrollTop = scroller.scrollHeight - rect.height;

            const diffX = x - this.scrollGrabX;
            const diffY = y - this.scrollGrabY;

            scroller.scrollTop = Math.max(0, Math.min(maxScrollTop, this.scrollGrabTop - diffY));
            scroller.scrollLeft = Math.max(0, Math.min(maxScrollLeft, this.scrollGrabLeft - diffX));
        },

        dropScroll: function (e) {
            if (!this.scrollGrabbed) {
                return;
            }
            this.scrollGrabbed = false;

            if (e.touches && e.touches.length > 0) {
                this.moveScrollByMouse(e.touches[0].pageX, e.touches[0].pageY);
            } else {
                this.moveScrollByMouse(e.pageX, e.pageY);
            }
        },

        moveScroll: function (e) {
            if (!this.scrollGrabbed) {
                return;
            }

            if (e.touches && e.touches.length > 0) {
                this.moveScrollByMouse(e.touches[0].pageX, e.touches[0].pageY);
            } else {
                this.moveScrollByMouse(e.pageX, e.pageY);
            }
        },

        centerScroll: function () {
            const scroller = this.$el.querySelector(".image-scroller");

            if (!scroller) {
                return;
            }

            scroller.scrollTop = (scroller.scrollHeight - scroller.getBoundingClientRect().height) / 2;
            scroller.scrollLeft = (scroller.scrollWidth - scroller.getBoundingClientRect().width) / 2;
        },

        computeImageDimensions() {
            if (!this.imageURL) {
                return;
            }

            const scroller = this.$el.querySelector(".image-scroller");

            if (!scroller) {
                return;
            }

            const scrollerDimensions = scroller.getBoundingClientRect();

            const fitDimensions = {
                width: this.width,
                height: this.height,
                fitWidth: true,
            };

            if (scrollerDimensions.width > scrollerDimensions.height) {
                fitDimensions.fitWidth = true;
                fitDimensions.height = scrollerDimensions.height;
                fitDimensions.width = (scrollerDimensions.height * this.width) / this.height;

                if (fitDimensions.width > scrollerDimensions.width) {
                    fitDimensions.fitWidth = false;
                    fitDimensions.width = scrollerDimensions.width;
                    fitDimensions.height = (scrollerDimensions.width * this.height) / this.width;
                }
            } else {
                fitDimensions.fitWidth = false;
                fitDimensions.width = scrollerDimensions.width;
                fitDimensions.height = (scrollerDimensions.width * this.height) / this.width;

                if (fitDimensions.height > scrollerDimensions.height) {
                    fitDimensions.fitWidth = true;
                    fitDimensions.height = scrollerDimensions.height;
                    fitDimensions.width = (scrollerDimensions.height * this.width) / this.height;
                }
            }

            if (this.fit) {
                let top = Math.max(0, (scrollerDimensions.height - fitDimensions.height) / 2);
                let left = Math.max(0, (scrollerDimensions.width - fitDimensions.width) / 2);

                this.imageTop = Math.floor(top) + "px";
                this.imageLeft = Math.floor(left) + "px";
                this.imageWidth = Math.floor(fitDimensions.width) + "px";
                this.imageHeight = Math.floor(fitDimensions.height) + "px";
            } else {
                let width = 0;
                let height = 0;

                if (fitDimensions.fitWidth) {
                    width = scrollerDimensions.width * (0.5 + this.scale * SCALE_RANGE);
                    height = (width * this.height) / this.width;
                } else {
                    height = scrollerDimensions.height * (0.5 + this.scale * SCALE_RANGE);
                    width = (height * this.width) / this.height;
                }

                let top = Math.max(0, (scrollerDimensions.height - height) / 2);
                let left = Math.max(0, (scrollerDimensions.width - width) / 2);

                this.imageTop = Math.floor(top) + "px";
                this.imageLeft = Math.floor(left) + "px";
                this.imageWidth = Math.floor(width) + "px";
                this.imageHeight = Math.floor(height) + "px";
            }
        },

        renderScale: function (v: number): string {
            return Math.round(50 + v * SCALE_RANGE_PERCENT) + "%";
        },
        enterTooltip: function (t: string) {
            this.helpTooltip = t;
        },

        leaveTooltip: function (t: string) {
            if (t === this.helpTooltip) {
                this.helpTooltip = "";
            }
        },

        showConfig: function (e) {
            this.displayConfig = !this.displayConfig;
            e.stopPropagation();
        },

        clickControls: function (e) {
            this.displayConfig = false;
            this.contextMenuShown = false;
            if (e) {
                e.stopPropagation();
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

        onUserScaleUpdated() {
            PlayerPreferences.SetScale(this.scale);
            this.computeImageDimensions();
            nextTick(this.centerScroll.bind(this));
        },

        changeScale: function (s: number) {
            this.scale = s;
            this.onUserScaleUpdated();
        },

        onUserFitUpdated() {
            PlayerPreferences.SetFit(this.fit);
            this.computeImageDimensions();
        },

        toggleFit: function () {
            this.fit = !this.fit;
            this.onUserFitUpdated();
        },

        onBackgroundChanged() {
            PlayerPreferences.SetImagePlayerBackground(this.background);
        },

        /* Player events */

        onImageLoaded: function () {
            this.loading = false;
        },

        playerMouseMove: function () {
            this.interactWithControls();
        },
        mouseLeavePlayer: function () {
            this.helpTooltip = "";
            this.displayConfig = false;
        },

        tick() {
            this.computeImageDimensions();
            if (!this.mouseInControls && this.helpTooltip && Date.now() - this.lastControlsInteraction > 2000) {
                this.helpTooltip = "";
            }
            if (!this.mouseInControls && this.scaleShown && Date.now() - this.lastControlsInteraction > 2000) {
                this.scaleShown = false;
            }
            if (!this.mouseInControls && this.cursorShown && Date.now() - this.lastControlsInteraction > 2000) {
                this.cursorShown = false;
            }
            if (this.helpTooltip && !this.showControls) {
                this.helpTooltip = "";
            }
        },

        interactWithControls() {
            this.lastControlsInteraction = Date.now();
            this.cursorShown = true;
        },

        enterControls: function () {
            this.mouseInControls = true;
        },

        leaveControls: function () {
            this.mouseInControls = false;
            this.helpTooltip = "";
            this.scaleShown = false;
        },

        clickPlayer: function () {
            if (this.displayConfig) {
                this.displayConfig = false;
            }
            this.interactWithControls();
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

        onMediaError: function () {
            if (AuthController.RefreshSessionCookie()) {
                MediaController.Load();
            } else {
                this.mediaError = true;
                this.loading = false;
                AuthController.CheckAuthStatusSilent();
            }
        },

        incrementImageScroll: function (a: number | string): boolean {
            if (this.fit) {
                return false;
            }

            const el = this.$el.querySelector(".image-scroller");

            if (!el) {
                return false;
            }

            const maxScroll = Math.max(0, el.scrollHeight - el.getBoundingClientRect().height);

            if (maxScroll <= 0) {
                return false;
            }

            if (typeof a === "number") {
                el.scrollTop = Math.min(Math.max(0, el.scrollTop + a), maxScroll);
            } else if (a === "home") {
                el.scrollTop = 0;
            } else if (a === "end") {
                el.scrollTop = maxScroll;
            }

            return true;
        },

        tryHorizontalScroll: function (a: number | string): boolean {
            if (this.fit) {
                return false;
            }
            const el = this.$el.querySelector(".image-scroller");

            if (!el) {
                return false;
            }

            const maxScroll = Math.max(0, el.scrollWidth - el.getBoundingClientRect().width);

            if (maxScroll <= 0) {
                return false;
            }

            if (typeof a === "number") {
                el.scrollLeft = Math.min(Math.max(0, el.scrollLeft + a), maxScroll);
            } else if (a === "home") {
                el.scrollLeft = 0;
            } else if (a === "end") {
                el.scrollLeft = maxScroll;
            }

            return true;
        },

        onKeyPress: function (event: KeyboardEvent) {
            if (
                AuthController.Locked ||
                !AppStatus.IsPlayerVisible() ||
                !event.key ||
                (event.ctrlKey && event.key !== "+" && event.key !== "-")
            ) {
                return false;
            }
            let caught = true;
            const shifting = event.shiftKey;
            switch (event.key) {
            case "A":
            case "a":
                this.manageAlbums();
                break;
            case "i":
            case "I":
                this.openExtendedDescription();
                break;
            case "t":
            case "T":
                this.openTags();
                break;
            case "S":
            case "s":
                this.showConfig();
                break;
            case "ArrowUp":
                this.incrementImageScroll(-40);
                break;
            case "ArrowDown":
                this.incrementImageScroll(40);
                break;
            case "ArrowRight":
                if (shifting || event.altKey || !this.tryHorizontalScroll(40)) {
                    if (this.next || this.pageNext) {
                        this.goNext();
                    } else {
                        caught = false;
                    }
                }
                break;
            case "ArrowLeft":
                if (shifting || event.altKey || !this.tryHorizontalScroll(-40)) {
                    if (this.prev || this.pagePrev) {
                        this.goPrev();
                    } else {
                        caught = false;
                    }
                }
                break;
            case "Home":
                if (event.altKey) {
                    caught = false;
                } else if (shifting) {
                    if (!this.tryHorizontalScroll("home")) {
                        caught = false;
                    }
                } else {
                    if (!this.incrementImageScroll("home")) {
                        caught = false;
                    }
                }
                break;
            case "End":
                if (event.altKey) {
                    caught = false;
                } else if (shifting) {
                    if (!this.tryHorizontalScroll("end")) {
                        caught = false;
                    }
                } else {
                    if (!this.incrementImageScroll("end")) {
                        caught = false;
                    }
                }
                break;
            case " ":
            case "K":
            case "k":
            case "Enter":
                this.toggleFit();
                this.scaleShown = true;
                this.helpTooltip = "scale";
                break;
            case "+":
                this.changeScale(Math.min(1, this.scale + (shifting ? SCALE_STEP_MIN : SCALE_STEP)));
                this.scaleShown = true;
                this.helpTooltip = "scale";
                this.fit = false;
                this.onUserFitUpdated();
                break;
            case "-":
                this.changeScale(Math.max(0, this.scale - (shifting ? SCALE_STEP_MIN : SCALE_STEP)));
                this.scaleShown = true;
                this.helpTooltip = "scale";
                this.fit = false;
                this.onUserFitUpdated();
                break;
            case "F":
            case "f":
                if (event.altKey || shifting) {
                    caught = false;
                } else {
                    this.toggleFullScreen();
                }
                break;
            case "C":
            case "c":
                this.showControlsState = !this.showControlsState;
                break;
            case "PageDown":
                if (this.prev || this.pagePrev) {
                    this.goPrev();
                } else {
                    caught = false;
                }
                break;
            case "PageUp":
                if (this.next || this.pageNext) {
                    this.goNext();
                } else {
                    caught = false;
                }
                break;
            default:
                caught = false;
            }

            if (caught) {
                this.interactWithControls();
            }

            return caught;
        },

        initializeImage() {
            if (!this.metadata) {
                return;
            }
            this.hasExtendedDescription = !!this.metadata.ext_desc_url;
            this.loading = true;
            this.currentResolution = PlayerPreferences.GetResolutionIndexImage(this.metadata);
            this.setImageURL();
        },

        onResolutionUpdated: function () {
            PlayerPreferences.SetResolutionIndexImage(this.metadata, this.currentResolution);
            this.setImageURL();
        },

        setImageURL() {
            this.mediaError = false;

            if (!this.metadata) {
                this.imageURL = "";
                this.loading = false;
                return;
            }

            if (this.currentResolution < 0) {
                if (this.metadata.encoded) {
                    this.imageURL = GetAssetURL(this.metadata.url);
                    this.imagePending = false;
                    this.imagePendingTask = 0;
                    this.width = this.metadata.width;
                    this.height = this.metadata.height;
                    this.setupAutoNextTimer();
                } else {
                    this.imageURL = "";
                    this.imagePending = true;
                    this.imagePendingTask = this.metadata.task;
                    this.loading = false;
                }
            } else {
                if (this.metadata.resolutions && this.metadata.resolutions.length > this.currentResolution) {
                    let res = this.metadata.resolutions[this.currentResolution];
                    if (res.ready) {
                        this.imageURL = GetAssetURL(res.url);
                        this.imagePending = false;
                        this.imagePendingTask = 0;
                        this.width = this.metadata.width;
                        this.height = this.metadata.height;
                        this.setupAutoNextTimer();
                    } else {
                        this.imageURL = "";
                        this.imagePending = true;
                        this.imagePendingTask = res.task;
                        this.loading = false;
                    }
                } else {
                    this.imageURL = "";
                    this.imagePending = true;
                    this.imagePendingTask = 0;
                    this.loading = false;
                }
            }

            this.computeImageDimensions();
            this.incrementImageScroll("home");
            this.tryHorizontalScroll("home");
        },

        setupAutoNextTimer: function () {
            if (this.$options.autoNextTimer) {
                clearTimeout(this.$options.autoNextTimer);
                this.$options.autoNextTimer = null;
            }
            const timerS = PlayerPreferences.ImageAutoNext;

            if (isNaN(timerS) || !isFinite(timerS) || timerS <= 0) {
                return;
            }

            if (!this.next && !this.pageNext) {
                return;
            }

            const ms = timerS * 1000;

            this.$options.autoNextTimer = setTimeout(() => {
                this.$options.autoNextTimer = null;
                if (this.displayConfig) {
                    this.setupAutoNextTimer();
                } else {
                    this.goNext();
                }
            }, ms);
        },

        onMouseWheel: function (e: WheelEvent) {
            if (e.ctrlKey) {
                e.preventDefault();
                e.stopPropagation();
                if (e.deltaY > 0) {
                    this.changeScale(Math.max(0, this.scale - SCALE_STEP));
                    this.scaleShown = true;
                    this.helpTooltip = "scale";
                    this.fit = false;
                    this.onUserFitUpdated();
                } else {
                    this.changeScale(Math.min(1, this.scale + SCALE_STEP));
                    this.scaleShown = true;
                    this.helpTooltip = "scale";
                    this.fit = false;
                    this.onUserFitUpdated();
                }
            }
        },

        onAlbumPrefetch: function () {
            if (AlbumsController.NextMediaData && AlbumsController.NextMediaData.type === MEDIA_TYPE_IMAGE) {
                if (this.currentResolution < 0) {
                    if (AlbumsController.NextMediaData.encoded) {
                        this.prefetchURL = GetAssetURL(AlbumsController.NextMediaData.url);
                    } else {
                        this.prefetchURL = "";
                    }
                } else {
                    if (
                        AlbumsController.NextMediaData.resolutions[this.currentResolution] &&
                        AlbumsController.NextMediaData.resolutions[this.currentResolution].ready
                    ) {
                        this.prefetchURL = GetAssetURL(AlbumsController.NextMediaData.resolutions[this.currentResolution].url);
                    } else {
                        this.prefetchURL = "";
                    }
                }
            } else {
                this.prefetchURL = "";
            }
        },

        handleMediaSessionEvent: function (event: { action: string; fastSeek: boolean; seekTime: number; seekOffset: number }) {
            if (!event || !event.action) {
                return;
            }
            switch (event.action) {
            case "nexttrack":
                if (this.next || this.pageNext) {
                    this.goNext();
                }
                break;
            case "previoustrack":
                if (this.prev || this.pagePrev) {
                    this.goPrev();
                }
                break;
            }
        },
    },
    mounted: function () {
        // Load player preferences
        this.fit = PlayerPreferences.PlayerFit;
        this.scale = PlayerPreferences.PlayerScale;
        this.background = PlayerPreferences.ImagePlayerBackground;

        this.$options.keyHandler = this.onKeyPress.bind(this);
        KeyboardManager.AddHandler(this.$options.keyHandler, 100);

        this.$options.timer = setInterval(this.tick.bind(this), Math.floor(1000 / 30));

        this.$options.exitFullScreenListener = this.onExitFullScreen.bind(this);
        document.addEventListener("fullscreenchange", this.$options.exitFullScreenListener);
        document.addEventListener("webkitfullscreenchange", this.$options.exitFullScreenListener);
        document.addEventListener("mozfullscreenchange", this.$options.exitFullScreenListener);
        document.addEventListener("MSFullscreenChange", this.$options.exitFullScreenListener);

        this.$options.dropScrollHandler = this.dropScroll.bind(this);
        document.addEventListener("mouseup", this.$options.dropScrollHandler);

        this.$options.moveScrollHandler = this.moveScroll.bind(this);

        document.addEventListener("mousemove", this.$options.moveScrollHandler);

        this.$options.onAlbumPrefetchH = this.onAlbumPrefetch.bind(this);
        AppEvents.AddEventListener("album-next-prefetch", this.$options.onAlbumPrefetchH);

        this.initializeImage();

        if (window.navigator && window.navigator.mediaSession) {
            navigator.mediaSession.setActionHandler("nexttrack", this.handleMediaSessionEvent.bind(this));
            navigator.mediaSession.setActionHandler("previoustrack", this.handleMediaSessionEvent.bind(this));
        }
    },
    beforeUnmount: function () {
        this.imageURL = "";
        clearInterval(this.$options.timer);

        document.removeEventListener("fullscreenchange", this.$options.exitFullScreenListener);
        document.removeEventListener("webkitfullscreenchange", this.$options.exitFullScreenListener);
        document.removeEventListener("mozfullscreenchange", this.$options.exitFullScreenListener);
        document.removeEventListener("MSFullscreenChange", this.$options.exitFullScreenListener);

        document.removeEventListener("mouseup", this.$options.dropScrollHandler);

        document.removeEventListener("mousemove", this.$options.moveScrollHandler);

        AppEvents.RemoveEventListener("album-next-prefetch", this.$options.onAlbumPrefetchH);

        KeyboardManager.RemoveHandler(this.$options.keyHandler);

        if (this.$options.autoNextTimer) {
            clearTimeout(this.$options.autoNextTimer);
            this.$options.autoNextTimer = null;
        }

        if (window.navigator && window.navigator.mediaSession) {
            navigator.mediaSession.setActionHandler("nexttrack", null);
            navigator.mediaSession.setActionHandler("previoustrack", null);
        }
    },
    watch: {
        rTick: function () {
            this.internalTick++;
            this.expandedTitle = false;
            this.initializeImage();
        },
        imageURL: function () {
            if (this.imageURL) {
                this.loading = true;
            }
        },
        next: function () {
            this.setupAutoNextTimer();
        },
        pageNext: function () {
            this.setupAutoNextTimer();
        },
    },
});
</script>
