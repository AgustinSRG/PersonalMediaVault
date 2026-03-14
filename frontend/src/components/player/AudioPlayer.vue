<template>
    <div
        ref="container"
        class="audio-player player-settings-no-trap"
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
    >
        <audio
            ref="audioElement"
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

        <canvas v-if="audioURL && animationColors !== 'none'" ref="canvasElement"></canvas>

        <div v-if="audioURL" class="audio-no-animation">
            <div v-if="thumbnail && showThumbnail" class="audio-no-animation-thumbnail-container">
                <img class="audio-no-animation-thumbnail" :src="getAssetURL(thumbnail)" :alt="title" :title="title" loading="lazy" />
            </div>
            <div class="audio-no-animation-title" :class="{ hidden: !showTitle }">{{ title }}</div>
        </div>

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
            v-if="(!loading && !audioURL && audioPending) || mediaError"
            :mid="mid"
            :tid="audioPendingTask"
            :res="-1"
            :error="mediaError"
            :error-message="mediaErrorMessage"
            :encoding-error="audioEncodeError"
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
            :type="'audio'"
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

        <PlayerTimelineTooltip v-if="tooltipShown" :x="tooltipX" :text="tooltipText" :time-slice="tooltipTimeSlice"></PlayerTimelineTooltip>

        <PlayerConfig
            v-if="displayConfig"
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
            type="audio"
            :x="contextMenuX"
            :y="contextMenuY"
            :url="audioURL"
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
    getAudioAnimationStyle,
    getCachedInitialTime,
    getPlayerMuted,
    getPlayerVolume,
    getShowAudioThumbnail,
    getShowAudioTitle,
    getTogglePlayDelay,
    setAudioAnimationStyle,
    setCachedInitialTime,
    setPlayerMuted,
    setPlayerVolume,
    setShowAudioThumbnail,
    setShowAudioTitle,
} from "@/local-storage/player-preferences";
import type { PropType } from "vue";
import { defineAsyncComponent, nextTick, onBeforeUnmount, onMounted, ref, shallowRef, useTemplateRef, watch } from "vue";
import VolumeControl from "./common/VolumeControl.vue";
import PlayerTopBar from "./common/PlayerTopBar.vue";
import { renderTimeSeconds } from "@/utils/time";
import { isTouchDevice } from "@/utils/touch";
import { getAssetURL } from "@/utils/api";
import { AUTO_LOOP_MIN_DURATION, loadCurrentMedia } from "@/global-state/media";
import { AppStatus } from "@/global-state/app-status";
import type { ColorThemeName } from "@/local-storage/app-preferences";
import { getTheme } from "@/local-storage/app-preferences";
import type { MediaData, MediaListItem } from "@/api/models";
import PlayerTooltip from "./common/PlayerTooltip.vue";
import { EVENT_NAME_THEME_CHANGED } from "@/global-state/app-events";
import PlayerControls from "./common/PlayerControls.vue";
import type { PlayerPlayFeedbackType } from "@/utils/player";
import PlayerPlayFeedback from "./common/PlayerPlayFeedback.vue";
import PlayerLoader from "./common/PlayerLoader.vue";
import PlayerAutoNextOverlay from "./common/PlayerAutoNextOverlay.vue";
import PlayerTimeline from "./common/PlayerTimeline.vue";
import PlayerTimelineTooltip from "./common/PlayerTimelineTooltip.vue";
import { useUserPermissions } from "@/composables/use-user-permissions";
import { PLAYER_KEYBOARD_HANDLER_PRIORITY, usePlayerCommon } from "@/composables/use-player-common";
import { useInterval } from "@/composables/use-interval";
import { onApplicationEvent } from "@/composables/on-app-event";
import { usePlayerCommonControls } from "@/composables/use-player-common-controls";
import { usePlayerAutoNext } from "@/composables/use-player-auto-next";
import { usePlayerNextOnEnd } from "@/composables/use-player-next-on-end";
import { usePlayerMediaSession } from "@/composables/use-player-media-session";
import { useGlobalKeyboardHandler } from "@/composables/use-global-keyboard-handler";
import { useI18n } from "@/composables/use-i18n";
import { usePlayerTimeSlices } from "@/composables/use-player-time-slices";
import { useTimeout } from "@/composables/use-timeout";
import { usePlayerSubtitles } from "@/composables/use-player-subtitles";
import { showSnackBar } from "@/global-state/snack-bar";
import { checkAuthenticationStatusSilent, isVaultLocked, refreshAuthenticationStatus } from "@/global-state/auth";

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

