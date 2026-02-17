<template>
    <div
        ref="container"
        class="home-page-row"
        :class="{
            moving: moving,
            expanded: expanded,
            mobile: isMobileSize,
            'moving-over': movingOver,
            'moving-self': movingSelf,
            current: isCurrentGroup,
            customizable: group.type == HomePageGroupTypes.Custom,
            'fast-transition': fastTransition,
            ['home-page-row-' + group.id]: true,
        }"
        :style="{ '--actual-row-size': rowSize + '', '--row-scroll-index': rowIndex + '', top: movingTop, left: movingLeft }"
        tabindex="-1"
        @keydown="onKeyDown"
    >
        <div class="home-page-row-inner">
            <div
                class="home-page-row-head"
                :draggable="!IS_TOUCH_DEVICE && !isMobileSize && !movePositionModalDisplay && !displayAddElement && !contextMenuShown"
                @dragstart="onDragStartRow"
            >
                <div class="home-page-row-title" :title="groupName">{{ groupName }}</div>
                <div v-if="editing" class="home-page-row-head-buttons">
                    <button
                        v-if="!loading && group.type == HomePageGroupTypes.Custom"
                        type="button"
                        class="page-header-btn"
                        :title="$t('Add elements')"
                        @click="addElements"
                    >
                        <i class="fas fa-plus"></i>
                    </button>
                    <button type="button" class="page-header-btn" :title="$t('Rename row')" @click="renameRow">
                        <i class="fas fa-pencil-alt"></i>
                    </button>
                    <button type="button" class="page-header-btn" :title="$t('Move row')" @click="moveRow">
                        <i class="fas fa-arrows-up-down-left-right"></i>
                    </button>
                    <button type="button" class="page-header-btn" :title="$t('Delete row')" @click="deleteRow">
                        <i class="fas fa-trash-alt"></i>
                    </button>
                </div>
            </div>

            <div
                v-if="(!editing || group.type !== HomePageGroupTypes.Custom) && !loadDisplay && firstLoaded && elements.length === 0"
                class="home-page-row-content home-page-row-loading"
            >
                <div v-if="pageSize > 0" class="search-result-item hidden">
                    <div class="search-result-thumb">
                        <div class="search-result-thumb-inner">
                            <div class="search-result-loader">
                                <i class="fa fa-spinner fa-spin"></i>
                            </div>
                        </div>
                    </div>
                    <div v-if="displayTitles" class="search-result-title">{{ $t("Loading") }}...</div>
                </div>
                <div v-if="editing && group.type !== HomePageGroupTypes.Custom" class="home-page-row-message">
                    <span>{{ $t("This row cannot be customized") }} </span>
                </div>
                <div v-else class="home-page-row-message home-page-row-message-empty">
                    <span><i class="fas fa-box-open"></i> {{ $t("This row is empty") }}</span>
                </div>
            </div>

            <div v-else-if="!visible" class="home-page-row-content home-page-row-loading">
                <div v-if="pageSize > 0" class="search-result-item hidden">
                    <div class="search-result-thumb">
                        <div class="search-result-thumb-inner">
                            <div class="search-result-loader">
                                <i class="fa fa-spinner fa-spin"></i>
                            </div>
                        </div>
                    </div>
                    <div v-if="displayTitles" class="search-result-title">{{ $t("Loading") }}...</div>
                </div>
            </div>

            <div v-else-if="loadDisplay || !firstLoaded" class="home-page-row-content home-page-row-loading">
                <div v-for="f in pageSize" :key="f" class="search-result-item" :class="{ hidden: !loadDisplay }">
                    <div class="search-result-thumb">
                        <div class="search-result-thumb-inner">
                            <div class="search-result-loader">
                                <i class="fa fa-spinner fa-spin"></i>
                            </div>
                        </div>
                    </div>
                    <div v-if="displayTitles" class="search-result-title">{{ $t("Loading") }}...</div>
                </div>
            </div>

            <div
                v-else
                class="home-page-row-content"
                :style="{
                    '--moving-element-width': draggingData.width + 'px',
                }"
                @scroll.passive="onContentScroll"
            >
                <div
                    v-for="(item, i) in elements"
                    :key="i"
                    class="search-result-item"
                    :class="{
                        current: !editing && isCurrent(item, currentMedia, isCurrentGroup),
                        dragging: dragging && draggingData.startPosition === i,
                        'dragging-over': dragging && draggingData.movingOver === i + 1,
                    }"
                >
                    <a
                        class="clickable"
                        :class="{ moveable: editing, ['home-page-row-element-' + i]: true }"
                        :href="getElementURL(item)"
                        target="_blank"
                        rel="noopener noreferrer"
                        @click="goToElement(item, i, $event)"
                        @focus="focusElementIndex(i)"
                        @contextmenu="showContextMenu(i, $event)"
                        @dragstart="onDragStart(i, $event)"
                    >
                        <HomePageItemThumbnail
                            :item="item"
                            :editing="editing"
                            :position="i"
                            @open-options="showContextMenu"
                        ></HomePageItemThumbnail>

                        <div v-if="displayTitles" class="search-result-title">
                            {{ getTitle(item) }}
                        </div>
                    </a>
                </div>

                <div
                    v-if="dragging && elements.length > 0 && draggingData.movingOver > elements.length"
                    class="home-page-moving-elements-extra-padding"
                ></div>

                <div
                    v-if="editing && group.type === HomePageGroupTypes.Custom && elements.length < LIMIT_CUSTOM_ROW_ELEMENTS"
                    class="search-result-item add-home-element"
                >
                    <div tabindex="0" class="search-result-thumb add-home-element-btn" :title="$t('Add elements')" @click="addElements">
                        <div class="search-result-thumb-inner">
                            <div class="search-result-loader">
                                <i class="fa fa-plus"></i>
                            </div>
                        </div>
                    </div>
                    <div v-if="displayTitles" class="search-result-title">{{ $t("Add elements") }}</div>
                </div>

                <div v-else-if="editing && group.type !== HomePageGroupTypes.Custom" class="home-page-row-message">
                    <span>{{ $t("This row cannot be customized") }}</span>
                </div>

                <div v-if="!editing && rowIndex > 0" class="home-page-row-go-left">
                    <button
                        type="button"
                        class="home-page-row-go-button"
                        :title="$t('Scroll to the left')"
                        @click="goLeft"
                        @dblclick="goFirst"
                    >
                        <i class="fas fa-chevron-left"></i>
                    </button>
                </div>

                <div v-if="!editing && rowIndex < rowSplitCount - 1" class="home-page-row-go-right">
                    <button
                        type="button"
                        class="home-page-row-go-button"
                        :title="$t('Scroll to the right')"
                        @click="goRight"
                        @dblclick="goLast"
                    >
                        <i class="fas fa-chevron-right"></i>
                    </button>
                </div>
            </div>

            <div v-if="!editing && isMobileSize && !expanded" class="home-page-row-go-bottom">
                <button type="button" class="home-page-row-go-button home-page-row-expand-button" :title="$t('Expand')" @click="expandRow">
                    <i class="fas fa-chevron-down"></i>
                </button>
            </div>
        </div>

        <div
            v-if="dragging && draggingElement"
            class="search-result-item home-element-dragging"
            :class="{ current: !editing && isCurrent(draggingElement, currentMedia, isCurrentGroup) }"
            :style="{
                left: draggingData.x + 'px',
                top: draggingData.y + 'px',
                width: draggingData.width + 'px',
                height: draggingData.height + 'px',
                'min-width': draggingData.width + 'px',
                'min-height': draggingData.height + 'px',
            }"
        >
            <a class="clickable" :href="getElementURL(draggingElement)" target="_blank" rel="noopener noreferrer">
                <HomePageItemThumbnail
                    :item="draggingElement"
                    :editing="editing"
                    :position="draggingData.startPosition"
                ></HomePageItemThumbnail>
                <div v-if="displayTitles" class="search-result-title">
                    {{ getTitle(draggingElement) }}
                </div>
            </a>
        </div>

        <HomePageElementContextMenu
            v-if="contextMenuShown"
            v-model:shown="contextMenuShown"
            :element-index="contextMenuSelectedIndex"
            :x="contextMenuX"
            :y="contextMenuY"
            @element-remove="onElementRemove"
            @change-pos="onChangePositionRequest"
        ></HomePageElementContextMenu>

        <HomePageRowAddElementModal
            v-if="displayAddElement"
            v-model:display="displayAddElement"
            :group-id="group.id"
            :group-elements="elements"
            @must-reload="load"
            @added-element="load"
        ></HomePageRowAddElementModal>

        <HomePageMoveElementModal
            v-if="movePositionModalDisplay"
            v-model:display="movePositionModalDisplay"
            :selected-position="movePositionModalSelectedIndex"
            :max-position="elements.length - 1"
            @move-element="onMoveElement"
        ></HomePageMoveElementModal>
    </div>
