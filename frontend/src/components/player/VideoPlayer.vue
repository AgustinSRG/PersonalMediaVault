<template>
    <div
        ref="container"
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
        <div ref="videoScroller" class="video-scroller" @mousedown="grabScrollWithMouse" @touchstart="onScrollerTouchStart">
            <video
                v-if="videoURL"
                ref="videoElement"
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
            ref="audioTrackElement"
            :key="rTick"
            :src="audioTrackURL"
            crossorigin="use-credentials"
            playsinline
            webkit-playsinline
            :muted="muted || !audioTrackURL"
            :volume.prop="volume"
            :playbackRate.prop="speed"
            @loadedmetadata="onAudioTrackLoadMetadata"
            @canplay="onAudioCanPlay"
        ></audio>

        <PlayerPlayFeedback v-model:feedback="feedback"></PlayerPlayFeedback>

        <PlayerLoader v-if="loading && !mediaError"></PlayerLoader>

        <PlayerAutoNextOverlay
            v-if="pendingNextEnd"
            :pending-next-end-seconds="pendingNextEndSeconds"
            @cancel="hideNextEnd"
            @next="goNext"
            @play="play"
            @loop="enableLoopAndPlay"
        ></PlayerAutoNextOverlay>

        <PlayerEncodingPending
            v-if="(!loading && !videoURL && videoPending) || mediaError"
            :mid="mid"
            :tid="videoPendingTask"
            :res="currentResolution"
            :error="mediaError"
            :error-message="mediaErrorMessage"
            :encoding-error="videoEncodeError"
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
            v-model:display="displayTagList"
            :context-open="contextMenuShown"
            @clicked="clickControls"
        ></TagsEditHelper>

        <DescriptionWidget
            v-if="displayDescription"
            v-model:display="displayDescription"
            :context-open="contextMenuShown"
            :title="title"
            @clicked="clickControls"
            @update-desc="refreshDescription"
        ></DescriptionWidget>

        <PlayerControls
            :type="'video'"
            :show-controls="showControls && userControls"
            :next="next"
            :prev="prev"
            :page-next="pageNext"
            :page-prev="pagePrev"
            :fullscreen="fullscreen"
            :playing="playing"
            :has-description="hasDescription"
            :has-attachments="hasAttachments"
            :has-related-media="hasRelatedMedia"
            @click="clickControls"
            @enter="enterControls"
            @leave="leaveControls"
            @enter-tooltip="enterTooltip"
            @leave-tooltip="leaveTooltip"
            @go-prev="goPrev"
            @play="togglePlayImmediate"
            @pause="togglePlayImmediate"
            @go-next="goNext"
            @open-description="openDescription"
            @open-attachments="showAttachments"
            @open-related-media="showRelatedMedia"
            @open-albums="manageAlbums"
            @open-config="showConfig"
            @toggle-full-screen="toggleFullScreen"
        >
            <VolumeControl
                ref="volumeControl"
                v-model:muted="muted"
                v-model:volume="volume"
                v-model:expanded="volumeShown"
                :min="min"
                :width="min ? VOLUME_CONTROL_WIDTH_MIN : VOLUME_CONTROL_WIDTH"
                @update:volume="onUserVolumeUpdated"
                @update:muted="onUserMutedUpdated"
                @enter="enterTooltip('volume')"
                @leave="leaveTooltip('volume')"
            ></VolumeControl>

            <div v-if="!min" class="player-time-label-container" :class="{ 'in-album': !!next || !!prev }">
                <span>{{ renderTimeSeconds(currentTime) }} / {{ renderTimeSeconds(duration) }}</span>
                <span v-if="currentTimeSlice" class="times-slice-name"><b class="separator"> - </b>{{ currentTimeSliceName }}</span>
            </div>
        </PlayerControls>

        <PlayerTooltip
            v-if="helpTooltip"
            :help-tooltip="helpTooltip"
            :hide-right-tooltip="displayConfig || displayAttachments || displayRelatedMedia"
            :next="next"
            :prev="prev"
            :page-next="pageNext"
            :page-prev="pagePrev"
            :muted="muted"
            :volume="volume"
        ></PlayerTooltip>

        <PlayerTimeline
            :hidden="!showControls || !userControls"
            :duration="duration"
            :buffered-time="bufferedTime"
            :current-time="currentTime"
            :time-slices="timeSlices"
            @enter="enterControls"
            @leave="mouseLeaveTimeline"
            @click="clickTimeline"
            @update-tooltip="updateTimelineTooltip"
            @grab="grabTimeline"
        ></PlayerTimeline>

        <PlayerTimelineTooltip
            v-if="tooltipShown"
            :x="tooltipX"
            :text="tooltipText"
            :time-slice="tooltipTimeSlice"
            :image="tooltipImage"
        ></PlayerTimelineTooltip>

        <PlayerConfig
            v-if="displayConfig"
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
        ></PlayerConfig>

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
            :shown="showControls && userControls"
            :fullscreen="fullscreen"
            :in-album="inAlbum"
            @update:expanded="onTopBarExpand"
            @click-player="clickControls"
        ></PlayerTopBar>

        <PlayerContextMenu
            v-if="contextMenuShown"
            ref="contextMenu"
            v-model:shown="contextMenuShown"
            v-model:loop="loop"
            v-model:slice-loop="sliceLoop"
            v-model:controls="userControls"
            v-model:time-slices-edit="timeSlicesEdit"
            type="video"
            :x="contextMenuX"
            :y="contextMenuY"
            :url="videoURL"
            :title="title"
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

