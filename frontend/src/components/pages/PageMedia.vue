<template>
    <div ref="container" class="page-inner scrollbar-stable">
        <div ref="autoFocusElement" class="search-results auto-focus" tabindex="-1">
            <PageMenu
                v-if="total > 0"
                :page-name="'media'"
                :order="order"
                :page="page"
                :pages="totalPages"
                :min="min"
                @goto="changePage"
            ></PageMenu>

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

            <PageMenu v-if="total > 0" :page-name="'media'" :page="page" :pages="totalPages" :min="min" @goto="changePage"></PageMenu>

            <div v-if="total > 0" class="search-results-total">{{ $t("Total") }}: {{ total }}</div>
        </div>
    </div>
</template>

<script setup lang="ts">
import {
    emitAppEvent,
    EVENT_NAME_APP_STATUS_CHANGED,
    EVENT_NAME_AUTH_CHANGED,
    EVENT_NAME_MEDIA_DELETE,
    EVENT_NAME_MEDIA_METADATA_CHANGE,
    EVENT_NAME_PAGE_NAV_NEXT,
    EVENT_NAME_PAGE_NAV_PREV,
    EVENT_NAME_UNAUTHORIZED,
} from "@/control/app-events";
import { AppStatus } from "@/control/app-status";
import { AuthController } from "@/control/auth";
import { makeNamedApiRequest, abortNamedApiRequest } from "@asanrom/request-browser";
import { setNamedTimeout, clearNamedTimeout } from "@/utils/named-timeouts";
import { computed, nextTick, onBeforeUnmount, onMounted, ref, useTemplateRef, watch } from "vue";
import PageMenu from "@/components/pages/common/PageMenu.vue";
import type { MediaListItem } from "@/api/models";
import { TagsController } from "@/control/tags";
import { orderSimple, packSearchParams, unPackSearchParams } from "@/utils/search-params";
import { PagesController } from "@/control/pages";
import { apiSearch } from "@/api/api-search";
import MediaItem from "../utils/MediaItem.vue";
import PageLoaderFiller from "./common/PageLoaderFiller.vue";
import { useI18n } from "@/composables/use-i18n";
import { onApplicationEvent } from "@/composables/on-app-event";
import { usePageLastRowPadding } from "@/composables/use-page-last-row-padding";
import { useRequestId } from "@/composables/use-request-id";
import { useGlobalKeyboardHandler } from "@/composables/use-global-keyboard-handler";

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

// Page order (type)
type PageOrder = "asc" | "desc";

// Page order
const order = ref<PageOrder>("desc");

// Page number
const page = ref(0);

/**
 * Updates the search parameters (page, order)
 */
const updateSearchParams = () => {
    const params = unPackSearchParams(searchParams.value);
    page.value = params.page;
    order.value = orderSimple(params.order);
};

updateSearchParams();

/**
 * Call when the search params are changed
 * in order to update them in the application status
 */
const onSearchParamsChanged = () => {
    searchParams.value = packSearchParams(page.value, order.value);
    AppStatus.ChangeSearchParams(searchParams.value);
};

// Total number of items
const total = ref(0);

// Total number of pages
const totalPages = ref(0);

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

    makeNamedApiRequest(loadRequestId, apiSearch(search.value || "", order.value, page.value, props.pageSize))
        .onSuccess((result) => {
            TagsController.OnMediaListReceived(result.page_items);

            pageItems.value = result.page_items;
            page.value = result.page_index;
            totalPages.value = result.page_count;
            total.value = result.total_count;

            clearNamedTimeout(loadRequestId);

            loading.value = false;

            firstLoaded.value = true;

            if (switchMediaOnLoad.value === "next") {
                switchMediaOnLoad.value = "";

                if (pageItems.value.length > 0) {
                    goToMedia(pageItems.value[0].id);
                }
            } else if (switchMediaOnLoad.value === "prev") {
                switchMediaOnLoad.value = "";

                if (pageItems.value.length > 0) {
                    goToMedia(pageItems.value[pageItems.value.length - 1].id);
                }
            }

            if (page.value < 0) {
                page.value = 0;
                load();
                return;
            } else if (page.value >= totalPages.value && totalPages.value > 0) {
                page.value = totalPages.value - 1;
                load();
                return;
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
        page.value = 0;
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

    if (AppStatus.SearchParams !== searchParams.value) {
        searchParams.value = AppStatus.SearchParams;
        updateSearchParams();
        load();
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
 * Changes the page number
 * @param p The page number
 */
const changePage = (p: number) => {
    page.value = p;
    onSearchParamsChanged();
    load();
};

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
    } else if (i === 0) {
        if (page.value > 0) {
            switchMediaOnLoad.value = "prev";
            changePage(page.value - 1);
        }
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
    } else if (i === pageItems.value.length - 1) {
        if (page.value < totalPages.value - 1) {
            switchMediaOnLoad.value = "next";
            changePage(page.value + 1);
        }
    }
};

onApplicationEvent(EVENT_NAME_PAGE_NAV_NEXT, nextMedia);

/**
 * Call to update the current media position in the page context
 */
const onCurrentMediaChanged = () => {
    const i = findCurrentMediaIndex();
    PagesController.OnPageLoad(i, pageItems.value.length, page.value, totalPages.value);
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
 * Scrolls the page to the top
 */
const scrollToTop = () => {
    if (container.value) {
        container.value.scrollTop = 0;
    }
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

// Padding for the last row
const { lastRowPadding } = usePageLastRowPadding(
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

    if (event.key === "PageUp" || (event.key === "ArrowLeft" && AppStatus.CurrentMedia < 0)) {
        if (page.value > 0) {
            changePage(page.value - 1);
        }
        return true;
    }

    if (event.key === "PageDown" || (event.key === "ArrowRight" && AppStatus.CurrentMedia < 0)) {
        if (page.value < totalPages.value - 1) {
            changePage(page.value + 1);
        }
        return true;
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

    return false;
}, KEYBOARD_HANDLER_PRIORITY);
</script>