</template>

<script setup lang="ts">
import type { HomePageElement, HomePageGroup } from "@/api/api-home";
import { apiHomeGetGroupElements, apiHomeGroupDeleteElement, apiHomeGroupMoveElement, getHomePageElementReference } from "@/api/api-home";
import { AppStatus } from "@/control/app-status";
import { TagsController } from "@/control/tags";
import { getFrontendUrl } from "@/utils/api";
import type { HomePageGroupStartMovingData } from "@/utils/home";
import {
    doHomePageSilentSaveAction,
    getDefaultGroupName,
    getHomePageBackStateRow,
    HomePageGroupTypes,
    setHomePageBackState,
} from "@/utils/home";
import { clearNamedTimeout, setNamedTimeout } from "@/utils/named-timeouts";
import { isTouchDevice } from "@/utils/touch";
import { abortNamedApiRequest, makeApiRequest, makeNamedApiRequest } from "@asanrom/request-browser";
import type { PropType } from "vue";
import { computed, defineAsyncComponent, nextTick, onBeforeUnmount, onMounted, reactive, ref, useTemplateRef, watch } from "vue";
import HomePageItemThumbnail from "./HomePageItemThumbnail.vue";
import { AuthController } from "@/control/auth";
import { apiSearch } from "@/api/api-search";
import { emitAppEvent, EVENT_NAME_HOME_SCROLL_CHANGED, EVENT_NAME_UNAUTHORIZED } from "@/control/app-events";
import { apiAlbumsGetAlbums } from "@/api/api-albums";
import { useRequestId } from "@/composables/use-request-id";
import { useInterval } from "@/composables/use-interval";
import { onApplicationEvent } from "@/composables/on-app-event";
import { useI18n } from "@/composables/use-i18n";