<script setup lang="ts">
import {
    CURRENT_TIME_UPDATE_DELAY,
    getCachedInitialTime,
    getPlayerMuted,
    getPlayerVolume,
    getSelectedAudioTrack,
    getTogglePlayDelay,
    getUserSelectedResolutionVideo,
    setCachedInitialTime,
    setPlayerMuted,
    setPlayerVolume,
    setUserSelectedResolutionVideo,
} from "@/local-storage/player-preferences";
import { defineAsyncComponent, nextTick, onBeforeUnmount, onMounted, ref, useTemplateRef, watch } from "vue";
import VolumeControl from "./common/VolumeControl.vue";
import PlayerTopBar from "./common/PlayerTopBar.vue";
import { renderTimeSeconds } from "@/utils/time";
import { isTouchDevice } from "@/utils/touch";
import { getAssetURL } from "@/utils/api";
import { AUTO_LOOP_MIN_DURATION, loadCurrentMedia } from "@/global-state/media";
import { checkAuthenticationStatusSilent, isVaultLocked, refreshAuthenticationStatus } from "@/global-state/auth";
import type { PropType } from "vue";
import type { MediaData, MediaListItem } from "@/api/models";
import PlayerTooltip from "./common/PlayerTooltip.vue";
import PlayerControls from "./common/PlayerControls.vue";
import type { PlayerPlayFeedbackType } from "@/utils/player";
import PlayerPlayFeedback from "./common/PlayerPlayFeedback.vue";
import PlayerLoader from "./common/PlayerLoader.vue";
import PlayerAutoNextOverlay from "./common/PlayerAutoNextOverlay.vue";
import PlayerTimeline from "./common/PlayerTimeline.vue";
import PlayerTimelineTooltip from "./common/PlayerTimelineTooltip.vue";
import { useUserPermissions } from "@/composables/use-user-permissions";
import { useI18n } from "@/composables/use-i18n";
import { PLAYER_KEYBOARD_HANDLER_PRIORITY, usePlayerCommon } from "@/composables/use-player-common";
import { usePlayerCommonControls } from "@/composables/use-player-common-controls";
import { usePlayerAutoNext } from "@/composables/use-player-auto-next";
import { usePlayerNextOnEnd } from "@/composables/use-player-next-on-end";
import { usePlayerTimeSlices } from "@/composables/use-player-time-slices";
import { useTimeout } from "@/composables/use-timeout";
import { useInterval } from "@/composables/use-interval";
import { usePlayerMediaSession } from "@/composables/use-player-media-session";
import { useGlobalKeyboardHandler } from "@/composables/use-global-keyboard-handler";
import type { PositionEvent } from "@/utils/position-event";
import { positionEventFromMouseEvent } from "@/utils/position-event";
import { onDocumentEvent } from "@/composables/on-document-event";
import { usePlayerSubtitles } from "@/composables/use-player-subtitles";
import { showSnackBar } from "@/global-state/snack-bar";
import { isPlayerVisible } from "@/global-state/navigation";

const PlayerContextMenu = defineAsyncComponent({
    loader: () => import("@/components/player/common/PlayerContextMenu.vue"),
});

const PlayerConfig = defineAsyncComponent({
    loader: () => import("@/components/player/config/PlayerConfig.vue"),
});

const PlayerSubtitles = defineAsyncComponent({
    loader: () => import("@/components/player/common/PlayerSubtitles.vue"),
});

const TimeSlicesEditHelper = defineAsyncComponent({
    loader: () => import("@/components/widgets/TimeSlicesEditHelper.vue"),
});

const PlayerEncodingPending = defineAsyncComponent({
    loader: () => import("@/components/player/common/PlayerEncodingPending.vue"),
});

const TagsEditHelper = defineAsyncComponent({
    loader: () => import("@/components/widgets/TagsEditHelper.vue"),
});

const DescriptionWidget = defineAsyncComponent({
    loader: () => import("@/components/widgets/DescriptionWidget.vue"),
});

const PlayerAttachmentsList = defineAsyncComponent({
    loader: () => import("@/components/player/common/PlayerAttachmentsList.vue"),
});

const PlayerRelatedMediaList = defineAsyncComponent({
    loader: () => import("@/components/player/common/PlayerRelatedMediaList.vue"),
});

// True if it is a touch device
const IS_TOUCH_DEVICE = isTouchDevice();

// Volume control width (px)
const VOLUME_CONTROL_WIDTH = 80;

// Volume control width when miniature mode (px)
const VOLUME_CONTROL_WIDTH_MIN = 50;

// Ref to the container element
const container = useTemplateRef("container");

// Ref to the video element
const videoElement = useTemplateRef("videoElement");

// Ref to the audio track element
const audioTrackElement = useTemplateRef("audioTrackElement");

// Ref to the context menu
const contextMenu = useTemplateRef("contextMenu");

// Ref to the video scroller component
const videoScroller = useTemplateRef("videoScroller");

// Translation
const { $t } = useI18n();

// User permissions
const { canWrite } = useUserPermissions();

// Full screen model
const fullscreen = defineModel<boolean>("fullscreen");

// Display tag list widget?
const displayTagList = defineModel<boolean>("displayTagList");

// Display description widget
const displayDescription = defineModel<boolean>("displayDescription");

// Display controls? (user setting)
const userControls = defineModel<boolean>("userControls");

// Props
const props = defineProps({
    /**
     * Media ID
     */
    mid: {
        type: Number,
        required: true,
    },

    /**
     * Media metadata
     */
    metadata: {
        type: Object as PropType<MediaData>,
        required: true,
    },

    /**
     * Reload tick
     */
    rTick: {
        type: Number,
        required: true,
    },

    /**
     * Next element in album
     */
    next: Object as PropType<MediaListItem | null>,

    /**
     * Previous element in album
     */
    prev: Object as PropType<MediaListItem | null>,

    /**
     * True if media is in album
     */
    inAlbum: Boolean,

    /**
     * Has next element in page?
     */
    pageNext: Boolean,

    /**
     * Has previous element in page?
     */
    pagePrev: Boolean,

    /**
     * Miniature mode
     */
    min: Boolean,

    /**
     * Forced loop?
     */
    loopForced: Boolean,

    /**
     * Value of the loop if forced
     */
    loopForcedValue: Boolean,

    /**
     * Auto play?
     */
    autoPlay: Boolean,
});

// Emits
const emit = defineEmits<{
    /**
     * Go to the next media
     */
    (e: "go-next"): void;

    /**
     * Go to the previous media
     */
    (e: "go-prev"): void;

    /**
     * The user wants to delete the media
     */
    (e: "delete"): void;

    /**
     * The user wants to open the albums list modal
     */
    (e: "albums-open"): void;

    /**
     * The user wants to upen the size stats modal
     */
    (e: "stats-open"): void;

    /**
     * The audio ended
     */
    (e: "ended"): void;

    /**
     * The audio ended
     */
    (e: "force-loop", loopVal: boolean): void;
}>();

// Player common features
const { expandedTitle, expandedAlbum, helpTooltip, enterTooltip, leaveTooltip, clearTooltip, goNext, goPrev, toggleFullScreen } =
    usePlayerCommon(props, emit, fullscreen);

