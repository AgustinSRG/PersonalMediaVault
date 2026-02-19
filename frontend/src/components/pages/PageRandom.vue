<template>
    <div ref="container" class="page-inner scrollbar-stable">
        <div ref="autoFocusElement" class="search-results auto-focus" tabindex="-1">
            <PageLoaderFiller v-if="loading" :page-size="pageSize" :display-titles="displayTitles"></PageLoaderFiller>

            <div v-else-if="total <= 0 && firstLoaded" class="search-results-msg-display">
                <div class="search-results-msg-icon">
                    <i v-if="search" class="fas fa-search"></i>
                    <i v-else class="fas fa-box-open"></i>
                </div>
                <div class="search-results-msg-text">
                    {{ search ? $t("Could not find any result") : $t("The vault is empty") }}
                </div>
                <div class="search-results-msg-btn">
                    <button type="button" class="btn btn-primary" @click="load"><i class="fas fa-sync-alt"></i> {{ $t("Refresh") }}</button>
                </div>
                <div v-if="search" class="search-results-msg-btn">
                    <button type="button" class="btn btn-primary" @click="clearSearch">
                        <i class="fas fa-times"></i> {{ $t("Clear search") }}
                    </button>
                </div>
                <div v-if="search" class="search-results-msg-btn">
                    <button type="button" class="btn btn-primary" @click="goAdvancedSearch">
                        <i class="fas fa-search"></i> {{ $t("Advanced search") }}
                    </button>
                </div>
            </div>

            <div v-else-if="total > 0" class="search-results-final-display">
                <MediaItem
                    v-for="(item, i) in pageItems"
                    :key="i"
                    :item="item"
                    :current="currentMedia == item.id"
                    :display-titles="displayTitles"
                    @go="goToMedia(item.id)"
                ></MediaItem>

                <div v-for="i in lastRowPadding" :key="'pad-last-' + i" class="search-result-item"></div>
            </div>
        </div>
    </div>
</template>

<script setup lang="ts">
import type { MediaListItem } from "@/api/models";
import {
    emitAppEvent,
    EVENT_NAME_APP_STATUS_CHANGED,
    EVENT_NAME_AUTH_CHANGED,
    EVENT_NAME_MEDIA_DELETE,
    EVENT_NAME_MEDIA_METADATA_CHANGE,
    EVENT_NAME_PAGE_NAV_NEXT,
    EVENT_NAME_PAGE_NAV_PREV,
    EVENT_NAME_RANDOM_PAGE_REFRESH,
    EVENT_NAME_UNAUTHORIZED,
} from "@/control/app-events";
import { AppStatus } from "@/control/app-status";
import { AuthController } from "@/control/auth";
import { TagsController } from "@/control/tags";
import { makeNamedApiRequest, abortNamedApiRequest } from "@asanrom/request-browser";
import { setNamedTimeout, clearNamedTimeout } from "@/utils/named-timeouts";
import { computed, nextTick, onBeforeUnmount, onMounted, ref, useTemplateRef, watch } from "vue";
import { MediaController } from "@/control/media";
import { PagesController } from "@/control/pages";
import { apiSearchRandom } from "@/api/api-search";
import MediaItem from "../utils/MediaItem.vue";
import PageLoaderFiller from "./common/PageLoaderFiller.vue";
import { useI18n } from "@/composables/use-i18n";
import { useRequestId } from "@/composables/use-request-id";
import { usePageLastRowPadding } from "@/composables/use-page-last-row-padding";
import { useGlobalKeyboardHandler } from "@/composables/use-global-keyboard-handler";
import { onApplicationEvent } from "@/composables/on-app-event";

// Ref to the container element
const container = useTemplateRef("container");

// Translation
const { $t } = useI18n();

// Props
const props = defineProps({
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
     * Page is in miniature mode
     */
    min: Boolean,

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
});

// Loading status
const loading = ref(false);

// True if it was loaded at least once
const firstLoaded = ref(false);

// Current media
const currentMedia = ref(AppStatus.CurrentMedia);

// Current search
const search = ref(AppStatus.CurrentSearch);

// Search parameters
const searchParams = ref(AppStatus.SearchParams);

// RNG seed
const seed = ref(AppStatus.RandomSeed);

// Total number of items
const total = ref(0);

// Items displayed in the current page
const pageItems = ref<MediaListItem[]>([]);