const HomePageRowAddElementModal = defineAsyncComponent({
    loader: () => import("@/components/modals/HomePageRowAddElementModal.vue"),
});

const HomePageElementContextMenu = defineAsyncComponent({
    loader: () => import("@/components/pages/common/HomePageElementContextMenu.vue"),
});

const HomePageMoveElementModal = defineAsyncComponent({
    loader: () => import("@/components/modals/HomePageMoveElementModal.vue"),
});

// Translation function
const { $t } = useI18n();

// True if the device uses a touch screen
const IS_TOUCH_DEVICE = isTouchDevice();

// Hard limit of elements that can be requested from the server
const LIMIT_CUSTOM_ROW_ELEMENTS = 256;

// Props
const props = defineProps({
    /**
     * Home page row being displayed
     */
    group: {
        type: Object as PropType<HomePageGroup>,
        required: true,
    },

    /**
     * Property to be used as a signal to reload
     */
    loadTick: Number,

    /**
     * True if the row is being displayed in mobile devices
     */
    isMobileSize: Boolean,

    /**
     * Row size
     */
    rowSize: {
        type: Number,
        required: true,
    },

    /**
     * Page size
     */
    pageSize: {
        type: Number,
        required: true,
    },

    /**
     * True to display the titles of the elements
     */
    displayTitles: Boolean,

    /**
     * True if edit mode
     */
    editing: Boolean,

    /**
     * True if the row is being moved
     */
    moving: Boolean,

    /**
     * Initial elements to display when moving
     */
    movingInitialElements: Array as PropType<HomePageElement[]>,

    /**
     * Initial scroll for elements being moved
     */
    movingInitialScroll: Number,

    /**
     * True if other row is being moved over this row
     */
    movingOver: Boolean,

    /**
     * True if the row itself is being moved
     */
    movingSelf: Boolean,

    /**
     * Style position property 'left', when moving
     */
    movingLeft: String,

    /**
     * Style position property 'top', when moving
     */
    movingTop: String,

    /**
     * Current media ID
     */
    currentMedia: Number,

    /**
     * True of the group is the current one
     */
    isCurrentGroup: Boolean,
});

// Name of the group
const groupName = computed(() => props.group.name || getDefaultGroupName(props.group.type, $t));

// Events
const emit = defineEmits<{
    /**
     * The user requested the row being renamed
     */
    (e: "request-rename", group: HomePageGroup): void;

    /**
     * The user requested the row being moved
     */
    (e: "request-move", group: HomePageGroup): void;

    /**
     * The user requested the row being deleted
     */
    (e: "request-delete", group: HomePageGroup): void;

    /**
     * The user started to move the row
     */
    (e: "start-moving", group: HomePageGroup, data: HomePageGroupStartMovingData): void;

    /**
     * The row loaded and is the current row
     */
    (e: "loaded-current"): void;

    /**
     * The home page must be reloaded
     */
    (e: "must-reload"): void;

    /**
     * Updates the position context of the current media
     */
    (e: "updated-prev-next", prevElement: number, nextElement: number, firstElement: number, lastElement: number): void;
}>();

/**
 * The user wants to rename the row
 */
const renameRow = () => {
    emit("request-rename", props.group);
};

/**
 * The user wants to move the row
 */
const moveRow = () => {
    emit("request-move", props.group);
};

/**
 * The user wants to delete the row
 */
const deleteRow = () => {
    emit("request-delete", props.group);
};

// True if visible
const visible = ref(false);

// True if the load process was triggered
const loadTriggered = ref(false);

