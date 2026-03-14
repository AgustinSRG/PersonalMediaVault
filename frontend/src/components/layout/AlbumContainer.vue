<template>
    <div class="album-container" tabindex="-1" @keydown="onHeaderKeyDown">
        <div v-if="!loading && loadedAlbum" class="album-header">
            <div class="album-header-title">
                <div class="album-title" :title="albumName"><i class="fas fa-list-ol"></i> {{ albumName }}</div>
                <button type="button" :title="$t('Close')" class="album-header-btn album-close-btn" @click="closePage">
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

                    <button type="button" :title="$t('Favorite')" class="album-header-btn" :class="{ toggled: isFav }" @click="toggleFav">
                        <i class="fas fa-star"></i>
                    </button>

                    <button
                        v-if="canWrite && albumListLength < 1024"
                        type="button"
                        :title="$t('Add media')"
                        class="album-header-btn"
                        @click="addMediaToAlbum"
                    >
                        <i class="fas fa-plus"></i>
                    </button>

                    <button v-if="canWrite" type="button" :title="$t('Rename')" class="album-header-btn" @click="renameAlbum">
                        <i class="fas fa-pencil-alt"></i>
                    </button>

                    <button v-if="canWrite" type="button" :title="$t('Thumbnail')" class="album-header-btn" @click="changeAlbumThumbnail">
                        <i class="fas fa-image"></i>
                    </button>

                    <button v-if="canWrite" type="button" :title="$t('Delete')" class="album-header-btn" @click="deleteAlbum">
                        <i class="fas fa-trash-alt"></i>
                    </button>
                </div>
                <div class="album-post-text">{{ renderedCurrentPos }} / {{ albumListLength }}</div>
            </div>
        </div>

        <div
            v-show="!loading && loadedAlbum"
            ref="albumBody"
            class="album-body"
            :class="{ 'is-dragging': dragging }"
            tabindex="-1"
            @scroll.passive="onScroll"
            @keydown="onBodyKeyDown"
        >
            <a
                v-for="item in albumList"
                :key="item.pos"
                :href="getMediaURL(item)"
                target="_blank"
                rel="noopener noreferrer"
                class="album-body-item"
                :class="{
                    current: item.pos === currentPos,
                    dragging: item.pos === draggingPosition,
                    'dragging-over': item.pos === draggingOverPosition,
                }"
                :title="item.title || $t('Untitled')"
                @click="clickMedia(item, $event)"
                @dragstart="onDrag(item, $event)"
            >
                <MediaItemAlbumThumbnail :item="item" :display-position="true"></MediaItemAlbumThumbnail>

                <div class="album-body-item-title">
                    {{ item.title || $t("Untitled") }}
                </div>

                <button
                    v-if="canWrite"
                    type="button"
                    :title="$t('Options')"
                    class="album-body-btn"
                    @click="showOptions(item, item.pos, $event)"
                    @mousedown="stopPropagationEvent"
                    @touchstart="stopPropagationEvent"
                >
                    <i class="fas fa-bars"></i>
                </button>
            </a>
            <div
                v-if="dragging && albumList.length > 0 && draggingOverPosition > albumList[albumList.length - 1].pos"
                class="dragging-padding-bottom"
            ></div>
        </div>

        <div
            v-if="dragging && draggingItem"
            class="album-dragging-helper"
            :style="{ top: mouseY - 65 + 'px', left: mouseX - 65 + 'px' }"
            @click="stopPropagationEvent"
        >
            <div
                :href="getMediaURL(draggingItem)"
                target="_blank"
                rel="noopener noreferrer"
                class="album-body-item"
                :class="{ current: draggingItem.pos === currentPos }"
                :title="draggingItem.title || $t('Untitled')"
            >
                <MediaItemAlbumThumbnail :item="draggingItem" :display-position="true"></MediaItemAlbumThumbnail>

                <div class="album-body-item-title">
                    {{ draggingItem.title || $t("Untitled") }}
                </div>

                <button v-if="canWrite" type="button" :title="$t('Options')" class="album-body-btn" disabled>
                    <i class="fas fa-bars"></i>
                </button>
            </div>
        </div>

        <AlbumContextMenu
            v-if="contextShown"
            v-model:shown="contextShown"
            :media-index="contextIndex"
            :album-length="albumListLength"
            :x="contextX"
            :y="contextY"
            @move-up="moveMediaUp"
            @move-down="moveMediaDown"
            @change-pos="changeMediaPos"
            @media-remove="removeMedia"
        ></AlbumContextMenu>

        <AlbumAddMediaModal v-if="displayUpload" v-model:display="displayUpload" :aid="albumId"></AlbumAddMediaModal>

        <AlbumRenameModal v-if="displayAlbumRename" v-model:display="displayAlbumRename"></AlbumRenameModal>

        <AlbumChangeThumbnailModal
            v-if="displayAlbumChangeThumbnail"
            v-model:display="displayAlbumChangeThumbnail"
        ></AlbumChangeThumbnailModal>

        <AlbumDeleteModal v-if="displayAlbumDelete" v-model:display="displayAlbumDelete"></AlbumDeleteModal>

        <AlbumGoToPosModal v-if="displayAlbumGoPos" v-model:display="displayAlbumGoPos"></AlbumGoToPosModal>

        <AlbumMovePosModal
            v-model:display="displayAlbumMovePos"
            :position-to-move="positionToMove"
            :album-list-length="albumListLength"
            @move="moveMedia"
        ></AlbumMovePosModal>

        <ErrorMessageModal v-if="errorDisplay" v-model:display="errorDisplay" :message="error"></ErrorMessageModal>

        <LoadingOverlay v-if="loading"></LoadingOverlay>
    </div>
