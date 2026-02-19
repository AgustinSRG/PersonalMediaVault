<template>
    <div ref="container" :class="{ 'page-inner': !inModal, 'scrollbar-stable': !inModal, 'page-in-modal': !!inModal }">
        <div class="search-results" tabindex="-1">
            <div class="search-results-options">
                <div class="search-results-option">
                    <input
                        ref="autoFocusElement"
                        v-model="filter"
                        type="text"
                        class="form-control form-control-full-width auto-focus"
                        autocomplete="off"
                        :placeholder="$t('Filter by name') + '...'"
                        @input="changeFilter"
                    />
                </div>
                <div class="search-results-option text-right">
                    <button v-if="canWrite && !inModal" type="button" class="btn btn-primary" @click="createAlbum">
                        <i class="fas fa-plus"></i> {{ $t("Create album") }}
                    </button>
                </div>
            </div>

            <PageMenu
                v-if="total > 0 && order !== 'rand'"
                :page-name="'albums'"
                :order="order"
                :page="page"
                :pages="totalPages"
                :min="min"
                @goto="changePage"
            ></PageMenu>

            <PageLoaderFiller v-if="loading" :page-size="pageSize" :display-titles="displayTitles"></PageLoaderFiller>

            <div v-else-if="total <= 0 && firstLoaded" class="search-results-msg-display">
                <div class="search-results-msg-icon">
                    <i v-if="filter" class="fas fa-search"></i>
                    <i class="fas fa-box-open"></i>
                </div>
                <div class="search-results-msg-text">
                    {{
                        filter
                            ? $t("Could not find any albums matching your filter")
                            : inModal
                              ? $t("Could not find any album")
                              : $t("This vault does not have any albums yet")
                    }}
                </div>
                <div class="search-results-msg-btn">
                    <button type="button" class="btn btn-primary" @click="refreshAlbums">
                        <i class="fas fa-sync-alt"></i> {{ $t("Refresh") }}
                    </button>
                </div>
                <div v-if="filter" class="search-results-msg-btn">
                    <button type="button" class="btn btn-primary" @click="clearFilter">
                        <i class="fas fa-times"></i> {{ $t("Clear filter") }}
                    </button>
                </div>
            </div>

            <div v-else-if="total > 0" class="search-results-final-display">
                <div v-for="item in pageItems" :key="item.id" class="search-result-item">
                    <a
                        class="clickable"
                        :href="getAlbumURL(item.id)"
                        target="_blank"
                        rel="noopener noreferrer"
                        @click="goToAlbum(item, $event)"
                    >
                        <AlbumItemThumbnail :item="item"></AlbumItemThumbnail>
                        <div v-if="displayTitles" class="search-result-title">
                            {{ item.name || $t("Untitled") }}
                        </div></a
                    >
                </div>

                <div v-for="i in lastRowPadding" :key="'pad-last-' + i" class="search-result-item"></div>
            </div>

            <PageMenu
                v-if="total > 0 && order !== 'rand'"
                :page-name="'albums'"
                :page="page"
                :pages="totalPages"
                :min="min"
                @goto="changePage"
            ></PageMenu>

            <div v-if="total > 0 && order !== 'rand'" class="search-results-total">{{ $t("Total") }}: {{ total }}</div>
        </div>

        <AlbumCreateModal v-model:display="displayAlbumCreate" @new-album="onNewAlbum"></AlbumCreateModal>
    </div>
</template>

