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
        @keydown="onKeyPress"
    >
        <div class="page-header">
            <button type="button" :title="$t('Expand')" class="page-header-btn page-expand-btn" @click="expandPage">
                <i class="fas fa-chevron-left"></i>
            </button>
            <div class="page-title" :title="renderTitle(page, search)"><i :class="getIcon(page)"></i> {{ renderTitle(page, search) }}</div>
            <button
                v-if="page === 'random' || (hasOrderAlbums(page) && order === 'rand')"
                type="button"
                :title="$t('Refresh')"
                class="page-header-btn"
                @click="triggerRefresh"
            >
                <i class="fas fa-sync-alt"></i>
            </button>
            <button
                v-if="hasOrderDate(page) && order === 'desc'"
                type="button"
                :title="$t('Show oldest')"
                class="page-header-btn"
                @click="toggleOrder"
            >
                <i class="fas fa-arrow-up-short-wide"></i>
            </button>
            <button
                v-if="hasOrderDate(page) && order !== 'desc'"
                type="button"
                :title="$t('Show most recent')"
                class="page-header-btn"
                @click="toggleOrder"
            >
                <i class="fas fa-arrow-down-short-wide"></i>
            </button>
            <button
                v-if="hasOrderAlbums(page) && order === 'desc'"
                type="button"
                :title="$t('Order alphabetically')"
                class="page-header-btn"
                @click="toggleOrder"
            >
                <i class="fas fa-arrow-down-a-z"></i>
            </button>
            <button
                v-else-if="hasOrderAlbums(page) && order === 'asc'"
                type="button"
                :title="$t('Random')"
                class="page-header-btn"
                @click="toggleOrder"
            >
                <i class="fas fa-shuffle"></i>
            </button>
            <button
                v-else-if="hasOrderAlbums(page)"
                type="button"
                :title="$t('Order by last modified date')"
                class="page-header-btn"
                @click="toggleOrder"
            >
                <i class="fas fa-arrow-down-short-wide"></i>
            </button>
            <button
                v-if="page === 'adv-search' && pageScroll > 0"
                type="button"
                :title="$t('Go to the top')"
                class="page-header-btn"
                @click="goToTop"
            >
                <i class="fas fa-angles-up"></i>
            </button>
            <button v-if="hasConfigOptions(page)" type="button" :title="$t('Page settings')" class="page-header-btn" @click="openConfig">
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
        ></PageHome>
        <PageSearch
            v-if="isDisplayed && page === 'search'"
            :display="isDisplayed && page === 'search'"
            :min="min"
            :page-size="pageSize"
            :display-titles="displayTitles"
            :row-size="rowSize"
            :row-size-min="rowSizeMin"
            :min-items-size="minItemSize"
            :max-items-size="maxItemSize"
        ></PageSearch>
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
        <PageAdvancedSearch
            v-if="isDisplayed && page === 'adv-search'"
            v-model:page-scroll="pageScroll"
            :display="isDisplayed && page === 'adv-search'"
            :min="min"
            :in-modal="false"
            :no-album="-1"
            :page-size="pageSize"
            :display-titles="displayTitles"
            :row-size="rowSize"
            :row-size-min="rowSizeMin"
            :min-items-size="minItemSize"
            :max-items-size="maxItemSize"
        ></PageAdvancedSearch>
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

<script lang="ts">
import { AppEvents } from "@/control/app-events";
import { AppStatus, EVENT_NAME_APP_STATUS_CHANGED } from "@/control/app-status";
import { defineAsyncComponent, defineComponent, nextTick } from "vue";

import { AuthController } from "@/control/auth";

import LoadingOverlay from "./LoadingOverlay.vue";
import { packSearchParams, unPackSearchParams } from "@/utils/search-params";
import { EVENT_NAME_ADVANCED_SEARCH_GO_TOP, EVENT_NAME_RANDOM_PAGE_REFRESH } from "@/control/pages";
import { EVENT_NAME_PAGE_PREFERENCES_UPDATED, getPagePreferences } from "@/control/app-preferences";

const PageHome = defineAsyncComponent({
    loader: () => import("@/components/pages/PageHome.vue"),
    loadingComponent: LoadingOverlay,
    delay: 200,
});