</template>

<script setup lang="ts">
import {
    emitAppEvent,
    EVENT_NAME_ALBUMS_CHANGED,
    EVENT_NAME_CURRENT_ALBUM_LOADING,
    EVENT_NAME_CURRENT_ALBUM_MEDIA_POSITION_UPDATED,
    EVENT_NAME_CURRENT_ALBUM_UPDATED,
    EVENT_NAME_FAVORITE_ALBUMS_UPDATED,
} from "@/global-state/app-events";
import { albumAddFav, albumIsFavorite, albumRemoveFav } from "@/local-storage/app-preferences";
import { AppStatus } from "@/global-state/app-status";
import { getFrontendUrl } from "@/utils/api";
import { makeApiRequest } from "@asanrom/request-browser";
import { computed, defineAsyncComponent, nextTick, onMounted, ref, useTemplateRef } from "vue";
import AlbumContextMenu from "../utils/AlbumContextMenu.vue";
import LoadingOverlay from "./LoadingOverlay.vue";
import AlbumMovePosModal from "@/components/modals/AlbumMovePosModal.vue";
import { BigListScroller } from "@/utils/big-list-scroller";
import { isTouchDevice } from "@/utils/touch";
import { apiAlbumsMoveMediaInAlbum, apiAlbumsRemoveMediaFromAlbum } from "@/api/api-albums";
import MediaItemAlbumThumbnail from "../utils/MediaItemAlbumThumbnail.vue";
import type { PositionedMediaListItem } from "@/api/models";
import { onApplicationEvent } from "@/composables/on-app-event";
import { useI18n } from "@/composables/use-i18n";
import { useUserPermissions } from "@/composables/use-user-permissions";
import { useInterval } from "@/composables/use-interval";
import { stopPropagationEvent } from "@/utils/events";
import { useGlobalKeyboardHandler } from "@/composables/use-global-keyboard-handler";
import { onDocumentEvent } from "@/composables/on-document-event";
import { useCommonRequestErrors } from "@/composables/use-common-request-errors";
import { showSnackBar } from "@/global-state/snack-bar";
import type { AlbumMediaPositionContext } from "@/global-state/album";
import {
    albumToggleLoop,
    albumToggleRandom,
    getCurrentAlbumData,
    getCurrentAlbumId,
    getCurrentAlbumMediaPositionContext,
    indicateAlbumMetadataChanged,
    isCurrentAlbumLoading,
    updateAlbumMediaPositionStatus,
} from "@/global-state/album";
import { isVaultLocked } from "@/global-state/auth";

