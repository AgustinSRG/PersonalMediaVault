<template>
    <div
        ref="container"
        class="player-top-bar"
        :class="{
            hidden: !shown,
            'with-album': inAlbum,
            'album-expand': albumExpanded,
            expanded: expanded && !albumExpanded,
            expanding: expanding,
            contracting: clickedContract,
        }"
        tabindex="-1"
        @click="clickTopBar"
        @dblclick="stopPropagationEvent"
        @mousedown="stopPropagationEvent"
        @touchstart="stopPropagationEvent"
        @contextmenu="stopPropagationEvent"
        @keydown="onKeyDown"
        @animationstart="onAnimationStart"
        @animationend="onAnimationEnd"
    >
        <div v-if="!albumExpanded" class="player-title-container">
            <div class="player-title-left">
                <button type="button" :title="$t('View Album')" class="player-btn" @click="expandAlbum">
                    <i class="fas fa-list-ol"></i>
                </button>
            </div>
            <div class="player-title">
                <div v-if="metadata">{{ metadata.title }}</div>
            </div>
            <div class="player-title-right">
                <button v-if="metadata && !expanded" type="button" :title="$t('Expand')" class="player-btn" @click="expandTitle">
                    <i class="fas fa-chevron-down"></i>
                </button>

                <button v-if="metadata && expanded" type="button" :title="$t('Close')" class="player-btn" @click="closeTitle">
                    <i class="fas fa-chevron-up"></i>
                </button>
            </div>
        </div>

        <PlayerAlbumFullScreen v-if="albumExpanded" @close="closeAlbum"></PlayerAlbumFullScreen>
        <PlayerMediaEditor v-if="expanded" @changed="onEditDone"></PlayerMediaEditor>
    </div>
</template>

<script setup lang="ts">
import type { PropType } from "vue";
import { defineAsyncComponent, nextTick, ref, useTemplateRef, watch } from "vue";
import { AuthController } from "@/control/auth";
import type { MediaData } from "@/api/models";
import { ExitPreventer } from "@/control/exit-prevent";
import { useI18n } from "@/composables/use-i18n";
import { stopPropagationEvent } from "@/utils/events";
import { useTimeout } from "@/composables/use-timeout";
import { useGlobalKeyboardHandler } from "@/composables/use-global-keyboard-handler";
import { loadCurrentMedia } from "@/control/media";

const PlayerAlbumFullScreen = defineAsyncComponent({
    loader: () => import("@/components/player/common/PlayerAlbumFullScreen.vue"),
});

const PlayerMediaEditor = defineAsyncComponent({
    loader: () => import("@/components/player/editor/PlayerMediaEditor.vue"),
});

// Ref to the container element
const container = useTemplateRef("container");

// Translation
const { $t } = useI18n();

// Props
const props = defineProps({
    /**
     * Media ID
     */
    mid: Number,

    /**
     * Media metadata
     */
    metadata: Object as PropType<MediaData>,

    /**
     * Is in album?
     */
    inAlbum: Boolean,

    /**
     * Display the top bar?
     */
    shown: Boolean,

    /**
     * Is full screen?
     */
    fullscreen: Boolean,
});

// True if player editor is expanded
const expanded = defineModel<boolean>("expanded");

// Is album expanded in full screen
const albumExpanded = defineModel<boolean>("albumExpanded");

// Emits
const emit = defineEmits<{
    /**
     * The top bar was clicked
     */
    (e: "click-player"): void;
}>();

// Changes made to the media in editor?
const dirty = ref(false);

// Expanding?
const expanding = ref(false);

// The user clicked to contract the top bar?
const clickedContract = ref(false);

// The album expanded can only be true when full screen
watch(
    () => props.fullscreen,
    () => {
        albumExpanded.value = false;
    },
);

// Timeout to load the media after dirty editor closes
const mediaLoadTimeout = useTimeout();

// Delay to load the media after
const MEDIA_LOAD_DELAY = 100;

// Watch expanded status in order to focus, or reload the media
watch(expanded, () => {
    clickedContract.value = !expanded.value;
    if (expanded.value) {
        nextTick(() => {
            const el = container.value?.querySelector(".player-media-editor") as HTMLElement;
            if (el) {
                el.focus();
            }
        });
    }
    if (dirty.value) {
        dirty.value = false;
        mediaLoadTimeout.set(() => {
            loadCurrentMedia();
        }, MEDIA_LOAD_DELAY);
    }
});

// Watch for albumExpanded in order to focus it
watch(albumExpanded, () => {
    clickedContract.value = false;
    if (albumExpanded.value) {
        nextTick(() => {
            const el = container.value?.querySelector(".player-album-container") as HTMLElement;
            if (el) {
                el.focus();
            }
        });
    }
});

/**
 * Event handler for 'click' on the top bar
 * @param e The click event
 */
const clickTopBar = (e: Event) => {
    e.stopPropagation();
    emit("click-player");
};

/**
 * Expands the top bar
 */
const expandTitle = () => {
    albumExpanded.value = false;
    expanded.value = true;
};

/**
 * Called when a change is made using the media editor
 */
const onEditDone = () => {
    dirty.value = true;
};

/**
 * Closes the top bar
 */
const closeTitle = () => {
    ExitPreventer.TryExit(() => {
        expanded.value = false;
    });
};

/**
 * Expands the album in full screen
 */
const expandAlbum = () => {
    albumExpanded.value = true;
    expanded.value = false;
};

/**
 * Closes the album in full screen
 */
const closeAlbum = () => {
    albumExpanded.value = false;
};

/**
 * Closes the top bar or the album in full screen
 */
const close = () => {
    closeTitle();
    closeAlbum();
};

/**
 * Event handler for 'keydown'
 * @param e The keyboard event
 */
const onKeyDown = (e: KeyboardEvent) => {
    if (!expanded.value && !albumExpanded.value) {
        return;
    }
    e.stopPropagation();
    if (e.key === "Escape") {
        e.preventDefault();
        close();
    }
};

/**
 * Event handler for 'animationstart'
 * @param event The animation event
 */
const onAnimationStart = (event: AnimationEvent) => {
    if (event.animationName == "player-top-bar-expand" && expanded.value) {
        expanding.value = true;
    }
};

/**
 * Event handler for 'animationend'
 * @param event The animation event
 */
const onAnimationEnd = (event: AnimationEvent) => {
    if (event.animationName == "player-top-bar-expand" && expanded.value) {
        expanding.value = false;

        const autoFocus = container.value?.querySelector(".auto-focus") as HTMLElement;
        if (autoFocus) {
            autoFocus.focus();
        }
    }
};

// Global keyboard handler
useGlobalKeyboardHandler((event: KeyboardEvent): boolean => {
    if (AuthController.Locked || !event.key || event.ctrlKey) {
        return false;
    }

    if (event.key.toUpperCase() === "E") {
        expandTitle();
        return true;
    }

    return false;
});
</script>
