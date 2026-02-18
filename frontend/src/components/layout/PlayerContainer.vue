<template>
    <div ref="container" class="player-container" :class="{ 'using-touch-device': touchDevice }" tabindex="-1">
        <EmptyPlayer
            v-if="!mediaData || mediaData.type === 0"
            v-model:fullscreen="fullScreen"
            :mid="mid"
            :status="status"
            :r-tick="tick"
            :prev="prev"
            :page-prev="hasPagePrev"
            :page-next="hasPageNext"
            :next="next"
            :in-album="isInAlbum"
            :album-loading="albumLoading"
            :can-write="canWrite"
            :min="minPlayer"
            @go-next="goNext"
            @go-prev="goPrev"
            @delete="openDelete"
        ></EmptyPlayer>
        <ImagePlayer
            v-if="mediaData && mediaData.type === 1"
            v-model:fullscreen="fullScreen"
            v-model:show-controls="showControls"
            v-model:display-tag-list="displayTagList"
            v-model:display-description="displayDescription"
            :mid="mid"
            :metadata="mediaData"
            :r-tick="tick"
            :prev="prev"
            :next="next"
            :page-prev="hasPagePrev"
            :page-next="hasPageNext"
            :in-album="isInAlbum"
            :can-write="canWrite"
            :min="minPlayer"
            @go-next="goNext"
            @go-prev="goPrev"
            @albums-open="openAlbums"
            @stats-open="openStats"
            @delete="openDelete"
        ></ImagePlayer>
        <VideoPlayer
            v-if="mediaData && mediaData.type === 2"
            v-model:fullscreen="fullScreen"
            v-model:user-controls="showControls"
            v-model:display-tag-list="displayTagList"
            v-model:display-description="displayDescription"
            :mid="mid"
            :metadata="mediaData"
            :r-tick="tick"
            :prev="prev"
            :next="next"
            :page-prev="hasPagePrev"
            :page-next="hasPageNext"
            :in-album="isInAlbum"
            :can-write="canWrite"
            :min="minPlayer"
            :loop-forced="loopForced"
            :loop-forced-value="loopForcedValue"
            :auto-play="!(displayAlbumList || displaySizeStats || displayUpload)"
            @go-next="goNext"
            @go-prev="goPrev"
            @albums-open="openAlbums"
            @stats-open="openStats"
            @force-loop="onForceLoop"
            @delete="openDelete"
        ></VideoPlayer>
        <AudioPlayer
            v-if="mediaData && mediaData.type === 3"
            v-model:fullscreen="fullScreen"
            v-model:display-tag-list="displayTagList"
            v-model:display-description="displayDescription"
            :mid="mid"
            :metadata="mediaData"
            :r-tick="tick"
            :prev="prev"
            :next="next"
            :page-prev="hasPagePrev"
            :page-next="hasPageNext"
            :in-album="isInAlbum"
            :can-write="canWrite"
            :min="minPlayer"
            :loop-forced="loopForced"
            :loop-forced-value="loopForcedValue"
            :auto-play="!(displayAlbumList || displaySizeStats || displayUpload)"
            @go-next="goNext"
            @go-prev="goPrev"
            @albums-open="openAlbums"
            @stats-open="openStats"
            @force-loop="onForceLoop"
            @delete="openDelete"
        ></AudioPlayer>

        <AlbumListModal v-if="displayAlbumList" v-model:display="displayAlbumList"></AlbumListModal>

        <SizeStatsModal v-if="displaySizeStats" v-model:display="displaySizeStats" :mid="mid"></SizeStatsModal>

        <MediaDeleteModal v-if="displayDelete" v-model:display="displayDelete"></MediaDeleteModal>
    </div>
</template>