const AlbumGoToPosModal = defineAsyncComponent({
    loader: () => import("@/components/modals/AlbumGoToPosModal.vue"),
});

const AlbumRenameModal = defineAsyncComponent({
    loader: () => import("@/components/modals/AlbumRenameModal.vue"),
});

const AlbumChangeThumbnailModal = defineAsyncComponent({
    loader: () => import("@/components/modals/AlbumChangeThumbnailModal.vue"),
});

const AlbumDeleteModal = defineAsyncComponent({
    loader: () => import("@/components/modals/AlbumDeleteModal.vue"),
});

const AlbumAddMediaModal = defineAsyncComponent({
    loader: () => import("@/components/modals/AlbumAddMediaModal.vue"),
});

const ErrorMessageModal = defineAsyncComponent({
    loader: () => import("@/components/modals/ErrorMessageModal.vue"),
});

// Translation
const { $t } = useI18n();

// User permissions
const { canWrite } = useUserPermissions();

// Reference to the album body element
const albumBody = useTemplateRef("albumBody");

// Display the upload modal?
const displayUpload = defineModel<boolean>("displayUpload");

/**
 * Opens the modal to add media into the album
 */
const addMediaToAlbum = () => {
    displayUpload.value = true;
};

onMounted(() => {
    if (displayUpload.value) {
        displayUpload.value = false;
    }
});

// Loading status
const loading = ref(isCurrentAlbumLoading());

// Current album ID
const albumId = ref(getCurrentAlbumId());

// Initial album data
const initialAlbumData = getCurrentAlbumData();

// Current album name
const albumName = ref(initialAlbumData ? initialAlbumData.name : "");

// Length of the album list
const albumListLength = ref(initialAlbumData ? initialAlbumData.list.length : 0);

// True if album is loaded
const loadedAlbum = ref(!!initialAlbumData);

// True if the album is favorite
const isFav = ref(albumIsFavorite(albumId.value));

// True if the album list must be scrolled
const mustScroll = ref(true);

// Initial album media position context
const initialAlbumMediaPositionContext = getCurrentAlbumMediaPositionContext();

// Current position in the album
const currentPos = ref(initialAlbumMediaPositionContext.pos);

// Rendered current position (to display)
const renderedCurrentPos = computed(() => {
    if (currentPos.value < 0) {
        return "?";
    } else {
        return "" + (currentPos.value + 1);
    }
});

// ID of the current media
const currentPosMedia = ref(currentPos.value >= 0 ? AppStatus.CurrentMedia : -1);

// Loop option
const loop = ref(false);

// Random order option
const random = ref(false);

onApplicationEvent(EVENT_NAME_CURRENT_ALBUM_UPDATED, (albumData) => {
    const currentAlbumId = getCurrentAlbumId();

    if (albumId.value !== currentAlbumId) {
        mustScroll.value = true;
    }

    albumId.value = currentAlbumId;

    loadedAlbum.value = !!albumData;

    albumName.value = albumData ? albumData.name : "";
    albumListLength.value = albumData ? albumData.list.length : 0;

    isFav.value = albumIsFavorite(currentAlbumId);

    updateAlbumList(!mustScroll.value);
});

onApplicationEvent(EVENT_NAME_CURRENT_ALBUM_LOADING, (l: boolean) => {
    if (l) {
        if (albumId.value !== getCurrentAlbumId()) {
            loading.value = true;
        }
    } else {
        loading.value = false;
    }
});

onApplicationEvent(EVENT_NAME_FAVORITE_ALBUMS_UPDATED, () => {
    isFav.value = albumIsFavorite(getCurrentAlbumId());
});

// Album list
const albumList = ref<PositionedMediaListItem[]>([]);

