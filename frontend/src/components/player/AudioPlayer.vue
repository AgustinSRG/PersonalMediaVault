<template>
    <div
        class="audio-player player-settings-no-trap"
        :class="{
            'player-min': min,
            'no-controls': !showControls,
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
    >
        <audio
            :src="audioURL || undefined"
            playsinline
            webkit-playsinline
            x-webkit-airplay="allow"
            :loop="(loop || isShort) && !sliceLoop"
            crossorigin="use-credentials"
            :muted="muted"
            :volume.prop="volume"
            :playbackRate.prop="speed"
            @ended="onEnded"
            @timeupdate="onAudioTimeUpdate"
            @canplay="onCanPlay"
            @loadedmetadata="onLoadMetaData"
            @waiting="onWaitForBuffer(true)"
            @playing="onWaitForBuffer(false)"
            @play="onPlay"
            @pause="onPause"
            @error="onMediaError"
        ></audio>

        <canvas v-if="audioURL && animationColors !== 'none'"></canvas>
        <div v-if="audioURL" class="audio-no-animation">
            <div v-if="thumbnail && showThumbnail" class="audio-no-animation-thumbnail-container">
                <img class="audio-no-animation-thumbnail" :src="getThumbnail(thumbnail)" :alt="title" :title="title" loading="lazy" />
            </div>
            <div class="audio-no-animation-title" :class="{ hidden: !showTitle }">{{ title }}</div>
        </div>

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
            v-if="(!loading && !audioURL && audioPending) || mediaError"
            :mid="mid"
            :tid="audioPendingTask"
            :res="-1"
            :error="mediaError"
            :error-message="mediaErrorMessage"
            :can-auto-reload="!expandedTitle && !expandedAlbum"
        ></PlayerEncodingPending>

        <PlayerSubtitles :show-controls="showControls" :subtitles="subtitles"></PlayerSubtitles>

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

        <DescriptionWidget
            v-if="displayDescription"
            v-model:display="displayDescriptionStatus"
            :context-open="contextMenuShown"
            @clicked="clickControls"
            @update-desc="refreshDescription"
        ></DescriptionWidget>

        <div
            class="player-controls"
            :class="{ hidden: !showControls }"
            @click="clickControls"
            @dblclick="stopPropagationEvent"
            @mousedown="stopPropagationEvent"
            @touchstart="stopPropagationEvent"
            @mouseenter="enterControls"
            @mouseleave="leaveControls"
            @contextmenu="stopPropagationEvent"
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
                    @click="togglePlay"
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
                    @click="togglePlay"
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
                    v-if="hasDescription || canWrite"
                    type="button"
                    :title="$t('Description')"
                    class="player-btn player-btn-hide-mobile"
                    @click="openDescription"
                    @mouseenter="enterTooltip('desc')"
                    @mouseleave="leaveTooltip('desc')"
                >
                    <i v-if="hasDescription" class="fas fa-file-lines"></i>
                    <i v-else class="fas fa-file-circle-plus"></i>
                </button>

                <button
                    v-if="hasAttachments"
                    type="button"
                    :title="$t('Attachments')"
                    class="player-btn player-btn-hide-mobile player-settings-no-trap"
                    @click="showAttachments"
                    @mouseenter="enterTooltip('attachments')"
                    @mouseleave="leaveTooltip('attachments')"
                >
                    <i class="fas fa-paperclip"></i>
                </button>

                <button
                    v-if="hasRelatedMedia"
                    type="button"
                    :title="$t('Related media')"
                    class="player-btn player-btn-hide-mobile player-settings-no-trap"
                    @click="showRelatedMedia"
                    @mouseenter="enterTooltip('related-media')"
                    @mouseleave="leaveTooltip('related-media')"
                >
                    <i class="fas fa-photo-film"></i>
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

        <div
            v-else-if="!displayConfig && !displayAttachments && !displayRelatedMedia && helpTooltip === 'desc'"
            class="player-tooltip player-help-tip-right"
        >
            {{ $t("Description") }}
        </div>

        <div
            v-else-if="!displayConfig && !displayAttachments && !displayRelatedMedia && helpTooltip === 'attachments'"
            class="player-tooltip player-help-tip-right"
        >
            {{ $t("Attachments") }}
        </div>

        <div
            v-else-if="!displayConfig && !displayAttachments && !displayRelatedMedia && helpTooltip === 'related-media'"
            class="player-tooltip player-help-tip-right"
        >
            {{ $t("Related media") }}
        </div>

        <div
            v-else-if="!displayConfig && !displayAttachments && !displayRelatedMedia && helpTooltip === 'config'"
            class="player-tooltip player-help-tip-right"
        >
            {{ $t("Player Configuration") }}
        </div>

        <div
            v-else-if="!displayConfig && !displayAttachments && !displayRelatedMedia && helpTooltip === 'albums'"
            class="player-tooltip player-help-tip-right"
        >
            {{ $t("Manage albums") }}
        </div>

        <div
            v-else-if="!displayConfig && !displayAttachments && !displayRelatedMedia && helpTooltip === 'full-screen'"
            class="player-tooltip player-help-tip-right"
        >
            {{ $t("Full screen") }}
        </div>

        <div
            v-else-if="!displayConfig && !displayAttachments && !displayRelatedMedia && helpTooltip === 'full-screen-exit'"
            class="player-tooltip player-help-tip-right"
        >
            {{ $t("Exit full screen") }}
        </div>

        <div
            class="player-timeline"
            :class="{ hidden: !showControls }"
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
            <div class="player-tooltip-text">{{ tooltipText }}</div>
            <div v-if="tooltipTimeSlice" class="player-tooltip-text">
                {{ tooltipTimeSlice }}
            </div>
        </div>

        <AudioPlayerConfig
            v-model:shown="displayConfig"
            v-model:speed="speed"
            v-model:loop="loop"
            v-model:next-end="nextEnd"
            v-model:auto-next-page-delay="autoNextPageDelay"
            v-model:anim-colors="animationColors"
            v-model:show-title="showTitle"
            v-model:show-thumbnail="showThumbnail"
            :r-tick="internalTick"
            :metadata="metadata"
            :is-short="isShort"
            :in-album="inAlbum"
            @update:loop="() => $emit('force-loop', loop)"
            @update:anim-colors="onUpdateAnimColors"
            @update:show-title="onUpdateShowTitle"
            @update:show-thumbnail="onUpdateShowThumbnail"
            @update:next-end="onUpdateNextEnd"
            @update:auto-next-page-delay="onUpdateAutoNextPageDelay"
            @enter="enterControls"
            @leave="leaveControls"
            @update-auto-next="setupAutoNextTimer"
        ></AudioPlayerConfig>

        <PlayerAttachmentsList
            v-if="metadata && metadata.attachments"
            v-model:shown="displayAttachments"
            :attachments="metadata.attachments"
            @enter="enterControls"
            @leave="leaveControls"
        >
        </PlayerAttachmentsList>

        <PlayerRelatedMediaList
            v-if="metadata && metadata.related"
            v-model:shown="displayRelatedMedia"
            :related-media="metadata.related"
            @enter="enterControls"
            @leave="leaveControls"
        >
        </PlayerRelatedMediaList>

        <PlayerTopBar
            v-if="metadata"
            v-model:expanded="expandedTitle"
            v-model:album-expanded="expandedAlbum"
            :mid="mid"
            :metadata="metadata"
            :shown="showControls"
            :fullscreen="fullscreen"
            :in-album="inAlbum"
            @update:expanded="onTopBarExpand"
            @click-player="clickControls"
        ></PlayerTopBar>

        <PlayerContextMenu
            v-model:shown="contextMenuShown"
            v-model:loop="loop"
            v-model:slice-loop="sliceLoop"
            v-model:time-slices-edit="timeSlicesEdit"
            type="audio"
            :x="contextMenuX"
            :y="contextMenuY"
            :url="audioURL"
            :title="title"
            :can-write="canWrite"
            :has-description="hasDescription"
            :has-slices="timeSlices && timeSlices.length > 0"
            :is-short="isShort"
            @update:loop="() => $emit('force-loop', loop)"
            @stats="openStats"
            @open-tags="openTags"
            @open-desc="openDescription"
        ></PlayerContextMenu>
    </div>
