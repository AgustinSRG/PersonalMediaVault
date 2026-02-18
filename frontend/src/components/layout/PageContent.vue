<template>
    <div
        class="page-content"
        tabindex="-1"
        :class="{ 'rounded-corners-cells': roundedCorners }"
        :style="{
            '--row-size': rowSize,
            '--row-size-min': rowSizeMin,
            '--min-cell-size': minItemSize + 'px',
            '--max-cell-size': maxItemSize + 'px',
            '--cell-padding': padding + 'px',
        }"
        @keydown="onKeyDown"
    >
        <div class="page-header">
            <button type="button" :title="$t('Expand')" class="page-header-btn page-expand-btn" @click="expandPage">
                <i class="fas fa-chevron-left"></i>
            </button>
            <div class="page-title" :title="title"><i :class="icon"></i> {{ title }}</div>
            <button
                v-if="(page === 'home' || page === 'media') && canWrite"
                type="button"
                :title="$t('Upload')"
                class="page-header-btn"
                @click="uploadMedia"
            >
                <i class="fas fa-upload"></i>
            </button>
            <button
                v-if="page === 'home' && !homeEditMode && canWrite && !min"
                type="button"
                :title="$t('Edit home page')"
                class="page-header-btn"
                @click="homeStartEdit"
            >
                <i class="fas fa-pencil-alt"></i>
            </button>
            <button
                v-if="page === 'home' && homeEditMode && canWrite"
                type="button"
                :title="$t('Finalize home page modifications')"
                class="page-header-btn"
                @click="homeFinishEdit"
            >
                <LoadingIcon icon="fas fa-check" :loading="savingHome"></LoadingIcon>
            </button>
            <button
                v-if="page === 'random' || (pageHasOrderAlbums && order === 'rand')"
                type="button"
                :title="$t('Refresh')"
                class="page-header-btn"
                @click="triggerRefresh"
            >
                <i class="fas fa-sync-alt"></i>
            </button>
            <button
                v-if="pageHasOrderDate && order === 'desc'"
                type="button"
                :title="$t('Show oldest')"
                class="page-header-btn"
                @click="toggleOrder"
            >
                <i class="fas fa-arrow-up-short-wide"></i>
            </button>
            <button
                v-if="pageHasOrderDate && order !== 'desc'"
                type="button"
                :title="$t('Show most recent')"
                class="page-header-btn"
                @click="toggleOrder"
            >
                <i class="fas fa-arrow-down-short-wide"></i>
            </button>
            <button
                v-if="pageHasOrderAlbums && order === 'desc'"
                type="button"
                :title="$t('Order alphabetically')"
                class="page-header-btn"
                @click="toggleOrder"
            >
                <i class="fas fa-arrow-down-a-z"></i>
            </button>
            <button
                v-else-if="pageHasOrderAlbums && order === 'asc'"
                type="button"
                :title="$t('Random')"
                class="page-header-btn"
                @click="toggleOrder"
            >
                <i class="fas fa-shuffle"></i>
            </button>
            <button
                v-else-if="pageHasOrderAlbums"
                type="button"
                :title="$t('Order by last modified date')"
                class="page-header-btn"
                @click="toggleOrder"
            >
                <i class="fas fa-arrow-down-short-wide"></i>
            </button>
            <button
                v-if="page === 'search' && pageScroll > 0"
                type="button"
                :title="$t('Go to the top')"
                class="page-header-btn"
                @click="goToTop"
            >
                <i class="fas fa-angles-up"></i>
            </button>
            <button v-if="pageHasConfigOptions" type="button" :title="$t('Page settings')" class="page-header-btn" @click="openConfig">
                <i class="fas fa-sliders"></i>
            </button>
            <button type="button" :title="$t('Close')" class="page-header-btn page-close-btn" @click="closePage">
                <i class="fas fa-times"></i>
            </button>
        </div>

        <PageHome
            v-if="isDisplayed && page === 'home'"
            :display="isDisplayed && page === 'home'"
            :min="min"
            :page-size="pageSize"
            :display-titles="displayTitles"
            :row-size="rowSize"
            :row-size-min="rowSizeMin"
            :min-items-size="minItemSize"
            :max-items-size="maxItemSize"
            :editing="homeEditMode"
        ></PageHome>
        <PageMedia
            v-if="isDisplayed && page === 'media'"
            :display="isDisplayed && page === 'media'"
            :min="min"
            :page-size="pageSize"
            :display-titles="displayTitles"
            :row-size="rowSize"
            :row-size-min="rowSizeMin"
            :min-items-size="minItemSize"
            :max-items-size="maxItemSize"
        ></PageMedia>
        <PageUpload v-if="isDisplayed && page === 'upload'" :display="isDisplayed && page === 'upload'"></PageUpload>
        <PageRandom
            v-if="isDisplayed && page === 'random'"
            :display="isDisplayed && page === 'random'"
            :min="min"
            :page-size="pageSize"
            :display-titles="displayTitles"
            :row-size="rowSize"
            :row-size-min="rowSizeMin"
            :min-items-size="minItemSize"
            :max-items-size="maxItemSize"
        ></PageRandom>
        <PageSearch
            v-if="isDisplayed && page === 'search'"
            v-model:page-scroll="pageScroll"
            :display="isDisplayed && page === 'search'"
            :min="min"
            :in-modal="false"
            :no-album="-1"
            :page-size="pageSize"
            :display-titles="displayTitles"
            :row-size="rowSize"
            :row-size-min="rowSizeMin"
            :min-items-size="minItemSize"
            :max-items-size="maxItemSize"
        ></PageSearch>
        <PageAlbums
            v-if="isDisplayed && page === 'albums'"
            :display="isDisplayed && page === 'albums'"
            :min="min"
            :page-size="pageSize"
            :display-titles="displayTitles"
            :row-size="rowSize"
            :row-size-min="rowSizeMin"
            :min-items-size="minItemSize"
            :max-items-size="maxItemSize"
        ></PageAlbums>

        <PageSettingsDropdown v-if="displayConfigModal" v-model:display="displayConfigModal" :page="page"></PageSettingsDropdown>
    </div>