// Initial window size for the list scroller
const INITIAL_WINDOW_SIZE = 100;

// Album list scroller
const listScroller = new BigListScroller(INITIAL_WINDOW_SIZE, {
    get: () => {
        return albumList.value;
    },
    set: (l) => {
        albumList.value = l;
    },
});

/**
 * Event handler for 'scroll' on the list container
 * @param e The scroll event
 */
const onScroll = (e: Event) => {
    closeOptionsMenu();
    listScroller.checkElementScroll(e.target as HTMLElement);
};

/**
 * Updates album list
 * @param preserveWindowScroll Preserve window scroll?
 */
const updateAlbumList = (preserveWindowScroll?: boolean) => {
    currentMenuOpen.value = -1;

    const currentWindowPosition = listScroller.windowPosition;

    let currentScroll = 0;

    const conEl = albumBody.value;

    if (conEl) {
        currentScroll = conEl.scrollTop;
    }

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

        if (preserveWindowScroll) {
            listScroller.setWindowPosition(currentWindowPosition);
            nextTick(() => {
                const conEl = albumBody.value;

                if (conEl) {
                    conEl.scrollTop = Math.min(currentScroll, conEl.scrollHeight - conEl.getBoundingClientRect().height);
                }
            });
        }
    }

    nextTick(() => {
        checkContainerHeight();
    });
};

/**
 * Call when album position updates
 */
const onAlbumPosUpdate = (ctx: AlbumMediaPositionContext) => {
    loop.value = ctx.loop;
    random.value = ctx.random;

    const newPosMedia = ctx.pos >= 0 ? AppStatus.CurrentMedia : -1;

    let newMustScroll = false;

    if (mustScroll.value || currentPosMedia.value !== newPosMedia) {
        mustScroll.value = false;
        newMustScroll = true;
    }

    currentPos.value = ctx.pos;
    currentPosMedia.value = newPosMedia;

    if (newMustScroll && currentPos.value >= 0) {
        nextTick(() => {
            listScroller.moveWindowToElement(currentPos.value);
            nextTick(() => {
                scrollToSelected();
            });
        });
    }
};

onApplicationEvent(EVENT_NAME_CURRENT_ALBUM_MEDIA_POSITION_UPDATED, onAlbumPosUpdate);

// Update the album list and position on component mount
onMounted(() => {
    updateAlbumList();

    mustScroll.value = true;

    onAlbumPosUpdate(getCurrentAlbumMediaPositionContext());
});

/**
 * Closes the page
 */
const closePage = () => {
    AppStatus.CloseAlbum();
};

/**
 * Toggles the loop option
 */
const toggleLoop = () => {
    if (albumToggleLoop()) {
        showSnackBar($t("Album loop enabled"));
    } else {
        showSnackBar($t("Album loop disabled"));
    }
};

/**
 * Toggles the random option
 */
const toggleRandom = () => {
    if (albumToggleRandom()) {
        showSnackBar($t("Album shuffle enabled"));
    } else {
        showSnackBar($t("Album shuffle disabled"));
    }
};

/**
 * Toggles the favorite option
 */
const toggleFav = () => {
    const currentAlbumId = getCurrentAlbumId();

    if (isFav.value) {
        isFav.value = false;
        albumRemoveFav(currentAlbumId);
        showSnackBar($t("Album removed from favorites"));
    } else {
        isFav.value = true;
        albumAddFav(currentAlbumId);
        showSnackBar($t("Album added to favorites"));
    }
};

/**
 * Checks the height of the album container
 */
const checkContainerHeight = () => {
    const cont = albumBody.value;

    if (!cont) {
        return;
    }

    listScroller.checkElementScroll(cont);

    const el = cont.querySelector(".album-body-item") as HTMLElement;

    if (!el) {
        return;
    }

    const changed = listScroller.checkScrollContainerHeight(cont, el);

    if (changed) {
        mustScroll.value = true;
        onAlbumPosUpdate(getCurrentAlbumMediaPositionContext());
    }
};

