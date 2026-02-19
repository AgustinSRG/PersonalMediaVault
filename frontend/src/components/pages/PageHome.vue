<template>
    <div
        ref="container"
        class="page-inner scrollbar-stable"
        :class="{ 'page-editing': editing, 'is-min': min }"
        :style="{
            '--moving-group-width': movingGroupData.width + 'px',
            '--moving-group-height': movingGroupData.height + 'px',
        }"
        @scroll.passive="onScroll"
    >
        <div ref="autoFocusElement" class="search-results auto-focus" tabindex="-1">
            <LoadingOverlay v-if="loading"></LoadingOverlay>

            <div v-else-if="!loading && groups.length == 0 && firstLoaded" class="search-results-msg-display">
                <div class="search-results-msg-icon">
                    <i class="fas fa-box-open"></i>
                </div>
                <div class="search-results-msg-text">
                    {{ $t("The home page is empty") }}
                </div>
                <div v-if="!editing" class="search-results-msg-btn">
                    <button type="button" class="btn btn-primary" @click="load"><i class="fas fa-sync-alt"></i> {{ $t("Refresh") }}</button>
                </div>
                <div v-else class="search-results-msg-btn">
                    <button type="button" class="btn btn-primary"><i class="fas fa-plus"></i> {{ $t("Add first row") }}</button>
                </div>
            </div>

            <div v-else>
                <div v-if="editing && groups.length < MAX_GROUPS_COUNT" class="home-add-row-form">
                    <button type="button" class="btn btn-primary btn-mr" @click="showAddRowPrepend">
                        <i class="fas fa-plus"></i> {{ $t("Add new row") }}
                    </button>
                </div>

                <HomePageRow
                    v-for="(g, i) in groups"
                    :key="g.id"
                    :group="g"
                    :row-size="actualRowSize"
                    :page-size="pageSize"
                    :display-titles="displayTitles"
                    :editing="editing"
                    :moving-over="movingGroup && movingGroupData.movingOver === i + 1"
                    :moving-self="movingGroup && movingGroupData.startPosition === i"
                    :current-media="currentMedia"
                    :is-current-group="currentGroup == g.id"
                    :load-tick="loadTick"
                    :is-mobile-size="isMobileSize"
                    @request-rename="showRenameRow"
                    @request-move="showMoveRow"
                    @request-delete="showDeleteRow"
                    @start-moving="onStartMoving"
                    @loaded-current="scrollToCurrentRow"
                    @must-reload="load"
                    @updated-prev-next="onPrevNextUpdated"
                ></HomePageRow>

                <div
                    v-if="movingGroup && groups.length > 0 && movingGroupData.movingOver > groups.length"
                    class="home-page-moving-groups-extra-padding"
                ></div>

                <HomePageRow
                    v-if="movingGroup && movingGroupData.group"
                    :moving-left="movingGroupData.x + 'px'"
                    :moving-top="movingGroupData.y + 'px'"
                    :group="movingGroupData.group"
                    :row-size="actualRowSize"
                    :page-size="pageSize"
                    :display-titles="displayTitles"
                    :editing="editing"
                    :moving="true"
                    :current-media="currentMedia"
                    :moving-initial-elements="initialMovingElements"
                    :moving-initial-scroll="initialMovingScroll"
                ></HomePageRow>

                <div v-if="editing && groups.length < MAX_GROUPS_COUNT" class="home-add-row-form">
                    <button type="button" class="btn btn-primary btn-mr" @click="showAddRow">
                        <i class="fas fa-plus"></i> {{ $t("Add new row") }}
                    </button>
                </div>
            </div>
        </div>

        <HomePageCreateRowModal
            v-if="displayRowAdd"
            v-model:display="displayRowAdd"
            :prepend="displayRowAddPrepend"
            @new-row="onRowAdded"
        ></HomePageCreateRowModal>

        <HomePageRenameRowModal
            v-if="displayRowRename"
            v-model:display="displayRowRename"
            :selected-row="selectedRow"
            :selected-row-type="selectedRowType"
            :selected-row-name="selectedRowName"
            @renamed="onRowRenamed"
            @must-reload="load"
        ></HomePageRenameRowModal>

        <HomePageMoveRowModal
            v-if="displayRowMove"
            v-model:display="displayRowMove"
            :selected-row="selectedRow"
            :selected-row-type="selectedRowType"
            :selected-row-name="selectedRowName"
            :selected-row-position="selectedRowPosition"
            :max-position="groups.length"
            @moved="onRowMoved"
            @must-reload="load"
        ></HomePageMoveRowModal>

        <HomePageDeleteRowModal
            v-if="displayRowDelete"
            v-model:display="displayRowDelete"
            :selected-row="selectedRow"
            :selected-row-type="selectedRowType"
            :selected-row-name="selectedRowName"
            @row-deleted="onRowDeleted"
            @must-reload="load"
        ></HomePageDeleteRowModal>
    </div>
