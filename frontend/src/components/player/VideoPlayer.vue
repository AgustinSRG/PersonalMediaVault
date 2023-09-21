<template>
    <div
        class="video-player player-settings-no-trap"
        :class="{
            'player-min': min,
            'no-controls': !showControls || !userControls,
            'full-screen': fullscreen,
        }"
        @mousemove="playerMouseMove"
        @mousedown="onPlayerMouseDown"
        @touchstart="onPlayerTouchStart"
        @dblclick="toggleFullScreen"
        @mouseleave="mouseLeavePlayer"
        @mouseup="playerMouseUp"
        @touchmove="playerMouseMove"
        @touchend.passive="playerMouseUp"
        @contextmenu="onContextMenu"
    >
        <video
            v-if="videoURL"
            :src="videoURL"
            crossorigin="use-credentials"
            :key="rTick"
            playsinline
            webkit-playsinline
            x-webkit-airplay="allow"
            :muted="muted || !!audioTrackURL"
            :loop="(loop || isShort) && !sliceLoop"
            :volume.prop="volume"
            :playbackRate.prop="speed"
            @ended="onEnded"
            @timeupdate="onVideoTimeUpdate"
            @canplay="onCanPlay"
            @loadedmetadata="onLoadMetaData"
            @waiting="onWaitForBuffer(true)"
            @playing="onWaitForBuffer(false)"
            @play="onPlay"
            @pause="onPause"
            @error="onMediaError"
        ></video>

        <audio
            v-if="audioTrackURL"
            :src="audioTrackURL"
            crossorigin="use-credentials"
            :key="rTick"
            playsinline
            webkit-playsinline
            :muted="muted || !audioTrackURL"
            :volume.prop="volume"
            :playbackRate.prop="speed"
            @loadedmetadata="onAudioLoadMetadata"
            @canplay="onAudioCanPlay"
        ></audio>

        <div class="player-feedback-container">
            <div class="player-feedback player-feedback-play" key="play" v-if="feedback === 'play'" @animationend="onFeedBackAnimationEnd">
                <div><i class="fas fa-play"></i></div>
            </div>
            <div
                class="player-feedback player-feedback-pause"
                key="pause"
                v-if="feedback === 'pause'"
                @animationend="onFeedBackAnimationEnd"
            >
                <div><i class="fas fa-pause"></i></div>
            </div>
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
            v-if="(!loading && !videoURL && videoPending) || mediaError"
            :mid="mid"
            :tid="videoPendingTask"
            :res="currentResolution"
            :error="mediaError"
        ></PlayerEncodingPending>

        <div class="player-subtitles-container" :class="{ 'controls-hidden': !showControls || !userControls }">
            <div
                class="player-subtitles"
                v-if="subtitles"
                v-html="subtitles"
                :class="{
                    'player-subtitles-s': subtitlesSize === 's',
                    'player-subtitles-m': subtitlesSize === 'm',
                    'player-subtitles-l': subtitlesSize === 'l',
                    'player-subtitles-xl': subtitlesSize === 'xl',
                    'player-subtitles-xxl': subtitlesSize === 'xxl',

                    'player-subtitles-bg-0': subtitlesBg === '0',
                    'player-subtitles-bg-25': subtitlesBg === '25',
                    'player-subtitles-bg-50': subtitlesBg === '50',
                    'player-subtitles-bg-75': subtitlesBg === '75',
                    'player-subtitles-bg-100': subtitlesBg === '100',
                }"
            ></div>
        </div>

        <div
            class="player-controls"
            :class="{ hidden: !showControls || !userControls }"
            @click="clickControls"
            @dblclick="stopPropagationEvent"
            @mousedown="stopPropagationEvent"
            @touchstart="stopPropagationEvent"
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

                <button
                    v-if="!playing"
                    type="button"
                    :title="$t('Play')"
                    class="player-btn player-play-btn"
                    @click="togglePlayImmediate"
                    @mouseenter="enterTooltip('play')"
                    @mouseleave="leaveTooltip('play')"
                >
                    <i class="fas fa-play"></i>
                </button>
                <button
                    v-if="playing"
                    type="button"
                    :title="$t('Pause')"
                    class="player-btn player-play-btn"
                    @click="togglePlayImmediate"
                    @mouseenter="enterTooltip('pause')"
                    @mouseleave="leaveTooltip('pause')"
                >
                    <i class="fas fa-pause"></i>
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

                <VolumeControl
                    ref="volumeControl"
                    :min="min"
                    :width="min ? 50 : 80"
                    v-model:muted="muted"
                    v-model:volume="volume"
                    v-model:expanded="volumeShown"
                    @update:volume="onUserVolumeUpdated"
                    @update:muted="onUserMutedUpdated"
                    @enter="enterTooltip('volume')"
                    @leave="leaveTooltip('volume')"
                ></VolumeControl>

                <div class="player-time-label-container" :class="{ 'in-album': !!next || !!prev }" v-if="!min">
                    <span>{{ renderTime(currentTime) }} / {{ renderTime(duration) }}</span>
                    <span v-if="currentTimeSlice" class="times-slice-name"><b class="separator"> - </b>{{ currentTimeSliceName }}</span>
                </div>
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

        <div v-if="helpTooltip === 'play'" class="player-tooltip player-help-tip-left">
            {{ $t("Play") }}
        </div>
        <div v-if="helpTooltip === 'pause'" class="player-tooltip player-help-tip-left">
            {{ $t("Pause") }}
        </div>

        <div v-if="prev && helpTooltip === 'prev'" class="player-tooltip player-help-tip-left">
            <PlayerMediaChangePreview :media="prev" :next="false"></PlayerMediaChangePreview>
        </div>

        <div v-if="next && helpTooltip === 'next'" class="player-tooltip player-help-tip-left">
            <PlayerMediaChangePreview :media="next" :next="true"></PlayerMediaChangePreview>
        </div>

        <div v-if="helpTooltip === 'volume'" class="player-tooltip player-help-tip-left">
            {{ $t("Volume") }} ({{ muted ? $t("Muted") : renderVolume(volume) }})
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

        <div
            class="player-timeline"
            :class="{ hidden: !showControls || !userControls }"
            @mouseenter="enterControls"
            @mouseleave="mouseLeaveTimeline"
            @mousemove="mouseMoveTimeline"
            @dblclick="stopPropagationEvent"
            @click="clickTimeline"
            @mousedown="grabTimeline"
            @touchstart="grabTimeline"
        >
            <div class="player-timeline-back"></div>
            <div class="player-timeline-buffer" :style="{ width: getTimelineBarWidth(bufferedTime, duration) }"></div>
            <div class="player-timeline-current" :style="{ width: getTimelineBarWidth(currentTime, duration) }"></div>

            <div
                v-for="ts in timeSlices"
                :key="ts"
                class="player-timeline-split"
                :class="{ 'start-split': ts.start <= 0 }"
                :style="{ left: getTimelineBarWidth(ts.start, duration) }"
            ></div>

            <div class="player-timeline-thumb" :style="{ left: getTimelineThumbLeft(currentTime, duration) }"></div>
        </div>

        <div v-if="tooltipShown" class="player-tooltip" :style="{ left: tooltipX + 'px' }">
            <div v-if="tooltipImage && !tooltipImageInvalid">
                <img class="player-tooltip-image" :src="tooltipImage" @error="onTooltipImageError" />
            </div>
            <div class="player-tooltip-text">{{ tooltipText }}</div>
            <div v-if="tooltipTimeSlice" class="player-tooltip-text">
                {{ tooltipTimeSlice }}
            </div>
        </div>

        <VideoPlayerConfig
            v-model:shown="displayConfig"
            v-model:speed="speed"
            v-model:loop="loop"
            @update:loop="() => this.$emit('force-loop', this.loop)"
            v-model:nextEnd="nextEnd"
            v-model:resolution="currentResolution"
            v-model:subSize="subtitlesSize"
            v-model:subBackground="subtitlesBg"
            v-model:subHTML="subtitlesHTML"
            @update:resolution="onResolutionUpdated"
            @update:subHTML="onUpdateSubHTML"
            @update:nextEnd="onUpdateNextEnd"
            :rTick="internalTick"
            :metadata="metadata"
            @enter="enterControls"
            @leave="leaveControls"
            v-model:audioTrack="audioTrack"
            @update:audioTrack="onUpdateAudioTrack"
            :isShort="isShort"
            @update-auto-next="setupAutoNextTimer"
        ></VideoPlayerConfig>

        <PlayerTopBar
            v-if="metadata"
            :mid="mid"
            :metadata="metadata"
            :shown="showControls && userControls"
            :fullscreen="fullscreen"
            v-model:expanded="expandedTitle"
            v-model:albumExpanded="expandedAlbum"
            @update:expanded="onTopBarExpand"
            :inAlbum="inAlbum"
            @click-player="clickControls"
        ></PlayerTopBar>

        <PlayerContextMenu
            type="video"
            v-model:shown="contextMenuShown"
            :x="contextMenuX"
            :y="contextMenuY"
            v-model:loop="loop"
            @update:loop="() => this.$emit('force-loop', this.loop)"
            :url="videoURL"
            :canWrite="canWrite"
            :hasExtendedDescription="hasExtendedDescription"
            @stats="openStats"
            v-model:sliceLoop="sliceLoop"
            :hasSlices="timeSlices && timeSlices.length > 0"
            :isShort="isShort"
            v-model:controls="userControlsState"
            @open-tags="openTags"
            @open-ext-desc="openExtendedDescription"
        ></PlayerContextMenu>
    </div>