// Interval to check the container height
const checkContainerTimer = useInterval();

// Interval to check container height (milliseconds)
const CHECK_CONTAINER_HEIGHT_INTERVAL = 1000;

onMounted(() => {
    checkContainerTimer.set(checkContainerHeight, CHECK_CONTAINER_HEIGHT_INTERVAL);
});

// Displays the album rename modal?
const displayAlbumRename = ref(false);

// Displays the modal to change the album thumbnail
const displayAlbumChangeThumbnail = ref(false);

// Displays the modal to delete the album
const displayAlbumDelete = ref(false);

// Displays the modal to go to a position in the album list
const displayAlbumGoPos = ref(false);

/**
 * Opens the modal to rename an album
 */
const renameAlbum = () => {
    displayAlbumRename.value = true;
};

/**
 * Opens the modal to change the album thumbnail
 */
const changeAlbumThumbnail = () => {
    displayAlbumChangeThumbnail.value = true;
};

/**
 * Opens the modal to delete the album
 */
const deleteAlbum = () => {
    displayAlbumDelete.value = true;
};

// Position of the item the context menu is open for
const currentMenuOpen = ref(-1);

// True if context menu shown
const contextShown = ref(false);

// Index of the media for the context menu
const contextIndex = ref(-1);

// X coordinate of context menu
const contextX = ref(0);

// Y coordinate of context menu
const contextY = ref(0);

// Displays the modal to move a media to another position
const displayAlbumMovePos = ref(false);

// Media position to move for the modal
const positionToMove = ref(0);

/**
 * Closes the context menu
 */
const closeOptionsMenu = () => {
    if (contextShown.value) {
        contextShown.value = false;
    }
};

/**
 * Shows the context menu
 * @param item The media item
 * @param i The index of the item
 * @param event The mouse event
 */
const showOptions = (item: PositionedMediaListItem, i: number, event: MouseEvent) => {
    event.preventDefault();
    event.stopPropagation();

    if (contextShown.value && currentMenuOpen.value === item.pos) {
        currentMenuOpen.value = -1;
        contextShown.value = false;
    } else {
        currentMenuOpen.value = item.pos;
        contextShown.value = true;
        contextIndex.value = i;

        const targetRect = (event.target as HTMLElement).getBoundingClientRect();

        contextX.value = targetRect.left + targetRect.width;

        if (targetRect.top > window.innerHeight / 2) {
            contextY.value = targetRect.top;
        } else {
            contextY.value = targetRect.top + targetRect.height;
        }
    }
};

/**
 * Moves the selected media up one position
 * @param i The index of the media
 */
const moveMediaUp = (i: number) => {
    if (i > 0) {
        moveMedia(i, i - 1);
    }
};

/**
 * Moves the selected media down one position
 * @param i The index of the media
 */
const moveMediaDown = (i: number) => {
    if (i < albumListLength.value - 1) {
        moveMedia(i, i + 1);
    }
};

/**
 * Opens a modal to change the position of a media element
 * @param i The index of the media
 */
const changeMediaPos = (i: number) => {
    positionToMove.value = i;
    displayAlbumMovePos.value = true;
};

// Request error
const { error, errorDisplay, setError, unauthorized, badRequest, accessDenied, notFound, serverError, networkError } =
    useCommonRequestErrors();

/**
 * Performs the request to move a media
 * @param oldIndex Old media index
 * @param newIndex New media index
 */