</template>

<script setup lang="ts">
import {
    emitAppEvent,
    EVENT_NAME_ALBUMS_CHANGED,
    EVENT_NAME_APP_STATUS_CHANGED,
    EVENT_NAME_AUTH_CHANGED,
    EVENT_NAME_HOME_SCROLL_CHANGED,
    EVENT_NAME_MEDIA_DELETE,
    EVENT_NAME_MEDIA_METADATA_CHANGE,
    EVENT_NAME_PAGE_NAV_NEXT,
    EVENT_NAME_PAGE_NAV_PREV,
    EVENT_NAME_UNAUTHORIZED,
} from "@/control/app-events";
import { AuthController } from "@/control/auth";
import { makeNamedApiRequest, abortNamedApiRequest, makeApiRequest } from "@asanrom/request-browser";
import { setNamedTimeout, clearNamedTimeout } from "@/utils/named-timeouts";
import { defineAsyncComponent, nextTick, onBeforeUnmount, onMounted, reactive, ref, useTemplateRef, watch } from "vue";
import { PagesController } from "@/control/pages";
import LoadingOverlay from "../layout/LoadingOverlay.vue";
import type { HomePageElement, HomePageGroup } from "@/api/api-home";
import { apiHomeGetGroups, apiHomeGroupMove } from "@/api/api-home";
import HomePageRow from "./common/HomePageRow.vue";
import { doHomePageSilentSaveAction, getHomePageBackStatePage, HomePageGroupTypes, type HomePageGroupStartMovingData } from "@/utils/home";
import { AppStatus } from "@/control/app-status";
import { useI18n } from "@/composables/use-i18n";
import { useRequestId } from "@/composables/use-request-id";
import { onApplicationEvent } from "@/composables/on-app-event";
import { useInterval } from "@/composables/use-interval";
import { onDocumentEvent } from "@/composables/on-document-event";
import { useGlobalKeyboardHandler } from "@/composables/use-global-keyboard-handler";

const HomePageCreateRowModal = defineAsyncComponent({
    loader: () => import("@/components/modals/HomePageCreateRowModal.vue"),
    loadingComponent: LoadingOverlay,
    delay: 200,
});

const HomePageRenameRowModal = defineAsyncComponent({
    loader: () => import("@/components/modals/HomePageRenameRowModal.vue"),
    loadingComponent: LoadingOverlay,
    delay: 200,
});

const HomePageMoveRowModal = defineAsyncComponent({
    loader: () => import("@/components/modals/HomePageMoveRowModal.vue"),
    loadingComponent: LoadingOverlay,
    delay: 200,
});

const HomePageDeleteRowModal = defineAsyncComponent({
    loader: () => import("@/components/modals/HomePageDeleteRowModal.vue"),
    loadingComponent: LoadingOverlay,
    delay: 200,
});

// Props
const props = defineProps({
    /**
     * Page is in miniature mode
     */
    min: Boolean,

    /**
     * Size of the page
     */
    pageSize: {
        type: Number,
        required: true,
    },

    /**
     * True to display titles
     */
    displayTitles: Boolean,

    /**
     * Preferred row size
     */
    rowSize: {
        type: Number,
        required: true,
    },

    /**
     * Preferred row size for miniature mode
     */
    rowSizeMin: {
        type: Number,
        required: true,
    },

    /**
     * Min size for items
     */
    minItemsSize: {
        type: Number,
        required: true,
    },

    /**
     * Max size for items
     */
    maxItemsSize: {
        type: Number,
        required: true,
    },

    /**
     * True if editing
     */
    editing: Boolean,
});