</template>

<script lang="ts">
import { PlayerPreferences } from "@/control/player-preferences";
import { defineComponent, nextTick } from "vue";

import VolumeControl from "./VolumeControl.vue";
import PlayerTopBar from "./PlayerTopBar.vue";
import PlayerMediaChangePreview from "./PlayerMediaChangePreview.vue";
import PlayerEncodingPending from "./PlayerEncodingPending.vue";

import { openFullscreen, closeFullscreen } from "../../utils/full-screen";
import { renderTimeSeconds } from "../../utils/time";
import { findTimeSlice, normalizeTimeSlices } from "../../utils/time-slices";
import { isTouchDevice } from "@/utils/touch";
import VideoPlayerConfig from "./VideoPlayerConfig.vue";
import PlayerContextMenu from "./PlayerContextMenu.vue";
import { GetAssetURL } from "@/utils/request";
import { useVModel } from "../../utils/v-model";
import { MediaController } from "@/control/media";
import { SubtitlesController } from "@/control/subtitles";
import { htmlToText } from "@/utils/html";
import { AppEvents } from "@/control/app-events";
import { getUniqueSubtitlesLoadTag, sanitizeSubtitlesHTML } from "@/utils/subtitles-html";
import { AppStatus } from "@/control/app-status";
import { KeyboardManager } from "@/control/keyboard";
import { AuthController } from "@/control/auth";
import { AUTO_LOOP_MIN_DURATION } from "@/utils/constants";

