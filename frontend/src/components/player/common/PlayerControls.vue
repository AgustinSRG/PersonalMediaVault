<template>
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
            <!-- Go Previous -->

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

            <!-- Play / Pause -->

            <button v-if="!canPlay" disabled type="button" :title="$t('Play')" class="player-btn">
                <i class="fas fa-play"></i>
            </button>
            <button
                v-else-if="!playing"
                type="button"
                :title="$t('Play')"
                class="player-btn player-play-btn"
                @click="play"
                @mouseenter="enterTooltip('play')"
                @mouseleave="leaveTooltip('play')"
            >
                <i class="fas fa-play"></i>
            </button>
            <button
                v-else
                type="button"
                :title="$t('Pause')"
                class="player-btn player-play-btn"
                @click="pause"
                @mouseenter="enterTooltip('pause')"
                @mouseleave="leaveTooltip('pause')"
            >
                <i class="fas fa-pause"></i>
            </button>

            <!-- Go Next -->

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

            <!-- Custom controls depending on the player -->

            <slot></slot>
        </div>

        <div class="player-controls-right">
            <!-- Description -->

            <button
                v-if="hasDescription"
                type="button"
                :title="$t('Description')"
                class="player-btn player-btn-hide-mobile"
                @click="openDescription"
                @mouseenter="enterTooltip('desc')"
                @mouseleave="leaveTooltip('desc')"
            >
                <i class="fas fa-file-lines"></i>
            </button>

            <!-- Attachments -->

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

            <!-- Related media -->

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

            <!-- Albums -->

            <button
                v-if="isNotEmpty"
                type="button"
                :title="$t('Manage albums')"
                class="player-btn"
                @click="manageAlbums"
                @mouseenter="enterTooltip('albums')"
                @mouseleave="leaveTooltip('albums')"
            >
                <i class="fas fa-list-ol"></i>
            </button>

            <!-- Configuration -->

            <button
                v-if="isNotEmpty"
                type="button"
                :title="$t('Player Configuration')"
                class="player-btn player-settings-no-trap"
                @click="showConfig"
                @mouseenter="enterTooltip('config')"
                @mouseleave="leaveTooltip('config')"
            >
                <i class="fas fa-cog"></i>
            </button>

            <!-- Full screen -->

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
                v-else
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
</template>

<script setup lang="ts">
import type { MediaListItem } from "@/api/models";
import { useI18n } from "@/composables/use-i18n";
import { stopPropagationEvent } from "@/utils/events";
import type { HelpTooltipType } from "@/utils/player-tooltip";
import type { PropType } from "vue";
import { computed } from "vue";

// Translation
const { $t } = useI18n();

// Props
const props = defineProps({
    type: {
        type: String as PropType<"empty" | "audio" | "video" | "image">,
        required: true,
    },

    /**
     * Show the controls?
     */
    showControls: {
        type: Boolean,
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
     * Has next element i n page?
     */
    pageNext: Boolean,

    /**
     * Has previous element in page?
     */
    pagePrev: Boolean,

    /**
     * Playing?
     */
    playing: Boolean,

    /**
     * Has description?
     */
    hasDescription: Boolean,

    /**
     * Has attachments?
     */
    hasAttachments: Boolean,

    /**
     * Has related media?
     */
    hasRelatedMedia: Boolean,

    /**
     * Is full screen
     */
    fullscreen: Boolean,
});

// True if the player is not EmptyPlayer
const isNotEmpty = computed(() => props.type !== "empty");

// True if the media can be played
const canPlay = computed(() => ["video", "audio"].includes(props.type));

// Emits
const emit = defineEmits<{
    /**
     * Mouse enters the controls
     */
    (e: "enter"): void;

    /**
     * Mouse leaves the controls
     */
    (e: "leave"): void;

    /**
     * Enter a tooltip
     */
    (e: "enter-tooltip", t: HelpTooltipType): void;

    /**
     * Leave a tooltip
     */
    (e: "leave-tooltip", t: HelpTooltipType): void;

    /**
     * Users clicked the controls
     */
    (e: "click", event: Event): void;

    /**
     * Goes to the previous element
     */
    (e: "go-prev"): void;

    /**
     * Play the media
     */
    (e: "play"): void;

    /**
     * Pause the media
     */
    (e: "pause"): void;

    /**
     * Goes to the next element
     */
    (e: "go-next"): void;

    /**
     * Opens the description widget
     */
    (e: "open-description"): void;

    /**
     * Opens the attachments
     */
    (e: "open-attachments"): void;

    /**
     * Opens related media
     */
    (e: "open-related-media"): void;

    /**
     * Opens albums
     */
    (e: "open-albums"): void;

    /**
     * Opens the player config
     */
    (e: "open-config"): void;

    /**
     * Toggles full screen
     */
    (e: "toggle-full-screen"): void;
}>();

/**
 * The mouse enters the controls
 */
const enterControls = () => {
    emit("enter");
};

/**
 * The mouse leaves the controls
 */
const leaveControls = () => {
    emit("leave");
};

/**
 * The user clicked the controls
 * @param e The click event
 */
const clickControls = (e: Event) => {
    emit("click", e);
};

/**
 * Goes to the previous element
 */
const goPrev = () => {
    if (props.prev || props.pagePrev) {
        emit("go-prev");
    }
};

/**
 * Goes to the next element
 */
const goNext = () => {
    if (props.next || props.pageNext) {
        emit("go-next");
    }
};

/**
 * Enters a tooltip
 * @param t The tooltip
 */
const enterTooltip = (t: HelpTooltipType) => {
    emit("enter-tooltip", t);
};

/**
 * Leaves a tooltip
 * @param t The tooltip
 */
const leaveTooltip = (t: HelpTooltipType) => {
    emit("leave-tooltip", t);
};

/**
 * Plays the media
 */
const play = () => {
    emit("play");
};

/**
 * Pauses the media
 */
const pause = () => {
    emit("pause");
};

/**
 * Opens the description widget
 */
const openDescription = () => {
    emit("open-description");
};

/**
 * Shows the attachments
 * @param e The click event
 */
const showAttachments = (e: Event) => {
    e.stopPropagation();

    emit("open-attachments");
};

/**
 * Shows the related media
 * @param e The click event
 */
const showRelatedMedia = (e: Event) => {
    e.stopPropagation();

    emit("open-related-media");
};

/**
 * Shows the albums modal
 */
const manageAlbums = () => {
    emit("open-albums");
};

/**
 * Shows the player configuration
 * @param e The click event
 */
const showConfig = (e: Event) => {
    e.stopPropagation();

    emit("open-config");
};

/**
 * Toggles full screen
 */
const toggleFullScreen = () => {
    emit("toggle-full-screen");
};
</script>