// Translation
const { $t } = useI18n();

// Max number of groups
const MAX_GROUPS_COUNT = 1024;

// Ref to the container element
const container = useTemplateRef("container");

// Loading status
const loading = ref(false);

// True if it was loaded at least once
const firstLoaded = ref(false);

// Load tick. When updated, the rows are reloaded.
const loadTick = ref(0);

// Home page groups
const groups = ref<HomePageGroup[]>([]);

// Load request ID
const loadRequestId = useRequestId();

// Delay to display the loader (milliseconds)
const LOADER_DISPLAY_DELAY = 330;

// Delay to retry loading (milliseconds)
const LOAD_RETRY_DELAY = 1500;

/**
 * Loads the data
 */
const load = () => {
    clearNamedTimeout(loadRequestId);
    abortNamedApiRequest(loadRequestId);

    scrollToTop();

    setNamedTimeout(loadRequestId, LOADER_DISPLAY_DELAY, () => {
        loading.value = true;
    });

    if (AuthController.Locked) {
        return; // Vault is locked
    }

    makeNamedApiRequest(loadRequestId, apiHomeGetGroups())
        .onSuccess((newGroups) => {
            clearNamedTimeout(loadRequestId);

            loading.value = false;
            firstLoaded.value = true;
            loadTick.value++;

            groups.value = newGroups;

            scrollToCurrentRow();
        })
        .onRequestError((err, handleErr) => {
            handleErr(err, {
                unauthorized: () => {
                    emitAppEvent(EVENT_NAME_UNAUTHORIZED);
                },
                temporalError: () => {
                    // Retry
                    loading.value = true;
                    setNamedTimeout(loadRequestId, LOAD_RETRY_DELAY, load);
                },
            });
        })
        .onUnexpectedError((err) => {
            console.error(err);
            // Retry
            loading.value = true;
            setNamedTimeout(loadRequestId, LOAD_RETRY_DELAY, load);
        });
};

onMounted(load);
onApplicationEvent(EVENT_NAME_AUTH_CHANGED, load);
onApplicationEvent(EVENT_NAME_MEDIA_METADATA_CHANGE, load);
onApplicationEvent(EVENT_NAME_MEDIA_DELETE, load);
onApplicationEvent(EVENT_NAME_ALBUMS_CHANGED, load);

// Reload when page size changes
watch(() => props.pageSize, load);

// Reload when edit mode changes
watch(
    () => props.editing,
    () => {
        movingGroup.value = false;

        storeCurrentScroll();

        load();
    },
);

// Current media
const currentMedia = ref(AppStatus.CurrentMedia);

// Current media group
const currentGroup = ref(AppStatus.CurrentHomePageGroup);

// Relative element to the current media
const currentGroupNext = ref(-1);
const currentGroupPrev = ref(-1);
const currentGroupFirst = ref(-1);
const currentGroupLast = ref(-1);

onApplicationEvent(EVENT_NAME_APP_STATUS_CHANGED, () => {
    const changed = currentMedia.value !== AppStatus.CurrentMedia || currentGroup.value !== AppStatus.CurrentHomePageGroup;

    currentMedia.value = AppStatus.CurrentMedia;
    currentGroup.value = AppStatus.CurrentHomePageGroup;

    if (changed) {
        currentGroupPrev.value = -1;
        currentGroupNext.value = -1;
        currentGroupFirst.value = -1;
        currentGroupLast.value = -1;
        PagesController.OnHomeGroupLoad(false, false);

        scrollToCurrentRow();
    }
});

/**
 * Updates the current media prev-next context
 * @param newCurrentGroupPrev Previous media
 * @param newCurrentGroupNext Next media
 * @param newCurrentGroupFirst First media
 * @param newCurrentGroupLast Last media
 */