<script setup lang="ts">
import {
    emitAppEvent,
    EVENT_NAME_CURRENT_ALBUM_LOADING,
    EVENT_NAME_CURRENT_ALBUM_MEDIA_POSITION_UPDATED,
    EVENT_NAME_GO_NEXT,
    EVENT_NAME_GO_PREV,
    EVENT_NAME_MEDIA_LOADING,
    EVENT_NAME_MEDIA_UPDATE,
    EVENT_NAME_PAGE_MEDIA_NAV_UPDATE,
    EVENT_NAME_PAGE_NAV_NEXT,
    EVENT_NAME_PAGE_NAV_PREV,
} from "@/control/app-events";
import { MediaController } from "@/control/media";
import { computed, defineAsyncComponent, onBeforeUnmount, onMounted, ref, useTemplateRef } from "vue";
import { AlbumsController } from "@/control/albums";
import { AppStatus } from "@/control/app-status";
import { closeFullscreen } from "@/utils/full-screen";
import { PagesController } from "@/control/pages";
import { isTouchDevice } from "@/utils/touch";
import LoadingOverlay from "./LoadingOverlay.vue";
import { useFocusTrap } from "@/composables/use-focus-trap";
import { useUserPermissions } from "@/composables/use-user-permissions";
import { onApplicationEvent } from "@/composables/on-app-event";

const EmptyPlayer = defineAsyncComponent({
    loader: () => import("@/components/player/EmptyPlayer.vue"),
    loadingComponent: LoadingOverlay,
    delay: 1000,
});

const AudioPlayer = defineAsyncComponent({
    loader: () => import("@/components/player/AudioPlayer.vue"),
    loadingComponent: LoadingOverlay,
    delay: 1000,
});

const VideoPlayer = defineAsyncComponent({
    loader: () => import("@/components/player/VideoPlayer.vue"),
    loadingComponent: LoadingOverlay,
    delay: 1000,
});

const ImagePlayer = defineAsyncComponent({
    loader: () => import("@/components/player/ImagePlayer.vue"),
    loadingComponent: LoadingOverlay,
    delay: 1000,
});

const AlbumListModal = defineAsyncComponent({
    loader: () => import("@/components/modals/AlbumListModal.vue"),
    loadingComponent: LoadingOverlay,
    delay: 1000,
});

const SizeStatsModal = defineAsyncComponent({
    loader: () => import("@/components/modals/SizeStatsModal.vue"),
    loadingComponent: LoadingOverlay,
    delay: 1000,
});

const MediaDeleteModal = defineAsyncComponent({
    loader: () => import("@/components/modals/MediaDeleteModal.vue"),
    loadingComponent: LoadingOverlay,
    delay: 1000,
});

// Props
defineProps({
    /**
     * Is the upload modal being displayed?
     */
    displayUpload: Boolean,
});

// User permissions
const { canWrite } = useUserPermissions();

// Ref to the container element
const container = useTemplateRef("container");

// Show player controls?
const showControls = ref(true);

// Display album list modal?
const displayAlbumList = ref(false);

// Display the size stats modal?
const displaySizeStats = ref(false);

// Display the tag list widget?
const displayTagList = ref(false);

// Display the description widget?
const displayDescription = ref(false);

// Display the deletion modal?
const displayDelete = ref(false);

// Is loop forced by the user
const loopForced = ref(false);

// The value for loop the user forced
const loopForcedValue = ref(false);

// Tick to reload the
const tick = ref(0);

// Media loading status
const loading = ref(MediaController.Loading);

onApplicationEvent(EVENT_NAME_MEDIA_LOADING, (l) => {
    loading.value = l;
});

// Current media ID
const mid = ref(MediaController.MediaId);

// Current media data
const mediaData = ref(MediaController.MediaData);

onApplicationEvent(EVENT_NAME_MEDIA_UPDATE, () => {
    displayDelete.value = false;

    mid.value = MediaController.MediaId;
    if (MediaController.MediaData !== mediaData.value) {
        mediaData.value = MediaController.MediaData;

        tick.value++;

        if (mid.value >= 0) {
            container.value.focus();
        }
    }
});

// Player statuses
type PlayerStatus = "loading" | "200" | "none" | "404";