// Player common controls
const {
    hasAttachments,
    displayAttachments,
    hasDescription,
    hasRelatedMedia,
    displayRelatedMedia,
    displayConfig,
    contextMenuShown,
    contextMenuX,
    contextMenuY,
    onContextMenu,
    openDescription,
    openStats,
    openTags,
    showAttachments,
    showRelatedMedia,
    showConfig,
    manageAlbums,
    refreshDescription,
    clickControls,
} = usePlayerCommonControls(props, emit, {
    displayTagList,
    displayDescription,
    contextMenu,
});

// Is a short audio? (forced loop, auto-next)
const isShort = ref(false);

// Auto-next
const { clearAutoNextTimer, setupAutoNextTimer, toggleAutoNext } = usePlayerAutoNext(
    props,
    isShort,
    {
        displayConfig,
        expandedTitle,
        displayAttachments,
        displayRelatedMedia,
    },
    goNext,
);

// Next-on-end
const {
    nextEnd,
    autoNextPageDelay,
    pendingNextEnd,
    pendingNextEndSeconds,
    showNextEnd,
    hideNextEnd,
    onUpdateNextEnd,
    onUpdateAutoNextPageDelay,
} = usePlayerNextOnEnd(props, goNext);

// Loading status
const loading = ref(true);

// Playing status
const playing = ref(true);

// Timeout to delay the play/pause action
const togglePlayDelayTimeout = useTimeout();

// True if auto-play was applied
const autoPlayApplied = ref(false);

// Title
const title = ref("");

// Video URL
const videoURL = ref("");

// True if video is pending of being encoded
const videoPending = ref(false);

// If pending of encoding, the ID of the encoding task
const videoPendingTask = ref(0);

// If the task resulted in an error, the error message
const videoEncodeError = ref("");

// True if the audio current time can be saved
const canSaveTime = ref(true);

// The current time (seconds)
const currentTime = ref(0);

// The duration (seconds)
const duration = ref(0);

// The total buffered time (seconds)
const bufferedTime = ref(0);

// True if ended
const ended = ref(false);

// True if timeline is being grabbed
const timelineGrabbed = ref(false);

// Timestamp of the last event that changed the timeline
const lastTimeChangedEvent = ref(0);

// Timeline tooltip shown?
const tooltipShown = ref(false);

// Timeline tooltip text
const tooltipText = ref("");

// Time slice for the timeline tooltip
const tooltipTimeSlice = ref("");

// Position of the timeline tooltip
const tooltipX = ref(0);

// X position after updating the timeline
const tooltipEventX = ref(0);

// Image to display for the timeline tooltip
const tooltipImage = ref("");

// True to show controls
const showControls = ref(true);

// Timestamp of the last interaction with the controls
const lastControlsInteraction = ref(Date.now());

// True if the mouse is on the controls
const mouseInControls = ref(false);

// Loop audio when ended?
const loop = ref(false);

// Index of the current selected resolution
const currentResolution = ref(-1);

// Current volume
const volume = ref(getPlayerVolume());

// Muted?
const muted = ref(getPlayerMuted());

// Volume control shown?
const volumeShown = ref(IS_TOUCH_DEVICE);

// Internal tick for player config
const internalTick = ref(0);

// Playback speed
const speed = ref(1);

// Video scale
const scale = ref(1);

// Play / Pause feedback
const feedback = ref<PlayerPlayFeedbackType>("");

// True if refresh of the audio is required to play it
const requiresRefresh = ref(false);

// Selected audio track
const audioTrack = ref(getSelectedAudioTrack());

// URL of the audio track file
const audioTrackURL = ref("");

// True if waiting to load data into buffer
const isWaiting = ref(false);

// Timestamp of the 'waiting' event reception in the video element
const waitingTimestamp = ref(0);

// True if media could not be loaded and resulted in an error
const mediaError = ref(false);

// The error message for the media load process
const mediaErrorMessage = ref("");

// Timestamp for player tap detection
const timeStartTap = ref(0);

// True if scroll has been grabbed
const scrollGrabbed = ref(false);

// Scroll grab coordinates
const scrollGrabX = ref(0);
const scrollGrabY = ref(0);
const scrollGrabTop = ref(0);
const scrollGrabLeft = ref(0);

// True if the scroll has been moved
const scrollMoved = ref(false);

/**
 * Initializes the video
 */
const initializeVideo = () => {
    if (!props.metadata) {
        return;
    }

    isShort.value = props.metadata.is_anim || props.metadata.duration <= AUTO_LOOP_MIN_DURATION;
    canSaveTime.value = !props.metadata.force_start_beginning;

    initializeTimeSlices();

    currentTime.value = canSaveTime.value ? getCachedInitialTime(props.mid) : 0;
    duration.value = props.metadata.duration || 0;

    speed.value = 1;
    scale.value = 1;

    setDefaultLoop();

    currentResolution.value = getUserSelectedResolutionVideo(props.metadata);

    loading.value = true;
    playing.value = true;

    setVideoURL();
    onUpdateAudioTrack();
};

onMounted(initializeVideo);

watch(
    () => props.rTick,
    () => {
        internalTick.value++;
        resetSubtitles();
        autoPlayApplied.value = false;
        initializeVideo();
    },
);

// Clear video URL before the component unmounts
onBeforeUnmount(() => {
    videoURL.value = "";
    onClearURL();
});

/**
 * Called to clear the video URL
 */
const onClearURL = () => {
    const videoElem = videoElement.value;
    if (videoElem) {
        delete videoElem.src;
    }
};

/**
 * Sets the video URL
 */
const setVideoURL = () => {
    hideNextEnd();

    clearAutoNextTimer();

    mediaError.value = false;
    mediaErrorMessage.value = "";

    if (!props.metadata) {
        videoURL.value = "";
        title.value = "";
        onClearURL();
        duration.value = 0;
        loading.value = false;
        return;
    }

    title.value = props.metadata.title;

    if (currentResolution.value < 0) {
        if (props.metadata.encoded) {
            videoURL.value = getAssetURL(props.metadata.url);
            videoPending.value = false;
            videoPendingTask.value = 0;
            videoEncodeError.value = "";
            setupAutoNextTimer();
        } else {
            videoURL.value = "";
            onClearURL();
            videoPending.value = true;
            videoPendingTask.value = props.metadata.task;
            videoEncodeError.value = props.metadata.error || "";
            duration.value = 0;
            loading.value = false;
        }
    } else {
        if (props.metadata.resolutions && props.metadata.resolutions.length > currentResolution.value) {
            const res = props.metadata.resolutions[currentResolution.value];
            if (res.ready) {
                videoURL.value = getAssetURL(res.url);
                videoPending.value = false;
                videoPendingTask.value = 0;
                videoEncodeError.value = "";
                setupAutoNextTimer();
            } else {
                videoURL.value = "";
                onClearURL();
                videoPending.value = true;
                videoPendingTask.value = res.task;
                videoEncodeError.value = res.error || "";
                duration.value = 0;
                loading.value = false;
            }
        } else {
            videoURL.value = "";
            onClearURL();
            videoPending.value = true;
            videoPendingTask.value = 0;
            videoEncodeError.value = "";
            duration.value = 0;
            loading.value = false;
        }
    }
};

