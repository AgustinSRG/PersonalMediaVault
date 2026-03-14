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
            <img v-if="prefetchURL" decoding="async" :src="prefetchURL" />
        </div>
        <div ref="imageScroller" class="image-scroller" :class="{ 'cursor-hidden': !cursorShown }" @mousedown="grabScrollWithMouse">
            <img
                v-if="imageURL"
                :key="rTick"
                class="player-img-element"
                :src="imageURL"
                :style="{
                    width: imageWidth,
                    height: imageHeight,
                    top: imageTop,
                    left: imageLeft,
                }"
                decoding="async"
                @load="onImageLoaded"
                @error="onMediaError"
            />

            <ImageNotes
                :visible="notesVisible"
                :editing="notesEditMode"
                :context-open="contextMenuShown"
                :width="imageWidth"
                :height="imageHeight"
                :top="imageTop"
                :left="imageLeft"
            ></ImageNotes>
        </div>

        <PlayerLoader v-if="loading && !mediaError"></PlayerLoader>

        <PlayerEncodingPending
            v-if="(!loading && !imageURL && imagePending) || mediaError"
            :mid="mid"
            :tid="imagePendingTask"
            :res="currentResolution"
            :error="mediaError"
            :encoding-error="imageEncodeError"
            :can-auto-reload="!expandedTitle && !expandedAlbum"
        ></PlayerEncodingPending>

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
            :type="'image'"
            :show-controls="showControls"
            :next="next"
            :prev="prev"
            :page-next="pageNext"
            :page-prev="pagePrev"
            :fullscreen="fullscreen"
            :has-description="hasDescription"
            :has-attachments="hasAttachments"
            :has-related-media="hasRelatedMedia"
            @click="clickControls"
            @enter="enterControls"
            @leave="leaveControls"
            @enter-tooltip="enterTooltip"
            @leave-tooltip="leaveTooltip"
            @go-prev="goPrev"
            @go-next="goNext"
            @open-description="openDescription"
            @open-attachments="showAttachments"
            @open-related-media="showRelatedMedia"
            @open-albums="manageAlbums"
            @open-config="showConfig"
            @toggle-full-screen="toggleFullScreen"
        >
            <ScaleControl
                ref="scaleControl"
                v-model:fit="fit"
                v-model:scale="scale"
                v-model:expanded="scaleShown"
                :min="min"
                :width="min ? SCALE_CONTROL_WIDTH_MIN : SCALE_CONTROL_WIDTH"
                @update:scale="onUserScaleUpdated"
                @update:fit="onUserFitUpdated"
                @enter="enterTooltip('scale')"
                @leave="leaveTooltip('scale')"
            ></ScaleControl>
        </PlayerControls>

        <PlayerTooltip
            v-if="helpTooltip"
            :help-tooltip="helpTooltip"
            :hide-right-tooltip="displayConfig || displayAttachments || displayRelatedMedia"
            :next="next"
            :prev="prev"
            :page-next="pageNext"
            :page-prev="pagePrev"
            :fit="fit"
            :scale="scale"
            :scale-range-percent="SCALE_RANGE_PERCENT"
        ></PlayerTooltip>

        <PlayerConfig
            v-if="displayConfig"
            v-model:shown="displayConfig"
            v-model:resolution="currentResolution"
            v-model:background="background"
            :r-tick="internalTick"
            :metadata="metadata"
            @update:resolution="onResolutionUpdated"
            @update:background="onBackgroundChanged"
            @update-auto-next="setupAutoNextTimer"
            @update-notes-visible="imageNotesVisibleUpdated"
            @enter="enterControls"
            @leave="leaveControls"
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
            :shown="showControls"
            :fullscreen="fullscreen"
            :in-album="inAlbum"
            @click-player="clickControls"
        ></PlayerTopBar>

        <PlayerContextMenu
            v-if="contextMenuShown"
            ref="contextMenu"
            v-model:shown="contextMenuShown"
            v-model:fit="fit"
            v-model:controls="showControls"
            v-model:notes-edit="notesEditMode"
            type="image"
            :x="contextMenuX"
            :y="contextMenuY"
            :url="imageURL"
            :title="title"
            :has-description="hasDescription"
            @update:fit="onUserFitUpdated"
            @stats="openStats"
            @open-tags="openTags"
            @open-desc="openDescription"
        ></PlayerContextMenu>
    </div>