// Current player status
const status = computed<PlayerStatus>(() => {
    if (loading.value) {
        return "loading";
    } else if (mediaData.value) {
        return "200";
    } else if (mid.value < 0) {
        return "none";
    } else {
        return "404";
    }
});

// Previous element in the album
const prev = ref(AlbumsController.CurrentPrev);

// Next element in the album
const next = ref(AlbumsController.CurrentNext);

// Is player coexisting with the album layout?
const isInAlbum = ref(AppStatus.CurrentAlbum >= 0);

onApplicationEvent(EVENT_NAME_CURRENT_ALBUM_MEDIA_POSITION_UPDATED, () => {
    prev.value = AlbumsController.CurrentPrev;
    next.value = AlbumsController.CurrentNext;
    isInAlbum.value = AppStatus.CurrentAlbum >= 0;
});

// Is the current album being loaded?
const albumLoading = ref(AlbumsController.CurrentAlbumLoading);

onApplicationEvent(EVENT_NAME_CURRENT_ALBUM_LOADING, (l) => {
    albumLoading.value = l;
});

// Does the media has a previous element within the current page?
const hasPagePrev = ref(PagesController.HasPagePrev);

// Does the media has a next element within the current page?
const hasPageNext = ref(PagesController.HasPageNext);

onApplicationEvent(EVENT_NAME_PAGE_MEDIA_NAV_UPDATE, (pageHasPrev: boolean, pageHasNext: boolean) => {
    hasPagePrev.value = pageHasPrev;
    hasPageNext.value = pageHasNext;
});

// Is a touch device?
const touchDevice = ref(false);

// Is a miniature player?
const minPlayer = ref(false);

/**
 * Checks the player size
 */
const checkPlayerSize = () => {
    if (!container.value) {
        return;
    }

    const rect = container.value.getBoundingClientRect();
    const width = rect.width;
    const height = rect.height;

    minPlayer.value = width < 480 || height < 360;

    touchDevice.value = isTouchDevice();
};

// Resize observer to update the player size
const resizeObserver = new ResizeObserver(checkPlayerSize);

onMounted(() => {
    if (container.value) {
        resizeObserver.observe(container.value);
    }
});

onBeforeUnmount(() => {
    resizeObserver.disconnect();
});

// Is player in full screen
const fullScreen = ref(false);

/**
 * Call when focus is lost
 */
const focusLost = () => {
    closeFullscreen();

    fullScreen.value = false;
};

// Focus trap for full screen
useFocusTrap(container, fullScreen, focusLost, null, true);

// Focus the player container on mounted
onMounted(() => {
    container.value?.focus();
});

/**
 * Navigates to the previous media element
 */
const goPrev = () => {
    if (prev.value) {
        AppStatus.ClickOnMedia(prev.value.id, false);
    } else if (hasPagePrev.value) {
        emitAppEvent(EVENT_NAME_PAGE_NAV_PREV);
    }
};

onApplicationEvent(EVENT_NAME_GO_PREV, goPrev);

/**
 * Navigates to the next media element
 */
const goNext = () => {
    if (next.value) {
        AppStatus.ClickOnMedia(next.value.id, false);
    } else if (hasPageNext.value) {
        emitAppEvent(EVENT_NAME_PAGE_NAV_NEXT);
    }
};

onApplicationEvent(EVENT_NAME_GO_NEXT, goNext);

/**
 * Opens the album list modal
 */
const openAlbums = () => {
    displayAlbumList.value = true;
};

/**
 * Opens the size stats modal
 */
const openStats = () => {
    displaySizeStats.value = true;
};

/**
 * Opens the deletion modal
 */
const openDelete = () => {
    displayDelete.value = true;
};

/**
 * Forces the loop
 * @param v The value of the loop setting chosen by the user
 */
const onForceLoop = (v: boolean) => {
    loopForced.value = true;
    loopForcedValue.value = v;
};
</script>