const moveMedia = (oldIndex: number, newIndex: number) => {
    const albumData = getCurrentAlbumData();

    if (!albumData) {
        return;
    }

    if (oldIndex < 0 || oldIndex >= albumData.list.length) {
        return;
    }

    const albumId = albumData.id;
    const mediaId = albumData.list[oldIndex].id;

    albumData.list.splice(newIndex, 0, albumData.list.splice(oldIndex, 1)[0]);

    emitAppEvent(EVENT_NAME_CURRENT_ALBUM_UPDATED, albumData);

    updateAlbumMediaPositionStatus();

    // Update in server

    makeApiRequest(apiAlbumsMoveMediaInAlbum(albumId, mediaId, newIndex))
        .onSuccess(() => {
            emitAppEvent(EVENT_NAME_ALBUMS_CHANGED);
        })
        .onRequestError((err, handleErr) => {
            // Revert changes
            indicateAlbumMetadataChanged(albumId, true);
            // Show error
            handleErr(err, {
                unauthorized,
                maxSizeReached: () => {
                    setError($t("The album reached the limit of 1024 elements. Please, consider creating another album."));
                },
                badRequest,
                accessDenied,
                notFound: () => {
                    notFound();
                    indicateAlbumMetadataChanged(albumId);
                },
                serverError,
                networkError,
            });
        })
        .onUnexpectedError((err) => {
            console.error(err);
            setError(err.message);
        });
};

/**
 * Removes the selected media from the album
 * @param i The index of the media
 */
const removeMedia = (i: number) => {
    const completeAlbumList = listScroller.list;

    const media = completeAlbumList[i];

    if (!media) {
        return;
    }

    const aid = albumId.value;

    makeApiRequest(apiAlbumsRemoveMediaFromAlbum(aid, media.id))
        .onSuccess(() => {
            showSnackBar($t("Successfully removed from album"));
            indicateAlbumMetadataChanged(aid, true);
            emitAppEvent(EVENT_NAME_ALBUMS_CHANGED);
        })
        .onRequestError((err, handleErr) => {
            handleErr(err, {
                unauthorized,
                accessDenied,
                notFound: () => {
                    notFound();
                    indicateAlbumMetadataChanged(aid);
                },
                serverError,
                networkError,
            });
        })
        .onUnexpectedError((err) => {
            console.error(err);
            setError(err.message);
        });
};

// True if an element is being dragged
const dragging = ref(false);

// Position of the element being dragged
const draggingPosition = ref(-1);

// Item being dragged
const draggingItem = ref<PositionedMediaListItem | null>(null);

// Position being dragged over
const draggingOverPosition = ref(-1);

// Current position of the mouse while dragging
const mouseX = ref(0);
const mouseY = ref(0);

// Interval to check the dragging position
const dragCheckInterval = useInterval();

// Delay to check the dragging position (milliseconds)
const DRAG_CHECK_DELAY = 40;

/**
 * Handler for 'dragstart' on a media element
 * @param item The media item
 * @param e The drag event
 */
const onDrag = (item: PositionedMediaListItem, e: DragEvent) => {
    e.preventDefault();

    if (isTouchDevice()) {
        return; // Disabled for touch devices
    }

    if (!canWrite.value) {
        return; // Cannot alter the album
    }

    dragging.value = true;
    draggingItem.value = item;
    draggingPosition.value = item.pos;
    draggingOverPosition.value = item.pos;

    dragCheckInterval.set(onDragCheck, DRAG_CHECK_DELAY);
};

/**
 * Gets the URL for a media element
 * @param item The media item
 * @returns The URL
 */
const getMediaURL = (item: PositionedMediaListItem) => {
    return getFrontendUrl({
        media: item.id,
        album: albumId.value,
    });
};

/**
 * Handler for when a media item is clicked
 * @param item The media item
 * @param e The click event
 */
const clickMedia = (item: PositionedMediaListItem, e?: MouseEvent) => {
    if (e) {
        e.preventDefault();
    }
    AppStatus.ClickOnMedia(item.id, false);
};

/**
 * Scrolls container to selected element
 */