watch(videoURL, () => {
    if (videoURL.value) {
        loading.value = true;
    }
});

/**
 * Called when the selected resolution changes
 */
const onResolutionUpdated = () => {
    setUserSelectedResolutionVideo(props.metadata, currentResolution.value);
    setVideoURL();
};

/**
 * Called to update the audio track file
 */
const onUpdateAudioTrack = () => {
    if (!audioTrack.value || !props.metadata || !props.metadata.audios) {
        audioTrackURL.value = "";
        return;
    }

    for (const audio of props.metadata.audios) {
        if (audio.id === audioTrack.value) {
            audioTrackURL.value = getAssetURL(audio.url);
            return;
        }
    }

    audioTrackURL.value = "";
};

/**
 * Sets the default state for the loop
 */
const setDefaultLoop = () => {
    if (props.loopForced) {
        loop.value = props.loopForcedValue;
    } else {
        loop.value = (!props.next && !props.pageNext && !props.inAlbum) || !nextEnd.value;
    }
};

// Reset default loop when next media changes
watch([() => props.next, () => props.pageNext], setDefaultLoop);

/* Subtitles */

const { subtitles, resetSubtitles, updateSubtitles } = usePlayerSubtitles(currentTime);

/* Controls */

/**
 * Toggles play/pause without delay
 */
const togglePlayImmediate = () => {
    togglePlayDelayTimeout.clear();

    if (playing.value) {
        feedback.value = "pause";
        pause();
    } else {
        feedback.value = "play";
        play();
    }

    displayConfig.value = false;
    displayAttachments.value = false;
    displayRelatedMedia.value = false;
};

/**
 * Toggles play/pause with delay
 */
const togglePlay = () => {
    const delay = getTogglePlayDelay();

    if (playing.value) {
        if (togglePlayDelayTimeout.isSet()) {
            togglePlayDelayTimeout.clear();
            feedback.value = "";
        } else if (delay > 0) {
            feedback.value = "pause";
            togglePlayDelayTimeout.set(pause, delay);
        } else {
            feedback.value = "pause";
            pause();
        }
    } else {
        if (togglePlayDelayTimeout.isSet()) {
            togglePlayDelayTimeout.clear();
            feedback.value = "";
        } else if (delay > 0) {
            feedback.value = "play";
            togglePlayDelayTimeout.set(play, delay);
        } else {
            feedback.value = "play";
            play();
        }
    }

    displayConfig.value = false;
    displayAttachments.value = false;
    displayRelatedMedia.value = false;
};

/**
 * Enables loop and replays the media
 */
const enableLoopAndPlay = () => {
    loop.value = true;
    emit("force-loop", loop.value);
    play();
};

/**
 * Plays the media
 */
const play = () => {
    hideNextEnd();

    if (requiresRefresh.value) {
        requiresRefresh.value = false;
        loadCurrentMedia();
        return;
    }

    playing.value = true;

    const video = videoElement.value;

    if (video) {
        video.play();
    }
};

/**
 * Pauses the media
 */
const pause = () => {
    playing.value = false;

    const video = videoElement.value;

    if (video) {
        video.pause();
    }

    if (!loading.value && canSaveTime.value && video && !video.ended) {
        setCachedInitialTime(props.mid, currentTime.value);
    }

    lastTimeChangedEvent.value = Date.now();

    interactWithControls();
};

/**
 * Sets the current time
 * @param time The current time
 * @param save True to save the current time
 */
const setTime = (time: number, save?: boolean) => {
    currentTimeSlice.value = null;

    time = Math.max(0, time);
    time = Math.min(time, duration.value);

    if (isNaN(time) || !isFinite(time) || time < 0) {
        return;
    }

    currentTime.value = time;

    const video = videoElement.value;

    if (video) {
        video.currentTime = time;
    }

    if (save && canSaveTime.value) {
        setCachedInitialTime(props.mid, currentTime.value);
        lastTimeChangedEvent.value = Date.now();
    }

    if (time < duration.value) {
        ended.value = false;
    }

    updateSubtitles();
    updateCurrentTimeSlice();
};

/**
 * Called when the top-bar expand status changes
 */
const onTopBarExpand = () => {
    if (expandedTitle.value) {
        hideNextEnd();
        pause();
    } else if (!expandedAlbum.value) {
        play();
    }
};

/**
 * Called when the user interacts with the controls
 */
const interactWithControls = () => {
    showControls.value = true;
    lastControlsInteraction.value = Date.now();
};

/**
 * Called when the mouse enters the controls
 */
const enterControls = () => {
    mouseInControls.value = true;
};

/**
 * Called when the mouse  leaves the controls
 */
const leaveControls = () => {
    mouseInControls.value = false;
    volumeShown.value = IS_TOUCH_DEVICE;
    clearTooltip();
};

/**
 * Called when the volume is updated by the user
 */
const onUserVolumeUpdated = () => {
    setPlayerVolume(volume.value);
};

/**
 * Changes the volume
 * @param v The new volume
 */
const changeVolume = (v: number) => {
    volume.value = v;
    onUserVolumeUpdated();
};

/**
 * Called when the muted state is changed by the user
 */
const onUserMutedUpdated = () => {
    setPlayerMuted(muted.value);
};

/**
 * Toggles the muted state
 */
const toggleMuted = () => {
    muted.value = !muted.value;
    onUserMutedUpdated();
};

/* Timeline */

/**
 * Click event handler for the timeline
 * @param e The click event
 */
const clickTimeline = (e: Event) => {
    contextMenu.value?.hide();
    displayConfig.value = false;
    displayAttachments.value = false;
    displayRelatedMedia.value = false;
    e.stopPropagation();
};

