<template>
    <div class="player-album-container" tabindex="-1">
        <div v-if="!loading && loadedAlbum" class="album-header">
            <div class="album-header-title">
                <div class="album-title"><i class="fas fa-list-ol"></i> {{ albumName }}</div>
                <button type="button" :title="$t('Close')" class="album-header-btn album-close-btn" @click="close">
                    <i class="fas fa-times"></i>
                </button>
            </div>

            <div class="album-header-controls">
                <div class="album-buttons">
                    <button type="button" :title="$t('Loop')" class="album-header-btn" :class="{ toggled: loop }" @click="toggleLoop">
                        <i class="fas fa-repeat"></i>
                    </button>

                    <button type="button" :title="$t('Random')" class="album-header-btn" :class="{ toggled: random }" @click="toggleRandom">
                        <i class="fas fa-shuffle"></i>
                    </button>
                </div>
                <div class="album-post-text">{{ renderedCurrentPos }} / {{ albumListLength }}</div>
            </div>
        </div>

        <div v-show="!loading && loadedAlbum" ref="albumBody" class="album-body" tabindex="-1" @scroll.passive="onScroll">
            <div
                v-for="item in albumList"
                :key="item.pos"
                class="album-body-item"
                :class="{ current: item.pos === currentPos }"
                :title="item.title || $t('Untitled')"
                tabindex="0"
                @click="clickMedia(item)"
                @keydown="clickOnEnter"
            >
                <MediaItemAlbumThumbnail :item="item" :display-position="true"></MediaItemAlbumThumbnail>

                <div class="album-body-item-title no-btn">
                    {{ item.title || $t("Untitled") }}
                </div>
            </div>
        </div>

        <div v-if="loading" class="album-loader">
            <div class="loading-overlay-loader">
                <div></div>
                <div></div>
                <div></div>
                <div></div>
            </div>
        </div>
    </div>
</template>

<script setup lang="ts">
import { AppStatus } from "@/control/app-status";
import { BigListScroller } from "@/utils/big-list-scroller";
import { computed, nextTick, onMounted, ref, useTemplateRef } from "vue";
import MediaItemAlbumThumbnail from "@/components/utils/MediaItemAlbumThumbnail.vue";
import type { Album, MediaListItem, PositionedMediaListItem } from "@/api/models";
import {
    EVENT_NAME_CURRENT_ALBUM_LOADING,
    EVENT_NAME_CURRENT_ALBUM_MEDIA_POSITION_UPDATED,
    EVENT_NAME_CURRENT_ALBUM_UPDATED,
} from "@/control/app-events";
import { useI18n } from "@/composables/use-i18n";
import { useInterval } from "@/composables/use-interval";
import { onApplicationEvent } from "@/composables/on-app-event";
import { clickOnEnter } from "@/utils/events";
import type { AlbumMediaPositionContext } from "@/control/album";
import {
    albumToggleLoop,
    albumToggleRandom,
    getCurrentAlbumData,
    getCurrentAlbumId,
    getCurrentAlbumMediaPositionContext,
    isCurrentAlbumLoading,
} from "@/control/album";
import { showSnackBar } from "@/control/snack-bar";

// Emits
const emit = defineEmits<{
    /**
     * Emitted when closed
     */
    (e: "close"): void;
}>();

// Translation
const { $t } = useI18n();

// Initial album data
const initialAlbumData = getCurrentAlbumData();

// True if album was loaded
const loadedAlbum = ref(!!initialAlbumData);

// Current album ID
const albumId = ref(getCurrentAlbumId());

// Current album name
const albumName = ref(initialAlbumData ? initialAlbumData.name : "");

// Length of the album list
const albumListLength = ref(initialAlbumData ? initialAlbumData.list.length : 0);

// List of album items
const albumList = ref<PositionedMediaListItem[]>();

// Loading status
const loading = ref(isCurrentAlbumLoading());

onApplicationEvent(EVENT_NAME_CURRENT_ALBUM_LOADING, (l) => {
    if (l) {
        if (albumId.value !== getCurrentAlbumId()) {
            loading.value = true;
        }
    } else {
        loading.value = false;
    }
});

// Initial album media position context
const initialAlbumMediaPositionContext = getCurrentAlbumMediaPositionContext();

// Current album position
const currentPos = ref(initialAlbumMediaPositionContext.pos);

// Rendered current position (to display it to the user)
const renderedCurrentPos = computed<string>(() => {
    if (currentPos.value < 0) {
        return "?";
    } else {
        return "" + (currentPos.value + 1);
    }
});