</template>

<script setup lang="ts">
import {
    emitAppEvent,
    EVENT_NAME_ADVANCED_SEARCH_GO_TOP,
    EVENT_NAME_APP_STATUS_CHANGED,
    EVENT_NAME_RANDOM_PAGE_REFRESH,
} from "@/control/app-events";
import { AppStatus } from "@/control/app-status";
import { computed, defineAsyncComponent, nextTick, ref, watch } from "vue";
import { AuthController } from "@/control/auth";
import LoadingOverlay from "./LoadingOverlay.vue";
import LoadingIcon from "../utils/LoadingIcon.vue";
import { packSearchParams, unPackSearchParams } from "@/utils/search-params";
import { waitForHomePageSilentSaveActions } from "@/utils/home";
import { usePagePreferences } from "@/composables/use-page-preferences";
import { useI18n } from "@/composables/use-i18n";
import { useUserPermissions } from "@/composables/use-user-permissions";
import { onApplicationEvent } from "@/composables/on-app-event";
import { useGlobalKeyboardHandler } from "@/composables/use-global-keyboard-handler";

const PageHome = defineAsyncComponent({
    loader: () => import("@/components/pages/PageHome.vue"),
    loadingComponent: LoadingOverlay,
    delay: 200,
});

const PageMedia = defineAsyncComponent({
    loader: () => import("@/components/pages/PageMedia.vue"),
    loadingComponent: LoadingOverlay,
    delay: 200,
});

const PageUpload = defineAsyncComponent({
    loader: () => import("@/components/pages/PageUpload.vue"),
    loadingComponent: LoadingOverlay,
    delay: 200,
});

const PageRandom = defineAsyncComponent({
    loader: () => import("@/components/pages/PageRandom.vue"),
    loadingComponent: LoadingOverlay,
    delay: 200,
});

const PageAlbums = defineAsyncComponent({
    loader: () => import("@/components/pages/PageAlbums.vue"),
    loadingComponent: LoadingOverlay,
    delay: 200,
});

const PageSearch = defineAsyncComponent({
    loader: () => import("@/components/pages/PageSearch.vue"),
    loadingComponent: LoadingOverlay,
    delay: 200,
});

const PageSettingsDropdown = defineAsyncComponent({
    loader: () => import("@/components/dropdowns/PageSettingsDropdown.vue"),
    loadingComponent: LoadingOverlay,
    delay: 1000,
});

// Translation function
const { $t } = useI18n();

// Props
defineProps({
    /**
     * Minimized style
     */
    min: Boolean,
});

// True if the page container is being displayed
const isDisplayed = ref((AppStatus.CurrentMedia < 0 || AppStatus.ListSplitMode) && AppStatus.CurrentAlbum < 0);

// Current page
const page = ref(AppStatus.CurrentPage);

// Current search
const search = ref(AppStatus.CurrentSearch);

// Current search params
const searchParams = ref(AppStatus.SearchParams);

// Current page number
const pageN = ref(1);

// Current page order
const order = ref<"desc" | "asc" | "rand">("desc");

/**
 * Updates the search parameters
 */
const updateSearchParams = () => {
    const params = unPackSearchParams(searchParams.value);
    pageN.value = params.page;
    order.value = params.order;
};

updateSearchParams();

onApplicationEvent(EVENT_NAME_APP_STATUS_CHANGED, () => {
    if (page.value !== AppStatus.CurrentPage && AppStatus.CurrentPage === "home") {
        homeEditMode.value = false;
    }

    page.value = AppStatus.CurrentPage;
    search.value = AppStatus.CurrentSearch;
    isDisplayed.value = (AppStatus.CurrentMedia < 0 || AppStatus.ListSplitMode) && AppStatus.CurrentAlbum < 0;

    searchParams.value = AppStatus.SearchParams;
    updateSearchParams();
});