/**
 * Grabs the timeline
 * @param x The X coordinate
 */
const grabTimeline = (x: number) => {
    timelineGrabbed.value = true;
    onTimelineSkip(x);
};

/**
 * Skips time based on the timeline click position
 * @param x The X coordinate
 */
const onTimelineSkip = (x: number) => {
    const bounds = (container.value?.querySelector(".player-timeline-back") as HTMLElement)?.getBoundingClientRect();

    if (!bounds) {
        return;
    }

    const offset = bounds.left;
    const width = bounds.width || 1;

    if (x < offset) {
        setTime(0);
    } else {
        const p = x - offset;
        const tP = Math.min(1, p / width);
        setTime(tP * duration.value);
    }
};

/**
 * Called when the mouse leaves the timeline
 */
const mouseLeaveTimeline = () => {
    if (!IS_TOUCH_DEVICE) {
        tooltipShown.value = false;
    }
    leaveControls();
};

/**
 * Gets the URL of a video preview for a specific time
 * @param time The time
 * @returns The URL of the image, or an empty string
 */
const getThumbnailForTime = (time: number): string => {
    if (duration.value <= 0 || !props.metadata || !props.metadata.video_previews || !props.metadata.video_previews_interval) {
        return "";
    }
    const thumbCount = Math.floor(duration.value / props.metadata.video_previews_interval);

    if (thumbCount <= 0) {
        return "";
    }

    let part = Math.floor(time / props.metadata.video_previews_interval);
    if (part > thumbCount) {
        part = thumbCount;
    }

    return getAssetURL(props.metadata.video_previews.replace("{INDEX}", "" + part));
};

/**
 * Updates the timeline tooltip
 * @param x The X coordinate
 */
const updateTimelineTooltip = (x: number) => {
    const bounds = (container.value?.querySelector(".player-timeline-back") as HTMLElement)?.getBoundingClientRect();

    if (!bounds) {
        return;
    }

    const offset = bounds.left;
    const width = bounds.width || 1;

    let time: number;

    if (x < offset) {
        time = 0;
    } else {
        const p = x - offset;
        const tP = Math.min(1, p / width);
        time = tP * duration.value;
    }

    tooltipShown.value = true;
    tooltipText.value = renderTimeSeconds(time);
    tooltipImage.value = getThumbnailForTime(time);
    tooltipTimeSlice.value = findTimeSliceName(time);
    tooltipEventX.value = x;

    nextTick(tick);
};

/* Time slices */

const {
    timeSlices,
    currentTimeSlice,
    currentTimeSliceName,
    timeSlicesEdit,
    sliceLoop,
    initializeTimeSlices,
    findTimeSliceName,
    updateCurrentTimeSlice,
    refreshTimeSlices,
} = usePlayerTimeSlices(props, currentTime, setTime);

/* Video scale */

/**
 * Renders scale
 * @param scale The scale
 * @returns The rendered value
 */
const renderScale = (scale: number): string => {
    if (scale > 1) {
        return Math.floor(scale * 100) + "%";
    } else if (scale < 1) {
        return Math.floor(scale * 100) + "%";
    } else {
        return $t("Normal");
    }
};

/**
 * Gets scale CSS style
 * @param scale The scale
 * @returns The value as percent
 */
const getScaleCss = (scale: number): string => {
    return Math.floor(scale * 100) + "%";
};

/**
 * Called wne scale value is updated by the user
 */
const onScaleUpdated = () => {
    if (scale.value > 1) {
        nextTick(centerScroll);
    }
};

/**
 * Mouse wheel event handler to change the player scale
 * @param e The mouse wheel event
 */
const onMouseWheel = (e: WheelEvent) => {
    if (e.ctrlKey) {
        e.preventDefault();
        e.stopPropagation();
        if (e.deltaY > 0) {
            scale.value = Math.max(1, scale.value - 0.1);
        } else {
            scale.value = Math.min(8, scale.value + 0.1);
        }

        showSnackBar($t("Scale") + ": " + renderScale(scale.value));
        onScaleUpdated();
    }
};

/**
 * Centers the scroller
 */
const centerScroll = () => {
    const scroller = videoScroller.value;

    if (!scroller) {
        return;
    }

    scroller.scrollTop = (scroller.scrollHeight - scroller.getBoundingClientRect().height) / 2;
    scroller.scrollLeft = (scroller.scrollWidth - scroller.getBoundingClientRect().width) / 2;
};

/**
 * Grabs scroll
 * @param e The position event
 */
const grabScroll = (e: PositionEvent) => {
    if (scale.value <= 1) {
        return; // Not scaled
    }

    leaveControls();

    if (displayConfig.value || contextMenuShown.value || displayAttachments.value || displayRelatedMedia.value) {
        displayConfig.value = false;
        contextMenu.value?.hide();
        displayAttachments.value = false;
        displayRelatedMedia.value = false;
        e.e.stopPropagation();
        return;
    }

    const scroller = videoScroller.value;

    if (!scroller) {
        return;
    }

    scrollGrabTop.value = scroller.scrollTop;
    scrollGrabLeft.value = scroller.scrollLeft;

    scrollGrabbed.value = true;
    scrollMoved.value = false;

    scrollGrabX.value = e.x;
    scrollGrabY.value = e.y;

    e.e.stopPropagation();
};

/**
 * Grabs scroll with the mouse
 * @param e The mouse event
 */
const grabScrollWithMouse = (e: MouseEvent) => {
    if (e.button !== 0) {
        return;
    }

    grabScroll(positionEventFromMouseEvent(e));
};

/**
 * Moves scroll using the mouse position
 * @param x The X coordinate
 * @param y The Y coordinate
 */
const moveScrollByMouse = (x: number, y: number) => {
    const scroller = videoScroller.value;

    if (!scroller) {
        return;
    }

    const rect = scroller.getBoundingClientRect();

    const maxScrollLeft = scroller.scrollWidth - rect.width;
    const maxScrollTop = scroller.scrollHeight - rect.height;

    const diffX = x - scrollGrabX.value;
    const diffY = y - scrollGrabY.value;

    scroller.scrollTop = Math.max(0, Math.min(maxScrollTop, scrollGrabTop.value - diffY));
    scroller.scrollLeft = Math.max(0, Math.min(maxScrollLeft, scrollGrabLeft.value - diffX));

    scrollMoved.value = true;
};