// Ref to the audio element
const audioElement = useTemplateRef("audioElement");

// Ref to the canvas element
const canvasElement = useTemplateRef("canvasElement");

// Ref to the context menu
const contextMenu = useTemplateRef("contextMenu");

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

// Thumbnail
const thumbnail = ref("");

// Audio file URL
const audioURL = ref("");

// Audio pending of being encoded?
const audioPending = ref(false);

// If pending, the ID of the encoding task
const audioPendingTask = ref(0);

// Encoding error if the audio could not be encoded
const audioEncodeError = ref("");

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

// True to show controls
const showControls = ref(true);

// Timestamp of the last interaction with the controls
const lastControlsInteraction = ref(Date.now());

// True if the mouse is on the controls
const mouseInControls = ref(false);

// Loop audio when ended?
const loop = ref(false);

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

// Play / Pause feedback
const feedback = ref<PlayerPlayFeedbackType>("");

// Configured animation colors
const animationColors = ref(getAudioAnimationStyle());

// Shows title
const showTitle = ref(getShowAudioTitle());

// Shows thumbnail
const showThumbnail = ref(getShowAudioThumbnail());

// True if refresh of the audio is required to play it
const requiresRefresh = ref(false);

// Current theme
const theme = ref(getTheme());

onApplicationEvent(EVENT_NAME_THEME_CHANGED, (t: ColorThemeName) => {
    theme.value = t;
});

// True if media could not be loaded and resulted in an error
const mediaError = ref(false);

// The error message for the media load process
const mediaErrorMessage = ref("");

// Timestamp for player tap detection
const timeStartTap = ref(0);

// Audio context to analyze the audio
const audioContext = shallowRef<AudioContext | null>(null);

// Audio source to analyze the audio
const audioSource = shallowRef<MediaElementAudioSourceNode | null>(null);

// Audio analyzer
const audioAnalyser = shallowRef<AnalyserNode | null>(null);

// Data produced by the audio analyzer
const audioAnalyserData = shallowRef<Uint8Array | null>(null);

// Interval to render the audio data
const rendererTimer = useInterval();

/**
 * Initializes the audio
 */
const initializeAudio = () => {
    if (!props.metadata) {
        return;
    }

    isShort.value = props.metadata.duration <= AUTO_LOOP_MIN_DURATION;
    canSaveTime.value = !props.metadata.force_start_beginning;

    initializeTimeSlices();

    currentTime.value = canSaveTime.value ? getCachedInitialTime(props.mid) : 0;
    duration.value = props.metadata.duration || 0;

    speed.value = 1;

    setDefaultLoop();

    loading.value = true;
    playing.value = true;

    clearAudioRenderer();
    setAudioURL();
};

onMounted(initializeAudio);

watch(
    () => props.rTick,
    () => {
        internalTick.value++;
        resetSubtitles();
        autoPlayApplied.value = false;
        initializeAudio();
    },
);

// Clear audio URL before the component unmounts
onBeforeUnmount(() => {
    audioURL.value = "";
    onClearURL();
});

/**
 * Called to clear the audio URL
 */
const onClearURL = () => {
    const audioElem = audioElement.value;
    if (audioElem) {
        delete audioElem.src;
    }
};

/**
 * Sets the audio URL to start playing the audio
 */
const setAudioURL = () => {
    hideNextEnd();

    clearAutoNextTimer();

    mediaError.value = false;
    mediaErrorMessage.value = "";

    if (!props.metadata) {
        audioURL.value = "";
        thumbnail.value = "";
        title.value = "";
        onClearURL();
        duration.value = 0;
        loading.value = false;
        clearAudioRenderer();
        audioElement.value?.load();
        return;
    }

    thumbnail.value = props.metadata.thumbnail;
    title.value = props.metadata.title;

    if (props.metadata.encoded) {
        audioURL.value = getAssetURL(props.metadata.url);
        audioPending.value = false;
        audioPendingTask.value = 0;
        audioEncodeError.value = "";
        audioElement.value?.load();
        setupAutoNextTimer();
    } else {
        audioURL.value = "";
        onClearURL();
        audioPending.value = true;
        audioPendingTask.value = props.metadata.task;
        audioEncodeError.value = props.metadata.error || "";
        duration.value = 0;
        loading.value = false;
        clearAudioRenderer();
        audioElement.value?.load();
    }
};

