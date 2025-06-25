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
        @touchmove="playerTouchMove"
        @touchend.passive="onPlayerTouchEnd"
        @contextmenu="onContextMenu"
        @wheel="onMouseWheel"
    >
        <div class="video-scroller" @mousedown="grabScroll" @touchstart="onScrollerTouchStart">
            <video
                v-if="videoURL"
                :key="rTick"
                class="player-video-element"
                :src="videoURL"
                crossorigin="use-credentials"
                playsinline
                webkit-playsinline
                x-webkit-airplay="allow"
                :muted="muted"
                :loop="(loop || isShort) && !sliceLoop"
                :volume.prop="audioTrackURL ? 0 : volume"
                :playbackRate.prop="speed"
                :style="{ width: getScaleCss(scale), height: getScaleCss(scale) }"
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
        </div>

        <audio
            v-if="audioTrackURL"
            :key="rTick"
            :src="audioTrackURL"
            crossorigin="use-credentials"
            playsinline
            webkit-playsinline
            :muted="muted || !audioTrackURL"
            :volume.prop="volume"
            :playbackRate.prop="speed"
            @loadedmetadata="onAudioLoadMetadata"
            @canplay="onAudioCanPlay"
        ></audio>

        <div class="player-feedback-container">
            <div v-if="feedback === 'play'" key="play" class="player-feedback player-feedback-play" @animationend="onFeedBackAnimationEnd">
                <div><i class="fas fa-play"></i></div>
            </div>
            <div
                v-if="feedback === 'pause'"
                key="pause"
                class="player-feedback player-feedback-pause"
                @animationend="onFeedBackAnimationEnd"
            >
                <div><i class="fas fa-pause"></i></div>
            </div>
        </div>

        <div v-if="loading && !mediaError" class="player-loader">
            <div class="player-lds-ring">
                <div></div>
                <div></div>
                <div></div>
                <div></div>
            </div>
        </div>

        <div
            v-if="pendingNextEnd"
            class="player-auto-next-overlay"
            @click="stopPropagationEvent"
            @mousedown="stopPropagationEvent"
            @touchstart="stopPropagationEvent"
        >
            <div class="next-end-container">
                <div class="next-end-wait-msg">
                    {{ $t("Next media will play in") }} <b>{{ pendingNextEndSeconds }}</b>
                </div>
                <div class="next-end-wait-buttons">
                    <button type="button" class="btn btn-primary" @click="hideNextEnd">
                        <i class="fas fa-times"></i> {{ $t("Cancel") }}
                    </button>
                    <button type="button" class="btn btn-primary" @click="goNext">
                        <i class="fas fa-forward-step"></i> {{ $t("Next") }}
                    </button>
                </div>
                <div class="next-end-wait-buttons">
                    <button type="button" class="btn btn-primary" @click="play">
                        <i class="fas fa-repeat"></i> {{ $t("Play again") }}
                    </button>
                    <button type="button" class="btn btn-primary" @click="enableLoopAndPlay">
                        <i class="fas fa-repeat"></i> {{ $t("Loop") }}
                    </button>
                </div>
            </div>
        </div>

        <PlayerEncodingPending
            v-if="(!loading && !videoURL && videoPending) || mediaError"
            :mid="mid"
            :tid="videoPendingTask"
            :res="currentResolution"
            :error="mediaError"
            :error-message="mediaErrorMessage"
            :can-auto-reload="!expandedTitle && !expandedAlbum"
        ></PlayerEncodingPending>

        <PlayerSubtitles :show-controls="showControls && userControls" :subtitles="subtitles"></PlayerSubtitles>

        <TimeSlicesEditHelper
            v-if="timeSlicesEdit"
            v-model:display="timeSlicesEdit"
            :current-time="currentTime"
            :context-open="contextMenuShown"
            @update-time-slices="refreshTimeSlices"
            @clicked="clickControls"
        ></TimeSlicesEditHelper>

        <TagsEditHelper
            v-if="displayTagList"
            v-model:display="displayTagListStatus"
            :context-open="contextMenuShown"
            @clicked="clickControls"
        ></TagsEditHelper>

        <ExtendedDescriptionWidget
            v-if="displayExtendedDescription"
            v-model:display="displayExtendedDescriptionStatus"
            :context-open="contextMenuShown"
            @clicked="clickControls"
            @update-ext-desc="refreshExtendedDescription"
        ></ExtendedDescriptionWidget>

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
                    class="player-btn player-btn-action-prev"
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
                    class="player-btn player-btn-action-next"
                    @click="goNext"
                    @mouseenter="enterTooltip('next')"
                    @mouseleave="leaveTooltip('next')"
                >
                    <i class="fas fa-forward-step"></i>
                </button>

                <VolumeControl
                    ref="volumeControl"
                    v-model:muted="muted"
                    v-model:volume="volume"
                    v-model:expanded="volumeShown"
                    :min="min"
                    :width="min ? 50 : 80"
                    @update:volume="onUserVolumeUpdated"
                    @update:muted="onUserMutedUpdated"
                    @enter="enterTooltip('volume')"
                    @leave="leaveTooltip('volume')"
                ></VolumeControl>

                <div v-if="!min" class="player-time-label-container" :class="{ 'in-album': !!next || !!prev }">
                    <span>{{ renderTime(currentTime) }} / {{ renderTime(duration) }}</span>
                    <span v-if="currentTimeSlice" class="times-slice-name"><b class="separator"> - </b>{{ currentTimeSliceName }}</span>
                </div>
            </div>

            <div class="player-controls-right">
                <button
                    v-if="hasExtendedDescription"
                    type="button"
                    :title="$t('Extended description')"
                    class="player-btn"
                    @click="openExtendedDescription"
                    @mouseenter="enterTooltip('ext-desc')"
                    @mouseleave="leaveTooltip('ext-desc')"
                >
                    <i class="fas fa-file-lines"></i>
                </button>

                <button
                    v-if="hasAttachments"
                    type="button"
                    :title="$t('Attachments')"
                    class="player-btn player-settings-no-trap"
                    @click="showAttachments"
                    @mouseenter="enterTooltip('attachments')"
                    @mouseleave="leaveTooltip('attachments')"
                >
                    <i class="fas fa-paperclip"></i>
                </button>

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

        <div v-else-if="helpTooltip === 'pause'" class="player-tooltip player-help-tip-left">
            {{ $t("Pause") }}
        </div>

        <div v-else-if="!prev && pagePrev && helpTooltip === 'prev'" class="player-tooltip player-help-tip-left">
            {{ $t("Previous") }}
        </div>

        <div v-else-if="!next && pageNext && helpTooltip === 'next'" class="player-tooltip player-help-tip-left">
            {{ $t("Next") }}
        </div>

        <div v-else-if="prev && helpTooltip === 'prev'" class="player-tooltip player-help-tip-left">
            <PlayerMediaChangePreview :media="prev" :next="false"></PlayerMediaChangePreview>
        </div>

        <div v-else-if="next && helpTooltip === 'next'" class="player-tooltip player-help-tip-left">
            <PlayerMediaChangePreview :media="next" :next="true"></PlayerMediaChangePreview>
        </div>

        <div v-else-if="helpTooltip === 'volume'" class="player-tooltip player-help-tip-left">
            {{ $t("Volume") }} ({{ muted ? $t("Muted") : renderVolume(volume) }})
        </div>

        <div v-else-if="!displayConfig && !displayAttachments && helpTooltip === 'ext-desc'" class="player-tooltip player-help-tip-right">
            {{ $t("Extended description") }}
        </div>

        <div
            v-else-if="!displayConfig && !displayAttachments && helpTooltip === 'attachments'"
            class="player-tooltip player-help-tip-right"
        >
            {{ $t("Attachments") }}
        </div>

        <div v-else-if="!displayConfig && !displayAttachments && helpTooltip === 'config'" class="player-tooltip player-help-tip-right">
            {{ $t("Player Configuration") }}
        </div>

        <div v-else-if="!displayConfig && !displayAttachments && helpTooltip === 'albums'" class="player-tooltip player-help-tip-right">
            {{ $t("Manage albums") }}
        </div>

        <div
            v-else-if="!displayConfig && !displayAttachments && helpTooltip === 'full-screen'"
            class="player-tooltip player-help-tip-right"
        >
            {{ $t("Full screen") }}
        </div>

        <div
            v-else-if="!displayConfig && !displayAttachments && helpTooltip === 'full-screen-exit'"
            class="player-tooltip player-help-tip-right"
        >
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
            @mousedown="grabTimelineByMouse"
            @touchstart="grabTimelineByTouch"
        >
            <div class="player-timeline-back"></div>
            <div class="player-timeline-buffer" :style="{ width: getTimelineBarWidth(bufferedTime, duration) }"></div>
            <div class="player-timeline-current" :style="{ width: getTimelineBarWidth(currentTime, duration) }"></div>

            <div
                v-for="(ts, tsi) in timeSlices"
                :key="tsi"
                class="player-timeline-split"
                :class="{ 'start-split': ts.start <= 0 }"
                :style="{ left: getTimelineBarWidth(ts.start, duration) }"
            ></div>

            <div class="player-timeline-thumb" :style="{ left: getTimelineThumbLeft(currentTime, duration) }"></div>
        </div>

        <div v-if="tooltipShown" class="player-tooltip" :style="{ left: tooltipX + 'px' }">
            <div v-if="tooltipImage && !tooltipImageInvalid" class="player-tooltip-image">
                <img :src="tooltipImage" @error="onTooltipImageError" @load="onTooltipImageLoaded" />
                <div v-if="tooltipImageLoading" class="player-tooltip-image-loading">
                    <div class="player-tooltip-image-loader">
                        <div></div>
                        <div></div>
                        <div></div>
                        <div></div>
                    </div>
                </div>
            </div>
            <div class="player-tooltip-text">{{ tooltipText }}</div>
            <div v-if="tooltipTimeSlice" class="player-tooltip-text">
                {{ tooltipTimeSlice }}
            </div>
        </div>

        <VideoPlayerConfig
            v-model:shown="displayConfig"
            v-model:speed="speed"
            v-model:scale="scale"
            v-model:loop="loop"
            v-model:next-end="nextEnd"
            v-model:auto-next-page-delay="autoNextPageDelay"
            v-model:resolution="currentResolution"
            v-model:audio-track="audioTrack"
            :r-tick="internalTick"
            :metadata="metadata"
            :is-short="isShort"
            :in-album="inAlbum"
            @update:scale="onScaleUpdated"
            @update:loop="() => $emit('force-loop', loop)"
            @update:resolution="onResolutionUpdated"
            @update:next-end="onUpdateNextEnd"
            @update:auto-next-page-delay="onUpdateAutoNextPageDelay"
            @enter="enterControls"
            @leave="leaveControls"
            @update:audio-track="onUpdateAudioTrack"
            @update-auto-next="setupAutoNextTimer"
        ></VideoPlayerConfig>

        <PlayerAttachmentsList
            v-if="metadata && metadata.attachments"
            v-model:shown="displayAttachments"
            :attachments="metadata.attachments"
            @enter="enterControls"
            @leave="leaveControls"
        >
        </PlayerAttachmentsList>

        <PlayerTopBar
            v-if="metadata"
            v-model:expanded="expandedTitle"
            v-model:album-expanded="expandedAlbum"
            :mid="mid"
            :metadata="metadata"
            :shown="showControls && userControls"
            :fullscreen="fullscreen"
            :in-album="inAlbum"
            @update:expanded="onTopBarExpand"
            @click-player="clickControls"
        ></PlayerTopBar>

        <PlayerContextMenu
            v-model:shown="contextMenuShown"
            v-model:loop="loop"
            v-model:slice-loop="sliceLoop"
            v-model:controls="userControlsState"
            v-model:time-slices-edit="timeSlicesEdit"
            type="video"
            :x="contextMenuX"
            :y="contextMenuY"
            :url="videoURL"
            :title="title"
            :can-write="canWrite"
            :has-extended-description="hasExtendedDescription"
            :has-slices="timeSlices && timeSlices.length > 0"
            :is-short="isShort"
            @update:loop="() => $emit('force-loop', loop)"
            @stats="openStats"
            @open-tags="openTags"
            @open-ext-desc="openExtendedDescription"
        ></PlayerContextMenu>
    </div>