</template>

<script lang="ts">
import {
    CURRENT_TIME_UPDATE_DELAY,
    getAudioAnimationStyle,
    getAutoNextOnEnd,
    getAutoNextPageDelay,
    getAutoNextTime,
    getCachedInitialTime,
    getPlayerMuted,
    getPlayerVolume,
    getShowAudioThumbnail,
    getShowAudioTitle,
    setAudioAnimationStyle,
    setAutoNextOnEnd,
    setAutoNextPageDelay,
    setCachedInitialTime,
    setPlayerMuted,
    setPlayerVolume,
    setShowAudioThumbnail,
    setShowAudioTitle,
} from "@/control/player-preferences";
import type { PropType } from "vue";
import { defineAsyncComponent, defineComponent, nextTick } from "vue";
import VolumeControl from "./VolumeControl.vue";
import PlayerMediaChangePreview from "./PlayerMediaChangePreview.vue";
import PlayerTopBar from "./PlayerTopBar.vue";
import { openFullscreen, closeFullscreen } from "../../utils/full-screen";
import { renderTimeSeconds } from "../../utils/time";
import type { NormalizedTimeSlice } from "../../utils/time-slices";
import { findTimeSlice, normalizeTimeSlices } from "../../utils/time-slices";
import { isTouchDevice } from "@/utils/touch";
import AudioPlayerConfig from "./AudioPlayerConfig.vue";
import PlayerContextMenu from "./PlayerContextMenu.vue";
import PlayerSubtitles from "./PlayerSubtitles.vue";
import { getAssetURL } from "@/utils/api";
import { useVModel } from "../../utils/v-model";
import { AUTO_LOOP_MIN_DURATION, MediaController, NEXT_END_WAIT_DURATION } from "@/control/media";
import { EVENT_NAME_SUBTITLES_UPDATE, SubtitlesController } from "@/control/subtitles";
import { AppStatus } from "@/control/app-status";
import { AuthController } from "@/control/auth";
import type { ColorThemeName } from "@/control/app-preferences";
import { EVENT_NAME_THEME_CHANGED, getTheme } from "@/control/app-preferences";
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