const onPrevNextUpdated = (
    newCurrentGroupPrev: number,
    newCurrentGroupNext: number,
    newCurrentGroupFirst: number,
    newCurrentGroupLast: number,
) => {
    currentGroupPrev.value = newCurrentGroupPrev;
    currentGroupNext.value = newCurrentGroupNext;
    currentGroupFirst.value = newCurrentGroupFirst;
    currentGroupLast.value = newCurrentGroupLast;

    PagesController.OnHomeGroupLoad(currentGroupPrev.value >= 0, currentGroupNext.value >= 0);
};

// Make sure to unload the page context before unmounting
onBeforeUnmount(() => {
    PagesController.OnPageUnload();
});

/**
 * Navigates to media element
 * @param id Media ID
 */
const goToMedia = (id: number) => {
    AppStatus.ClickOnMedia(id, true, currentGroup.value);
};

/**
 * Navigates to previous media
 */
const prevMedia = () => {
    if (currentGroupPrev.value >= 0) {
        goToMedia(currentGroupPrev.value);
    }
};

onApplicationEvent(EVENT_NAME_PAGE_NAV_PREV, prevMedia);

/**
 * navigates to the next media
 */
const nextMedia = () => {
    if (currentGroupNext.value >= 0) {
        goToMedia(currentGroupNext.value);
    }
};

onApplicationEvent(EVENT_NAME_PAGE_NAV_NEXT, nextMedia);

// Container width
const windowWidth = ref(0);

// Window width under is is considered mobile size
const WINDOW_MOBILE_SIZE = 600;

// True if the window has a mobile size
const isMobileSize = ref(document.documentElement.clientWidth <= WINDOW_MOBILE_SIZE);

// Actual row size
const actualRowSize = ref(props.rowSize || 1);

/**
 * Called when the window size changes
 */
const updateWindowWidth = () => {
    if (!container.value) {
        return;
    }

    windowWidth.value = container.value.getBoundingClientRect().width;
    isMobileSize.value = document.documentElement.clientWidth <= WINDOW_MOBILE_SIZE;
    updateActualRowSize();
    onScroll();
};

/**
 * Updated the actual size for the rows
 */
const updateActualRowSize = () => {
    const preferRowSize = (props.min ? props.rowSizeMin : props.rowSize) || 1;

    let itemsWidth = windowWidth.value / preferRowSize;

    itemsWidth = Math.min(itemsWidth, Math.min(windowWidth.value, props.maxItemsSize || 0));

    itemsWidth = Math.max(1, Math.max(itemsWidth, props.minItemsSize || 0));

    actualRowSize.value = Math.ceil(windowWidth.value / itemsWidth);

    if (currentMedia.value) {
        scrollToCurrentRow();
    }
};

watch(
    [() => props.rowSize, () => props.rowSizeMin, () => props.minItemsSize, () => props.maxItemsSize, () => props.min],
    updateActualRowSize,
);

const windowResizeObserver = new ResizeObserver(updateWindowWidth);

onMounted(() => {
    if (container.value) {
        windowResizeObserver.observe(container.value);
    }

    updateWindowWidth();
});

onBeforeUnmount(() => {
    windowResizeObserver.disconnect();
});

// Stored scroll in order to maintain it
const storedScroll = ref(0);

// Should scroll be restored?
const shouldRestoreStoreScroll = ref(false);

/**
 * Scrolls the page to the top,
 * or to the stored scroll level.
 */
const scrollToTop = () => {
    let scroll = 0;

    if (shouldRestoreStoreScroll.value) {
        scroll = storedScroll.value * (container.value?.scrollHeight || 1);
        shouldRestoreStoreScroll.value = false;
    }

    if (container.value) {
        container.value.scrollTop = scroll;
    }
};

/**
 * Stores current scroll level
 * in order to restore it later
 */
const storeCurrentScroll = () => {
    storedScroll.value = (container.value?.scrollTop || 0) / (container.value?.scrollHeight || 1);
    shouldRestoreStoreScroll.value = true;
};

// Scrolling to current row?
let scrollingToCurrent = false;

/**
 * Scrolls page to current row
 */