const PageSearch = defineAsyncComponent({
    loader: () => import("@/components/pages/PageSearch.vue"),
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

const PageAdvancedSearch = defineAsyncComponent({
    loader: () => import("@/components/pages/PageAdvancedSearch.vue"),
    loadingComponent: LoadingOverlay,
    delay: 200,
});

const PageSettingsDropdown = defineAsyncComponent({
    loader: () => import("@/components/dropdowns/PageSettingsDropdown.vue"),
    loadingComponent: LoadingOverlay,
    delay: 1000,
});

export default defineComponent({
    name: "PageContent",
    components: {
        PageHome,
        PageSearch,
        PageAlbums,
        PageUpload,
        PageRandom,
        PageAdvancedSearch,
        PageSettingsDropdown,
    },
    props: {
        min: Boolean,
    },
    emits: [],
    data: function () {
        const pagePreferences = getPagePreferences(AppStatus.CurrentPage);
        return {
            isDisplayed: (AppStatus.CurrentMedia < 0 || AppStatus.ListSplitMode) && AppStatus.CurrentAlbum < 0,
            page: AppStatus.CurrentPage,
            search: AppStatus.CurrentSearch,

            pageN: 1,
            order: "desc",
            searchParams: AppStatus.SearchParams,

            displayConfigModal: false,

            pageSize: pagePreferences.pageSize,

            rowSize: pagePreferences.rowSize,
            rowSizeMin: pagePreferences.rowSizeMin,

            minItemSize: pagePreferences.minItemSize,
            maxItemSize: pagePreferences.maxItemSize,

            padding: pagePreferences.padding,

            displayTitles: pagePreferences.displayTitles,
            roundedCorners: pagePreferences.roundedCorners,

            pageScroll: 0,
        };
    },
    mounted: function () {
        this.$listenOnAppEvent(EVENT_NAME_APP_STATUS_CHANGED, this.updatePage.bind(this));

        this.$listenOnAppEvent(EVENT_NAME_PAGE_PREFERENCES_UPDATED, this.updatePagePreferences.bind(this));

        this.$addKeyboardHandler(this.handleGlobalKey.bind(this), 10);

        this.updateSearchParams();
    },
    methods: {
        updatePage: function () {
            this.page = AppStatus.CurrentPage;
            this.search = AppStatus.CurrentSearch;
            this.isDisplayed = (AppStatus.CurrentMedia < 0 || AppStatus.ListSplitMode) && AppStatus.CurrentAlbum < 0;

            this.searchParams = AppStatus.SearchParams;
            this.updateSearchParams();
            this.updatePagePreferences();
        },

        expandPage: function () {
            AppStatus.ExpandPage();
        },

        closePage: function () {
            AppStatus.ClosePage();
            const player: any = document.querySelector(".player-container");
            if (player) {
                player.focus();
            }
        },

        hasConfigOptions: function (p: string): boolean {
            switch (p) {
                case "home":
                case "search":
                case "adv-search":
                case "albums":
                case "random":
                    return true;
                default:
                    return false;
            }
        },

        hasOrderDate: function (p: string): boolean {
            switch (p) {
                case "home":
                case "search":
                    return true;
                default:
                    return false;
            }
        },

        hasOrderAlbums: function (p: string): boolean {
            switch (p) {
                case "albums":
                    return true;
                default:
                    return false;
            }
        },

        renderTitle: function (p, s) {
            switch (p) {
                case "home":
                    return this.$t("Home");
                case "search":
                    return this.$t("Search results") + ": " + s;
                case "adv-search":
                    return this.$t("Advanced search");
                case "upload":
                    return this.$t("Upload media");
                case "albums":
                    return this.$t("Albums list");
                case "random":
                    return this.$t("Random results");
                default:
                    return "";
            }
        },

        getIcon: function (p) {
            switch (p) {
                case "home":
                    return "fas fa-home";
                case "search":
                case "adv-search":
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
        },

        handleGlobalKey: function (event: KeyboardEvent): boolean {
            if (AuthController.Locked || !AppStatus.IsPageVisible() || !event.key || event.ctrlKey) {
                return false;
            }

            if (event.key.toUpperCase() === "Q") {
                this.closePage();
                return true;
            }

            if (event.key.toUpperCase() === "BACKSPACE") {
                this.expandPage();
                return true;
            }

            return false;
        },

        updateSearchParams: function () {
            const params = unPackSearchParams(this.searchParams);
            this.pageN = params.page;
            this.order = params.order;
        },

        onSearchParamsChanged: function () {
            this.searchParams = packSearchParams(this.pageN, this.order);
            AppStatus.ChangeSearchParams(this.searchParams);
        },

        triggerRefresh: function () {
            AppEvents.Emit(EVENT_NAME_RANDOM_PAGE_REFRESH);
            nextTick(() => {
                const elementToFocus: any = document.querySelector(".page-content .search-results");
                if (elementToFocus) {
                    elementToFocus.focus();
                }
            });
        },

        openConfig: function () {
            this.displayConfigModal = !this.displayConfigModal;
        },

        toggleOrder: function () {
            if (this.order === "desc") {
                this.order = "asc";
            } else if (this.order === "asc") {
                this.order = this.hasOrderAlbums(this.page) ? "rand" : "desc";
            } else {
                this.order = "desc";
            }
            this.onSearchParamsChanged();
        },

        updatePagePreferences: function () {
            const pagePreferences = getPagePreferences(AppStatus.CurrentPage);

            this.pageSize = pagePreferences.pageSize;

            this.rowSize = pagePreferences.rowSize;
            this.rowSizeMin = pagePreferences.rowSizeMin;

            this.minItemSize = pagePreferences.minItemSize;
            this.maxItemSize = pagePreferences.maxItemSize;

            this.padding = pagePreferences.padding;

            this.displayTitles = pagePreferences.displayTitles;
            this.roundedCorners = pagePreferences.roundedCorners;
        },

        goToTop: function () {
            AppEvents.Emit(EVENT_NAME_ADVANCED_SEARCH_GO_TOP);
        },

        onKeyPress: function (e: KeyboardEvent) {
            if (e.key === "ArrowDown" || e.key === "ArrowUp") {
                e.stopPropagation();
            }
        },
    },
});
</script>