// Album loop
const loop = ref(initialAlbumMediaPositionContext.loop);

// Album random
const random = ref(initialAlbumMediaPositionContext.random);

// Window size for the scroller
const INITIAL_WINDOW_SIZE = 100;

// Scroller
const listScroller = new BigListScroller(INITIAL_WINDOW_SIZE, {
    get: () => {
        return albumList.value;
    },
    set: (l) => {
        albumList.value = l;
    },
});

/**
 * Handler for 'scroll' on the container
 * @param e Ths scroll event
 */
const onScroll = (e: Event) => {
    listScroller.checkElementScroll(e.target as HTMLElement);
};

/**
 * Updates the album list
 */
const updateAlbumList = () => {
    listScroller.reset();

    if (loadedAlbum.value) {
        let i = 0;

        const originalList = getCurrentAlbumData()?.list || [];

        listScroller.addElements(
            originalList.map((m) => {
                return {
                    pos: i++,
                    id: m.id,
                    type: m.type,
                    title: m.title,
                    thumbnail: m.thumbnail,
                    duration: m.duration,
                    tags: m.tags,
                };
            }),
        );
    }
};

updateAlbumList();

onApplicationEvent(EVENT_NAME_CURRENT_ALBUM_UPDATED, (albumData: Album | null) => {
    albumId.value = getCurrentAlbumId();
    loadedAlbum.value = !!albumData;
    albumName.value = albumData ? albumData.name : "";
    albumListLength.value = albumData ? albumData.list.length : 0;

    updateAlbumList();
});

/**
 * Called when the album position changes
 */
const onAlbumPosUpdate = (ctx: AlbumMediaPositionContext) => {
    loop.value = ctx.loop;
    random.value = ctx.random;
    currentPos.value = ctx.pos;

    listScroller.moveWindowToElement(currentPos.value);

    nextTick(scrollToSelected);
};

onMounted(onAlbumPosUpdate);
onApplicationEvent(EVENT_NAME_CURRENT_ALBUM_MEDIA_POSITION_UPDATED, onAlbumPosUpdate);

// Ref to the album body element
const albumBody = useTemplateRef("albumBody");

/**
 * Scrolls to the selected media
 */
const scrollToSelected = () => {
    const cont = albumBody.value;

    if (!cont) {
        return;
    }

    const contBounds = cont.getBoundingClientRect();

    const el = cont.querySelector(".album-body-item.current") as HTMLElement;

    if (!el) {
        return;
    }

    const elBounds = el.getBoundingClientRect();
    const elTopRel = elBounds.top - contBounds.top + cont.scrollTop;

    const expectedTop = contBounds.height / 2 - elBounds.height / 2;

    const scroll = Math.max(0, Math.min(cont.scrollHeight - contBounds.height, elTopRel - expectedTop));

    cont.scrollTop = scroll;
};

/**
 * Checks container height
 */
const checkContainerHeight = () => {
    const cont = albumBody.value;

    if (!cont) {
        return;
    }

    const el = cont.querySelector(".album-body-item") as HTMLElement;

    if (!el) {
        return;
    }

    const changed = listScroller.checkScrollContainerHeight(cont, el);

    if (changed) {
        onAlbumPosUpdate(getCurrentAlbumMediaPositionContext());
    }
};

// Delay to check the container periodically
const CONTAINER_CHECK_DELAY = 1000;

// Interval to check the container
const checkContainerTimer = useInterval();

onMounted(() => {
    checkContainerTimer.set(checkContainerHeight, CONTAINER_CHECK_DELAY);
});

/**
 * Automatically focuses the container
 */
const autoFocus = () => {
    nextTick(() => {
        albumBody.value?.focus();
        checkContainerHeight();
    });
};

onMounted(autoFocus);

/**
 * Toggles the album loop
 */
const toggleLoop = () => {
    if (albumToggleLoop()) {
        showSnackBar($t("Album loop enabled"));
    } else {
        showSnackBar($t("Album loop disabled"));
    }
};

/**
 * Toggles the album random order
 */
const toggleRandom = () => {
    if (albumToggleRandom()) {
        showSnackBar($t("Album shuffle enabled"));
    } else {
        showSnackBar($t("Album shuffle disabled"));
    }
};

/**
 * Closes the album menu
 */
const close = () => {
    emit("close");
};

/**
 * Clicks on media element
 * @param item The media element
 */
const clickMedia = (item: MediaListItem) => {
    AppStatus.ClickOnMedia(item.id, false);
    close();
};
</script>