const scrollToCurrentRow = () => {
    if (scrollingToCurrent) {
        return;
    }
    scrollingToCurrent = true;

    const backState = getHomePageBackStatePage();

    if (backState !== null && currentGroup.value === -1) {
        nextTick(() => {
            scrollingToCurrent = false;
            const currentElem = container.value?.querySelector(".home-page-row-" + backState);
            if (currentElem) {
                currentElem.scrollIntoView();
            }
        });
        return;
    }

    nextTick(() => {
        scrollingToCurrent = false;
        const currentElem = container.value?.querySelector(".home-page-row.current");
        if (currentElem) {
            currentElem.scrollIntoView();
        }
    });
};

/**
 * Called whenever the page is scrolled
 */
const onScroll = () => {
    emitAppEvent(EVENT_NAME_HOME_SCROLL_CHANGED);
};

// Element to be auto-focused on load
const autoFocusElement = useTemplateRef("autoFocusElement");

/**
 * Automatically focuses the appropriate element on load
 */
const autoFocus = () => {
    nextTick(() => {
        autoFocusElement.value?.focus();
    });
};

onMounted(autoFocus);

// Display modal to add a row
const displayRowAdd = ref(false);

// Should the added row be prepended?
const displayRowAddPrepend = ref(false);

/**
 * Opens the modal to add a row
 */
const showAddRow = () => {
    displayRowAdd.value = true;
    displayRowAddPrepend.value = false;
};

/**
 * Opens the modal to add a row (prepend mode)
 */
const showAddRowPrepend = () => {
    displayRowAdd.value = true;
    displayRowAddPrepend.value = true;
};

/**
 * Called when a row is added
 * @param row The row
 * @param prepend Has been prepended?
 */
const onRowAdded = (row: HomePageGroup, prepend: boolean) => {
    if (prepend) {
        groups.value.unshift(row);
    } else {
        groups.value.push(row);
    }
};

// Selected row
const selectedRow = ref(-1);
const selectedRowType = ref(HomePageGroupTypes.Custom);
const selectedRowName = ref("");
const selectedRowPosition = ref(0);

// True to display the modal to rename a row
const displayRowRename = ref(false);

/**
 * Opens the modal to rename a row
 * @param group The row
 */
const showRenameRow = (group: HomePageGroup) => {
    displayRowRename.value = true;
    selectedRow.value = group.id;
    selectedRowName.value = group.name;
    selectedRowType.value = group.type;
};

/**
 * Called when a row is renamed
 * @param id The row ID
 * @param newName The new name
 */
const onRowRenamed = (id: number, newName: string) => {
    for (const g of groups.value) {
        if (g.id === id) {
            g.name = newName;
            break;
        }
    }
};

// True to display the modal to move a row
const displayRowMove = ref(false);

/**
 * Opens the modal to move a row
 * @param group The row
 */
const showMoveRow = (group: HomePageGroup) => {
    let startPosition = 0;
    for (let i = 0; i < groups.value.length; i++) {
        if (groups.value[i].id === group.id) {
            startPosition = i;
            break;
        }
    }

    displayRowMove.value = true;
    selectedRow.value = group.id;
    selectedRowPosition.value = startPosition;
    selectedRowName.value = group.name;
    selectedRowType.value = group.type;
};

/**
 * Called when row is moved
 * @param id The row ID
 * @param position The new position
 */
const onRowMoved = (id: number, position: number) => {
    position = Math.max(0, Math.min(position, groups.value.length));
    for (let i = 0; i < groups.value.length; i++) {
        if (groups.value[i].id === id) {
            groups.value.splice(position, 0, groups.value.splice(i, 1)[0]);
            return;
        }
    }
};

// True to display the modal to delete a row
const displayRowDelete = ref(false);

/**
 * Thens the modal to delete a row
 * @param group The row
 */
const showDeleteRow = (group: HomePageGroup) => {
    displayRowDelete.value = true;
    selectedRow.value = group.id;
    selectedRowName.value = group.name;
    selectedRowType.value = group.type;
};

/**
 * Called when a row is deleted
 * @param id The row ID
 */
const onRowDeleted = (id: number) => {
    for (let i = 0; i < groups.value.length; i++) {
        if (groups.value[i].id === id) {
            groups.value.splice(i, 1);
            return;
        }
    }
};

// Interval to check the drag status
const dragCheckInterval = useInterval();

// Elements before movement
let initialMovingElements: HomePageElement[] | null = null;

// Scroll before row movement
let initialMovingScroll = 0;