// Title of the page
const title = computed<string>(() => {
    switch (page.value) {
        case "home":
            return $t("Home");
        case "media":
            return $t("Media") + (search.value ? " (" + $t("Tag") + ": " + search.value + ")" : "");
        case "search":
            return $t("Find media");
        case "upload":
            return $t("Upload media");
        case "albums":
            return $t("Albums list");
        case "random":
            return $t("Random media") + (search.value ? " (" + $t("Tag") + ": " + search.value + ")" : "");
        default:
            return "";
    }
});

const icon = computed(() => {
    switch (page.value) {
        case "home":
            return "fas fa-home";
        case "media":
            return "fas fa-photo-film";
        case "search":
            return "fas fa-search";
        case "upload":
            return "fas fa-upload";
        case "albums":
            return "fas fa-list";
        case "random":
            return "fas fa-shuffle";
        default:
            return "";
    }
});

// True if the page has config options, to display the button
const pageHasConfigOptions = computed<boolean>(() => {
    switch (page.value) {
        case "home":
        case "media":
        case "search":
        case "albums":
        case "random":
            return true;
        default:
            return false;
    }
});

// True of the page can be order by date
const pageHasOrderDate = computed<boolean>(() => {
    switch (page.value) {
        case "media":
            return true;
        default:
            return false;
    }
});

// True if the page has the ordering of albums
const pageHasOrderAlbums = computed<boolean>(() => {
    switch (page.value) {
        case "albums":
            return true;
        default:
            return false;
    }
});

// Page preferences
const { pageSize, rowSize, rowSizeMin, minItemSize, maxItemSize, padding, displayTitles, roundedCorners } = usePagePreferences(page, true);

// Edit mode for home page
const homeEditMode = ref(false);

// True if home page is being saved
const savingHome = ref(false);

// User permissions
const { canWrite } = useUserPermissions();

watch(canWrite, () => {
    if (!canWrite.value) {
        homeEditMode.value = false;
    }
});

/**
 * Expands the page
 */
const expandPage = () => {
    AppStatus.ExpandPage();
};

/**
 * Closes the page and focus the player (if any)
 */
const closePage = () => {
    AppStatus.ClosePage();
    const player: any = document.querySelector(".player-container");
    if (player) {
        player.focus();
    }
};

// Display the page configuration modal?
const displayConfigModal = ref(false);

/**
 * Opens the page configuration modal
 */
const openConfig = () => {
    displayConfigModal.value = !displayConfigModal.value;
};

/**
 * Call this when making any changes to the search parameters,
 * in order to update the application status
 */
const onSearchParamsChanged = () => {
    searchParams.value = packSearchParams(pageN.value, order.value);
    AppStatus.ChangeSearchParams(searchParams.value);
};

/**
 * Toggles the page ordering direction
 */
const toggleOrder = () => {
    if (order.value === "desc") {
        order.value = "asc";
    } else if (order.value === "asc") {
        order.value = pageHasOrderAlbums.value ? "rand" : "desc";
    } else {
        order.value = "desc";
    }

    onSearchParamsChanged();
};

/**
 * Triggers the refresh of the random results, changing the seed
 */
const triggerRefresh = () => {
    emitAppEvent(EVENT_NAME_RANDOM_PAGE_REFRESH);

    nextTick(() => {
        const elementToFocus: any = document.querySelector(".page-content .search-results");
        if (elementToFocus) {
            elementToFocus.focus();
        }
    });
};

/**
 * Starts editing the home page
 */
const homeStartEdit = () => {
    homeEditMode.value = true;
};

/**
 * Finish editing the home page
 */
const homeFinishEdit = () => {
    savingHome.value = true;

    waitForHomePageSilentSaveActions(() => {
        savingHome.value = false;
        homeEditMode.value = false;
    });
};

/**
 * Navigates to the upload page
 */
const uploadMedia = () => {
    AppStatus.GoToPageConditionalSplit("upload");
};

// Current page scroll
const pageScroll = ref(0);

/**
 * Goes to the top of the scrolled section of the page
 */
const goToTop = () => {
    emitAppEvent(EVENT_NAME_ADVANCED_SEARCH_GO_TOP);
};

/**
 * Event handler for 'keydown'
 * @param e The keyboard event
 */
const onKeyDown = (e: KeyboardEvent) => {
    if (e.key === "ArrowDown" || e.key === "ArrowUp") {
        e.stopPropagation();
    }
};

// Priority for the global keyboard handler
const KEYBOARD_HANDLER_PRIORITY = 10;

// Global keyboard handler
useGlobalKeyboardHandler((event: KeyboardEvent): boolean => {
    if (AuthController.Locked || !AppStatus.IsPageVisible() || !event.key || event.ctrlKey) {
        return false;
    }

    if (event.key.toUpperCase() === "Q") {
        closePage();
        return true;
    }

    if (event.key.toUpperCase() === "BACKSPACE") {
        expandPage();
        return true;
    }

    return false;
}, KEYBOARD_HANDLER_PRIORITY);
</script>