</template>

<script setup lang="ts">
import {
    getImageBackgroundStyle,
    getImageFit,
    getImageNotesVisible,
    getImageScale,
    getUserSelectedResolutionImage,
    setImageBackgroundStyle,
    setImageFit,
    setImageScale,
    setUserSelectedResolutionImage,
} from "@/local-storage/player-preferences";
import type { PropType } from "vue";
import { defineAsyncComponent, nextTick, onBeforeUnmount, onMounted, ref, useTemplateRef, watch } from "vue";
import ScaleControl from "./common/ScaleControl.vue";
import PlayerTopBar from "./common/PlayerTopBar.vue";
import ImageNotes from "./common/ImageNotes.vue";
import { isTouchDevice } from "@/utils/touch";
import { getAssetURL } from "@/utils/api";
import { checkAuthenticationStatusSilent, isVaultLocked, refreshAuthenticationStatus } from "@/global-state/auth";
import { AppStatus } from "@/global-state/app-status";
import type { MediaData, MediaListItem } from "@/api/models";
import { MEDIA_TYPE_IMAGE } from "@/constants";
import PlayerTooltip from "./common/PlayerTooltip.vue";
import { EVENT_NAME_NEXT_PRE_FETCH } from "@/global-state/app-events";
import PlayerControls from "./common/PlayerControls.vue";
import PlayerLoader from "./common/PlayerLoader.vue";
import { PLAYER_KEYBOARD_HANDLER_PRIORITY, usePlayerCommon } from "@/composables/use-player-common";
import { useInterval } from "@/composables/use-interval";
import { usePlayerMediaSession } from "@/composables/use-player-media-session";
import { useUserPermissions } from "@/composables/use-user-permissions";
import { onApplicationEvent } from "@/composables/on-app-event";
import { useGlobalKeyboardHandler } from "@/composables/use-global-keyboard-handler";
import type { PositionEvent } from "@/utils/position-event";
import { positionEventFromMouseEvent } from "@/utils/position-event";
import { onDocumentEvent } from "@/composables/on-document-event";
import { usePlayerCommonControls } from "@/composables/use-player-common-controls";
import { usePlayerAutoNext } from "@/composables/use-player-auto-next";
import { getAlbumNextPrefetchData } from "@/global-state/album";

const PlayerContextMenu = defineAsyncComponent({
    loader: () => import("@/components/player/common/PlayerContextMenu.vue"),
});