// TRue if a group is being moved by dragging it
const movingGroup = ref(false);

// X position of the mouse while dragging
const mouseX = ref(0);

// Y position of the mouse while dragging
const mouseY = ref(0);

// Data to keep track of the move-by-drag process
const movingGroupData = reactive({
    group: null as HomePageGroup,

    startX: 0,
    startY: 0,

    offsetX: 0,
    offsetY: 0,

    width: 0,
    height: 0,

    x: 0,
    y: 0,

    startPosition: -1,
    movingOver: -1,
});

// Drag check delay (milliseconds)
const DRAG_CHECK_DELAY = 40;

/**
 * Call to start moving a group
 * @param group The group
 * @param moveData The initial move data
 */
const onStartMoving = (group: HomePageGroup, moveData: HomePageGroupStartMovingData) => {
    let startPosition = -1;
    for (let i = 0; i < groups.value.length; i++) {
        if (groups.value[i].id === group.id) {
            startPosition = i;
            break;
        }
    }

    mouseX.value = moveData.startX;
    mouseY.value = moveData.startY;

    movingGroup.value = true;
    movingGroupData.group = group;
    movingGroupData.startPosition = startPosition;

    movingGroupData.startX = moveData.startX;
    movingGroupData.startY = moveData.startY;

    movingGroupData.offsetX = moveData.offsetX;
    movingGroupData.offsetY = moveData.offsetY;

    movingGroupData.width = moveData.width;
    movingGroupData.height = moveData.height;

    movingGroupData.x = moveData.startX - moveData.offsetX;
    movingGroupData.y = moveData.startY - moveData.offsetY;

    initialMovingElements = moveData.initialElements;
    initialMovingScroll = moveData.initialScroll;

    updateMovingOver();

    dragCheckInterval.set(onDragCheck, DRAG_CHECK_DELAY);
};

/**
 * Updates the element being moved over
 */
const updateMovingOver = () => {
    const topAddButtonForm = container.value?.querySelector(".home-add-row-form");
    const containerEl = container.value;

    if (!containerEl || !topAddButtonForm) {
        movingGroupData.movingOver = -1;
        return;
    }

    const offsetTop = containerEl.getBoundingClientRect().top + topAddButtonForm.getBoundingClientRect().height;
    const scrollTop = containerEl.scrollTop || 0;

    const y = movingGroupData.y + Math.round(movingGroupData.height / 2);

    const height = movingGroupData.height;
    const expectedIndex = Math.round((y - offsetTop + scrollTop) / height);

    movingGroupData.movingOver = Math.min(groups.value.length + 1, Math.max(1, 1 + expectedIndex));
};

/**
 * Checks the dragging status
 */
const onDragCheck = () => {
    const con = container.value;

    if (!con) {
        return;
    }

    const conBounds = con.getBoundingClientRect();

    if (mouseX.value >= conBounds.left - movingGroupData.height) {
        // Auto scroll

        const relTop = (mouseY.value - conBounds.top) / (conBounds.height || 1);
        const scrollStep = Math.floor(conBounds.height / 20);

        if (relTop <= 0.1) {
            con.scrollTop = Math.max(0, con.scrollTop - scrollStep);
        } else if (relTop >= 0.9) {
            con.scrollTop = Math.min(con.scrollHeight - conBounds.height, con.scrollTop + scrollStep);
        }
    }

    updateMovingOver();
};

// Listener for 'mousemove'
onDocumentEvent("mousemove", (event: MouseEvent) => {
    if (!movingGroup.value) {
        return;
    }

    if (typeof event.pageX !== "number" || typeof event.pageY !== "number") {
        return;
    }

    if (isNaN(event.pageX) || isNaN(event.pageY)) {
        return;
    }

    mouseX.value = event.pageX;
    mouseY.value = event.pageY;

    movingGroupData.x = event.pageX - movingGroupData.offsetX;
    movingGroupData.y = event.pageY - movingGroupData.offsetY;
});