onDocumentEvent("mousemove", (e: MouseEvent) => {
    if (!scrollGrabbed.value) {
        return;
    }

    moveScrollByMouse(e.pageX, e.pageY);
});

onDocumentEvent("mouseup", (e: MouseEvent) => {
    if (e.button !== 0) {
        return;
    }

    if (!scrollGrabbed.value) {
        return;
    }

    scrollGrabbed.value = false;

    if (!scrollMoved.value) {
        clickPlayer();
        return;
    }

    scrollMoved.value = false;

    moveScrollByMouse(e.pageX, e.pageY);
});

/**
 * Event handler for 'touchstart' on the player
 * @param e The event
 */
const onScrollerTouchStart = (e: Event) => {
    if (scale.value <= 1) {
        return;
    }

    e.stopPropagation();
};

/**
 * Increments the vertical scroll
 * @param a An increment, 'home' to set it to the beginning, 'end' to set it to the end
 * @returns True if applied, false if it cannot be applied
 */
const incrementVerticalScroll = (a: number | "home" | "end"): boolean => {
    if (scale.value <= 1) {
        return false;
    }

    const el = videoScroller.value;

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
};

/**
 * Increments the horizontal scroll
 * @param a An increment, 'home' to set it to the beginning, 'end' to set it to the end
 * @returns True if applied, false if it cannot be applied
 */
const incrementHorizontalScroll = (a: number | "home" | "end"): boolean => {
    if (scale.value <= 1) {
        return false;
    }

    const el = videoScroller.value;

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
};

/* Player mouse / touch events */

/**
 * Called when the mouse leaves the player
 */
const mouseLeavePlayer = () => {
    timelineGrabbed.value = false;

    if (
        !playing.value ||
        expandedTitle.value ||
        expandedAlbum.value ||
        displayConfig.value ||
        displayAttachments.value ||
        displayRelatedMedia.value
    ) {
        return;
    }

    showControls.value = false;
    volumeShown.value = IS_TOUCH_DEVICE;
    clearTooltip();
};

/**
 * Event handler for 'mouseup'
 * @param e The mouse event
 */
const playerMouseUp = (e: MouseEvent) => {
    if (timelineGrabbed.value) {
        timelineGrabbed.value = false;
        onTimelineSkip(e.pageX);

        if (IS_TOUCH_DEVICE) {
            tooltipShown.value = false;
        }
    }
};

/**
 * Event handler for 'mousemove'
 * @param e The mouse event
 */
const playerMouseMove = (e: MouseEvent) => {
    interactWithControls();

    if (timelineGrabbed.value) {
        onTimelineSkip(e.pageX);
    }
};

/**
 * Event handler for 'touchmove'
 * @param e The touch event
 */
const playerTouchMove = (e: TouchEvent) => {
    interactWithControls();

    if (timelineGrabbed.value && e.touches[0]) {
        onTimelineSkip(e.touches[0].pageX);
        updateTimelineTooltip(e.touches[0].pageX);
    }
};

/**
 * Event handler for 'click'
 */
const clickPlayer = () => {
    leaveControls();

    if (displayConfig.value || contextMenuShown.value || displayAttachments.value || displayRelatedMedia.value) {
        displayConfig.value = false;
        contextMenu.value?.hide();
        displayAttachments.value = false;
        displayRelatedMedia.value = false;
    } else {
        togglePlay();
    }

    interactWithControls();
};

/**
 * Event handler for 'mousedown'
 * @param e The mouse event
 */
const onPlayerMouseDown = (e: MouseEvent) => {
    if (IS_TOUCH_DEVICE) {
        return;
    }

    if (e.button !== 0) {
        return;
    }

    if (contextMenuShown.value) {
        e.stopPropagation();
    }

    clickPlayer();
};

/**
 * Event handler for 'touchstart'
 * @param e The touch event
 */
const onPlayerTouchStart = (e: TouchEvent) => {
    if (contextMenuShown.value) {
        e.stopPropagation();
    }

    leaveControls();

    if (displayConfig.value || contextMenuShown.value || displayAttachments.value || displayRelatedMedia.value) {
        displayConfig.value = false;
        contextMenu.value?.hide();
        displayAttachments.value = false;
        displayRelatedMedia.value = false;
    } else {
        timeStartTap.value = Date.now();
    }

    interactWithControls();
};

// Max duration of a tap
const MAX_TIME_TAP = 500;

/**
 * Event handler for 'touchend'
 * @param e The touch event
 */
const onPlayerTouchEnd = (e: TouchEvent) => {
    if (timelineGrabbed.value) {
        timelineGrabbed.value = false;
        if (e.touches[0]) {
            onTimelineSkip(e.touches[0].pageX);
        }
    }

    tooltipShown.value = false;

    if (timeStartTap.value && Date.now() - timeStartTap.value < MAX_TIME_TAP) {
        togglePlay();
        timeStartTap.value = 0;
    }
};

/* Player media events */

/**
 * Event handler for 'loadedmetadata' on the audio track element
 */
const onAudioTrackLoadMetadata = () => {
    const audioElement = audioTrackElement.value;

    if (!audioElement) {
        return;
    }

    if (typeof audioElement.duration !== "number" || isNaN(audioElement.duration) || !isFinite(audioElement.duration)) {
        return;
    }

    const duration = audioElement.duration;

    if (typeof currentTime.value === "number" && !isNaN(currentTime.value) && isFinite(currentTime.value) && currentTime.value >= 0) {
        audioElement.currentTime = Math.min(currentTime.value, duration);
    }
};

/**
 * Event handler for 'loadedmetadata' on the video element
 * @param ev The event
 */
const onLoadMetaData = (ev: Event) => {
    const videoElement = ev.target as HTMLVideoElement;

    if (!videoElement) {
        return;
    }

    if (typeof videoElement.duration !== "number" || isNaN(videoElement.duration) || !isFinite(videoElement.duration)) {
        return;
    }

    duration.value = videoElement.duration;

    if (typeof currentTime.value === "number" && !isNaN(currentTime.value) && isFinite(currentTime.value) && currentTime.value >= 0) {
        videoElement.currentTime = Math.min(currentTime.value, duration.value);
        updateSubtitles();
        updateCurrentTimeSlice();
    }
};

/**
 * Event handler for 'timeupdate' on the video element
 * @param ev The event
 */