const PlayerConfig = defineAsyncComponent({
    loader: () => import("@/components/player/config/PlayerConfig.vue"),
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

const PlayerRelatedMediaList = defineAsyncComponent({
    loader: () => import("@/components/player/common/PlayerRelatedMediaList.vue"),
});

const PlayerAttachmentsList = defineAsyncComponent({
    loader: () => import("@/components/player/common/PlayerAttachmentsList.vue"),
});

// True if it is a touch device
const IS_TOUCH_DEVICE = isTouchDevice();

// Scale control width (px)
const SCALE_CONTROL_WIDTH = 100;

// Scale control width when miniature mode (px)
const SCALE_CONTROL_WIDTH_MIN = 70;

// Scale range
const SCALE_RANGE = 2;

// Scale range (%)
const SCALE_RANGE_PERCENT = SCALE_RANGE * 100;

// Scale step
const SCALE_STEP = 0.1 / SCALE_RANGE;

// Scale step (small increments)
const SCALE_STEP_MIN = 0.01 / SCALE_RANGE;

// Ref to the image scroller element
const imageScroller = useTemplateRef("imageScroller");

// Ref to the context menu
const contextMenu = useTemplateRef("contextMenu");

// User permissions
const { canWrite } = useUserPermissions();

// Full screen model
const fullscreen = defineModel<boolean>("fullscreen");

// Show controls?
const showControls = defineModel<boolean>("showControls");

// Display tag list widget?
const displayTagList = defineModel<boolean>("displayTagList");

// Display description widget
const displayDescription = defineModel<boolean>("displayDescription");

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

    /**
     * The user wants to open the albums list modal
     */
    (e: "albums-open"): void;

    /**
     * The user wants to upen the size stats modal
     */
    (e: "stats-open"): void;
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

// Auto-next
const { clearAutoNextTimer, setupAutoNextTimer, toggleAutoNext } = usePlayerAutoNext(
    props,
    ref(true),
    {
        displayConfig,
        expandedTitle,
        displayAttachments,
        displayRelatedMedia,
    },
    goNext,
);

// Loading status
const loading = ref(true);

// Title
const title = ref("");

// Image URL
const imageURL = ref("");

// Image pending of encoding
const imagePending = ref(false);

// If the image is pending of encoding, ID of the task
const imagePendingTask = ref(0);

// Error message of the image encoding failed
const imageEncodeError = ref("");

// Index of the current selected resolution
const currentResolution = ref(-1);

// Image width
const width = ref(0);

// Image height
const height = ref(0);

// Image coordinates (scroll & scale)
const imageTop = ref("0");
const imageLeft = ref("0");
const imageWidth = ref("auto");
const imageHeight = ref("auto");

// Timestamp of the last interaction with the player controls
const lastControlsInteraction = ref(Date.now());

// True if the mouse is in the player controls
const mouseInControls = ref(false);

// Image scale
const scale = ref(getImageScale());

// Fit image
const fit = ref(getImageFit());

// True to show the scale control
const scaleShown = ref(IS_TOUCH_DEVICE);

// Image background
const background = ref(getImageBackgroundStyle());

// Internal tick for player config
const internalTick = ref(0);

// Grabbed scroll?
const scrollGrabbed = ref(false);

// Grabbed image scroll coordinates
const scrollGrabX = ref(0);
const scrollGrabY = ref(0);
const scrollGrabTop = ref(0);
const scrollGrabLeft = ref(0);

// Show cursor?
const cursorShown = ref(false);

// URL of the next image to pre-fetch it
const prefetchURL = ref("");

// Image notes visible?
const notesVisible = ref(getImageNotesVisible());

// Editing image notes?
const notesEditMode = ref(false);

// Does the media could not load due to an error?
const mediaError = ref(false);

/**
 * Initializes the image
 */
const initializeImage = () => {
    if (!props.metadata) {
        return;
    }

    loading.value = true;
    currentResolution.value = getUserSelectedResolutionImage(props.metadata);
    setImageURL();
};

// Reload image when reload tick changes
watch(
    () => props.rTick,
    () => {
        internalTick.value++;
        initializeImage();
    },
);

// Initialize on mounted
onMounted(initializeImage);

// Set the loading status of the image when the URL changes
watch(imageURL, () => {
    if (imageURL.value) {
        loading.value = true;
    }
});

// Reset image URL if unmounted to prevent load / error events
onBeforeUnmount(() => {
    imageURL.value = "";
});

/**
 * Sets the image URL and other parameters
 * from the image metadata
 */
const setImageURL = () => {
    clearAutoNextTimer();

    mediaError.value = false;

    if (!props.metadata) {
        imageURL.value = "";
        title.value = "";
        loading.value = false;
        return;
    }

    title.value = props.metadata.title;

    if (currentResolution.value < 0) {
        if (props.metadata.encoded) {
            imageURL.value = getAssetURL(props.metadata.url);
            imagePending.value = false;
            imagePendingTask.value = 0;
            imageEncodeError.value = "";
            width.value = props.metadata.width;
            height.value = props.metadata.height;
            setupAutoNextTimer();
        } else {
            imageURL.value = "";
            imagePending.value = true;
            imagePendingTask.value = props.metadata.task;
            imageEncodeError.value = props.metadata.error || "";
            loading.value = false;
        }
    } else {
        if (props.metadata.resolutions && props.metadata.resolutions.length > currentResolution.value) {
            const res = props.metadata.resolutions[currentResolution.value];
            if (res.ready) {
                imageURL.value = getAssetURL(res.url);
                imagePending.value = false;
                imagePendingTask.value = 0;
                imageEncodeError.value = "";
                width.value = props.metadata.width;
                height.value = props.metadata.height;
                setupAutoNextTimer();
            } else {
                imageURL.value = "";
                imagePending.value = true;
                imagePendingTask.value = res.task;
                imageEncodeError.value = res.error || "";
                loading.value = false;
            }
        } else {
            imageURL.value = "";
            imagePending.value = true;
            imagePendingTask.value = 0;
            imageEncodeError.value = "";
            loading.value = false;
        }
    }

    computeImageDimensions();
    incrementImageScroll("home");
    tryHorizontalScroll("home");
};

/**
 * Computes the image dimensions
 */
const computeImageDimensions = () => {
    if (!imageURL.value) {
        return;
    }

    const scroller = imageScroller.value;

    if (!scroller) {
        return;
    }

    const scrollerDimensions = scroller.getBoundingClientRect();

    const fitDimensions = {
        width: width.value,
        height: height.value,
        fitWidth: true,
    };

    if (scrollerDimensions.width > scrollerDimensions.height) {
        fitDimensions.fitWidth = true;
        fitDimensions.height = scrollerDimensions.height;
        fitDimensions.width = (scrollerDimensions.height * width.value) / height.value;

        if (fitDimensions.width > scrollerDimensions.width) {
            fitDimensions.fitWidth = false;
            fitDimensions.width = scrollerDimensions.width;
            fitDimensions.height = (scrollerDimensions.width * height.value) / width.value;
        }
    } else {
        fitDimensions.fitWidth = false;
        fitDimensions.width = scrollerDimensions.width;
        fitDimensions.height = (scrollerDimensions.width * height.value) / width.value;

        if (fitDimensions.height > scrollerDimensions.height) {
            fitDimensions.fitWidth = true;
            fitDimensions.height = scrollerDimensions.height;
            fitDimensions.width = (scrollerDimensions.height * width.value) / height.value;
        }
    }

    if (fit.value) {
        const top = Math.max(0, (scrollerDimensions.height - fitDimensions.height) / 2);
        const left = Math.max(0, (scrollerDimensions.width - fitDimensions.width) / 2);

        imageTop.value = Math.floor(top) + "px";
        imageLeft.value = Math.floor(left) + "px";
        imageWidth.value = Math.floor(fitDimensions.width) + "px";
        imageHeight.value = Math.floor(fitDimensions.height) + "px";
    } else {
        let w: number;
        let h: number;

        if (fitDimensions.fitWidth) {
            w = scrollerDimensions.width * (0.5 + scale.value * SCALE_RANGE);
            h = (w * height.value) / width.value;
        } else {
            h = scrollerDimensions.height * (0.5 + scale.value * SCALE_RANGE);
            w = (h * width.value) / height.value;
        }

        const top = Math.max(0, (scrollerDimensions.height - h) / 2);
        const left = Math.max(0, (scrollerDimensions.width - w) / 2);

        imageTop.value = Math.floor(top) + "px";
        imageLeft.value = Math.floor(left) + "px";
        imageWidth.value = Math.floor(w) + "px";
        imageHeight.value = Math.floor(h) + "px";
    }
};

/**
 * Called when image successfully loaded
 */
const onImageLoaded = () => {
    loading.value = false;
};

/**
 * Called when the media cannot be loaded
 */
const onMediaError = () => {
    if (!refreshAuthenticationStatus()) {
        mediaError.value = true;

        loading.value = false;

        checkAuthenticationStatusSilent();
    }
};

/**
 * Called when the selected resolution updates
 */
const onResolutionUpdated = () => {
    setUserSelectedResolutionImage(props.metadata, currentResolution.value);
    setImageURL();
};

/**
 * Increments image vertical scroll
 * @param a An increment, 'home' to set it to the beginning, 'end' to set it to the end
 * @returns True if applied, false if it cannot be applied
 */
const incrementImageScroll = (a: number | "home" | "end"): boolean => {
    if (fit.value) {
        return false;
    }

    const el = imageScroller.value;

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
 * Tries to increment the horizontal scroll of the image
 * @param a An increment, 'home' to set it to the beginning, 'end' to set it to the end
 * @returns True if applied, false if it cannot be applied
 */
const tryHorizontalScroll = (a: number | "home" | "end"): boolean => {
    if (fit.value) {
        return false;
    }

    const el = imageScroller.value;

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

/**
 * Called when the user clicks the player
 */
const clickPlayer = () => {
    if (displayConfig.value) {
        displayConfig.value = false;
    }

    if (displayAttachments.value) {
        displayAttachments.value = false;
    }

    if (displayRelatedMedia.value) {
        displayRelatedMedia.value = false;
    }

    interactWithControls();
};

/**
 * Called when the user interacted with the player controls
 */
const interactWithControls = () => {
    lastControlsInteraction.value = Date.now();
    cursorShown.value = true;
};

/**
 * Called when the user enters the controls
 */
const enterControls = () => {
    mouseInControls.value = true;
};

/**
 * Called when the user leaves the controls
 */
const leaveControls = () => {
    mouseInControls.value = false;
    clearTooltip();
    scaleShown.value = IS_TOUCH_DEVICE;
};

/**
 * Called when image scale is updated by the user
 */
const onUserScaleUpdated = () => {
    setImageScale(scale.value);
    computeImageDimensions();
    nextTick(centerScroll);
};

/**
 * Sets the scale
 * @param s The scale
 */
const changeScale = (s: number) => {
    scale.value = s;
    onUserScaleUpdated();
};

/**
 * Called when the image fit option is changed by the user
 */
const onUserFitUpdated = () => {
    setImageFit(fit.value);
    computeImageDimensions();
};

/**
 * Toggles the fit image option
 */
const toggleFit = () => {
    fit.value = !fit.value;
    onUserFitUpdated();
};

/**
 * Called when the background option is changed
 */
const onBackgroundChanged = () => {
    setImageBackgroundStyle(background.value);
};

/**
 * Called when the image notes visibility changes
 * @param v The image notes visibility
 */
const imageNotesVisibleUpdated = (v: boolean) => {
    notesVisible.value = v;
};

/**
 * Centers the scroll
 */
const centerScroll = () => {
    const scroller = imageScroller.value;

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
    if (e.target) {
        const target = e.target;
        if (target.classList && target.classList.contains("image-scroller")) {
            return;
        }
    }

    if (displayConfig.value || contextMenuShown.value || displayAttachments.value || displayRelatedMedia.value) {
        displayConfig.value = false;
        contextMenu.value?.hide();
        displayAttachments.value = false;
        displayRelatedMedia.value = false;
        e.e.stopPropagation();
        return;
    }

    const scroller = imageScroller.value;

    if (!scroller) {
        return;
    }

    scrollGrabTop.value = scroller.scrollTop;
    scrollGrabLeft.value = scroller.scrollLeft;

    scrollGrabbed.value = true;

    scrollGrabX.value = e.x;
    scrollGrabY.value = e.y;
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
 * Moves scroll using the mouse
 * @param x The X coordinate
 * @param y The Y coordinate
 */
const moveScrollByMouse = (x: number, y: number) => {
    const scroller = imageScroller.value;

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

    moveScrollByMouse(e.pageX, e.pageY);
});

/**
 * Event handler for "mousemove" on the player
 */
const playerMouseMove = () => {
    interactWithControls();
};

/**
 * Event handler for "mouseleave" on the player
 */
const mouseLeavePlayer = () => {
    clearTooltip();
};

/**
 * Mouse wheel event handler for the player
 * @param e The mouse wheel event
 */
const onMouseWheel = (e: WheelEvent) => {
    if (e.ctrlKey) {
        e.preventDefault();
        e.stopPropagation();
        if (e.deltaY > 0) {
            changeScale(Math.max(0, scale.value - SCALE_STEP));
            scaleShown.value = true;
            helpTooltip.value = "scale";
            fit.value = false;
            onUserFitUpdated();
        } else {
            changeScale(Math.min(1, scale.value + SCALE_STEP));
            scaleShown.value = true;
            helpTooltip.value = "scale";
            fit.value = false;
            onUserFitUpdated();
        }
    }
};

// Delay for ticks (30 ticks/sec)
const TICK_DELAY = Math.floor(1000 / 30);

// Ticks
useInterval().set(() => {
    computeImageDimensions();

    const now = Date.now();

    if (!mouseInControls.value && helpTooltip.value && now - lastControlsInteraction.value > 2000) {
        clearTooltip();
    }

    if (!mouseInControls.value && scaleShown.value && now - lastControlsInteraction.value > 2000 && scaleShown.value !== IS_TOUCH_DEVICE) {
        scaleShown.value = IS_TOUCH_DEVICE;
    }

    if (!mouseInControls.value && cursorShown.value && now - lastControlsInteraction.value > 2000) {
        cursorShown.value = false;
    }

    if (helpTooltip.value && !showControls.value) {
        clearTooltip();
    }
}, TICK_DELAY);

// MediaSession actions handler
usePlayerMediaSession(["nexttrack", "previoustrack"], (event: MediaSessionActionDetails) => {
    if (!event || !event.action) {
        return;
    }
    switch (event.action) {
        case "nexttrack":
            goNext();
            break;
        case "previoustrack":
            goPrev();
            break;
    }
});

useGlobalKeyboardHandler((event: KeyboardEvent) => {
    if (isVaultLocked() || !AppStatus.IsPlayerVisible() || !event.key || (event.ctrlKey && event.key !== "+" && event.key !== "-")) {
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
        case "ArrowUp":
            incrementImageScroll(-40);
            break;
        case "ArrowDown":
            incrementImageScroll(40);
            break;
        case "ArrowLeft":
            if (shifting || event.altKey || !tryHorizontalScroll(-40)) {
                if (props.prev || props.pagePrev) {
                    goPrev();
                } else {
                    caught = false;
                }
            }
            break;
        case "ArrowRight":
            if (shifting || event.altKey || !tryHorizontalScroll(40)) {
                if (props.next || props.pageNext) {
                    goNext();
                } else {
                    caught = false;
                }
            }
            break;
        case "Home":
            if (event.altKey) {
                caught = false;
            } else if (shifting) {
                if (!tryHorizontalScroll("home")) {
                    caught = false;
                }
            } else {
                if (!incrementImageScroll("home")) {
                    caught = false;
                }
            }
            break;
        case "End":
            if (event.altKey) {
                caught = false;
            } else if (shifting) {
                if (!tryHorizontalScroll("end")) {
                    caught = false;
                }
            } else {
                if (!incrementImageScroll("end")) {
                    caught = false;
                }
            }
            break;
        case " ":
        case "K":
        case "k":
        case "Enter":
            toggleFit();
            scaleShown.value = true;
            helpTooltip.value = "scale";
            break;
        case "+":
            changeScale(Math.min(1, scale.value + (shifting ? SCALE_STEP_MIN : SCALE_STEP)));
            scaleShown.value = true;
            helpTooltip.value = "scale";
            fit.value = false;
            onUserFitUpdated();
            break;
        case "-":
            changeScale(Math.max(0, scale.value - (shifting ? SCALE_STEP_MIN : SCALE_STEP)));
            scaleShown.value = true;
            helpTooltip.value = "scale";
            fit.value = false;
            onUserFitUpdated();
            break;
        case "C":
        case "c":
            showControls.value = !showControls.value;
            break;
        case "n":
        case "N":
            if (canWrite.value) {
                notesEditMode.value = !notesEditMode.value;
            } else {
                notesEditMode.value = false;
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

/**
 * Called to pre-fetch the next image in an album, for faster load times
 */
const onAlbumPrefetch = () => {
    const nextMediaData = getAlbumNextPrefetchData();

    if (nextMediaData && nextMediaData.type === MEDIA_TYPE_IMAGE) {
        if (currentResolution.value < 0) {
            if (nextMediaData.encoded) {
                prefetchURL.value = getAssetURL(nextMediaData.url);
            } else {
                prefetchURL.value = "";
            }
        } else {
            if (nextMediaData.resolutions[currentResolution.value] && nextMediaData.resolutions[currentResolution.value].ready) {
                prefetchURL.value = getAssetURL(nextMediaData.resolutions[currentResolution.value].url);
            } else {
                prefetchURL.value = "";
            }
        }
    } else {
        prefetchURL.value = "";
    }
};

onApplicationEvent(EVENT_NAME_NEXT_PRE_FETCH, onAlbumPrefetch);
onMounted(onAlbumPrefetch);
</script>