export default defineComponent({
    components: {
        VolumeControl,
        VideoPlayerConfig,
        PlayerMediaChangePreview,
        PlayerTopBar,
        PlayerContextMenu,
        PlayerEncodingPending,
    },
    name: "VideoPlayer",
    emits: [
        "go-next",
        "go-prev",
        "ended",
        "update:fullscreen",
        "albums-open",
        "stats-open",
        "tags-open",
        "ext-desc-open",
        "force-loop",
        "delete",
    ],
    props: {
        mid: Number,
        metadata: Object,
        rTick: Number,

        fullscreen: Boolean,

        next: Object,
        prev: Object,
        inAlbum: Boolean,

        pageNext: Boolean,
        pagePrev: Boolean,

        canWrite: Boolean,

        userControls: Boolean,

        min: Boolean,

        loopForced: Boolean,
        loopForcedValue: Boolean,

        autoPlay: Boolean,
    },
    setup(props) {
        return {
            fullScreenState: useVModel(props, "fullscreen"),
            userControlsState: useVModel(props, "userControls"),
        };
    },
    data: function () {
        return {
            playing: false,
            loading: true,

            autoPlayApplied: false,

            videoURL: "",
            videoPending: false,
            videoPendingTask: 0,

            canSaveTime: true,

            displayConfig: false,

            currentTime: 0,
            duration: 0,
            bufferedTime: 0,
            ended: false,
            timelineGrabbed: false,
            lastTimeChangedEvent: 0,

            // Timeline tooltip
            tooltipShown: false,
            tooltipText: "",
            tooltipTimeSlice: "",
            tooltipX: 0,
            tooltipEventX: 0,
            tooltipImage: "",
            tooltipImageInvalid: false,

            showControls: true,
            lastControlsInteraction: Date.now(),
            mouseInControls: false,

            loop: false,
            nextEnd: false,
            sliceLoop: false,

            isShort: false,

            currentResolution: -1,

            volume: 1,
            muted: false,
            volumeShown: isTouchDevice(),
            internalTick: 0,

            speed: 1,

            feedback: "",

            helpTooltip: "",

            expandedTitle: false,
            expandedAlbum: false,

            contextMenuX: 0,
            contextMenuY: 0,
            contextMenuShown: false,

            requiresRefresh: false,

            subtitles: "",
            subtitlesStart: -1,
            subtitlesEnd: -1,
            subtitlesSize: "l",
            subtitlesBg: "75",
            subtitlesHTML: false,

            audioTrack: PlayerPreferences.SelectedAudioTrack,
            audioTrackURL: "",

            timeSlices: [],
            currentTimeSlice: null,
            currentTimeSliceName: "",
            currentTimeSliceStart: 0,
            currentTimeSliceEnd: 0,

            waitingTimestamp: 0,
            isWaiting: false,

            mediaError: false,

            hasExtendedDescription: false,
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

        onPlayerMouseDown: function (e: MouseEvent) {
            if (e.button !== 0) {
                return;
            }
            if (this.contextMenuShown) {
                e.stopPropagation();
            }
            this.clickPlayer();
        },

        onPlayerTouchStart: function (e) {
            if (this.contextMenuShown) {
                e.stopPropagation();
            }
            this.clickPlayer();
        },

        clickPlayer: function () {
            this.leaveControls();
            if (this.displayConfig || this.contextMenuShown) {
                this.displayConfig = false;
                this.contextMenuShown = false;
            } else {
                this.togglePlay();
            }
            this.interactWithControls();
        },

        renderVolume: function (v: number): string {
            return Math.round(v * 100) + "%";
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

        getThumbnailForTime: function (time: number) {
            if (this.duration <= 0 || !this.metadata || !this.metadata.video_previews || !this.metadata.video_previews_interval) {
                return "";
            }
            const thumbCount = Math.floor(this.duration / this.metadata.video_previews_interval);

            if (thumbCount <= 0) {
                return "";
            }

            let part = Math.floor(time / this.metadata.video_previews_interval);
            if (part > thumbCount) {
                part = thumbCount;
            }

            return GetAssetURL(this.metadata.video_previews.replace("{INDEX}", "" + part));
        },

        onResolutionUpdated: function () {
            PlayerPreferences.SetResolutionIndex(this.metadata, this.currentResolution);
            this.setVideoURL();
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

        onUserVolumeUpdated() {
            PlayerPreferences.SetVolume(this.volume);
        },

        changeVolume: function (v: number) {
            this.volume = v;
            this.onUserVolumeUpdated();
        },

        onUserMutedUpdated() {
            PlayerPreferences.SetMuted(this.muted);
        },

        toggleMuted: function () {
            this.muted = !this.muted;
            this.onUserMutedUpdated();
        },

        /* Player events */

        onAudioLoadMetadata: function () {
            const audioElement = this.$el.querySelector("audio");
            if (!audioElement) {
                return;
            }

            if (typeof audioElement.duration !== "number" || isNaN(audioElement.duration) || !isFinite(audioElement.duration)) {
                return;
            }

            const duration = audioElement.duration;

            if (typeof this.currentTime === "number" && !isNaN(this.currentTime) && isFinite(this.currentTime) && this.currentTime >= 0) {
                audioElement.currentTime = Math.min(this.currentTime, duration);
            }
        },

        onLoadMetaData: function (ev) {
            const videoElement = ev.target;

            if (!videoElement) {
                return;
            }

            if (typeof videoElement.duration !== "number" || isNaN(videoElement.duration) || !isFinite(videoElement.duration)) {
                return;
            }

            this.duration = videoElement.duration;

            if (typeof this.currentTime === "number" && !isNaN(this.currentTime) && isFinite(this.currentTime) && this.currentTime >= 0) {
                videoElement.currentTime = Math.min(this.currentTime, this.duration);
                this.updateSubtitles();
                this.updateCurrentTimeSlice();
            }
        },
        onVideoTimeUpdate: function (ev) {
            if (this.loading) return;
            const videoElement = ev.target;
            if (
                !videoElement ||
                typeof videoElement.currentTime !== "number" ||
                isNaN(videoElement.currentTime) ||
                !isFinite(videoElement.currentTime)
            ) {
                return;
            }
            this.currentTime = videoElement.currentTime;
            this.duration = videoElement.duration;
            if (this.canSaveTime && Date.now() - this.lastTimeChangedEvent > 5000) {
                PlayerPreferences.SetInitialTime(this.mid, this.currentTime);
                this.lastTimeChangedEvent = Date.now();
            }
            this.updateSubtitles();
            this.updateCurrentTimeSlice();

            if (this.audioTrackURL) {
                const audioElement = this.$el.querySelector("audio");
                if (audioElement) {
                    const correspondingTime = Math.min(videoElement.currentTime, audioElement.duration);

                    if (Math.abs(correspondingTime - audioElement.currentTime) > this.speed / 10) {
                        audioElement.currentTime = Math.min(videoElement.currentTime, audioElement.duration);
                    }

                    if (audioElement.paused !== videoElement.paused && videoElement.currentTime <= audioElement.duration) {
                        if (videoElement.paused) {
                            audioElement.pause();
                        } else {
                            audioElement.play();
                        }
                    }
                }
            }
        },
        onCanPlay: function () {
            this.loading = false;
            if (this.autoPlayApplied) {
                return;
            }
            if (!this.autoPlay || this.expandedTitle) {
                this.autoPlayApplied = true;
                this.playing = false;
                return;
            }
            const player = this.getVideoElement();
            if (!player) {
                return;
            }
            const promise = player.play();
            if (promise) {
                promise.catch(() => {
                    this.playing = false;
                    this.requiresRefresh = true;
                });
            }
            this.autoPlayApplied = true;
        },

        onAudioCanPlay: function () {
            const audioElement = this.$el.querySelector("audio");
            if (!audioElement) {
                return;
            }
            const videoElement = this.getVideoElement();
            if (!videoElement) {
                return;
            }

            const correspondingTime = Math.min(videoElement.currentTime, audioElement.duration);

            if (Math.abs(correspondingTime - audioElement.currentTime) > this.speed / 10) {
                audioElement.currentTime = Math.min(videoElement.currentTime, audioElement.duration);
            }

            if (audioElement.paused !== videoElement.paused && videoElement.currentTime <= audioElement.duration) {
                if (videoElement.paused) {
                    audioElement.pause();
                } else {
                    const promise = audioElement.play();

                    if (promise) {
                        promise.catch(() => {
                            this.pause();
                        });
                    }
                }
            }
        },

        onWaitForBuffer: function (b: boolean) {
            if (b) {
                this.isWaiting = true;
                this.waitingTimestamp = Date.now();
            } else {
                this.loading = false;
                this.isWaiting = false;
            }
        },
        onEnded: function () {
            this.loading = false;
            if (this.currentTimeSlice && this.sliceLoop) {
                this.setTime(this.currentTimeSlice.start, false);
                this.play();
                return;
            }
            if (this.canSaveTime) {
                PlayerPreferences.SetInitialTime(this.mid, 0);
            }
            if (!this.loop && !this.isShort) {
                this.pause();
                this.ended = true;
                if (this.nextEnd) {
                    this.goNext();
                }
            }
        },

        playerMouseUp: function (e) {
            if (this.timelineGrabbed) {
                if (e.touches && e.touches.length > 0) {
                    this.onTimelineSkip(e.touches[0].pageX, e.touches[0].pageY);
                } else {
                    this.onTimelineSkip(e.pageX, e.pageY);
                }
                this.timelineGrabbed = false;
            }
        },
        playerMouseMove: function (e) {
            this.interactWithControls();

            if (this.timelineGrabbed) {
                if (e.touches && e.touches.length > 0) {
                    this.onTimelineSkip(e.touches[0].pageX, e.touches[0].pageY);
                } else {
                    this.onTimelineSkip(e.pageX, e.pageY);
                }
            }
        },
        mouseLeavePlayer: function () {
            this.timelineGrabbed = false;
            if (!this.playing || this.expandedTitle || this.expandedAlbum) return;
            this.showControls = false;
            this.volumeShown = isTouchDevice();
            this.helpTooltip = "";
            this.displayConfig = false;
        },

        tick() {
            if (!this.loading && this.isWaiting && Date.now() - this.waitingTimestamp > 1000) {
                this.loading = true;
            }

            if (this.showControls && !this.mouseInControls && this.playing && !this.expandedTitle && !this.expandedAlbum) {
                if (Date.now() - this.lastControlsInteraction > 2000) {
                    this.showControls = false;
                    this.volumeShown = false;
                    this.helpTooltip = "";
                    this.displayConfig = false;
                }
            }

            const video = this.getVideoElement();

            if (video && video.buffered.length > 0) {
                this.bufferedTime = video.buffered.end(video.buffered.length - 1);
            } else {
                this.bufferedTime = 0;
            }

            if (this.tooltipShown) {
                const tooltip = this.$el.querySelector(".player-tooltip");
                if (tooltip) {
                    let x = this.tooltipEventX;
                    const toolTipWidth = tooltip.getBoundingClientRect().width;
                    const leftPlayer = this.$el.getBoundingClientRect().left;
                    const widthPlayer = this.$el.getBoundingClientRect().width;

                    x = x - Math.floor(toolTipWidth / 2);
                    if (x + toolTipWidth > leftPlayer + widthPlayer - 20) {
                        x = leftPlayer + widthPlayer - 20 - toolTipWidth;
                    }
                    if (x < leftPlayer + 10) {
                        x = leftPlayer + 10;
                    }
                    this.tooltipX = x - leftPlayer;
                }
            }
        },

        getVideoElement() {
            return this.$el.querySelector("video");
        },

        interactWithControls() {
            this.showControls = true;
            this.lastControlsInteraction = Date.now();
        },

        enterControls: function () {
            this.mouseInControls = true;
        },

        leaveControls: function () {
            this.mouseInControls = false;
            this.volumeShown = isTouchDevice();
            this.helpTooltip = "";
        },

        togglePlayImmediate() {
            if (this._handles.togglePlayDelayTimeout) {
                clearTimeout(this._handles.togglePlayDelayTimeout);
                this._handles.togglePlayDelayTimeout = null;
            }
            if (this.playing) {
                this.feedback = "pause";
                this.pause();
            } else {
                this.feedback = "play";
                this.play();
            }

            this.displayConfig = false;
        },

        togglePlay() {
            if (this.playing) {
                if (this._handles.togglePlayDelayTimeout) {
                    clearTimeout(this._handles.togglePlayDelayTimeout);
                    this._handles.togglePlayDelayTimeout = null;
                    this.feedback = "";
                } else if (PlayerPreferences.PlayerTogglePlayDelay > 0) {
                    this.feedback = "pause";
                    this._handles.togglePlayDelayTimeout = setTimeout(() => {
                        this._handles.togglePlayDelayTimeout = null;
                        this.pause();
                    }, PlayerPreferences.PlayerTogglePlayDelay);
                } else {
                    this.feedback = "pause";
                    this.pause();
                }
            } else {
                if (this._handles.togglePlayDelayTimeout) {
                    clearTimeout(this._handles.togglePlayDelayTimeout);
                    this._handles.togglePlayDelayTimeout = null;
                    this.feedback = "";
                } else if (PlayerPreferences.PlayerTogglePlayDelay > 0) {
                    this.feedback = "play";
                    this._handles.togglePlayDelayTimeout = setTimeout(() => {
                        this._handles.togglePlayDelayTimeout = null;
                        this.play();
                    }, PlayerPreferences.PlayerTogglePlayDelay);
                } else {
                    this.feedback = "play";
                    this.play();
                }
            }

            this.displayConfig = false;
        },

        play: function () {
            if (this.requiresRefresh) {
                this.requiresRefresh = false;
                MediaController.Load();
                return;
            }
            const video = this.getVideoElement();
            this.playing = true;
            if (video) {
                video.play();
            }
        },
        pause: function () {
            const video = this.getVideoElement();
            this.playing = false;

            if (video) {
                video.pause();
            }

            if (this.canSaveTime && video && !video.ended) {
                PlayerPreferences.SetInitialTime(this.mid, this.currentTime);
            }

            this.lastTimeChangedEvent = Date.now();

            this.interactWithControls();
        },

        onPlay: function () {
            this.playing = true;
        },

        onPause: function () {
            this.playing = false;
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

        clickTimeline: function (e) {
            this.displayConfig = false;
            this.contextMenuShown = false;
            e.stopPropagation();
        },

        /* Timeline */

        grabTimeline: function (e: MouseEvent & TouchEvent) {
            e.stopPropagation();
            if (e.touches && e.touches.length > 0) {
                this.timelineGrabbed = true;
                this.onTimelineSkip(e.touches[0].pageX, e.touches[0].pageY);
            } else if (e.button === 0) {
                this.timelineGrabbed = true;
                this.onTimelineSkip(e.pageX, e.pageY);
            }
        },
        getTimelineBarWidth: function (time, duration) {
            if (duration > 0) {
                return Math.min((time / duration) * 100, 100) + "%";
            } else {
                return "0";
            }
        },
        getTimelineThumbLeft: function (time, duration) {
            if (duration > 0) {
                return "calc(" + Math.min((time / duration) * 100, 100) + "% - 7px)";
            } else {
                return "-7px";
            }
        },
        onTimelineSkip: function (x: number) {
            const offset = this.$el.querySelector(".player-timeline-back").getBoundingClientRect().left;
            const width = this.$el.querySelector(".player-timeline-back").getBoundingClientRect().width || 1;
            if (x < offset) {
                this.setTime(0);
            } else {
                const p = x - offset;
                const tP = Math.min(1, p / width);
                this.setTime(tP * this.duration);
            }
        },
        mouseLeaveTimeline: function () {
            this.tooltipShown = false;
            this.leaveControls();
        },
        mouseMoveTimeline: function (event) {
            const x = event.pageX;
            const offset = this.$el.querySelector(".player-timeline-back").getBoundingClientRect().left;
            const width = this.$el.querySelector(".player-timeline-back").getBoundingClientRect().width || 1;

            let time: number;
            if (x < offset) {
                time = 0;
            } else {
                const p = x - offset;
                const tP = Math.min(1, p / width);
                time = tP * this.duration;
            }

            this.tooltipShown = true;
            this.tooltipText = this.renderTime(time);
            this.tooltipTimeSlice = this.findTimeSliceName(time);
            const oldTooltipImage = this.tooltipImage;
            this.tooltipImage = this.getThumbnailForTime(time);
            if (oldTooltipImage !== this.tooltipImage) {
                this.tooltipImageInvalid = false;
            }
            this.tooltipEventX = x;

            nextTick(this.tick.bind(this));
        },
        renderTime: function (s: number): string {
            return renderTimeSeconds(s);
        },

        setTime: function (time, save) {
            this.currentTimeSlice = null;
            time = Math.max(0, time);
            time = Math.min(time, this.duration);

            if (isNaN(time) || !isFinite(time) || time < 0) {
                return;
            }

            this.currentTime = time;

            const video = this.getVideoElement();

            if (video) {
                video.currentTime = time;
            }

            if (save && this.canSaveTime) {
                PlayerPreferences.SetInitialTime(this.mid, this.currentTime);
                this.lastTimeChangedEvent = Date.now();
            }

            if (time < this.duration) {
                this.ended = false;
            }

            this.updateSubtitles();
            this.updateCurrentTimeSlice();
        },

        onKeyPress: function (event: KeyboardEvent): boolean {
            if (AuthController.Locked || !AppStatus.IsPlayerVisible() || !event.key || event.ctrlKey) {
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
                case "M":
                case "m":
                    this.toggleMuted();
                    this.volumeShown = true;
                    this.helpTooltip = "volume";
                    break;
                case " ":
                case "K":
                case "k":
                case "Enter":
                    this.togglePlayImmediate();
                    break;
                case "ArrowUp":
                    if (!shifting) {
                        this.changeVolume(Math.min(1, this.volume + 0.05));
                    } else {
                        this.changeVolume(Math.min(1, this.volume + 0.01));
                    }
                    if (this.muted) {
                        this.muted = false;
                        this.onUserMutedUpdated();
                    }
                    this.volumeShown = true;
                    this.helpTooltip = "volume";
                    break;
                case "ArrowDown":
                    if (!shifting) {
                        this.changeVolume(Math.max(0, this.volume - 0.05));
                    } else {
                        this.changeVolume(Math.max(0, this.volume - 0.01));
                    }
                    if (this.muted) {
                        this.muted = false;
                        this.onUserMutedUpdated();
                    }
                    this.volumeShown = true;
                    this.helpTooltip = "volume";
                    break;
                case "F":
                case "f":
                    if (event.altKey || shifting) {
                        caught = false;
                    } else {
                        this.toggleFullScreen();
                    }
                    break;
                case "ArrowRight":
                    if (shifting || event.altKey) {
                        if (this.next || this.pageNext) {
                            this.goNext();
                        } else {
                            caught = false;
                        }
                    } else {
                        this.setTime(this.currentTime + 5, true);
                    }
                    break;
                case "ArrowLeft":
                    if (shifting || event.altKey) {
                        if (this.prev || this.pagePrev) {
                            this.goPrev();
                        } else {
                            caught = false;
                        }
                    } else {
                        this.setTime(this.currentTime - 5, true);
                    }
                    break;
                case ".":
                    if (!this.playing) {
                        this.setTime(this.currentTime - 1 / 30);
                    }
                    break;
                case ",":
                    if (!this.playing) {
                        this.setTime(this.currentTime + 1 / 30);
                    }
                    break;
                case "Home":
                    if (event.altKey || shifting) {
                        caught = false;
                    } else {
                        this.setTime(0, true);
                    }
                    break;
                case "End":
                    if (event.altKey || shifting) {
                        caught = false;
                    } else {
                        this.setTime(this.duration, true);
                    }
                    break;
                case "X":
                case "x":
                    this.sliceLoop = !this.sliceLoop;
                    if (this.sliceLoop) {
                        AppEvents.Emit("snack", this.$t("Slice loop enabled"));
                    } else {
                        AppEvents.Emit("snack", this.$t("Slice loop disabled"));
                    }
                    break;
                case "l":
                case "L":
                    if (event.altKey || shifting || this.isShort) {
                        caught = false;
                    } else {
                        this.loop = !this.loop;
                        if (this.loop) {
                            AppEvents.Emit("snack", this.$t("Loop enabled"));
                        } else {
                            AppEvents.Emit("snack", this.$t("Loop disabled"));
                        }
                        this.$emit("force-loop", this.loop);
                    }
                    break;
                case "b":
                case "B":
                    if (this.currentTimeSlice) {
                        this.setTime(this.currentTimeSlice.start, true);
                    }
                    break;
                case "j":
                case "J":
                    if (this.currentTimeSlice) {
                        this.setTime(this.currentTimeSlice.end, true);
                    }
                    break;
                case "C":
                case "c":
                    this.userControlsState = !this.userControlsState;
                    break;
                case "Delete":
                    this.$emit("delete");
                    break;
                default:
                    caught = false;
            }

            if (caught) {
                this.interactWithControls();
            }

            return caught;
        },

        initializeVideo() {
            if (!this.metadata) {
                return;
            }
            this.isShort = this.metadata.duration <= AUTO_LOOP_MIN_DURATION;
            this.canSaveTime = !this.metadata.force_start_beginning;
            this.hasExtendedDescription = !!this.metadata.ext_desc_url;
            this.timeSlices = normalizeTimeSlices(
                (this.metadata.time_slices || []).sort((a, b) => {
                    if (a.time < b.time) {
                        return -1;
                    } else if (a.time > b.time) {
                        return 1;
                    } else {
                        return 0;
                    }
                }),
                this.metadata.duration,
            );
            this.currentTimeSlice = null;
            this.currentTimeSliceName = "";
            this.currentTimeSliceStart = 0;
            this.currentTimeSliceEnd = 0;
            this.sliceLoop = false;
            this.currentTime = this.canSaveTime ? PlayerPreferences.GetInitialTime(this.mid) : 0;
            this.duration = 0;
            this.speed = 1;
            this.setDefaultLoop();
            this.currentResolution = PlayerPreferences.GetResolutionIndex(this.metadata);
            this.loading = true;
            this.playing = true;
            this.setVideoURL();
            this.onUpdateAudioTrack();
        },

        onClearURL: function () {
            const videoElem = this.$el.querySelector("video");
            if (videoElem) {
                videoElem.src = "";
            }
        },

        setVideoURL() {
            this.mediaError = false;
            if (!this.metadata) {
                this.videoURL = "";
                this.onClearURL();
                this.duration = 0;
                this.loading = false;
                return;
            }

            if (this.currentResolution < 0) {
                if (this.metadata.encoded) {
                    this.videoURL = GetAssetURL(this.metadata.url);
                    this.videoPending = false;
                    this.videoPendingTask = 0;
                    this.setupAutoNextTimer();
                } else {
                    this.videoURL = "";
                    this.onClearURL();
                    this.videoPending = true;
                    this.videoPendingTask = this.metadata.task;
                    this.duration = 0;
                    this.loading = false;
                }
            } else {
                if (this.metadata.resolutions && this.metadata.resolutions.length > this.currentResolution) {
                    const res = this.metadata.resolutions[this.currentResolution];
                    if (res.ready) {
                        this.videoURL = GetAssetURL(res.url);
                        this.videoPending = false;
                        this.videoPendingTask = 0;
                        this.setupAutoNextTimer();
                    } else {
                        this.videoURL = "";
                        this.onClearURL();
                        this.videoPending = true;
                        this.videoPendingTask = res.task;
                        this.duration = 0;
                        this.loading = false;
                    }
                } else {
                    this.videoURL = "";
                    this.onClearURL();
                    this.videoPending = true;
                    this.videoPendingTask = 0;
                    this.duration = 0;
                    this.loading = false;
                }
            }
        },

        onFeedBackAnimationEnd: function () {
            this.feedback = "";
        },

        onTooltipImageError: function () {
            this.tooltipImageInvalid = true;
        },

        reloadSubtitles: function () {
            this.subtitles = "";
            this.subtitlesStart = -1;
            this.subtitlesEnd = -1;
            this.updateSubtitles();
        },

        updateSubtitles: function () {
            if (this.currentTime >= this.subtitlesStart && this.currentTime <= this.subtitlesEnd) {
                return;
            }
            const sub = SubtitlesController.GetSubtitlesLine(this.currentTime);
            if (sub) {
                if (this.subtitlesHTML) {
                    const subTag = getUniqueSubtitlesLoadTag();
                    this._handles.subTag = subTag;
                    sanitizeSubtitlesHTML(sub.text).then((text) => {
                        if (this._handles.subTag === subTag) {
                            this.subtitles = text;
                        }
                    });
                } else {
                    this._handles.subTag = "";
                    this.subtitles = htmlToText(sub.text);
                }
                this.subtitlesStart = sub.start;
                this.subtitlesEnd = sub.end;
            } else {
                this._handles.subTag = "";
                this.subtitles = "";
                this.subtitlesStart = 0;
                this.subtitlesEnd = 0;
            }
        },

        findTimeSliceName: function (time: number) {
            const slice = findTimeSlice(this.timeSlices, time);
            if (slice) {
                return slice.name;
            } else {
                return "";
            }
        },

        updateCurrentTimeSlice: function () {
            if (this.currentTimeSlice && this.sliceLoop && this.currentTime >= this.currentTimeSlice.end) {
                this.setTime(this.currentTimeSlice.start, false);
                return;
            }
            const slice = findTimeSlice(this.timeSlices, this.currentTime);
            if (slice) {
                this.currentTimeSlice = slice;
                this.currentTimeSliceName = slice.name;
                this.currentTimeSliceStart = slice.start;
                this.currentTimeSliceEnd = slice.end;
            } else {
                this.currentTimeSlice = null;
                this.currentTimeSliceName = "";
                this.currentTimeSliceStart = 0;
                this.currentTimeSliceEnd = 0;
            }
        },

        onUpdateAudioTrack: function () {
            if (!this.audioTrack || !this.metadata || !this.metadata.audios) {
                this.audioTrackURL = "";
                return;
            }

            for (const audio of this.metadata.audios) {
                if (audio.id === this.audioTrack) {
                    this.audioTrackURL = GetAssetURL(audio.url);
                    return;
                }
            }

            this.audioTrackURL = "";
        },

        onUpdateSubHTML: function () {
            PlayerPreferences.SetSubtitlesHTML(this.subtitlesHTML);
            this.reloadSubtitles();
        },

        onUpdateNextEnd: function () {
            PlayerPreferences.SetNextOnEnd(this.nextEnd);
        },

        handleMediaSessionEvent: function (event: { action: string; fastSeek: boolean; seekTime: number; seekOffset: number }) {
            if (!event || !event.action) {
                return;
            }
            switch (event.action) {
                case "play":
                    this.play();
                    break;
                case "pause":
                    this.pause();
                    break;
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

        onMediaError: function () {
            if (!AuthController.RefreshAuthStatus()) {
                this.mediaError = true;
                this.loading = false;
                AuthController.CheckAuthStatusSilent();
            }
        },

        setDefaultLoop: function () {
            if (this.loopForced) {
                this.loop = this.loopForcedValue;
            } else {
                this.loop = (!this.next && !this.pageNext) || !this.nextEnd;
            }
        },

        onTopBarExpand: function () {
            if (this.expandedTitle) {
                this.pause();
            }
        },

        setupAutoNextTimer: function () {
            if (this._handles.autoNextTimer) {
                clearTimeout(this._handles.autoNextTimer);
                this._handles.autoNextTimer = null;
            }

            if (!this.isShort) {
                return;
            }

            const timerS = PlayerPreferences.ImageAutoNext;

            if (isNaN(timerS) || !isFinite(timerS) || timerS <= 0) {
                return;
            }

            if (!this.next && !this.pageNext) {
                return;
            }

            const ms = timerS * 1000;

            this._handles.autoNextTimer = setTimeout(() => {
                this._handles.autoNextTimer = null;
                if (this.displayConfig || this.expandedTitle) {
                    this.setupAutoNextTimer();
                } else {
                    this.goNext();
                }
            }, ms);
        },

    },
    mounted: function () {
        this._handles = Object.create(null);

        // Load player preferences
        this.muted = PlayerPreferences.PlayerMuted;
        this.volume = PlayerPreferences.PlayerVolume;
        this.subtitlesSize = PlayerPreferences.SubtitlesSize;
        this.subtitlesBg = PlayerPreferences.SubtitlesBackground;
        this.subtitlesHTML = PlayerPreferences.SubtitlesHTML;
        this.nextEnd = PlayerPreferences.NextOnEnd;

        this._handles.keyHandler = this.onKeyPress.bind(this);
        KeyboardManager.AddHandler(this._handles.keyHandler, 100);

        this._handles.timer = setInterval(this.tick.bind(this), 100);

        this._handles.exitFullScreenListener = this.onExitFullScreen.bind(this);
        document.addEventListener("fullscreenchange", this._handles.exitFullScreenListener);
        document.addEventListener("webkitfullscreenchange", this._handles.exitFullScreenListener);
        document.addEventListener("mozfullscreenchange", this._handles.exitFullScreenListener);
        document.addEventListener("MSFullscreenChange", this._handles.exitFullScreenListener);

        this._handles.subtitlesReloadH = this.reloadSubtitles.bind(this);
        AppEvents.AddEventListener("subtitles-update", this._handles.subtitlesReloadH);

        this.initializeVideo();

        if (window.navigator && window.navigator.mediaSession) {
            navigator.mediaSession.setActionHandler("play", this.handleMediaSessionEvent.bind(this));
            navigator.mediaSession.setActionHandler("pause", this.handleMediaSessionEvent.bind(this));
            navigator.mediaSession.setActionHandler("nexttrack", this.handleMediaSessionEvent.bind(this));
            navigator.mediaSession.setActionHandler("previoustrack", this.handleMediaSessionEvent.bind(this));
        }
    },
    beforeUnmount: function () {
        this.videoURL = "";
        this.onClearURL();

        clearInterval(this._handles.timer);

        if (this._handles.togglePlayDelayTimeout) {
            clearTimeout(this._handles.togglePlayDelayTimeout);
            this._handles.togglePlayDelayTimeout = null;
        }

        if (this._handles.autoNextTimer) {
            clearTimeout(this._handles.autoNextTimer);
            this._handles.autoNextTimer = null;
        }

        document.removeEventListener("fullscreenchange", this._handles.exitFullScreenListener);
        document.removeEventListener("webkitfullscreenchange", this._handles.exitFullScreenListener);
        document.removeEventListener("mozfullscreenchange", this._handles.exitFullScreenListener);
        document.removeEventListener("MSFullscreenChange", this._handles.exitFullScreenListener);

        AppEvents.RemoveEventListener("subtitles-update", this._handles.subtitlesReloadH);
        KeyboardManager.RemoveHandler(this._handles.keyHandler);

        if (window.navigator && window.navigator.mediaSession) {
            navigator.mediaSession.setActionHandler("play", null);
            navigator.mediaSession.setActionHandler("pause", null);
            navigator.mediaSession.setActionHandler("nexttrack", null);
            navigator.mediaSession.setActionHandler("previoustrack", null);
        }
    },
    watch: {
        rTick: function () {
            this.internalTick++;
            this.expandedTitle = false;
            this.subtitles = "";
            this.subtitlesStart = -1;
            this.subtitlesEnd = -1;
            this.autoPlayApplied = false;
            this.initializeVideo();
        },
        videoURL: function () {
            if (this.videoURL) {
                this.loading = true;
            }
        },
        next: function () {
            this.setDefaultLoop();
            this.setupAutoNextTimer();
        },
        pageNext: function () {
            this.setDefaultLoop();
            this.setupAutoNextTimer();
        },
    },
});
</script>