const onVideoTimeUpdate = (ev: Event) => {
    hideNextEnd();

    if (loading.value) {
        return;
    }

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

    currentTime.value = videoElement.currentTime;
    duration.value = videoElement.duration;

    if (!loading.value && canSaveTime.value && Date.now() - lastTimeChangedEvent.value > CURRENT_TIME_UPDATE_DELAY) {
        setCachedInitialTime(props.mid, currentTime.value);
        lastTimeChangedEvent.value = Date.now();
    }

    updateSubtitles();
    updateCurrentTimeSlice();

    if (audioTrackURL.value) {
        const audio = audioTrackElement.value;
        if (audio) {
            const correspondingTime = Math.min(videoElement.currentTime, audio.duration);

            if (Math.abs(correspondingTime - audio.currentTime) > speed.value / 10) {
                audio.currentTime = Math.min(videoElement.currentTime, audio.duration);
            }

            if (audio.paused !== videoElement.paused && videoElement.currentTime <= audio.duration) {
                if (videoElement.paused) {
                    audio.pause();
                } else {
                    audio.play();
                }
            }
        }
    }
};

/**
 * Event handler for 'canplay' on the video element
 */
const onCanPlay = () => {
    loading.value = false;

    if (autoPlayApplied.value) {
        if (playing.value) {
            const player = videoElement.value;

            if (!player) {
                return;
            }

            const promise = player.play();

            if (promise) {
                promise.catch(() => {
                    playing.value = false;
                    requiresRefresh.value = true;
                });
            }
        }
        return;
    }

    if (!props.autoPlay || expandedTitle.value) {
        autoPlayApplied.value = true;
        playing.value = false;
        return;
    }

    const player = videoElement.value;

    if (!player) {
        return;
    }

    const promise = player.play();

    if (promise) {
        promise.catch(() => {
            playing.value = false;
            requiresRefresh.value = true;
        });
    }

    autoPlayApplied.value = true;
};

/**
 * Event handler for 'canplay' on the audio track element
 */
const onAudioCanPlay = () => {
    const audio = audioTrackElement.value;

    if (!audio) {
        return;
    }

    const video = videoElement.value;

    if (!video) {
        return;
    }

    const correspondingTime = Math.min(video.currentTime, audio.duration);

    if (Math.abs(correspondingTime - audio.currentTime) > speed.value / 10) {
        audio.currentTime = Math.min(video.currentTime, audio.duration);
    }

    if (audio.paused !== video.paused && video.currentTime <= audio.duration) {
        if (video.paused) {
            audio.pause();
        } else {
            const promise = audio.play();

            if (promise) {
                promise.catch(() => {
                    pause();
                });
            }
        }
    }
};

/**
 * Handler for 'waiting' or 'playing'
 * @param b True if waiting, false if playing
 */
const onWaitForBuffer = (b: boolean) => {
    if (b) {
        isWaiting.value = true;
        waitingTimestamp.value = Date.now();
    } else {
        loading.value = false;
        isWaiting.value = false;
    }
};

/**
 * Event handler for 'ended'
 */
const onEnded = () => {
    loading.value = false;

    if (currentTimeSlice.value && sliceLoop.value) {
        setTime(currentTimeSlice.value.start, false);
        play();
        return;
    }

    if (canSaveTime.value) {
        setCachedInitialTime(props.mid, 0);
    }

    if (!loop.value && !isShort.value) {
        pause();

        ended.value = true;

        if (nextEnd.value) {
            if (props.next) {
                goNext();
            } else if (props.pageNext) {
                if (autoNextPageDelay.value) {
                    showNextEnd();
                } else {
                    goNext();
                }
            }
        }
    }
};

/**
 * Event handler for 'play'
 */
const onPlay = () => {
    playing.value = true;
};

/**
 * Event handler for 'pause'
 */
const onPause = () => {
    playing.value = false;
};

/**
 * Handler for the 'error' event
 */
const onMediaError = () => {
    if (!videoURL.value) {
        return;
    }
    if (!refreshAuthenticationStatus()) {
        mediaError.value = true;
        updateMediaErrorMessage();

        loading.value = false;

        checkAuthenticationStatusSilent();
    }
};

/**
 * Updates the media error message
 */
const updateMediaErrorMessage = () => {
    mediaErrorMessage.value = "";

    const mediaElem = videoElement.value;

    if (!mediaElem) {
        return;
    }

    const err = mediaElem.error;

    if (!err) {
        return;
    }

    mediaErrorMessage.value = err.message || "";

    if (mediaErrorMessage.value) {
        console.error(mediaErrorMessage.value);
    }
};

/* Ticks */

// Delay to change the loading status after the 'waiting' event is received
const WAITING_LOAD_DELAY = 1000;

// Delay to hide controls (milliseconds)
const CONTROLS_HIDE_DELAY = 2000;

/**
 * Checks the player status.
 * Ran every tick.
 */
const tick = () => {
    if (!loading.value && isWaiting.value && Date.now() - waitingTimestamp.value > WAITING_LOAD_DELAY) {
        loading.value = true;
    }

    if (showControls.value && !mouseInControls.value && playing.value && !expandedTitle.value && !expandedAlbum.value) {
        if (Date.now() - lastControlsInteraction.value > CONTROLS_HIDE_DELAY) {
            showControls.value = false;
            volumeShown.value = IS_TOUCH_DEVICE;
            helpTooltip.value = "";
            displayConfig.value = false;
            displayAttachments.value = false;
            displayRelatedMedia.value = false;
        }
    }

    const video = videoElement.value;

    if (video && video.buffered.length > 0) {
        bufferedTime.value = video.buffered.end(video.buffered.length - 1);
    } else {
        bufferedTime.value = 0;
    }

    if (tooltipShown.value) {
        const tooltip = container.value?.querySelector(".player-tooltip");
        const containerBounds = container.value?.getBoundingClientRect();

        if (tooltip && containerBounds) {
            let x = tooltipEventX.value;

            const toolTipWidth = tooltip.getBoundingClientRect().width;
            const leftPlayer = containerBounds.left;
            const widthPlayer = containerBounds.width;

            x = x - Math.floor(toolTipWidth / 2);

            if (x + toolTipWidth > leftPlayer + widthPlayer - 20) {
                x = leftPlayer + widthPlayer - 20 - toolTipWidth;
            }

            if (x < leftPlayer + 10) {
                x = leftPlayer + 10;
            }

            tooltipX.value = x - leftPlayer;
        }
    }
};