const scrollToSelected = () => {
    const cont = albumBody.value;

    if (!cont) {
        return;
    }

    const contBounds = cont.getBoundingClientRect();

    const el = cont.querySelector(".album-body-item.current");

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
 * Checks the dragging position
 */
const onDragCheck = () => {
    const con = albumBody.value;

    if (!con) {
        return;
    }

    const conBounds = con.getBoundingClientRect();

    if (mouseX.value >= conBounds.left - 420) {
        // Auto scroll

        const relTop = (mouseY.value - conBounds.top) / (conBounds.height || 1);
        const scrollStep = Math.floor(conBounds.height / 10);

        if (relTop <= 0.1) {
            con.scrollTop = Math.max(0, con.scrollTop - scrollStep);
        } else if (relTop >= 0.9) {
            con.scrollTop = Math.min(con.scrollHeight - conBounds.height, con.scrollTop + scrollStep);
        }
    }

    // Check drop position
    checkDropPosition();
};

/**
 * Checks the position where the dragging element will be dropped
 */
const checkDropPosition = () => {
    const con = albumBody.value;

    if (!con) {
        return;
    }

    const conBounds = con.getBoundingClientRect();

    if (albumList.value.length === 0) {
        return;
    }

    const item = con.querySelector(".album-body-item:not(.dragging)") as HTMLElement;

    if (!item) {
        return;
    }

    if (mouseX.value < conBounds.left - 420) {
        draggingOverPosition.value = -1;
        return;
    }

    const itemHeight = item.getBoundingClientRect().height;

    const firstPos = albumList.value[0].pos;
    const lastPos = albumList.value[albumList.value.length - 1].pos;
    const containerScrollTop = con.scrollTop;

    draggingOverPosition.value = Math.min(
        lastPos + 1,
        Math.max(firstPos, firstPos + Math.round((mouseY.value - conBounds.top + containerScrollTop) / itemHeight)),
    );
};

onDocumentEvent("mousemove", (e: MouseEvent) => {
    mouseX.value = e.pageX;
    mouseY.value = e.pageY;
});

onDocumentEvent("mouseup", () => {
    if (!dragging.value) {
        return;
    }

    dragCheckInterval.clear();

    // Check drop position

    checkDropPosition();

    // Move element if needed

    if (draggingOverPosition.value >= 0 && draggingPosition.value >= 0) {
        const initialPosition = draggingPosition.value;
        const finalPosition =
            draggingOverPosition.value > draggingPosition.value ? draggingOverPosition.value - 1 : draggingOverPosition.value;

        if (initialPosition !== finalPosition) {
            moveMedia(initialPosition, finalPosition);
        }
    }

    // Reset dragging
    dragging.value = false;
    draggingPosition.value = -1;
    draggingOverPosition.value = -1;
});

/**
 * Event handler for 'keydown' on the album header element
 * @param e The keyboard event
 */
const onHeaderKeyDown = (e: KeyboardEvent) => {
    if (e.key === "ArrowUp" || e.key === "ArrowDown") {
        e.stopPropagation();

        albumBody.value?.focus();
    }
};

/**
 * Event handler for 'keydown' on the album body element
 * @param e The keyboard event
 */
const onBodyKeyDown = (e: KeyboardEvent) => {
    if (e.key === "ArrowUp" || e.key === "ArrowDown") {
        e.stopPropagation();
    }
};

// Priority for the global keyboard handler
const KEYBOARD_HANDLER_PRIORITY = 10;

// Global keyboard handler
useGlobalKeyboardHandler((event: KeyboardEvent): boolean => {
    if (isVaultLocked() || AppStatus.CurrentLayout !== "album" || !event.key || event.ctrlKey) {
        return false;
    }

    const completeAlbumList = listScroller.list;

    switch (event.key) {
        case "l":
        case "L":
            toggleLoop();
            return true;
        case "r":
        case "R":
            toggleRandom();
            return true;

        case "q":
        case "Q":
            closePage();
            return true;
        case "f":
        case "F":
            toggleFav();
            return true;
        case "g":
        case "G":
            displayAlbumGoPos.value = true;
            return true;
        case "Home":
            if (completeAlbumList.length > 0) {
                clickMedia(completeAlbumList[0]);
            }
            return true;
        case "End":
            if (completeAlbumList.length > 0) {
                clickMedia(completeAlbumList[completeAlbumList.length - 1]);
            }
            return true;
        default:
            return false;
    }
}, KEYBOARD_HANDLER_PRIORITY);
</script>