// Listener for 'mouseup'
onDocumentEvent("mouseup", (event: MouseEvent) => {
    if (!movingGroup.value) {
        return;
    }

    dragCheckInterval.clear();

    event.stopPropagation();

    movingGroup.value = false;

    if (!movingGroupData.group) {
        return;
    }

    const groupId = movingGroupData.group.id;

    let position =
        movingGroupData.movingOver > movingGroupData.startPosition + 1 ? movingGroupData.movingOver - 1 : movingGroupData.movingOver;

    position = Math.max(0, Math.min(groups.value.length, position - 1));

    const startPosition = movingGroupData.startPosition;

    if (startPosition === -1 || position === startPosition) {
        return;
    }

    doSilentMove(groupId, position);
    groups.value.splice(position, 0, groups.value.splice(startPosition, 1)[0]);
});

/**
 * Silently moves the row by calling the API
 * @param rowId The row ID
 * @param position The new row position
 */
const doSilentMove = (rowId: number, position: number) => {
    doHomePageSilentSaveAction((callback) => {
        makeApiRequest(apiHomeGroupMove(rowId, position))
            .onSuccess(() => {
                callback();
            })
            .onCancel(() => {
                callback();
            })
            .onRequestError((err, handleErr) => {
                callback();
                handleErr(err, {
                    unauthorized: () => {
                        emitAppEvent(EVENT_NAME_UNAUTHORIZED);
                    },
                    accessDenied: () => {
                        AuthController.CheckAuthStatus();
                    },
                    notFound: () => {
                        load();
                    },
                    temporalError: () => {
                        load();
                    },
                });
            })
            .onUnexpectedError((err) => {
                callback();
                console.error(err);
                load();
            });
    });
};

// Priority for the global keyboard handler
const KEYBOARD_HANDLER_PRIORITY = 20;

// Global keyboard handler
useGlobalKeyboardHandler((event: KeyboardEvent): boolean => {
    if (AuthController.Locked || !AppStatus.IsPageVisible() || !event.key || event.ctrlKey || props.editing) {
        return false;
    }

    if (event.shiftKey && event.key === "PageUp") {
        const activeElement = document.activeElement as HTMLElement;

        if (activeElement && activeElement.classList.contains("home-page-row") && activeElement.previousElementSibling) {
            const sibling = activeElement.previousElementSibling as HTMLElement;
            sibling.focus();
        } else if (
            activeElement &&
            activeElement.parentElement &&
            activeElement.parentElement.parentElement &&
            activeElement.parentElement.parentElement.parentElement &&
            activeElement.parentElement.parentElement.parentElement.parentElement
        ) {
            const rowElement = activeElement.parentElement.parentElement.parentElement.parentElement;

            if (rowElement.classList.contains("home-page-row") && rowElement.previousElementSibling) {
                const sibling = rowElement.previousElementSibling as HTMLElement;
                sibling.focus();
            }
        }

        return true;
    }

    if (event.shiftKey && event.key === "PageDown") {
        const activeElement = document.activeElement as HTMLElement;

        if (activeElement && activeElement.classList.contains("home-page-row") && activeElement.nextElementSibling) {
            const sibling = activeElement.nextElementSibling as HTMLElement;
            sibling.focus();
        } else if (
            activeElement &&
            activeElement.parentElement &&
            activeElement.parentElement.parentElement &&
            activeElement.parentElement.parentElement.parentElement &&
            activeElement.parentElement.parentElement.parentElement.parentElement
        ) {
            const rowElement = activeElement.parentElement.parentElement.parentElement.parentElement;

            if (rowElement.classList.contains("home-page-row") && rowElement.nextElementSibling) {
                const sibling = rowElement.nextElementSibling as HTMLElement;
                sibling.focus();
            }
        }

        return true;
    }

    if (event.key === "Home") {
        if (currentGroupFirst.value > 0) {
            goToMedia(currentGroupFirst.value);
        }
        return true;
    }

    if (event.key === "End") {
        if (currentGroupLast.value > 0) {
            goToMedia(currentGroupLast.value);
        }
        return true;
    }

    if (event.key === "PageUp" || event.key === "ArrowLeft") {
        prevMedia();
        return true;
    }

    if (event.key === "PageDown" || event.key === "ArrowRight") {
        nextMedia();
        return true;
    }

    return false;
}, KEYBOARD_HANDLER_PRIORITY);
</script>
