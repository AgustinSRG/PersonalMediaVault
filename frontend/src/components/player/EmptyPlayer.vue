<template>
    <div
        class="empty-player"
        :class="{
            'player-min': min,
            'full-screen': fullscreen,
        }"
        @dblclick="toggleFullScreen"
        @mouseleave="mouseLeavePlayer"
    >
        <div v-if="status === 'loading' || (status === 'none' && albumLoading)" class="player-loader">
            <div class="player-lds-ring">
                <div></div>
                <div></div>
                <div></div>
                <div></div>
            </div>
        </div>

        <div v-else-if="status === '404'" class="player-error-container">
            <div class="player-info-icon"><i class="fas fa-ban"></i></div>
            <div class="player-error">{{ $t("Media asset does not exist or was removed from the vault") }}</div>
        </div>

        <div v-else-if="status === 'none' && !albumLoading" class="player-error-container">
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

        <PlayerTooltip
            v-if="helpTooltip"
            :help-tooltip="helpTooltip"
            :hide-right-tooltip="false"
            :next="next"
            :prev="prev"
            :page-next="pageNext"
            :page-prev="pagePrev"
            :has-description="false"
        ></PlayerTooltip>

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

<script setup lang="ts">
import type { PropType } from "vue";
import { AppStatus } from "@/control/app-status";
import { AuthController } from "@/control/auth";
import type { MediaListItem } from "@/api/models";
import PlayerTooltip from "./common/PlayerTooltip.vue";
import PlayerTopBar from "./common/PlayerTopBar.vue";
import { useI18n } from "@/composables/use-i18n";
import type { PlayerLoadStatus } from "@/utils/player";
import { PLAYER_KEYBOARD_HANDLER_PRIORITY, usePlayerCommon } from "@/composables/use-player-common";
import { useGlobalKeyboardHandler } from "@/composables/use-global-keyboard-handler";
import { stopPropagationEvent } from "@/utils/events";

// Translation
const { $t } = useI18n();

// Full screen model
const fullscreen = defineModel<boolean>("fullscreen");

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
     * Load status
     */
    status: {
        type: String as PropType<PlayerLoadStatus>,
        required: true,
    },

    /**
     * Loading album?
     */
    albumLoading: Boolean,

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
     * Has next element i n page?
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
}>();

// Player common features
const { expandedTitle, expandedAlbum, helpTooltip, enterTooltip, leaveTooltip, clearTooltip, goNext, goPrev, toggleFullScreen } =
    usePlayerCommon(props, emit, fullscreen);

/**
 * Event handler for 'mouseleave' in the player
 */
const mouseLeavePlayer = () => {
    clearTooltip();
};

/**
 * Event handler for 'mouseleave' in the controls
 */
const leaveControls = () => {
    clearTooltip();
};

// Global keyboard handler
useGlobalKeyboardHandler((event: KeyboardEvent): boolean => {
    if (AuthController.Locked || !AppStatus.IsPlayerVisible() || !event.key || event.ctrlKey) {
        return false;
    }

    let caught = true;

    switch (event.key) {
        case "ArrowLeft":
            if (props.prev || props.pagePrev) {
                goPrev();
            } else {
                caught = false;
            }
            break;

        case "ArrowRight":
            if (props.next || props.pageNext) {
                goNext();
            } else {
                caught = false;
            }
            break;
        default:
            caught = false;
    }

    return caught;
}, PLAYER_KEYBOARD_HANDLER_PRIORITY);
</script>