const DescriptionWidget = defineAsyncComponent({
    loader: () => import("@/components/player/DescriptionWidget.vue"),
});

const PlayerAttachmentsList = defineAsyncComponent({
    loader: () => import("@/components/player/PlayerAttachmentsList.vue"),
});

const PlayerRelatedMediaList = defineAsyncComponent({
    loader: () => import("@/components/player/PlayerRelatedMediaList.vue"),
});

export default defineComponent({
    name: "AudioPlayer",
    components: {
        VolumeControl,
        AudioPlayerConfig,
        PlayerMediaChangePreview,
        PlayerTopBar,
        PlayerContextMenu,
        PlayerEncodingPending,
        TimeSlicesEditHelper,
        TagsEditHelper,
        DescriptionWidget,
        PlayerAttachmentsList,
        PlayerSubtitles,
        PlayerRelatedMediaList,
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

        min: Boolean,

        loopForced: Boolean,
        loopForcedValue: Boolean,

        autoPlay: Boolean,

        displayTagList: Boolean,
        displayDescription: Boolean,
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
        "update:displayDescription",
    ],
    setup(props) {
        return {
            timer: null as ReturnType<typeof setInterval> | null,
            autoNextTimer: null as ReturnType<typeof setInterval> | null,
            audioContext: null as AudioContext,
            audioSource: null as MediaElementAudioSourceNode | null,
            audioAnalyser: null as AnalyserNode,
            audioAnalyserData: null as Uint8Array,
            rendererTimer: null as ReturnType<typeof setInterval> | null,
            nextEndTimer: null as ReturnType<typeof setInterval> | null,
            mediaSessionId: getUniqueStringId(),
            fullScreenState: useVModel(props, "fullscreen"),
            displayTagListStatus: useVModel(props, "displayTagList"),
            displayDescriptionStatus: useVModel(props, "displayDescription"),
        };
    },
    data: function () {
        return {
            playing: false,
            loading: true,

            autoPlayApplied: false,

            audioURL: "",
            audioPending: false,
            audioPendingTask: 0,

            thumbnail: "",
            title: "",

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

            showControls: true,
            lastControlsInteraction: Date.now(),
            mouseInControls: false,

            loop: false,
            nextEnd: false,
            autoNextPageDelay: false,

            isShort: false,

            sliceLoop: false,

            volume: 1,
            muted: false,
            volumeShown: isTouchDevice(),
            internalTick: 0,

            speed: 1,

            feedback: "",

            helpTooltip: "",

            animationColors: "",

            showTitle: true,
            showThumbnail: true,

            expandedTitle: false,
            expandedAlbum: false,

            contextMenuX: 0,
            contextMenuY: 0,
            contextMenuShown: false,

            requiresRefresh: false,

            subtitles: "",
            subtitlesStart: -1,
            subtitlesEnd: -1,

            timeSlices: [] as NormalizedTimeSlice[],
            currentTimeSlice: null as NormalizedTimeSlice | null,
            currentTimeSliceName: "",
            currentTimeSliceStart: 0,
            currentTimeSliceEnd: 0,

            timeSlicesEdit: false,

            theme: getTheme(),

            mediaError: false,
            mediaErrorMessage: "",

            hasDescription: false,

            hasAttachments: false,
            displayAttachments: false,

            hasRelatedMedia: false,
            displayRelatedMedia: false,

            pendingNextEnd: false,
            pendingNextEndSeconds: 0,

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
            this.initializeAudio();
        },
        audioURL: function () {
            if (this.audioURL) {
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
        this.animationColors = getAudioAnimationStyle();
        this.showTitle = getShowAudioTitle();
        this.showThumbnail = getShowAudioThumbnail();
        this.nextEnd = getAutoNextOnEnd();
        this.autoNextPageDelay = getAutoNextPageDelay();

        this.$addKeyboardHandler(this.onKeyPress.bind(this), 100);

        this.timer = setInterval(this.tick.bind(this), 100);

        this.$listenOnDocumentEvent("fullscreenchange", this.onExitFullScreen.bind(this));

        this.$listenOnAppEvent(EVENT_NAME_SUBTITLES_UPDATE, this.reloadSubtitles.bind(this));

        this.$listenOnAppEvent(EVENT_NAME_THEME_CHANGED, this.themeUpdated.bind(this));

        this.initializeAudio();

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
        this.audioURL = "";
        this.onClearURL();
        clearInterval(this.timer);

        if (this.autoNextTimer) {
            clearTimeout(this.autoNextTimer);
            this.autoNextTimer = null;
        }

        if (this.nextEndTimer) {
            clearTimeout(this.nextEndTimer);
            this.nextEndTimer = null;
        }

        this.clearAudioRenderer();
        this.closeAudioContext();

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

        openDescription: function () {
            if (!this.hasDescription && !this.canWrite) {
                return;
            }
            this.displayDescriptionStatus = true;
        },

        showAttachments: function (e?: Event) {
            if (e) {
                e.stopPropagation();
            }
            this.displayAttachments = !this.displayAttachments;
            this.displayRelatedMedia = false;
            this.displayConfig = false;
        },

        showRelatedMedia: function (e?: Event) {
            if (e) {
                e.stopPropagation();
            }

            this.displayRelatedMedia = !this.displayRelatedMedia;
            this.displayAttachments = false;
            this.displayConfig = false;
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
            this.displayRelatedMedia = false;
            if (e) {
                e.stopPropagation();
            }
        },

        clickControls: function (e: Event) {
            this.displayConfig = false;
            this.contextMenuShown = false;
            this.displayAttachments = false;
            this.displayRelatedMedia = false;
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

        onUpdateAnimColors: function () {
            setAudioAnimationStyle(this.animationColors);
        },

        onUpdateShowTitle: function () {
            setShowAudioTitle(this.showTitle);
        },

        onUpdateShowThumbnail: function () {
            setShowAudioThumbnail(this.showThumbnail);
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

        onLoadMetaData: function (ev) {
            const audioElement = ev.target;

            if (!audioElement) {
                return;
            }

            if (typeof audioElement.duration !== "number" || isNaN(audioElement.duration) || !isFinite(audioElement.duration)) {
                return;
            }

            this.duration = audioElement.duration;

            if (typeof this.currentTime === "number" && !isNaN(this.currentTime) && isFinite(this.currentTime) && this.currentTime >= 0) {
                audioElement.currentTime = Math.min(this.currentTime, this.duration);
                this.updateSubtitles();
                this.updateCurrentTimeSlice();
            }
        },
        onAudioTimeUpdate: function (ev) {
            this.hideNextEnd();
            if (this.loading) return;

            const audioElement = ev.target;

            if (
                !audioElement ||
                typeof audioElement.currentTime !== "number" ||
                isNaN(audioElement.currentTime) ||
                !isFinite(audioElement.currentTime) ||
                typeof audioElement.duration !== "number" ||
                isNaN(audioElement.duration) ||
                !isFinite(audioElement.duration)
            ) {
                return;
            }

            this.currentTime = audioElement.currentTime;
            this.duration = audioElement.duration;

            if (!this.loading && this.canSaveTime && Date.now() - this.lastTimeChangedEvent > CURRENT_TIME_UPDATE_DELAY) {
                setCachedInitialTime(this.mid, this.currentTime);
                this.lastTimeChangedEvent = Date.now();
            }

            this.updateSubtitles();
            this.updateCurrentTimeSlice();
        },
        onCanPlay: function () {
            this.loading = false;
            if (this.autoPlayApplied) {
                if (this.playing) {
                    const player = this.getAudioElement();
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
            const player = this.getAudioElement();
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
        onWaitForBuffer: function (b: boolean) {
            this.loading = b;
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
            if (
                !this.playing ||
                this.expandedTitle ||
                this.expandedAlbum ||
                this.displayConfig ||
                this.displayAttachments ||
                this.displayRelatedMedia
            )
                return;
            this.showControls = false;
            this.volumeShown = isTouchDevice();
            this.helpTooltip = "";
        },

        tick() {
            if (this.showControls && !this.mouseInControls && this.playing && !this.expandedTitle && !this.expandedAlbum) {
                if (Date.now() - this.lastControlsInteraction > 2000) {
                    this.showControls = false;
                    this.volumeShown = isTouchDevice();
                    this.helpTooltip = "";
                    this.displayConfig = false;
                    this.displayAttachments = false;
                    this.displayRelatedMedia = false;
                }
            }

            const audio = this.getAudioElement();

            if (audio && audio.buffered.length > 0) {
                this.bufferedTime = audio.buffered.end(audio.buffered.length - 1);
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

        getAudioElement() {
            return this.$el.querySelector("audio");
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

        togglePlay() {
            if (this.playing) {
                this.pause();
                this.feedback = "pause";
            } else {
                this.play();
                this.feedback = "play";
            }

            this.displayConfig = false;
            this.displayAttachments = false;
            this.displayRelatedMedia = false;
        },

        clickPlayer: function () {
            this.leaveControls();
            if (this.displayConfig || this.contextMenuShown || this.displayAttachments || this.displayRelatedMedia) {
                this.displayConfig = false;
                this.contextMenuShown = false;
                this.displayAttachments = false;
                this.displayRelatedMedia = false;
            } else {
                this.togglePlay();
            }
            this.interactWithControls();
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
            if (this.displayConfig || this.contextMenuShown || this.displayAttachments || this.displayRelatedMedia) {
                this.displayConfig = false;
                this.contextMenuShown = false;
                this.displayAttachments = false;
                this.displayRelatedMedia = false;
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

        play: function () {
            this.hideNextEnd();
            if (this.requiresRefresh) {
                this.requiresRefresh = false;
                MediaController.Load();
                return;
            }
            const audio = this.getAudioElement();
            this.playing = true;
            if (audio) {
                audio.play();
            }
            this.updateMediaSessionPlaybackState();
        },
        pause: function () {
            const audio = this.getAudioElement();
            this.playing = false;

            if (audio) {
                audio.pause();
            }

            this.updateMediaSessionPlaybackState();

            if (!this.loading && this.canSaveTime && audio && !audio.ended) {
                setCachedInitialTime(this.mid, this.currentTime);
            }

            this.lastTimeChangedEvent = Date.now();

            this.interactWithControls();
        },

        onPlay: function () {
            this.playing = true;
            this.setupAudioRenderer();
            this.updateMediaSessionPlaybackState();
        },

        onPause: function () {
            this.playing = false;
            this.clearAudioRenderer();
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
            this.displayRelatedMedia = false;
            e.stopPropagation();
        },

        /* Timeline */

        grabTimelineByMouse: function (e: MouseEvent & TouchEvent) {
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
            if (!isTouchDevice()) {
                this.tooltipShown = false;
            }
            this.leaveControls();
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
            this.tooltipEventX = x;

            nextTick(this.tick.bind(this));
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

            const audio = this.getAudioElement();

            if (audio) {
                audio.currentTime = time;
            }

            if (this.canSaveTime && save) {
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
                    this.openDescription();
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
                    this.togglePlay();
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
                    } else {
                        this.setTime(this.currentTime + 5, true);
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
                case "n":
                case "N":
                    if (this.canWrite) {
                        this.timeSlicesEdit = true;
                    }
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

        initializeAudio() {
            if (!this.metadata) {
                return;
            }
            this.isShort = this.metadata.duration <= AUTO_LOOP_MIN_DURATION;
            this.canSaveTime = !this.metadata.force_start_beginning;
            this.hasDescription = !!this.metadata.description_url;
            this.hasAttachments = this.metadata.attachments && this.metadata.attachments.length > 0;
            this.hasRelatedMedia = this.metadata.related && this.metadata.related.length > 0;
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
            this.setDefaultLoop();
            this.loading = true;
            this.playing = true;
            this.updateMediaSessionPlaybackState();
            this.clearAudioRenderer();
            this.setAudioURL();
        },

        onClearURL: function () {
            const audioElem = this.$el.querySelector("audio");
            if (audioElem) {
                audioElem.src = "";
            }
        },

        setAudioURL() {
            this.hideNextEnd();

            if (this.autoNextTimer) {
                clearTimeout(this.autoNextTimer);
                this.autoNextTimer = null;
            }

            this.mediaError = false;
            this.mediaErrorMessage = "";
            if (!this.metadata) {
                this.audioURL = "";
                this.thumbnail = "";
                this.title = "";
                this.onClearURL();
                this.duration = 0;
                this.loading = false;
                this.clearAudioRenderer();
                this.getAudioElement().load();
                return;
            }

            this.thumbnail = this.metadata.thumbnail;
            this.title = this.metadata.title;

            if (this.metadata.encoded) {
                this.audioURL = getAssetURL(this.metadata.url);
                this.audioPending = false;
                this.audioPendingTask = 0;
                this.getAudioElement().load();
                this.setupAutoNextTimer();
            } else {
                this.audioURL = "";
                this.onClearURL();
                this.audioPending = true;
                this.audioPendingTask = this.metadata.task;
                this.duration = 0;
                this.loading = false;
                this.clearAudioRenderer();
                this.getAudioElement().load();
            }
        },

        clearAudioRenderer: function () {
            if (this.rendererTimer) {
                clearInterval(this.rendererTimer);
                this.rendererTimer = null;
            }

            if (this.audioSource) {
                this.audioSource.disconnect();
            }

            if (this.audioAnalyser) {
                this.audioAnalyser.disconnect();
            }

            this.audioAnalyser = null;
            this.audioAnalyserData = null;
        },

        setupAudioRenderer: function () {
            if (!this.audioContext) {
                this.setupAudioContext();
            }

            this.clearAudioRenderer();

            if (this.audioURL) {
                const context = this.audioContext;
                const source = this.audioSource;

                const analyser = context.createAnalyser();

                this.audioSource = source;
                this.audioAnalyser = analyser;
                source.connect(analyser);
                analyser.connect(context.destination);

                analyser.fftSize = 256;

                this.audioAnalyserData = new Uint8Array(analyser.frequencyBinCount);

                this.rendererTimer = setInterval(this.audioAnimationFrame.bind(this), Math.floor(1000 / 30));
            } else {
                this.audioContext = null;
                this.audioSource = null;
                this.audioAnalyser = null;
                this.audioAnalyserData = null;
            }
        },

        setupAudioContext: function () {
            const context = new AudioContext();
            this.audioContext = context;
            const source = context.createMediaElementSource(this.getAudioElement());
            this.audioSource = source;
        },

        closeAudioContext: function () {
            if (this.audioContext) {
                this.audioContext.close();
            }

            this.audioContext = null;
            this.audioSource = null;
        },

        audioAnimationFrame: function () {
            if (!this.playing || this.animationColors === "none") {
                return;
            }

            const analyser = this.audioAnalyser;
            const canvas = this.$el.querySelector("canvas");

            if (!analyser || !canvas) {
                return;
            }

            const rect = canvas.getBoundingClientRect();
            if (canvas.width !== rect.width || canvas.height !== rect.height) {
                canvas.width = rect.width;
                canvas.height = rect.height;
            }

            let bufferLength = analyser.frequencyBinCount;

            const dataArray = this.audioAnalyserData;

            if (!dataArray) {
                return;
            }

            analyser.getByteFrequencyData(dataArray);

            bufferLength = Math.floor(Math.max(1, bufferLength / 2));

            const WIDTH = canvas.width;
            const HEIGHT = canvas.height;

            const barWidth = Math.max(1, (WIDTH - (bufferLength - 1)) / bufferLength);

            const ctx = canvas.getContext("2d") as CanvasRenderingContext2D;

            ctx.fillStyle = this.theme === "light" ? "#fff" : "#000";
            ctx.fillRect(0, 0, WIDTH, HEIGHT);

            let x = 0;

            for (let i = 0; i < bufferLength; i++) {
                const barHeight = dataArray[i];

                switch (this.animationColors) {
                    case "gradient":
                        {
                            if (this.theme === "light") {
                                const r = Math.min(255, (barHeight + 25 * (i / bufferLength)) * 2);
                                const g = Math.min(255, 250 * (i / bufferLength) * 2);
                                const b = 100;
                                ctx.fillStyle = "rgb(" + r + "," + g + "," + b + ")";
                            } else {
                                const r = Math.min(255, barHeight + 25 * (i / bufferLength));
                                const g = 250 * (i / bufferLength);
                                const b = 50;
                                ctx.fillStyle = "rgb(" + r + "," + g + "," + b + ")";
                            }
                        }
                        break;
                    case "rainbow":
                        {
                            ctx.fillStyle =
                                "hsl(" + (i * 300) / bufferLength + ", " + "60%" + ", " + (0.3 + 0.3 * (barHeight / 255)) * 100 + "%)";
                        }
                        break;
                    default:
                        ctx.fillStyle = this.theme === "light" ? "rgba(0, 0, 0, 0.5)" : "rgba(255, 255, 255, 0.5)";
                }

                const trueHeight = Math.floor(HEIGHT * (barHeight / 255));

                ctx.fillRect(x, HEIGHT - trueHeight, barWidth, trueHeight);

                x += barWidth + 1;
            }
        },

        onFeedBackAnimationEnd: function () {
            this.feedback = "";
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

        onUpdateNextEnd: function () {
            setAutoNextOnEnd(this.nextEnd);
        },

        onUpdateAutoNextPageDelay: function () {
            setAutoNextPageDelay(this.autoNextPageDelay);
        },

        themeUpdated: function (theme: ColorThemeName) {
            this.theme = theme;
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
            if (!this.audioURL) {
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

            const mediaElem = this.getAudioElement();

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
                if (this.displayConfig || this.expandedTitle || this.displayAttachments || this.displayRelatedMedia) {
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

        refreshDescription: function () {
            this.hasDescription = !!this.metadata.description_url;
        },

        getThumbnail(thumb: string) {
            return getAssetURL(thumb);
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