// Tick delay (milliseconds)
const TICK_DELAY = 100;

// Run ticks continuously
useInterval().set(tick, TICK_DELAY);

/* Media session */

usePlayerMediaSession(
    ["play", "pause", "nexttrack", "previoustrack", "seekbackward", "seekforward", "seekto"],
    (event: MediaSessionActionDetails) => {
        if (!event || !event.action) {
            return;
        }
        switch (event.action) {
            case "play":
                play();
                break;
            case "pause":
                pause();
                break;
            case "nexttrack":
                goNext();
                break;
            case "previoustrack":
                goPrev();
                break;
            case "seekbackward":
                setTime(currentTime.value - Math.max(1, Math.abs(event.seekOffset || 5)), true);
                break;
            case "seekforward":
                setTime(currentTime.value + Math.max(1, Math.abs(event.seekOffset || 5)), true);
                break;
            case "seekto":
                if (typeof event.seekTime === "number" && !isNaN(event.seekTime) && isFinite(event.seekTime)) {
                    setTime(event.seekTime, !event.fastSeek);
                }
                break;
        }
    },
    playing,
);

/* Keyboard handler */

useGlobalKeyboardHandler((event: KeyboardEvent): boolean => {
    if (isVaultLocked() || !isPlayerVisible() || !event.key || (event.ctrlKey && event.key !== "+" && event.key !== "-")) {
        return false;
    }

    let caught = true;

    const shifting = event.shiftKey;

    switch (event.key) {
        case "A":
        case "a":
            manageAlbums();
            break;
        case "i":
        case "I":
            openDescription();
            break;
        case "t":
        case "T":
            openTags();
            break;
        case "S":
        case "s":
            showConfig();
            break;
        case "M":
        case "m":
            toggleMuted();
            volumeShown.value = true;
            helpTooltip.value = "volume";
            break;
        case " ":
        case "K":
        case "k":
        case "Enter":
            togglePlayImmediate();
            break;
        case "ArrowUp":
            if (!shifting) {
                if (incrementVerticalScroll(-40)) {
                    break;
                }
                changeVolume(Math.min(1, volume.value + 0.05));
            } else {
                changeVolume(Math.min(1, volume.value + 0.01));
            }
            if (muted.value) {
                muted.value = false;
                onUserMutedUpdated();
            }
            volumeShown.value = true;
            helpTooltip.value = "volume";
            break;
        case "ArrowDown":
            if (!shifting) {
                if (incrementVerticalScroll(40)) {
                    break;
                }
                changeVolume(Math.max(0, volume.value - 0.05));
            } else {
                changeVolume(Math.max(0, volume.value - 0.01));
            }
            if (muted.value) {
                muted.value = false;
                onUserMutedUpdated();
            }
            volumeShown.value = true;
            helpTooltip.value = "volume";
            break;
        case "ArrowLeft":
            if (shifting || event.altKey) {
                if (props.prev || props.pagePrev) {
                    goPrev();
                } else {
                    caught = false;
                }
            } else if (!incrementHorizontalScroll(-40)) {
                if (isShort.value) {
                    if (props.prev || props.pagePrev) {
                        goPrev();
                    } else {
                        caught = false;
                    }
                } else {
                    setTime(currentTime.value - 5, true);
                }
            }
            break;

        case "ArrowRight":
            if (shifting || event.altKey) {
                if (props.next || props.pageNext) {
                    goNext();
                } else {
                    caught = false;
                }
            } else if (!incrementHorizontalScroll(40)) {
                if (isShort.value) {
                    if (props.next || props.pageNext) {
                        goNext();
                    } else {
                        caught = false;
                    }
                } else {
                    setTime(currentTime.value + 5, true);
                }
            }
            break;

        case ".":
            if (!playing.value) {
                setTime(currentTime.value - 1 / 30);
            }
            break;
        case ",":
            if (!playing.value) {
                setTime(currentTime.value + 1 / 30);
            }
            break;
        case "Home":
            if (event.altKey) {
                caught = false;
            } else if (shifting) {
                if (!incrementVerticalScroll("home")) {
                    caught = false;
                }
            } else if (!incrementHorizontalScroll("home")) {
                setTime(0, true);
            }
            break;
        case "End":
            if (event.altKey) {
                caught = false;
            } else if (shifting) {
                if (!incrementVerticalScroll("end")) {
                    caught = false;
                }
            } else if (!incrementHorizontalScroll("end")) {
                setTime(duration.value, true);
            }
            break;
        case "V":
        case "v":
            sliceLoop.value = !sliceLoop.value;

            if (sliceLoop.value) {
                showSnackBar($t("Slice loop enabled"));
            } else {
                showSnackBar($t("Slice loop disabled"));
            }
            break;
        case "l":
        case "L":
            if (event.altKey || shifting || isShort.value) {
                caught = false;
            } else {
                loop.value = !loop.value;

                if (loop.value) {
                    showSnackBar($t("Loop enabled"));
                } else {
                    showSnackBar($t("Loop disabled"));
                }

                emit("force-loop", loop.value);
            }
            break;
        case "b":
        case "B":
            if (currentTimeSlice.value) {
                setTime(currentTimeSlice.value.start, true);
            }
            break;
        case "j":
        case "J":
            if (currentTimeSlice.value) {
                setTime(currentTimeSlice.value.end, true);
            }
            break;
        case "C":
        case "c":
            userControls.value = !userControls.value;
            break;
        case "+":
            scale.value = Math.min(8, scale.value + (shifting ? 0.01 : 0.1));

            showSnackBar($t("Scale") + ": " + renderScale(scale.value));

            onScaleUpdated();
            break;
        case "-":
            scale.value = Math.max(1, scale.value - (shifting ? 0.01 : 0.1));

            showSnackBar($t("Scale") + ": " + renderScale(scale.value));

            onScaleUpdated();
            break;
        case "n":
        case "N":
            if (canWrite.value) {
                timeSlicesEdit.value = true;
            }
            break;
        case "x":
        case "X":
            toggleAutoNext();
            break;
        case "z":
        case "Z":
            scale.value = 1;
            showSnackBar($t("Scale") + ": " + renderScale(scale.value));
            onScaleUpdated();
            break;
        default:
            caught = false;
    }

    if (caught) {
        interactWithControls();
    }

    return caught;
}, PLAYER_KEYBOARD_HANDLER_PRIORITY);
</script>