</template>

<script lang="ts">
import {
    CURRENT_TIME_UPDATE_DELAY,
    getAutoNextOnEnd,
    getAutoNextPageDelay,
    getAutoNextTime,
    getCachedInitialTime,
    getPlayerMuted,
    getPlayerVolume,
    getSelectedAudioTrack,
    getTogglePlayDelay,
    getUserSelectedResolutionVideo,
    setAutoNextOnEnd,
    setAutoNextPageDelay,
    setCachedInitialTime,
    setPlayerMuted,
    setPlayerVolume,
    setUserSelectedResolutionVideo,
} from "@/control/player-preferences";
import { defineAsyncComponent, defineComponent, nextTick } from "vue";

import VolumeControl from "./VolumeControl.vue";
import PlayerTopBar from "./PlayerTopBar.vue";
import PlayerMediaChangePreview from "./PlayerMediaChangePreview.vue";
import { openFullscreen, closeFullscreen } from "../../utils/full-screen";
import { renderTimeSeconds } from "../../utils/time";
import type { NormalizedTimeSlice } from "../../utils/time-slices";
import { findTimeSlice, normalizeTimeSlices } from "../../utils/time-slices";
import { isTouchDevice } from "@/utils/touch";
import VideoPlayerConfig from "./VideoPlayerConfig.vue";
import PlayerAttachmentsList from "./PlayerAttachmentsList.vue";
import PlayerContextMenu from "./PlayerContextMenu.vue";
import PlayerSubtitles from "./PlayerSubtitles.vue";
import { getAssetURL } from "@/utils/api";
import { useVModel } from "../../utils/v-model";
import { AUTO_LOOP_MIN_DURATION, MediaController, NEXT_END_WAIT_DURATION } from "@/control/media";
import { EVENT_NAME_SUBTITLES_UPDATE, SubtitlesController } from "@/control/subtitles";
import { AppStatus } from "@/control/app-status";
import { AuthController } from "@/control/auth";
import type { PropType } from "vue";
import type { MediaData, MediaListItem } from "@/api/models";
import { PagesController } from "@/control/pages";
import { getUniqueStringId } from "@/utils/unique-id";
import { addMediaSessionActionHandler, clearMediaSessionActionHandlers } from "@/utils/media-session";