watch(audioURL, () => {
    if (audioURL.value) {
        loading.value = true;
    }
});

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
 * Plays the audio
 */
const play = () => {
    hideNextEnd();

    if (requiresRefresh.value) {
        requiresRefresh.value = false;
        loadCurrentMedia();
        return;
    }

    playing.value = true;

    const audio = audioElement.value;

    if (audio) {
        audio.play();
    }
};

/**
 * Pauses the audio
 */
const pause = () => {
    playing.value = false;

    const audio = audioElement.value;

    if (audio) {
        audio.pause();
    }

    if (!loading.value && canSaveTime.value && audio && !audio.ended) {
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

    const audio = audioElement.value;

    if (audio) {
        audio.currentTime = time;
    }

    if (canSaveTime.value && save) {
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
 * Called when the animation colors are updated by the user
 */
const onUpdateAnimColors = () => {
    setAudioAnimationStyle(animationColors.value);
};

/**
 * Called when the "Show title" is updated by the user
 */
const onUpdateShowTitle = () => {
    setShowAudioTitle(showTitle.value);
};

/**
 * Called when the "Show thumbnail" is updated by the user
 */
const onUpdateShowThumbnail = () => {
    setShowAudioThumbnail(showThumbnail.value);
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

/* Audio renderer */

/**
 * Clears the audio renderer
 */
const clearAudioRenderer = () => {
    rendererTimer.clear();

    if (audioSource.value) {
        audioSource.value.disconnect();
    }

    if (audioAnalyser.value) {
        audioAnalyser.value.disconnect();
    }

    audioAnalyser.value = null;
    audioAnalyserData.value = null;
};

onBeforeUnmount(clearAudioRenderer);

/**
 * Closes the audio context used for the render
 */
const closeAudioContext = () => {
    if (audioContext.value) {
        audioContext.value.close();
    }

    audioContext.value = null;
    audioSource.value = null;
};

onBeforeUnmount(closeAudioContext);

/**
 * Sets up the audio renderer
 */
const setupAudioRenderer = () => {
    if (!audioContext.value) {
        setupAudioContext();
    }

    if (!audioContext.value) {
        return;
    }

    clearAudioRenderer();

    if (audioURL.value) {
        const context = audioContext.value;
        const source = audioSource.value;

        const analyser = context.createAnalyser();

        audioSource.value = source;
        audioAnalyser.value = analyser;
        source.connect(analyser);
        analyser.connect(context.destination);

        analyser.fftSize = 256;

        audioAnalyserData.value = new Uint8Array(analyser.frequencyBinCount);

        // Set renderer at 30 fps
        rendererTimer.set(audioAnimationFrame, Math.floor(1000 / 30));
    } else {
        audioContext.value = null;
        audioSource.value = null;
        audioAnalyser.value = null;
        audioAnalyserData.value = null;
    }
};

/**
 * Sets up the audio context for analysis
 */
const setupAudioContext = () => {
    if (!audioElement.value) {
        return;
    }

    const context = new AudioContext();
    audioContext.value = context;
    const source = context.createMediaElementSource(audioElement.value);
    audioSource.value = source;
};

/**
 * Renders the audio animation.
 * Called every frame
 */
const audioAnimationFrame = () => {
    if (!playing.value || animationColors.value === "none") {
        return;
    }

    const analyser = audioAnalyser.value;
    const canvas = canvasElement.value;

    if (!analyser || !canvas) {
        return;
    }

    const rect = canvas.getBoundingClientRect();
    if (canvas.width !== rect.width || canvas.height !== rect.height) {
        canvas.width = rect.width;
        canvas.height = rect.height;
    }

    let bufferLength = analyser.frequencyBinCount;

    const dataArray = audioAnalyserData.value;

    if (!dataArray) {
        return;
    }

    analyser.getByteFrequencyData(dataArray as Uint8Array<ArrayBuffer>);

    bufferLength = Math.floor(Math.max(1, bufferLength / 2));

    const WIDTH = canvas.width;
    const HEIGHT = canvas.height;

    const barWidth = Math.max(1, (WIDTH - (bufferLength - 1)) / bufferLength);

    const ctx = canvas.getContext("2d") as CanvasRenderingContext2D;

    ctx.fillStyle = theme.value === "light" ? "#fff" : "#000";
    ctx.fillRect(0, 0, WIDTH, HEIGHT);

    let x = 0;

    for (let i = 0; i < bufferLength; i++) {
        const barHeight = dataArray[i];

        switch (animationColors.value) {
            case "gradient":
                {
                    if (theme.value === "light") {
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
                    ctx.fillStyle = "hsl(" + (i * 300) / bufferLength + ", " + "60%" + ", " + (0.3 + 0.3 * (barHeight / 255)) * 100 + "%)";
                }
                break;
            default:
                ctx.fillStyle = theme.value === "light" ? "rgba(0, 0, 0, 0.5)" : "rgba(255, 255, 255, 0.5)";
        }

        const trueHeight = Math.floor(HEIGHT * (barHeight / 255));

        ctx.fillRect(x, HEIGHT - trueHeight, barWidth, trueHeight);

        x += barWidth + 1;
    }
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
 * Event handler for 'loadedmetadata'
 * @param ev The event
 */
const onLoadMetaData = (ev: Event) => {
    const audioElement = ev.target as HTMLAudioElement;

    if (!audioElement) {
        return;
    }

    if (typeof audioElement.duration !== "number" || isNaN(audioElement.duration) || !isFinite(audioElement.duration)) {
        return;
    }

    duration.value = audioElement.duration;

    if (typeof currentTime.value === "number" && !isNaN(currentTime.value) && isFinite(currentTime.value) && currentTime.value >= 0) {
        audioElement.currentTime = Math.min(currentTime.value, duration.value);
        updateSubtitles();
        updateCurrentTimeSlice();
    }
};

/**
 * Event handler for 'timeupdate'
 * @param ev The event
 */
const onAudioTimeUpdate = (ev: Event) => {
    hideNextEnd();

    if (loading.value) {
        return;
    }

    const audioElement = ev.target as HTMLAudioElement;

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

    currentTime.value = audioElement.currentTime;
    duration.value = audioElement.duration;

    if (!loading.value && canSaveTime.value && Date.now() - lastTimeChangedEvent.value > CURRENT_TIME_UPDATE_DELAY) {
        setCachedInitialTime(props.mid, currentTime.value);
        lastTimeChangedEvent.value = Date.now();
    }

    updateSubtitles();
    updateCurrentTimeSlice();
};

/**
 * Event handler for 'canplay'
 */
const onCanPlay = () => {
    loading.value = false;

    if (autoPlayApplied.value) {
        if (playing.value) {
            const player = audioElement.value;
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
    const player = audioElement.value;

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
 * Handler for 'waiting' or 'playing'
 * @param b True if waiting, false if playing
 */
const onWaitForBuffer = (b: boolean) => {
    loading.value = b;
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
    setupAudioRenderer();
};

/**
 * Event handler for 'pause'
 */
const onPause = () => {
    playing.value = false;
    clearAudioRenderer();
};

/**
 * Handler for the 'error' event
 */
const onMediaError = () => {
    if (!audioURL.value) {
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

    const mediaElem = audioElement.value;

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

// Delay to hide controls (milliseconds)
const CONTROLS_HIDE_DELAY = 2000;

/**
 * Checks the player status.
 * Ran every tick.
 */
const tick = () => {
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

    const audio = audioElement.value;

    if (audio && audio.buffered.length > 0) {
        bufferedTime.value = audio.buffered.end(audio.buffered.length - 1);
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
    if (isVaultLocked() || !AppStatus.IsPlayerVisible() || !event.key || event.ctrlKey) {
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
            } else {
                setTime(currentTime.value - 5, true);
            }
            break;
        case "ArrowRight":
            if (shifting || event.altKey) {
                if (props.next || props.pageNext) {
                    goNext();
                } else {
                    caught = false;
                }
            } else {
                setTime(currentTime.value + 5, true);
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
            if (event.altKey || shifting) {
                caught = false;
            } else {
                setTime(0, true);
            }
            break;
        case "End":
            if (event.altKey || shifting) {
                caught = false;
            } else {
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
        default:
            caught = false;
    }

    if (caught) {
        interactWithControls();
    }

    return caught;
}, PLAYER_KEYBOARD_HANDLER_PRIORITY);
</script>