// Current index of the row being displayed
const rowIndex = ref(0);

// Number of split parts of the row
const rowSplitCount = ref(1);

// True to disable animations for transitions
const fastTransition = ref(false);

// List of loaded elements
const elements = ref<HomePageElement[]>([]);

// True if the row is expanded (for mobile)
const expanded = ref(false);

/**
 * Expands the row
 */
const expandRow = () => {
    expanded.value = true;
};

// Ref to the container element
const container = useTemplateRef("container");

// True if loading
const loading = ref(true);

// True if the row was loaded at least one time
const firstLoaded = ref(false);

// True to display the loader
const loadDisplay = ref(false);

// Load request ID
const loadRequestId = useRequestId();

/**
 * Unloads the row
 */
const unload = () => {
    clearNamedTimeout(loadRequestId);
    abortNamedApiRequest(loadRequestId);
    loadTriggered.value = false;
    loading.value = true;
    firstLoaded.value = false;
    elements.value = [];
};

/**
 * Checks if the row should be loaded
 * @param forced To force the load of the row
 */
const checkLoad = (forced?: boolean) => {
    const containerParent = container.value?.parentElement?.parentElement?.parentElement as HTMLElement;

    if (!containerParent) {
        return;
    }

    const containerBounds = containerParent.getBoundingClientRect();
    const rowBounds = (container.value as HTMLElement).getBoundingClientRect();

    const graceSize = rowBounds.height / 2;

    const visibleTop = rowBounds.top >= containerBounds.top - graceSize && rowBounds.top <= containerBounds.bottom + graceSize;
    const visibleBottom = rowBounds.bottom >= containerBounds.top - graceSize && rowBounds.bottom <= containerBounds.bottom + graceSize;

    const isVisible = visibleTop || visibleBottom;

    visible.value = isVisible;

    if (loadTriggered.value) {
        if (forced) {
            load();
        } else {
            const outOfReach =
                !isVisible &&
                Math.abs(containerBounds.top + containerBounds.height / 2 - (rowBounds.top + rowBounds.height / 2)) >
                    containerBounds.height * 2;

            if (outOfReach) {
                unload();
            }
        }
        return;
    }

    if (isVisible) {
        loadTriggered.value = true;
        load();
    }
};

// When the home page is scrolled, check if the row should be loaded
onApplicationEvent(EVENT_NAME_HOME_SCROLL_CHANGED, () => {
    contextMenuShown.value = false;
    checkLoad();
});

// When mounted, check for load
onMounted(() => {
    checkLoad(true);
});

// If page size changes,, check for load
watch(
    () => props.pageSize,
    () => {
        checkLoad(true);
    },
);

// If the load tick signal updates, load
watch(
    () => props.loadTick,
    () => {
        expanded.value = false; // When reloading, reset the expande status
        checkLoad(true);
    },
);

watch(
    () => props.movingSelf,
    () => {
        nextTick(() => {
            checkLoad();
        });
    },
);

/**
 * Loads the row data
 */
const load = () => {
    loading.value = true;

    if (props.moving && props.movingInitialElements) {
        onElementsLoaded(props.movingInitialElements);

        if (props.movingInitialScroll) {
            applyInitialMovingScroll();
        }
        return;
    }

    switch (props.group.type) {
        case HomePageGroupTypes.Custom:
            loadCustomElements();
            break;
        case HomePageGroupTypes.RecentMedia:
            loadRecentMedia();
            break;
        case HomePageGroupTypes.RecentAlbums:
            loadRecentAlbums();
            break;
        default:
            onElementsLoaded([]);
    }
};

/**
 * Call when elements are loaded
 * in order to update them.
 * @param newElements New list of elements
 */
const onElementsLoaded = (newElements: HomePageElement[]) => {
    elements.value = newElements;
    firstLoaded.value = true;
    loading.value = false;
    loadDisplay.value = false;

    rowIndex.value = 0;

    const backState = getHomePageBackStateRow(props.group.id);

    if (backState !== null) {
        rowIndex.value = Math.floor(backState / (props.rowSize || 1));
    }

    updateRowSplits();
    updateCurrentMedia();

    if (props.isCurrentGroup) {
        emit("loaded-current");
    }
};

// Delay to retry after error (milliseconds)
const LOAD_RETRY_DELAY = 1500;

// Delay to display the loader (milliseconds)
const LOADER_DISPLAY_DELAY = 330;

/**
 * Loads custom elements for the row
 */