<script setup lang="ts">
import {
    emitAppEvent,
    EVENT_NAME_ALBUMS_CHANGED,
    EVENT_NAME_APP_STATUS_CHANGED,
    EVENT_NAME_AUTH_CHANGED,
    EVENT_NAME_RANDOM_PAGE_REFRESH,
    EVENT_NAME_UNAUTHORIZED,
} from "@/control/app-events";
import { AppStatus } from "@/control/app-status";
import { getFrontendUrl } from "@/utils/api";
import { makeNamedApiRequest, abortNamedApiRequest } from "@asanrom/request-browser";
import { setNamedTimeout, clearNamedTimeout } from "@/utils/named-timeouts";
import type { PropType } from "vue";
import { computed, nextTick, onMounted, ref, useTemplateRef, watch } from "vue";
import PageMenu from "@/components/pages/common/PageMenu.vue";
import { AuthController } from "@/control/auth";
import AlbumCreateModal from "../modals/AlbumCreateModal.vue";
import { filterToWords, matchSearchFilter, normalizeString } from "@/utils/normalize";
import { packSearchParams, unPackSearchParams } from "@/utils/search-params";
import type { AlbumListItem } from "@/api/models";
import { PagesController } from "@/control/pages";
import { apiAlbumsGetAlbums } from "@/api/api-albums";
import { isTouchDevice } from "@/utils/touch";
import { shuffleArray } from "@/utils/shuffle";
import PageLoaderFiller from "./common/PageLoaderFiller.vue";
import AlbumItemThumbnail from "../utils/AlbumItemThumbnail.vue";
import { useI18n } from "@/composables/use-i18n";
import { usePageLastRowPadding } from "@/composables/use-page-last-row-padding";
import { useUserPermissions } from "@/composables/use-user-permissions";
import { useRequestId } from "@/composables/use-request-id";
import { onApplicationEvent } from "@/composables/on-app-event";
import { useGlobalKeyboardHandler } from "@/composables/use-global-keyboard-handler";

// Ref to the container element
const container = useTemplateRef("container");

// Translation
const { $t } = useI18n();

// User permissions
const { canWrite } = useUserPermissions();