// Action to perform when page loads
const switchMediaOnLoad = ref<"" | "next" | "prev">("");

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

    makeNamedApiRequest(loadRequestId, apiSearchRandom(search.value, seed.value, props.pageSize))
        .onSuccess((result) => {
            const s = new Set();

            pageItems.value = result.page_items.filter((i) => {
                if (s.has(i.id)) {
                    return false;
                }
                s.add(i.id);
                return true;
            });

            TagsController.OnMediaListReceived(pageItems.value);

            total.value = pageItems.value.length;

            clearNamedTimeout(loadRequestId);

            loading.value = false;
            firstLoaded.value = true;

            if (switchMediaOnLoad.value === "next") {
                switchMediaOnLoad.value = "";
                if (pageItems.value.length > 0) {
                    if (MediaController.MediaId === pageItems.value[0].id) {
                        MediaController.Load();
                    } else {
                        goToMedia(pageItems.value[0].id);
                    }
                }
            } else if (switchMediaOnLoad.value === "prev") {
                switchMediaOnLoad.value = "";
                if (pageItems.value.length > 0) {
                    if (MediaController.MediaId === pageItems.value[pageItems.value.length - 1].id) {
                        MediaController.Load();
                    } else {
                        goToMedia(pageItems.value[pageItems.value.length - 1].id);
                    }
                }
            }
            scrollToCurrentMedia();
            onCurrentMediaChanged();
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

// Reload when page size changes
watch(
    () => props.pageSize,
    () => {
        load();
    },
);

onApplicationEvent(EVENT_NAME_APP_STATUS_CHANGED, () => {
    const changed = currentMedia.value !== AppStatus.CurrentMedia;
    currentMedia.value = AppStatus.CurrentMedia;

    let mustLoad = false;

    if (AppStatus.CurrentSearch !== search.value) {
        search.value = AppStatus.CurrentSearch;
        mustLoad = true;
    }

    if (AppStatus.SearchParams !== searchParams.value || AppStatus.RandomSeed !== seed.value) {
        seed.value = AppStatus.RandomSeed;
        searchParams.value = AppStatus.SearchParams;
        mustLoad = true;
    }

    if (mustLoad) {
        load();
    }

    if (changed) {
        scrollToCurrentMedia();
    }

    onCurrentMediaChanged();
});

/**
 * Navigates to a media element
 * @param mid The ID of the media
 */
const goToMedia = (mid: number) => {
    AppStatus.ClickOnMedia(mid, true);
};

/**
 * Fins the index of the current media
 * @returns The index, or -1 (if not found)
 */
const findCurrentMediaIndex = (): number => {
    for (let i = 0; i < pageItems.value.length; i++) {
        if (pageItems.value[i].id === currentMedia.value) {
            return i;
        }
    }
    return -1;
};

/**
 * Navigates to the previous media
 */
const prevMedia = () => {
    const i = findCurrentMediaIndex();
    if (i !== -1 && i > 0) {
        goToMedia(pageItems.value[i - 1].id);
    } else if (i === -1 && pageItems.value.length > 0) {
        goToMedia(pageItems.value[0].id);
    } else {
        switchMediaOnLoad.value = "prev";
        refreshSeed();
    }
};

onApplicationEvent(EVENT_NAME_PAGE_NAV_PREV, prevMedia);

/**
 * Navigates to the next media
 */
const nextMedia = () => {
    const i = findCurrentMediaIndex();
    if (i !== -1 && i < pageItems.value.length - 1) {
        goToMedia(pageItems.value[i + 1].id);
    } else if (i === -1 && pageItems.value.length > 0) {
        goToMedia(pageItems.value[0].id);
    } else {
        switchMediaOnLoad.value = "next";
        refreshSeed();
    }
};

onApplicationEvent(EVENT_NAME_PAGE_NAV_NEXT, nextMedia);

/**
 * Call to update the current media position in the page context
 */
const onCurrentMediaChanged = () => {
    const i = findCurrentMediaIndex();
    PagesController.OnPageLoad(i, pageItems.value.length, 1, 3);
};

// Make sure to unload the page context when the component unmounts
onBeforeUnmount(() => {
    PagesController.OnPageUnload();
});

/**
 * Clears the search
 */
const clearSearch = () => {
    AppStatus.ClearSearch();
};

/**
 * Navigates to advanced search
 */
const goAdvancedSearch = () => {
    AppStatus.GoToPage("search");
};

/**
 * Refreshes the seed
 */
const refreshSeed = () => {
    AppStatus.RefreshSeed();
};

onApplicationEvent(EVENT_NAME_RANDOM_PAGE_REFRESH, refreshSeed);

/**
 * Scrolls the page to the top
 */
const scrollToTop = () => {
    if (container.value) {
        container.value.scrollTop = 0;
    }
};

/**
 * Scrolls to the current media element
 */
const scrollToCurrentMedia = () => {
    nextTick(() => {
        const currentElem = container.value?.querySelector(".search-result-item.current") as HTMLElement;
        if (currentElem) {
            currentElem.scrollIntoView();
        }
    });
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

// Padding for the last row
const lastRowPadding = usePageLastRowPadding(
    props,
    container,
    computed(() => pageItems.value.length),
);

// Priority for the global keyboard handler
const KEYBOARD_HANDLER_PRIORITY = 20;

// Global keyboard handler
useGlobalKeyboardHandler((event: KeyboardEvent): boolean => {
    if (AuthController.Locked || !AppStatus.IsPageVisible() || !event.key || event.ctrlKey) {
        return false;
    }

    if (event.key === "Home") {
        if (pageItems.value.length > 0) {
            goToMedia(pageItems.value[0].id);
        }
        return true;
    }

    if (event.key === "End") {
        if (pageItems.value.length > 0) {
            goToMedia(pageItems.value[pageItems.value.length - 1].id);
        }
        return true;
    }

    if (event.key === "ArrowLeft") {
        prevMedia();
        return true;
    }

    if (event.key === "ArrowRight") {
        nextMedia();
        return true;
    }

    if (event.key.toUpperCase() === "R") {
        refreshSeed();
        return true;
    }

    return false;
}, KEYBOARD_HANDLER_PRIORITY);
</script>
