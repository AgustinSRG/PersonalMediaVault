<template>
    <div
        class="page-content"
        tabindex="-1"
        :class="{
            'items-fit-many': pageItemsFit <= 0,
            'items-size-small': pageItemsSize === 'small',
            'items-size-big': pageItemsSize === 'big',
            'items-size-normal': pageItemsSize !== 'small' && pageItemsSize !== 'big',
        }"
        :style="{ '--page-items-fit': pageItemsFit }"
    >
        <div class="page-header">
            <button type="button" :title="$t('Expand')" class="page-header-btn page-expand-btn" @click="expandPage">
                <i class="fas fa-chevron-left"></i>
            </button>
            <div
                class="page-title"
                :title="renderTitle(page, search)"
                :class="{
                    'kind-home': page === 'home' || page === 'search',
                    'kind-albums': page === 'albums',
                    'kind-random': page === 'random',
                    'kind-upload': page === 'upload',
                }"
            >
                <i :class="getIcon(page)"></i> {{ renderTitle(page, search) }}
            </div>
            <button v-if="page === 'random'" type="button" :title="$t('Refresh')" class="page-header-btn" @click="triggerRefresh">
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
                v-if="hasOrderAlbums(page) && order === 'desc'"
                type="button"
                :title="$t('Order alphabetically')"
                class="page-header-btn"
                @click="toggleOrder"
            >
                <i class="fas fa-arrow-down-a-z"></i>
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
                v-if="hasOrderAlbums(page) && order !== 'desc'"
                type="button"
                :title="$t('Order by last modified date')"
                class="page-header-btn"
                @click="toggleOrder"
            >
                <i class="fas fa-arrow-down-short-wide"></i>
            </button>
            <button v-if="hasConfigOptions(page)" type="button" :title="$t('Configuration')" class="page-header-btn" @click="openConfig">
                <i class="fas fa-cog"></i>
            </button>
            <button type="button" :title="$t('Close')" class="page-header-btn page-close-btn" @click="closePage">
                <i class="fas fa-times"></i>
            </button>
        </div>

        <PageHome v-if="isDisplayed && page === 'home'" :display="isDisplayed && page === 'home'" :min="min"></PageHome>
        <PageSearch v-if="isDisplayed && page === 'search'" :display="isDisplayed && page === 'search'" :min="min"></PageSearch>
        <PageUpload v-if="isDisplayed && page === 'upload'" :display="isDisplayed && page === 'upload'"></PageUpload>
        <PageRandom v-if="isDisplayed && page === 'random'" :display="isDisplayed && page === 'random'"></PageRandom>
        <PageAdvancedSearch
            v-if="isDisplayed && page === 'adv-search'"
            :display="isDisplayed && page === 'adv-search'"
            :inModal="false"
            :noAlbum="-1"
        ></PageAdvancedSearch>
        <PageAlbums v-if="isDisplayed && page === 'albums'" :display="isDisplayed && page === 'albums'" :min="min"></PageAlbums>

        <PageConfigModal v-if="displayConfigModal" v-model:display="displayConfigModal"></PageConfigModal>
    </div>
</template>

<script lang="ts">
import { AppEvents } from "@/control/app-events";
import { AppStatus } from "@/control/app-status";
import { defineAsyncComponent, defineComponent } from "vue";

import { AuthController } from "@/control/auth";
import { KeyboardManager } from "@/control/keyboard";

import LoadingOverlay from "./LoadingOverlay.vue";
import { AppPreferences } from "@/control/app-preferences";

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

const PageConfigModal = defineAsyncComponent({
    loader: () => import("@/components/modals/PageConfigModal.vue"),
    loadingComponent: LoadingOverlay,
    delay: 1000,
});

export default defineComponent({
    components: {
        PageHome,
        PageSearch,
        PageAlbums,
        PageUpload,
        PageRandom,
        PageAdvancedSearch,
        PageConfigModal,
    },
    name: "PageContent",
    emits: [],
    props: {
        min: Boolean,
    },
    data: function () {
        return {
            isDisplayed: (AppStatus.CurrentMedia < 0 || AppStatus.ListSplitMode) && AppStatus.CurrentAlbum < 0,
            page: AppStatus.CurrentPage,
            search: AppStatus.CurrentSearch,

            pageN: 1,
            order: "desc",
            searchParams: AppStatus.SearchParams,

            displayConfigModal: false,

            pageItemsFit: AppPreferences.PageItemsFit,
            pageItemsSize: AppPreferences.PageItemsSize,
        };
    },
    methods: {
        updatePage: function () {
            this.page = AppStatus.CurrentPage;
            this.search = AppStatus.CurrentSearch;
            this.isDisplayed = (AppStatus.CurrentMedia < 0 || AppStatus.ListSplitMode) && AppStatus.CurrentAlbum < 0;

            this.searchParams = AppStatus.SearchParams;
            this.updateSearchParams();
        },

        expandPage: function () {
            AppStatus.ExpandPage();
            this.$el.focus();
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
            const params = AppStatus.UnPackSearchParams(this.searchParams);
            this.pageN = params.page;
            this.order = params.order;
        },

        onSearchParamsChanged: function () {
            this.searchParams = AppStatus.PackSearchParams(this.pageN, this.order);
            AppStatus.ChangeSearchParams(this.searchParams);
        },

        triggerRefresh: function () {
            AppEvents.Emit("random-page-refresh");
        },

        openConfig: function () {
            this.displayConfigModal = true;
        },

        toggleOrder: function () {
            if (this.order === "desc") {
                this.order = "asc";
            } else {
                this.order = "desc";
            }
            this.onSearchParamsChanged();
        },

        updatePageItemsPreferences: function () {
            this.pageItemsFit = AppPreferences.PageItemsFit;
            this.pageItemsSize = AppPreferences.PageItemsSize;
        },
    },
    mounted: function () {
        this.$options.pageUpdater = this.updatePage.bind(this);
        AppEvents.AddEventListener("app-status-update", this.$options.pageUpdater);

        this.$options.updatePageItemsPreferencesH = this.updatePageItemsPreferences.bind(this);
        AppEvents.AddEventListener("page-items-pref-updated", this.$options.updatePageItemsPreferencesH);

        this.$options.handleGlobalKeyH = this.handleGlobalKey.bind(this);
        KeyboardManager.AddHandler(this.$options.handleGlobalKeyH, 10);

        this.updateSearchParams();
    },
    beforeUnmount: function () {
        AppEvents.RemoveEventListener("app-status-update", this.$options.pageUpdater);
        AppEvents.RemoveEventListener("page-items-pref-updated", this.$options.updatePageItemsPreferencesH);
        KeyboardManager.RemoveHandler(this.$options.handleGlobalKeyH);
    },
});
</script>