// Props
const props = defineProps({
    /**
     * Page is being displayed in a modal
     */
    inModal: Boolean,

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
     * Set of albums to remove from the list
     */
    removeAlbumsFromList: Object as PropType<Set<number>>,

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

// Emits
const emit = defineEmits<{
    /**
     * The user selected an album
     */
    (e: "select-album", albumId: number, callback: () => void): void;
}>();

// Loading status
const loading = ref(false);

// True if it was loaded at least once
const firstLoaded = ref(false);

// List of albums
const albumsList = ref<AlbumListItem[]>([]);

// Filter for albums
const filter = ref(props.inModal ? "" : PagesController.AlbumsPageSearch);

// Order for albums (type)
type AlbumsOrder = "asc" | "desc" | "rand";

// Order fort albums
const order = ref<AlbumsOrder>("desc");

// Search parameters
const searchParams = ref(props.inModal ? "" : AppStatus.SearchParams);

// Page number
const page = ref(0);

/**
 * Pages the search parameters (page, order)
 */
const updateSearchParams = () => {
    if (props.inModal) {
        return;
    }
    const params = unPackSearchParams(searchParams.value);
    page.value = params.page;
    order.value = params.order;
};

updateSearchParams();

onApplicationEvent(EVENT_NAME_APP_STATUS_CHANGED, () => {
    if (props.inModal) {
        return;
    }
    if (AppStatus.SearchParams !== searchParams.value) {
        searchParams.value = AppStatus.SearchParams;
        updateSearchParams();
        updateList();
    }
});

/**
 * Call when the search parameters must be updated
 */
const onSearchParamsChanged = () => {
    if (props.inModal) {
        return;
    }
    searchParams.value = packSearchParams(page.value, order.value);
    AppStatus.ChangeSearchParams(searchParams.value);
};

// Total number of albums
const total = ref(0);

// Total number of pages
const totalPages = ref(0);

// Items displayed in the current page
const pageItems = ref<AlbumListItem[]>([]);

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

    makeNamedApiRequest(loadRequestId, apiAlbumsGetAlbums())
        .onSuccess((result) => {
            albumsList.value = result;

            clearNamedTimeout(loadRequestId);

            loading.value = false;
            firstLoaded.value = true;

            updateList();
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
onApplicationEvent(EVENT_NAME_ALBUMS_CHANGED, load);

// Reload when page size changes
watch(
    () => props.pageSize,
    () => {
        page.value = 0;
        load();
    },
);

/**
 * Refreshes the albums list
 */
const refreshAlbums = () => {
    load();
};

/**
 * Updates the albums list for the page
 */
const updateList = () => {
    const filterName = normalizeString(filter.value + "")
        .trim()
        .toLowerCase();

    let pageList = albumsList.value.map((a: AlbumListItem) => {
        return {
            id: a.id,
            name: a.name,
            nameLowerCase: a.name.trim().toLowerCase(),
            size: a.size,
            thumbnail: a.thumbnail,
            lm: a.lm,
        };
    });

    if (filterName) {
        const filterWords = filterToWords(filterName);
        pageList = pageList.filter((a) => {
            return matchSearchFilter(a.name, filterName, filterWords) >= 0;
        });
    }

    if (props.removeAlbumsFromList) {
        const blacklist = props.removeAlbumsFromList;
        pageList = pageList.filter((a) => {
            return !blacklist.has(a.id);
        });
    }

    if (order.value === "asc") {
        pageList = pageList.sort((a, b) => {
            if (a.nameLowerCase < b.nameLowerCase) {
                return -1;
            } else {
                return 1;
            }
        });
    } else if (order.value === "rand") {
        pageList = shuffleArray(pageList);
    } else {
        pageList = pageList.sort((a, b) => {
            if (a.lm > b.lm) {
                return -1;
            } else if (b.lm > a.lm) {
                return 1;
            } else if (a.id < b.id) {
                return 1;
            } else {
                return -1;
            }
        });
    }

    total.value = pageList.length;

    const pageSize = Math.max(1, props.pageSize);

    totalPages.value = Math.floor(total.value / pageSize);

    if (total.value % pageSize > 0) {
        totalPages.value++;
    }

    if (order.value === "rand") {
        page.value = 0;
    } else {
        page.value = Math.max(0, Math.min(page.value, totalPages.value - 1));
    }

    const skip = pageSize * page.value;

    pageItems.value = pageList.slice(skip, skip + pageSize);

    onSearchParamsChanged();
};

// Update the list when the page is forced to refresh
// for the random order
onApplicationEvent(EVENT_NAME_RANDOM_PAGE_REFRESH, updateList);

/**
 * Call when the filter changes
 */
const changeFilter = () => {
    if (!props.inModal) {
        PagesController.AlbumsPageSearch = filter.value;
    }
    page.value = 0;
    updateList();
};

/**
 * Clears the filter
 */
const clearFilter = () => {
    filter.value = "";
    changeFilter();
};

/**
 * Changes the page number
 * @param p The page number
 */
const changePage = (p: number) => {
    page.value = p;
    onSearchParamsChanged();
    updateList();
};

/**
 * Gets the URL of an album
 * @param albumId The album ID
 * @returns The URL
 */
const getAlbumURL = (albumId: number): string => {
    return getFrontendUrl({
        album: albumId,
    });
};

/**
 * Navigates to an album
 * @param album The album element
 * @param e The click event on the link
 */
const goToAlbum = (album: AlbumListItem, e: Event) => {
    if (e) {
        e.preventDefault();
    }

    if (props.inModal) {
        emit("select-album", album.id, () => {
            updateList();
        });
        return;
    }

    AppStatus.ClickOnAlbum(album.id);
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
    if (isTouchDevice()) {
        return;
    }
    nextTick(() => {
        autoFocusElement.value?.focus();
        autoFocusElement.value?.select();
    });
};

onMounted(autoFocus);

// Display modal to create an album?
const displayAlbumCreate = ref(false);

/**
 * Opens the modal to create an album
 */
const createAlbum = () => {
    displayAlbumCreate.value = true;
};

/**
 * Called when an album is created using the album creation modal
 * @param albumId The album ID
 */
const onNewAlbum = (albumId: number) => {
    AppStatus.ClickOnAlbum(albumId);
};

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

    if (event.key === "+") {
        createAlbum();
        return true;
    }

    if (event.key.toUpperCase() === "R") {
        refreshAlbums();
        return true;
    }

    return false;
}, KEYBOARD_HANDLER_PRIORITY);
</script>