const loadCustomElements = () => {
    clearNamedTimeout(loadRequestId);
    abortNamedApiRequest(loadRequestId);

    setNamedTimeout(loadRequestId, LOADER_DISPLAY_DELAY, () => {
        loadDisplay.value = true;
    });

    if (AuthController.Locked) {
        return; // Vault is locked
    }

    makeNamedApiRequest(loadRequestId, apiHomeGetGroupElements(props.group.id))
        .onSuccess((result) => {
            TagsController.OnMediaListReceived(result.filter((r) => !!r.media).map((r) => r.media));
            clearNamedTimeout(loadRequestId);

            onElementsLoaded(result);
        })
        .onRequestError((err, handleErr) => {
            handleErr(err, {
                unauthorized: () => {
                    emitAppEvent(EVENT_NAME_UNAUTHORIZED);
                },
                notFound: () => {
                    emit("must-reload");
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

/**
 * Loads recent media elements for the row
 */
const loadRecentMedia = () => {
    clearNamedTimeout(loadRequestId);
    abortNamedApiRequest(loadRequestId);

    setNamedTimeout(loadRequestId, LOADER_DISPLAY_DELAY, () => {
        loadDisplay.value = true;
    });

    if (AuthController.Locked) {
        return; // Vault is locked
    }

    makeNamedApiRequest(loadRequestId, apiSearch("", "desc", 0, props.pageSize))
        .onSuccess((result) => {
            TagsController.OnMediaListReceived(result.page_items);
            clearNamedTimeout(loadRequestId);

            onElementsLoaded(
                result.page_items.map((r) => {
                    return {
                        media: r,
                    };
                }),
            );
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

/**
 * Loads recent albums for the row
 */
const loadRecentAlbums = () => {
    clearNamedTimeout(loadRequestId);
    abortNamedApiRequest(loadRequestId);

    setNamedTimeout(loadRequestId, LOADER_DISPLAY_DELAY, () => {
        loadDisplay.value = true;
    });

    if (AuthController.Locked) {
        return; // Vault is locked
    }

    makeNamedApiRequest(loadRequestId, apiAlbumsGetAlbums())
        .onSuccess((result) => {
            clearNamedTimeout(loadRequestId);

            onElementsLoaded(
                result
                    .sort((a, b) => {
                        if (a.lm > b.lm) {
                            return -1;
                        } else if (b.lm > a.lm) {
                            return 1;
                        } else if (a.id < b.id) {
                            return 1;
                        } else {
                            return -1;
                        }
                    })
                    .slice(0, props.pageSize || props.rowSize)
                    .map((a) => {
                        return { album: a };
                    }),
            );
        })
        .onRequestError((err, handleErr) => {
            handleErr(err, {
                unauthorized: () => {
                    emitAppEvent(EVENT_NAME_UNAUTHORIZED);
                },
                temporalError: () => {
                    // Retry
                    loading.value = true;
                    setNamedTimeout(loadRequestId, LOADER_DISPLAY_DELAY, load);
                },
            });
        })
        .onUnexpectedError((err) => {
            console.error(err);
            // Retry
            loading.value = true;
            setNamedTimeout(loadRequestId, LOADER_DISPLAY_DELAY, load);
        });
};

/**
 * Updates the row splits based
 * on the current list of elements.
 */
const updateRowSplits = () => {
    rowSplitCount.value = Math.ceil(elements.value.length / (props.rowSize || 1)) || 1;

    rowIndex.value = Math.min(rowIndex.value, rowSplitCount.value - 1);

    if (!props.editing) {
        nextTick(() => {
            const scrollContainer = container.value?.querySelector(".home-page-row-content") as HTMLElement;
            if (scrollContainer) {
                scrollContainer.scrollLeft = 0;
            }
        });
    }
};

// If row size changes, update the row splits
watch(
    () => props.rowSize,
    () => {
        rowIndex.value = 0;

        // To ensure the scroll is applied fast,
        // disable the animations for 1 tick
        fastTransition.value = true;
        nextTick(() => {
            fastTransition.value = false;
        });

        updateRowSplits();
        updateCurrentMedia();
    },
);

/**
 * Applies initial scroll,
 * necessary when the row is being moved
 */
const applyInitialMovingScroll = () => {
    nextTick(() => {
        const scrollContainer = container.value?.querySelector(".home-page-row-content") as HTMLElement;
        if (scrollContainer) {
            scrollContainer.scrollLeft = props.movingInitialScroll || 0;
        }
    });
};

/**
 * Updates the current media position
 */
const updateCurrentMedia = () => {
    if (!props.isCurrentGroup || props.currentMedia === -1) {
        return;
    }

    let currentMediaPos = -1;

    for (let i = 0; i < elements.value.length; i++) {
        if (elements.value[i].media && elements.value[i].media.id === props.currentMedia) {
            currentMediaPos = i;
            break;
        }
    }

    let firstElement = -1;
    let lastElement = -1;

    let prevElement = -1;
    let nextElement = -1;

    if (currentMediaPos >= 0) {
        rowIndex.value = Math.floor(currentMediaPos / (props.rowSize || 1));

        for (let i = 0; i < elements.value.length; i++) {
            const element = elements.value[i];

            if (firstElement === -1 && element.media) {
                firstElement = element.media.id;
            }

            if (element.media) {
                lastElement = element.media.id;
            }

            if (i < currentMediaPos && element.media) {
                prevElement = element.media.id;
            } else if (nextElement === -1 && i > currentMediaPos && element.media) {
                nextElement = element.media.id;
            }
        }
    }

    emit("updated-prev-next", prevElement, nextElement, firstElement, lastElement);
};

// Update the current media position if the current media changes
watch(
    () => props.currentMedia,
    () => {
        updateCurrentMedia();
    },
);

/**
 * Gets the URL of an element
 * @param element The element
 * @returns The URL
 */
const getElementURL = (element: HomePageElement): string => {
    if (element.media) {
        return getFrontendUrl({
            media: element.media.id,
        });
    } else if (element.album) {
        return getFrontendUrl({
            album: element.album.id,
        });
    } else {
        return "#";
    }
};

/**
 * Gets the title of an element
 * @param element The element
 * @returns The title
 */
const getTitle = (element: HomePageElement): string => {
    if (element.media) {
        return element.media.title || $t("Untitled");
    } else if (element.album) {
        return element.album.name || $t("Untitled");
    } else {
        return "-";
    }
};

/**
 * Checks if an element is the current element
 * @param item The element
 * @param currentMedia The current media ID
 * @param isCurrentGroup True if the group is current
 * @returns True if the element is current
 */
const isCurrent = (item: HomePageElement, currentMedia: number, isCurrentGroup: boolean): boolean => {
    return isCurrentGroup && item.media && item.media.id === currentMedia;
};

/**
 * The user want to navigate to a specific element
 * @param element The element
 * @param i The index of the element
 * @param e The click event (if applicable)
 */
const goToElement = (element: HomePageElement, i: number, e?: Event) => {
    if (e) {
        e.preventDefault();
    }

    if (props.editing) {
        return;
    }

    if (element.media) {
        AppStatus.ClickOnMedia(element.media.id, true, props.group.id);
    } else if (element.album) {
        setHomePageBackState(props.group.id, i);
        AppStatus.ClickOnAlbum(element.album.id);
    }
};

/**
 * Moves the row index to the first element
 */
const goFirst = () => {
    rowIndex.value = 0;
};

/**
 * Moves the row index to the left
 */
const goLeft = () => {
    rowIndex.value = Math.max(0, rowIndex.value - 1);
};

/**
 * Moves the row index to the right
 */
const goRight = () => {
    rowIndex.value = Math.min(rowIndex.value + 1, rowSplitCount.value - 1);
};

/**
 * Moves the row index to the last element
 */
const goLast = () => {
    rowIndex.value = rowSplitCount.value - 1;
};

/**
 * Focuses a specific element,
 * moving the row index to the necessary position
 * @param i The element index
 */
const focusElementIndex = (i: number) => {
    rowIndex.value = Math.floor(i / (props.rowSize || 1));
};

/**
 * Moves the focus to the first element of the row
 */
const moveFocusToFirstRowElement = () => {
    const firstRowElementIndex = rowIndex.value * (props.rowSize || 1);
    const firstRowElement = container.value?.querySelector(".home-page-row-element-" + firstRowElementIndex) as HTMLElement;
    if (firstRowElement) {
        firstRowElement.focus();
    }
};

// True to display the modal to add elements
const displayAddElement = ref(false);

/**
 * Displays the modal to add elements to the row
 */
const addElements = () => {
    displayAddElement.value = true;
};

// True to display the context menu for en element
const contextMenuShown = ref(false);

// Index of the element being selected for the context menu
const contextMenuSelectedIndex = ref(-1);

// X coordinate of the context menu
const contextMenuX = ref(0);

// Y coordinate of the context menu
const contextMenuY = ref(0);

/**
 * Shows the context menu for an element
 * @param i The element index
 * @param event The event (Context event or click event)
 */
const showContextMenu = (i: number, event: MouseEvent) => {
    if (!props.editing) {
        return;
    }

    event.preventDefault();

    if (contextMenuShown.value && i === contextMenuSelectedIndex.value) {
        if (event.type === "click") {
            contextMenuShown.value = false;
        }
        return;
    }

    contextMenuSelectedIndex.value = i;
    contextMenuX.value = event.pageX;
    contextMenuY.value = event.pageY;
    contextMenuShown.value = true;
};

/**
 * The user wants to remove an element
 * @param i The element index
 */
const onElementRemove = (i: number) => {
    const element = elements.value[i];

    if (!element) {
        return;
    }

    elements.value.splice(i, 1);

    doSilentDelete(element);
};

/**
 * Silently removes the element from the row
 * @param element The element to remove
 */
const doSilentDelete = (element: HomePageElement) => {
    doHomePageSilentSaveAction((callback) => {
        makeApiRequest(apiHomeGroupDeleteElement(props.group.id, getHomePageElementReference(element)))
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
                    notCustomGroup: () => {
                        emit("must-reload");
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

// True to display the modal to move an element
const movePositionModalDisplay = ref(false);

// The index of the element being selected for the move modal
const movePositionModalSelectedIndex = ref(-1);

/**
 * The user wants to move an element
 * @param i The element index
 */
const onChangePositionRequest = (i: number) => {
    movePositionModalSelectedIndex.value = i;
    movePositionModalDisplay.value = true;
};

/**
 * The user requested to move an element
 * @param oldPos The old position of the element
 * @param newPos The new position of the element
 */
const onMoveElement = (oldPos: number, newPos: number) => {
    const element = elements.value[oldPos];

    if (!element) {
        return;
    }

    elements.value.splice(newPos, 0, elements.value.splice(oldPos, 1)[0]);

    doSilentMove(element, newPos);
};

/**
 * Silently moves the element by calling the API
 * @param element The element to move
 * @param position The new position for the element
 */
const doSilentMove = (element: HomePageElement, position: number) => {
    doHomePageSilentSaveAction((callback) => {
        makeApiRequest(apiHomeGroupMoveElement(props.group.id, getHomePageElementReference(element), position))
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
                    notCustomGroup: () => {
                        emit("must-reload");
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

/**
 * Event handler for 'dragstart' on the row
 * @param event The drag event
 */
const onDragStartRow = (event: DragEvent) => {
    event.preventDefault();

    if (!props.editing) {
        return;
    }

    if (isTouchDevice() || props.isMobileSize) {
        return;
    }

    const startX = event.pageX;
    const startY = event.pageY;

    const bounds = (container.value as HTMLElement).getBoundingClientRect();

    const scrollContainer = container.value?.querySelector(".home-page-row-content") as HTMLElement;
    const initialScroll = scrollContainer ? scrollContainer.scrollLeft || 0 : 0;

    const data: HomePageGroupStartMovingData = {
        startX,
        startY,
        offsetX: startX - bounds.left,
        offsetY: startY - bounds.top,
        width: bounds.width,
        height: bounds.height,
        initialElements: loading.value ? null : elements.value.slice(),
        initialScroll,
    };

    emit("start-moving", props.group, data);
};

/**
 * Event handler for 'scroll' on the context element
 */
const onContentScroll = () => {
    if (contextMenuShown.value) {
        contextMenuShown.value = false;
    }
    if (!props.editing) {
        const scrollContainer = container.value?.querySelector(".home-page-row-content") as HTMLElement;
        if (scrollContainer) {
            scrollContainer.scrollLeft = 0;
        }
    }
};

/**
 * Event handler for 'keydown'
 * @param event The keyboard event
 */
const onKeyDown = (event: KeyboardEvent) => {
    if (event.key === "ArrowLeft" || event.key === "PageUp") {
        event.preventDefault();

        goLeft();

        moveFocusToFirstRowElement();
    } else if (event.key === "ArrowRight" || event.key === "PageDown") {
        event.preventDefault();

        goRight();

        moveFocusToFirstRowElement();
    } else if (event.key === "Home") {
        event.preventDefault();

        goFirst();

        moveFocusToFirstRowElement();
    } else if (event.key === "End") {
        event.preventDefault();

        goLast();

        const lastRowElement = container.value?.querySelector(".home-page-row-element-" + (elements.value.length - 1)) as HTMLElement;
        if (lastRowElement) {
            lastRowElement.focus();
        }
    }
};

// Interval to check for dragging
const dragCheckInterval = useInterval();

// True if dragging an element
const dragging = ref(false);

// Element being dragged
const draggingElement = ref<HomePageElement | null>(null);

// Dragging data (coordinates)
const draggingData = reactive({
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

// Current X coordinate of the mouse
const mouseX = ref(0);

// Current Y coordinate of the mouse
const mouseY = ref(0);

// Delay to check for dragging (milliseconds)
const DRAG_CHECK_DELAY = 40;

/**
 * Event handler for 'dragstart' on an element
 * @param i The element index
 * @param event The drag event
 */
const onDragStart = (i: number, event: DragEvent) => {
    if (!props.editing) {
        return;
    }

    event.preventDefault();
    event.stopPropagation();

    const element = elements.value[i];

    if (!element) {
        return;
    }

    mouseX.value = event.pageX;
    mouseY.value = event.pageY;

    const elClickable = container.value?.querySelector(".home-page-row-element-" + i) as HTMLElement;

    if (!elClickable) {
        return;
    }

    const target = elClickable.parentElement as HTMLElement;

    if (!target) {
        return;
    }

    const targetBounds = target.getBoundingClientRect();

    const startX = event.pageX;
    const startY = event.pageY;

    const offsetX = startX - targetBounds.left;
    const offsetY = startY - targetBounds.top;

    dragging.value = true;
    draggingElement.value = element;

    draggingData.startX = startX;
    draggingData.startY = startY;

    draggingData.offsetX = offsetX;
    draggingData.offsetY = offsetY;

    draggingData.width = targetBounds.width;
    draggingData.height = targetBounds.height;

    draggingData.x = startX - offsetX;
    draggingData.y = startY - offsetY;

    draggingData.startPosition = i;

    dragCheckInterval.clear();

    setDocumentEventListeners();

    dragCheckInterval.set(onDragCheck, DRAG_CHECK_DELAY);

    updateMovingOver();
};

/**
 * Event handler for 'mousemove' on document
 * @param event The mouse event
 */
const onDocumentMouseMove = (event: MouseEvent) => {
    if (!dragging.value) {
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

    draggingData.x = event.pageX - draggingData.offsetX;
    draggingData.y = event.pageY - draggingData.offsetY;
};

/**
 * Event handler for 'mouseup' on document
 * @param event The mouse event
 */
const onDocumentMouseUp = (event: MouseEvent) => {
    if (!dragging.value) {
        return;
    }

    dragCheckInterval.clear();

    clearDocumentEventListeners();

    event.stopPropagation();

    dragging.value = false;

    if (!draggingElement.value) {
        return;
    }

    let position = draggingData.movingOver > draggingData.startPosition + 1 ? draggingData.movingOver - 1 : draggingData.movingOver;

    position = Math.max(0, Math.min(elements.value.length, position - 1));

    const startPosition = draggingData.startPosition;

    if (startPosition === -1 || position === startPosition) {
        return;
    }

    doSilentMove(draggingElement.value, position);
    elements.value.splice(position, 0, elements.value.splice(startPosition, 1)[0]);
};

// True if listening on document events
let listeningOnDocumentEvents = false;

/**
 * Sets the document event listeners
 */
const setDocumentEventListeners = () => {
    clearDocumentEventListeners();

    document.addEventListener("mousemove", onDocumentMouseMove);
    document.addEventListener("mouseup", onDocumentMouseUp);

    listeningOnDocumentEvents = true;
};

/**
 * Clears the document event listeners
 */
const clearDocumentEventListeners = () => {
    if (listeningOnDocumentEvents) {
        document.removeEventListener("mousemove", onDocumentMouseMove);
        document.removeEventListener("mouseup", onDocumentMouseUp);
        listeningOnDocumentEvents = false;
    }
};

onBeforeUnmount(clearDocumentEventListeners);

/**
 * Updates the element being moved over
 */
const updateMovingOver = () => {
    const rowContentEl = container.value?.querySelector(".home-page-row-content") as HTMLElement;

    if (!rowContentEl) {
        draggingData.movingOver = -1;
        return;
    }

    const containerBounds = rowContentEl.getBoundingClientRect();

    const offsetLeft = containerBounds.left;
    const scrollLeft = rowContentEl.scrollLeft || 0;

    const x = draggingData.x + Math.round(draggingData.width / 2);
    const y = draggingData.y + Math.round(draggingData.height / 2);

    if (
        y + Math.round(draggingData.height / 2) < containerBounds.top ||
        y - Math.round(draggingData.height / 2) > containerBounds.top + containerBounds.height
    ) {
        draggingData.movingOver = -1;
        return;
    }

    const width = draggingData.width;
    const expectedIndex = Math.round((x - offsetLeft + scrollLeft) / width);

    draggingData.movingOver = Math.min(elements.value.length + 1, Math.max(1, 1 + expectedIndex));
};

/**
 * Checks the dragging status
 */
const onDragCheck = () => {
    const rowContentEl = container.value?.querySelector(".home-page-row-content");

    if (!rowContentEl) {
        return;
    }

    const conBounds = rowContentEl.getBoundingClientRect();

    if (mouseX.value >= conBounds.left - draggingData.width) {
        // Auto scroll

        const relTop = (mouseX.value - conBounds.left) / (conBounds.width || 1);
        const scrollStep = Math.ceil(conBounds.width / 25);

        if (relTop <= 0.1) {
            rowContentEl.scrollLeft = Math.max(0, rowContentEl.scrollLeft - scrollStep);
        } else if (relTop >= 0.9) {
            rowContentEl.scrollLeft = Math.min(rowContentEl.scrollWidth - conBounds.width, rowContentEl.scrollLeft + scrollStep);
        }
    }

    updateMovingOver();
};
</script>