const TimeSlicesEditHelper = defineAsyncComponent({
    loader: () => import("@/components/player/TimeSlicesEditHelper.vue"),
});

const PlayerEncodingPending = defineAsyncComponent({
    loader: () => import("@/components/player/PlayerEncodingPending.vue"),
});

const TagsEditHelper = defineAsyncComponent({
    loader: () => import("@/components/player/TagsEditHelper.vue"),
});

const ExtendedDescriptionWidget = defineAsyncComponent({
    loader: () => import("@/components/player/ExtendedDescriptionWidget.vue"),
});

export default defineComponent({
    name: "VideoPlayer",
    components: {
        VolumeControl,
        VideoPlayerConfig,
        PlayerMediaChangePreview,
        PlayerTopBar,
        PlayerContextMenu,
        PlayerEncodingPending,
        TimeSlicesEditHelper,
        TagsEditHelper,
        ExtendedDescriptionWidget,
        PlayerAttachmentsList,
        PlayerSubtitles,
    },
    props: {
        mid: Number,
        metadata: Object as PropType<MediaData>,
        rTick: Number,

        fullscreen: Boolean,

        next: Object as PropType<MediaListItem | null>,
        prev: Object as PropType<MediaListItem | null>,
        inAlbum: Boolean,

        pageNext: Boolean,
        pagePrev: Boolean,

        canWrite: Boolean,

        userControls: Boolean,

        min: Boolean,

        loopForced: Boolean,
        loopForcedValue: Boolean,

        autoPlay: Boolean,

        displayTagList: Boolean,
        displayExtendedDescription: Boolean,
    },
    emits: [
        "go-next",
        "go-prev",
        "ended",
        "update:fullscreen",
        "albums-open",
        "stats-open",
        "force-loop",
        "delete",
        "update:displayTagList",
        "update:displayExtendedDescription",
    ],
    setup(props) {
        return {
            togglePlayDelayTimeout: null as ReturnType<typeof setTimeout> | null,
            autoNextTimer: null as ReturnType<typeof setInterval> | null,
            subTag: "",
            nextEndTimer: null as ReturnType<typeof setInterval> | null,
            timer: null as ReturnType<typeof setInterval> | null,
            mediaSessionId: getUniqueStringId(),
            fullScreenState: useVModel(props, "fullscreen"),
            userControlsState: useVModel(props, "userControls"),
            displayTagListStatus: useVModel(props, "displayTagList"),
            displayExtendedDescriptionStatus: useVModel(props, "displayExtendedDescription"),
        };
    },
    data: function () {
        return {
            playing: false,
            loading: true,

            title: "",

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
            tooltipImageLoading: false,
            tooltipImageInvalid: false,

            showControls: true,
            lastControlsInteraction: Date.now(),
            mouseInControls: false,

            loop: false,
            nextEnd: false,
            autoNextPageDelay: false,
            sliceLoop: false,

            isShort: false,

            currentResolution: -1,

            volume: 1,
            muted: false,
            volumeShown: isTouchDevice(),
            internalTick: 0,

            speed: 1,
            scale: 1,

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

            audioTrack: getSelectedAudioTrack(),
            audioTrackURL: "",

            timeSlices: [] as NormalizedTimeSlice[],
            currentTimeSlice: null as NormalizedTimeSlice | null,
            currentTimeSliceName: "",
            currentTimeSliceStart: 0,
            currentTimeSliceEnd: 0,

            timeSlicesEdit: false,

            waitingTimestamp: 0,
            isWaiting: false,

            mediaError: false,
            mediaErrorMessage: "",

            hasExtendedDescription: false,

            hasAttachments: false,
            displayAttachments: false,

            pendingNextEnd: false,
            pendingNextEndSeconds: 0,

            scrollGrabbed: false,
            scrollGrabX: 0,
            scrollGrabY: 0,
            scrollGrabTop: 0,
            scrollGrabLeft: 0,
            scrollMoved: false,

            timeStartTap: 0,
        };
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
            if (!this.next && !this.pageNext) {
                this.hideNextEnd();
            }
        },
        pageNext: function () {
            this.setDefaultLoop();
            this.setupAutoNextTimer();
            if (!this.next && !this.pageNext) {
                this.hideNextEnd();
            }
        },
    },
    mounted: function () {
        // Load player preferences
        this.muted = getPlayerMuted();
        this.volume = getPlayerVolume();
        this.nextEnd = getAutoNextOnEnd();
        this.autoNextPageDelay = getAutoNextPageDelay();

        this.$addKeyboardHandler(this.onKeyPress.bind(this), 100);

        this.timer = setInterval(this.tick.bind(this), 100);

        this.$listenOnDocumentEvent("fullscreenchange", this.onExitFullScreen.bind(this));

        this.$listenOnAppEvent(EVENT_NAME_SUBTITLES_UPDATE, this.reloadSubtitles.bind(this));

        this.$listenOnDocumentEvent("mouseup", this.dropScroll.bind(this));

        this.$listenOnDocumentEvent("mousemove", this.moveScroll.bind(this));

        this.initializeVideo();

        if (window.navigator && window.navigator.mediaSession) {
            MediaController.MediaSessionId = this.mediaSessionId;
            clearMediaSessionActionHandlers();

            addMediaSessionActionHandler(
                ["play", "pause", "nexttrack", "previoustrack", "seekbackward", "seekforward", "seekto"],
                this.handleMediaSessionEvent.bind(this),
            );

            this.updateMediaSessionPlaybackState();
        }
    },
    beforeUnmount: function () {
        this.videoURL = "";
        this.onClearURL();

        clearInterval(this.timer);

        if (this.togglePlayDelayTimeout) {
            clearTimeout(this.togglePlayDelayTimeout);
            this.togglePlayDelayTimeout = null;
        }

        if (this.autoNextTimer) {
            clearTimeout(this.autoNextTimer);
            this.autoNextTimer = null;
        }

        if (this.nextEndTimer) {
            clearTimeout(this.nextEndTimer);
            this.nextEndTimer = null;
        }

        if (window.navigator && window.navigator.mediaSession && MediaController.MediaSessionId === this.mediaSessionId) {
            clearMediaSessionActionHandlers();
            navigator.mediaSession.playbackState = "none";
            MediaController.MediaSessionId = "";
        }
    },
    methods: {
        onContextMenu: function (e: MouseEvent) {
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
            this.displayTagListStatus = true;
        },

        openExtendedDescription: function () {
            if (!this.hasExtendedDescription && !this.canWrite) {
                return;
            }
            this.displayExtendedDescriptionStatus = true;
        },

        showAttachments: function (e?: Event) {
            if (e) {
                e.stopPropagation();
            }
            this.displayAttachments = !this.displayAttachments;
            this.displayConfig = false;
        },

        onPlayerMouseDown: function (e: MouseEvent) {
            if (isTouchDevice()) {
                return;
            }
            if (e.button !== 0) {
                return;
            }
            if (this.contextMenuShown) {
                e.stopPropagation();
            }
            this.clickPlayer();
        },

        onPlayerTouchStart: function (e: TouchEvent) {
            if (this.contextMenuShown) {
                e.stopPropagation();
            }
            this.leaveControls();
            if (this.displayConfig || this.contextMenuShown || this.displayAttachments) {
                this.displayConfig = false;
                this.contextMenuShown = false;
                this.displayAttachments = false;
            } else {
                this.timeStartTap = Date.now();
            }
            this.interactWithControls();
        },

        onPlayerTouchEnd: function (e: TouchEvent) {
            if (this.timelineGrabbed) {
                this.timelineGrabbed = false;
                if (e.touches[0]) {
                    this.onTimelineSkip(e.touches[0].pageX);
                }
            }

            this.tooltipShown = false;

            if (this.timeStartTap && Date.now() - this.timeStartTap < 500) {
                this.togglePlay();
                this.timeStartTap = 0;
            }
        },

        clickPlayer: function () {
            this.leaveControls();
            if (this.displayConfig || this.contextMenuShown || this.displayAttachments) {
                this.displayConfig = false;
                this.contextMenuShown = false;
                this.displayAttachments = false;
            } else {
                this.togglePlay();
            }
            this.interactWithControls();
        },

        renderVolume: function (v: number): string {
            return Math.round(v * 100) + "%";
        },

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

        showConfig: function (e?: Event) {
            this.displayConfig = !this.displayConfig;
            this.displayAttachments = false;
            if (e) {
                e.stopPropagation();
            }
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

            return getAssetURL(this.metadata.video_previews.replace("{INDEX}", "" + part));
        },

        onResolutionUpdated: function () {
            setUserSelectedResolutionVideo(this.metadata, this.currentResolution);
            this.setVideoURL();
        },

        clickControls: function (e: Event) {
            this.displayConfig = false;
            this.contextMenuShown = false;
            this.displayAttachments = false;
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
            setPlayerVolume(this.volume);
        },

        changeVolume: function (v: number) {
            this.volume = v;
            this.onUserVolumeUpdated();
        },

        onUserMutedUpdated() {
            setPlayerMuted(this.muted);
        },

        toggleMuted: function () {
            this.muted = !this.muted;
            this.onUserMutedUpdated();
        },

        /* Player events */

        onAudioLoadMetadata: function () {
            const audioElement = this.$el.querySelector("audio") as HTMLAudioElement;
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

        onLoadMetaData: function (ev: Event) {
            const videoElement = ev.target as HTMLVideoElement;

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
        onVideoTimeUpdate: function (ev: Event) {
            this.hideNextEnd();
            if (this.loading) return;
            const videoElement = ev.target as HTMLVideoElement;
            if (
                !videoElement ||
                typeof videoElement.currentTime !== "number" ||
                isNaN(videoElement.currentTime) ||
                !isFinite(videoElement.currentTime) ||
                typeof videoElement.duration !== "number" ||
                isNaN(videoElement.duration) ||
                !isFinite(videoElement.duration)
            ) {
                return;
            }
            this.currentTime = videoElement.currentTime;
            this.duration = videoElement.duration;
            if (!this.loading && this.canSaveTime && Date.now() - this.lastTimeChangedEvent > CURRENT_TIME_UPDATE_DELAY) {
                setCachedInitialTime(this.mid, this.currentTime);
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
                if (this.playing) {
                    const player = this.getVideoElement();
                    if (!player) {
                        return;
                    }
                    const promise = player.play();
                    if (promise) {
                        promise.catch(() => {
                            this.playing = false;
                            this.requiresRefresh = true;
                            this.updateMediaSessionPlaybackState();
                        });
                    }
                }
                return;
            }
            if (!this.autoPlay || this.expandedTitle) {
                this.autoPlayApplied = true;
                this.playing = false;
                this.updateMediaSessionPlaybackState();
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
                    this.updateMediaSessionPlaybackState();
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
                setCachedInitialTime(this.mid, 0);
            }
            if (!this.loop && !this.isShort) {
                this.pause();
                this.ended = true;
                if (this.nextEnd) {
                    if (this.next) {
                        this.goNext();
                    } else if (this.pageNext) {
                        if (this.autoNextPageDelay) {
                            this.showNextEnd();
                        } else {
                            this.goNext();
                        }
                    }
                }
            }
        },

        playerMouseUp: function (e: MouseEvent) {
            if (this.timelineGrabbed) {
                this.timelineGrabbed = false;
                this.onTimelineSkip(e.pageX);

                if (isTouchDevice()) {
                    this.tooltipShown = false;
                }
            }
        },

        playerMouseMove: function (e: MouseEvent) {
            this.interactWithControls();

            if (this.timelineGrabbed) {
                this.onTimelineSkip(e.pageX);
            }
        },

        playerTouchMove: function (e: TouchEvent) {
            this.interactWithControls();

            if (this.timelineGrabbed && e.touches[0]) {
                this.onTimelineSkip(e.touches[0].pageX);
                this.updateTimelineTooltip(e.touches[0].pageX);
            }
        },

        mouseLeavePlayer: function () {
            this.timelineGrabbed = false;
            if (!this.playing || this.expandedTitle || this.expandedAlbum || this.displayConfig || this.displayAttachments) return;
            this.showControls = false;
            this.volumeShown = isTouchDevice();
            this.helpTooltip = "";
        },

        tick() {
            if (!this.loading && this.isWaiting && Date.now() - this.waitingTimestamp > 1000) {
                this.loading = true;
            }

            if (this.showControls && !this.mouseInControls && this.playing && !this.expandedTitle && !this.expandedAlbum) {
                if (Date.now() - this.lastControlsInteraction > 2000) {
                    this.showControls = false;
                    this.volumeShown = isTouchDevice();
                    this.helpTooltip = "";
                    this.displayConfig = false;
                    this.displayAttachments = false;
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
            if (this.togglePlayDelayTimeout) {
                clearTimeout(this.togglePlayDelayTimeout);
                this.togglePlayDelayTimeout = null;
            }
            if (this.playing) {
                this.feedback = "pause";
                this.pause();
            } else {
                this.feedback = "play";
                this.play();
            }

            this.displayConfig = false;
            this.displayAttachments = false;
        },

        togglePlay() {
            const delay = getTogglePlayDelay();
            if (this.playing) {
                if (this.togglePlayDelayTimeout) {
                    clearTimeout(this.togglePlayDelayTimeout);
                    this.togglePlayDelayTimeout = null;
                    this.feedback = "";
                } else if (delay > 0) {
                    this.feedback = "pause";
                    this.togglePlayDelayTimeout = setTimeout(() => {
                        this.togglePlayDelayTimeout = null;
                        this.pause();
                    }, delay);
                } else {
                    this.feedback = "pause";
                    this.pause();
                }
            } else {
                if (this.togglePlayDelayTimeout) {
                    clearTimeout(this.togglePlayDelayTimeout);
                    this.togglePlayDelayTimeout = null;
                    this.feedback = "";
                } else if (delay > 0) {
                    this.feedback = "play";
                    this.togglePlayDelayTimeout = setTimeout(() => {
                        this.togglePlayDelayTimeout = null;
                        this.play();
                    }, delay);
                } else {
                    this.feedback = "play";
                    this.play();
                }
            }

            this.displayConfig = false;
            this.displayAttachments = false;
        },

        play: function () {
            this.hideNextEnd();
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
            this.updateMediaSessionPlaybackState();
        },
        pause: function () {
            const video = this.getVideoElement();
            this.playing = false;

            if (video) {
                video.pause();
            }

            this.updateMediaSessionPlaybackState();

            if (!this.loading && this.canSaveTime && video && !video.ended) {
                setCachedInitialTime(this.mid, this.currentTime);
            }

            this.lastTimeChangedEvent = Date.now();

            this.interactWithControls();
        },

        onPlay: function () {
            this.playing = true;
            this.updateMediaSessionPlaybackState();
        },

        onPause: function () {
            this.playing = false;
            this.updateMediaSessionPlaybackState();
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

        clickTimeline: function (e: Event) {
            this.displayConfig = false;
            this.contextMenuShown = false;
            this.displayAttachments = false;
            e.stopPropagation();
        },

        /* Timeline */

        grabTimelineByMouse: function (e: MouseEvent) {
            e.stopPropagation();
            if (e.button === 0) {
                this.timelineGrabbed = true;
                this.onTimelineSkip(e.pageX);
            }
        },

        grabTimelineByTouch: function (e: TouchEvent) {
            e.stopPropagation();

            if (!e.touches[0]) {
                return;
            }

            this.timelineGrabbed = true;
            this.onTimelineSkip(e.touches[0].pageX);

            this.updateTimelineTooltip(e.touches[0].pageX);
        },

        updateTimelineTooltip: function (x: number) {
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
                this.tooltipImageLoading = true;
            }
            this.tooltipEventX = x;

            nextTick(this.tick.bind(this));
        },

        getTimelineBarWidth: function (time: number, duration: number) {
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
            if (!isTouchDevice()) {
                this.tooltipShown = false;
            }
            this.leaveControls();
        },
        mouseMoveTimeline: function (event: MouseEvent) {
            if (isTouchDevice()) {
                return;
            }

            this.updateTimelineTooltip(event.pageX);
        },
        renderTime: function (s: number): string {
            return renderTimeSeconds(s);
        },

        setTime: function (time: number, save?: boolean) {
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
                setCachedInitialTime(this.mid, this.currentTime);
                this.lastTimeChangedEvent = Date.now();
            }

            if (time < this.duration) {
                this.ended = false;
            }

            this.updateSubtitles();
            this.updateCurrentTimeSlice();
        },

        onKeyPress: function (event: KeyboardEvent): boolean {
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
                        if (this.incrementVerticalScroll(-40)) {
                            break;
                        }
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
                        if (this.incrementVerticalScroll(40)) {
                            break;
                        }
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
                case "ArrowLeft":
                    if (shifting || event.altKey) {
                        if (this.prev || this.pagePrev) {
                            this.goPrev();
                        } else {
                            caught = false;
                        }
                    } else if (!this.incrementHorizontalScroll(-40)) {
                        if (this.isShort) {
                            if (this.prev || this.pagePrev) {
                                this.goPrev();
                            } else {
                                caught = false;
                            }
                        } else {
                            this.setTime(this.currentTime - 5, true);
                        }
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
                    if (shifting || event.altKey) {
                        if (this.next || this.pageNext) {
                            this.goNext();
                        } else {
                            caught = false;
                        }
                    } else if (!this.incrementHorizontalScroll(40)) {
                        if (this.isShort) {
                            if (this.next || this.pageNext) {
                                this.goNext();
                            } else {
                                caught = false;
                            }
                        } else {
                            this.setTime(this.currentTime + 5, true);
                        }
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
                    if (event.altKey) {
                        caught = false;
                    } else if (shifting) {
                        if (!this.incrementVerticalScroll("home")) {
                            caught = false;
                        }
                    } else if (!this.incrementHorizontalScroll("home")) {
                        this.setTime(0, true);
                    }
                    break;
                case "End":
                    if (event.altKey) {
                        caught = false;
                    } else if (shifting) {
                        if (!this.incrementVerticalScroll("end")) {
                            caught = false;
                        }
                    } else if (!this.incrementHorizontalScroll("end")) {
                        this.setTime(this.duration, true);
                    }
                    break;
                case "X":
                case "x":
                    this.sliceLoop = !this.sliceLoop;
                    if (this.sliceLoop) {
                        PagesController.ShowSnackBar(this.$t("Slice loop enabled"));
                    } else {
                        PagesController.ShowSnackBar(this.$t("Slice loop disabled"));
                    }
                    break;
                case "l":
                case "L":
                    if (event.altKey || shifting || this.isShort) {
                        caught = false;
                    } else {
                        this.loop = !this.loop;
                        if (this.loop) {
                            PagesController.ShowSnackBar(this.$t("Loop enabled"));
                        } else {
                            PagesController.ShowSnackBar(this.$t("Loop disabled"));
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
                case "+":
                    this.scale = Math.min(8, this.scale + (shifting ? 0.01 : 0.1));
                    PagesController.ShowSnackBar(this.$t("Scale") + ": " + this.renderScale(this.scale));
                    this.onScaleUpdated();
                    break;
                case "-":
                    this.scale = Math.max(1, this.scale - (shifting ? 0.01 : 0.1));
                    PagesController.ShowSnackBar(this.$t("Scale") + ": " + this.renderScale(this.scale));
                    this.onScaleUpdated();
                    break;
                case "n":
                case "N":
                    if (this.canWrite) {
                        this.timeSlicesEdit = true;
                    }
                    break;
                case "z":
                case "Z":
                    this.scale = 1;
                    PagesController.ShowSnackBar(this.$t("Scale") + ": " + this.renderScale(this.scale));
                    this.onScaleUpdated();
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
            this.isShort = this.metadata.is_anim || this.metadata.duration <= AUTO_LOOP_MIN_DURATION;
            this.canSaveTime = !this.metadata.force_start_beginning;
            this.hasExtendedDescription = !!this.metadata.ext_desc_url;
            this.hasAttachments = this.metadata.attachments && this.metadata.attachments.length > 0;
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
            this.currentTime = this.canSaveTime ? getCachedInitialTime(this.mid) : 0;
            this.duration = this.metadata.duration || 0;
            this.speed = 1;
            this.scale = 1;
            this.setDefaultLoop();
            this.currentResolution = getUserSelectedResolutionVideo(this.metadata);
            this.loading = true;
            this.playing = true;
            this.updateMediaSessionPlaybackState();
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
            this.hideNextEnd();

            if (this.autoNextTimer) {
                clearTimeout(this.autoNextTimer);
                this.autoNextTimer = null;
            }

            this.mediaError = false;
            this.mediaErrorMessage = "";
            if (!this.metadata) {
                this.videoURL = "";
                this.title = "";
                this.onClearURL();
                this.duration = 0;
                this.loading = false;
                return;
            }

            this.title = this.metadata.title;

            if (this.currentResolution < 0) {
                if (this.metadata.encoded) {
                    this.videoURL = getAssetURL(this.metadata.url);
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
                        this.videoURL = getAssetURL(res.url);
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

        onTooltipImageLoaded: function () {
            this.tooltipImageLoading = false;
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
                this.subtitles = sub.text;
                this.subtitlesStart = sub.start;
                this.subtitlesEnd = sub.end;
            } else {
                this.subTag = "";
                this.subtitles = "";
                this.subtitlesStart = 0;
                this.subtitlesEnd = 0;
            }
        },

        findTimeSliceName: function (time: number) {
            const slice = findTimeSlice(this.timeSlices, time);
            if (slice) {
                return slice.name + " (" + renderTimeSeconds(slice.end - slice.start) + ")";
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
                    this.audioTrackURL = getAssetURL(audio.url);
                    return;
                }
            }

            this.audioTrackURL = "";
        },

        onUpdateNextEnd: function () {
            setAutoNextOnEnd(this.nextEnd);
        },

        onUpdateAutoNextPageDelay: function () {
            setAutoNextPageDelay(this.autoNextPageDelay);
        },

        handleMediaSessionEvent: function (event: MediaSessionActionDetails) {
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
                case "seekbackward":
                    this.setTime(this.currentTime - Math.max(1, Math.abs(event.seekOffset || 5)), true);
                    break;
                case "seekforward":
                    this.setTime(this.currentTime + Math.max(1, Math.abs(event.seekOffset || 5)), true);
                    break;
                case "seekto":
                    if (typeof event.seekTime === "number" && !isNaN(event.seekTime) && isFinite(event.seekTime)) {
                        this.setTime(event.seekTime, !event.fastSeek);
                    }
                    break;
            }
        },

        onMediaError: function () {
            if (!this.videoURL) {
                return;
            }
            if (!AuthController.RefreshAuthStatus()) {
                this.mediaError = true;
                this.updateMediaErrorMessage();
                this.loading = false;
                AuthController.CheckAuthStatusSilent();
            }
        },

        updateMediaErrorMessage: function () {
            this.mediaErrorMessage = "";

            const mediaElem = this.getVideoElement();

            if (!mediaElem) {
                return;
            }

            const err = mediaElem.error;

            if (!err) {
                return;
            }

            this.mediaErrorMessage = err.message || "";
            if (this.mediaErrorMessage) {
                console.error(this.mediaErrorMessage);
            }
        },

        setDefaultLoop: function () {
            if (this.loopForced) {
                this.loop = this.loopForcedValue;
            } else {
                this.loop = (!this.next && !this.pageNext && !this.inAlbum) || !this.nextEnd;
            }
        },

        onTopBarExpand: function () {
            if (this.expandedTitle) {
                this.hideNextEnd();
                this.pause();
            } else if (!this.expandedAlbum) {
                this.play();
            }
        },

        setupAutoNextTimer: function () {
            if (this.autoNextTimer) {
                clearTimeout(this.autoNextTimer);
                this.autoNextTimer = null;
            }

            if (!this.isShort) {
                return;
            }

            const timerS = getAutoNextTime();

            if (isNaN(timerS) || !isFinite(timerS) || timerS <= 0) {
                return;
            }

            if (!this.next && !this.pageNext) {
                return;
            }

            const ms = timerS * 1000;

            this.autoNextTimer = setTimeout(() => {
                this.autoNextTimer = null;
                if (this.displayConfig || this.expandedTitle || this.displayAttachments || !this.playing) {
                    this.setupAutoNextTimer();
                } else {
                    this.goNext();
                }
            }, ms);
        },

        showNextEnd: function () {
            this.pendingNextEnd = true;
            this.pendingNextEndSeconds = NEXT_END_WAIT_DURATION;

            if (this.nextEndTimer) {
                clearTimeout(this.nextEndTimer);
                this.nextEndTimer = null;
            }

            this.nextEndTimer = setTimeout(this.tickNextEnd.bind(this), 1000);
        },

        tickNextEnd: function () {
            this.nextEndTimer = null;

            this.pendingNextEndSeconds = Math.max(0, this.pendingNextEndSeconds - 1);

            if (this.pendingNextEndSeconds <= 0) {
                this.pendingNextEnd = false;
                this.goNext();
                return;
            }

            this.nextEndTimer = setTimeout(this.tickNextEnd.bind(this), 1000);
        },

        hideNextEnd: function () {
            this.pendingNextEnd = false;

            if (this.nextEndTimer) {
                clearTimeout(this.nextEndTimer);
                this.nextEndTimer = null;
            }
        },

        enableLoopAndPlay: function () {
            this.loop = true;
            this.$emit("force-loop", this.loop);
            this.play();
        },

        refreshTimeSlices: function () {
            const metadata = MediaController.MediaData;

            if (!metadata) {
                return;
            }

            this.timeSlices = normalizeTimeSlices(
                (metadata.time_slices || []).sort((a, b) => {
                    if (a.time < b.time) {
                        return -1;
                    } else if (a.time > b.time) {
                        return 1;
                    } else {
                        return 0;
                    }
                }),
                metadata.duration,
            );
            this.currentTimeSlice = null;
            this.currentTimeSliceName = "";
            this.currentTimeSliceStart = 0;
            this.currentTimeSliceEnd = 0;
            this.updateCurrentTimeSlice();
        },

        refreshExtendedDescription: function () {
            this.hasExtendedDescription = !!this.metadata.ext_desc_url;
        },

        renderScale: function (scale: number) {
            if (scale > 1) {
                return Math.floor(scale * 100) + "%";
            } else if (scale < 1) {
                return Math.floor(scale * 100) + "%";
            } else {
                return this.$t("Normal");
            }
        },

        getScaleCss: function (scale: number) {
            return Math.floor(scale * 100) + "%";
        },

        onScaleUpdated: function () {
            if (this.scale > 1) {
                nextTick(this.centerScroll.bind(this));
            }
        },

        grabScroll: function (e: TouchEvent | MouseEvent) {
            if ("button" in e && e.button !== 0) {
                return;
            }

            if (this.scale <= 1) {
                return; // Not scaled
            }

            this.leaveControls();

            if (this.displayConfig || this.contextMenuShown || this.displayAttachments) {
                this.displayConfig = false;
                this.contextMenuShown = false;
                this.displayAttachments = false;
                e.stopPropagation();
                return;
            }

            const scroller = this.$el.querySelector(".video-scroller");

            if (!scroller) {
                return;
            }

            this.scrollGrabTop = scroller.scrollTop;
            this.scrollGrabLeft = scroller.scrollLeft;

            this.scrollGrabbed = true;
            this.scrollMoved = false;

            if ("touches" in e && e.touches.length > 0) {
                this.scrollGrabX = e.touches[0].pageX;
                this.scrollGrabY = e.touches[0].pageY;
            } else if ("pageX" in e) {
                this.scrollGrabX = e.pageX;
                this.scrollGrabY = e.pageY;
            }

            e.stopPropagation();
        },

        moveScrollByMouse: function (x: number, y: number) {
            const scroller = this.$el.querySelector(".video-scroller");

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
            this.scrollMoved = true;
        },

        dropScroll: function (e: TouchEvent | MouseEvent) {
            if ("button" in e && e.button !== 0) {
                return;
            }

            if (!this.scrollGrabbed) {
                return;
            }

            this.scrollGrabbed = false;

            if (!this.scrollMoved) {
                this.clickPlayer();
                return;
            }

            this.scrollMoved = false;

            if ("touches" in e && e.touches.length > 0) {
                this.moveScrollByMouse(e.touches[0].pageX, e.touches[0].pageY);
            } else if ("pageX" in e) {
                this.moveScrollByMouse(e.pageX, e.pageY);
            }
        },

        moveScroll: function (e: MouseEvent | TouchEvent) {
            if (!this.scrollGrabbed) {
                return;
            }

            if ("touches" in e && e.touches.length > 0) {
                this.moveScrollByMouse(e.touches[0].pageX, e.touches[0].pageY);
            } else if ("pageX" in e) {
                this.moveScrollByMouse(e.pageX, e.pageY);
            }
        },

        centerScroll: function () {
            const scroller = this.$el.querySelector(".video-scroller");

            if (!scroller) {
                return;
            }

            scroller.scrollTop = (scroller.scrollHeight - scroller.getBoundingClientRect().height) / 2;
            scroller.scrollLeft = (scroller.scrollWidth - scroller.getBoundingClientRect().width) / 2;
        },

        onMouseWheel: function (e: WheelEvent) {
            if (e.ctrlKey) {
                e.preventDefault();
                e.stopPropagation();
                if (e.deltaY > 0) {
                    this.scale = Math.max(1, this.scale - 0.1);
                } else {
                    this.scale = Math.min(8, this.scale + 0.1);
                }

                PagesController.ShowSnackBar(this.$t("Scale") + ": " + this.renderScale(this.scale));
                this.onScaleUpdated();
            }
        },

        incrementVerticalScroll: function (a: number | string): boolean {
            if (this.scale <= 1) {
                return false;
            }

            const el = this.$el.querySelector(".video-scroller");

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

        incrementHorizontalScroll: function (a: number | string): boolean {
            if (this.scale <= 1) {
                return false;
            }

            const el = this.$el.querySelector(".video-scroller");

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

        onScrollerTouchStart: function (e: Event) {
            if (this.scale <= 1) {
                return;
            }

            e.stopPropagation();
        },

        updateMediaSessionPlaybackState: function () {
            if (!window.navigator || !window.navigator.mediaSession) {
                return;
            }

            if (MediaController.MediaSessionId !== this.mediaSessionId) {
                return;
            }

            navigator.mediaSession.playbackState = this.playing ? "playing" : "paused";
        },
    },
});
</script>
