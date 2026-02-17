<template>
    <div class="page-inner scrollbar-stable" :class="{ hidden: !display }">
        <div class="search-results auto-focus" tabindex="-1">
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

<script lang="ts">
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
import { defineComponent, nextTick } from "vue";
import { MediaController } from "@/control/media";
import { packSearchParams, unPackSearchParams } from "@/utils/search-params";
import { PagesController } from "@/control/pages";
import { getUniqueStringId } from "@/utils/unique-id";
import { apiSearchRandom } from "@/api/api-search";
import MediaItem from "../utils/MediaItem.vue";
import PageLoaderFiller from "./common/PageLoaderFiller.vue";

export default defineComponent({
    name: "PageRandom",
    components: {
        MediaItem,
        PageLoaderFiller,
    },
    props: {
        display: Boolean,
        min: Boolean,
        pageSize: Number,
        displayTitles: Boolean,

        rowSize: Number,
        rowSizeMin: Number,
        minItemsSize: Number,
        maxItemsSize: Number,
    },
    setup() {
        return {
            loadRequestId: getUniqueStringId(),
            windowResizeObserver: null as ResizeObserver,
        };
    },
    data: function () {
        return {
            loading: false,
            firstLoaded: false,

            search: AppStatus.CurrentSearch,

            order: "desc",
            searchParams: AppStatus.SearchParams,
            seed: AppStatus.RandomSeed,
            page: 0,

            currentMedia: AppStatus.CurrentMedia,

            pageItems: [] as MediaListItem[],
            total: 0,

            loadingFiller: [] as number[],

            switchMediaOnLoad: "",

            tagVersion: TagsController.TagsVersion,

            windowWidth: 0,
        };
    },
    computed: {
        lastRowPadding() {
            const containerWidth = this.windowWidth;

            const itemWidth = Math.max(
                this.minItemsSize,
                Math.min(
                    this.maxItemsSize,
                    this.min ? containerWidth / Math.max(1, this.rowSizeMin) : containerWidth / Math.max(1, this.rowSize),
                ),
            );

            const elementsFitInRow = Math.max(1, Math.floor(containerWidth / Math.max(1, itemWidth)));

            return Math.max(0, elementsFitInRow - (this.pageItems.length % elementsFitInRow));
        },
    },
    watch: {
        display: function () {
            this.load();
            if (this.display) {
                this.autoFocus();
            }
        },
        pageSize: function () {
            this.updatePageSize();
        },
    },
    mounted: function () {
        this.$addKeyboardHandler(this.handleGlobalKey.bind(this), 20);

        this.$listenOnAppEvent(EVENT_NAME_AUTH_CHANGED, this.load.bind(this));

        this.$listenOnAppEvent(EVENT_NAME_MEDIA_METADATA_CHANGE, this.load.bind(this));
        this.$listenOnAppEvent(EVENT_NAME_MEDIA_DELETE, this.load.bind(this));
        this.$listenOnAppEvent(EVENT_NAME_APP_STATUS_CHANGED, this.onAppStatusChanged.bind(this));

        this.$listenOnAppEvent(EVENT_NAME_PAGE_NAV_NEXT, this.nextMedia.bind(this));

        this.$listenOnAppEvent(EVENT_NAME_PAGE_NAV_PREV, this.prevMedia.bind(this));

        this.$listenOnAppEvent(EVENT_NAME_RANDOM_PAGE_REFRESH, this.refreshSeed.bind(this));

        this.updateSearchParams();
        this.load();

        if (this.display) {
            this.autoFocus();
        }

        this.windowWidth = this.$el.getBoundingClientRect().width;

        this.windowResizeObserver = new ResizeObserver(this.updateWindowWidth.bind(this));
        this.windowResizeObserver.observe(this.$el);
    },
    beforeUnmount: function () {
        clearNamedTimeout(this.loadRequestId);
        abortNamedApiRequest(this.loadRequestId);
        PagesController.OnPageUnload();
        this.windowResizeObserver.disconnect();
    },
    methods: {
        scrollToTop: function () {
            this.$el.scrollTop = 0;
        },

        autoFocus: function () {
            nextTick(() => {
                const el = this.$el.querySelector(".auto-focus");
                if (el) {
                    el.focus();
                    if (el.select) {
                        el.select();
                    }
                }
            });
        },

        refreshSeed: function () {
            AppStatus.RefreshSeed();
        },

        load: function () {
            clearNamedTimeout(this.loadRequestId);
            abortNamedApiRequest(this.loadRequestId);

            if (!this.display) {
                return;
            }

            this.scrollToTop();

            setNamedTimeout(this.loadRequestId, 330, () => {
                this.loading = true;
            });

            if (AuthController.Locked) {
                return; // Vault is locked
            }

            makeNamedApiRequest(this.loadRequestId, apiSearchRandom(this.search, this.seed, this.pageSize))
                .onSuccess((result) => {
                    const s = new Set();
                    this.pageItems = result.page_items.filter((i) => {
                        if (s.has(i.id)) {
                            return false;
                        }
                        s.add(i.id);
                        return true;
                    });
                    TagsController.OnMediaListReceived(this.pageItems);
                    this.total = this.pageItems.length;
                    clearNamedTimeout(this.loadRequestId);
                    this.loading = false;
                    this.firstLoaded = true;
                    if (this.switchMediaOnLoad === "next") {
                        this.switchMediaOnLoad = "";
                        if (this.pageItems.length > 0) {
                            if (MediaController.MediaId === this.pageItems[0].id) {
                                MediaController.Load();
                            } else {
                                this.goToMedia(this.pageItems[0].id);
                            }
                        }
                    } else if (this.switchMediaOnLoad === "prev") {
                        this.switchMediaOnLoad = "";
                        if (this.pageItems.length > 0) {
                            if (MediaController.MediaId === this.pageItems[this.pageItems.length - 1].id) {
                                MediaController.Load();
                            } else {
                                this.goToMedia(this.pageItems[this.pageItems.length - 1].id);
                            }
                        }
                    }
                    this.scrollToCurrentMedia();
                    this.onCurrentMediaChanged();
                })
                .onRequestError((err, handleErr) => {
                    handleErr(err, {
                        unauthorized: () => {
                            emitAppEvent(EVENT_NAME_UNAUTHORIZED);
                        },
                        temporalError: () => {
                            // Retry
                            this.loading = true;
                            setNamedTimeout(this.loadRequestId, 1500, this.load.bind(this));
                        },
                    });
                })
                .onUnexpectedError((err) => {
                    console.error(err);
                    // Retry
                    this.loading = true;
                    setNamedTimeout(this.loadRequestId, 1500, this.load.bind(this));
                });
        },

        updatePageSize: function () {
            this.updateLoadingFiller();
            this.load();
        },

        onAppStatusChanged: function () {
            const changed = this.currentMedia !== AppStatus.CurrentMedia;
            this.currentMedia = AppStatus.CurrentMedia;

            let mustLoad = false;

            if (AppStatus.CurrentSearch !== this.search) {
                this.search = AppStatus.CurrentSearch;
                mustLoad = true;
            }

            if (AppStatus.SearchParams !== this.searchParams || AppStatus.RandomSeed !== this.seed) {
                this.seed = AppStatus.RandomSeed;
                this.searchParams = AppStatus.SearchParams;
                this.updateSearchParams();
                mustLoad = true;
            }

            if (mustLoad) {
                this.load();
            }

            if (changed) {
                this.scrollToCurrentMedia();
            }
            this.onCurrentMediaChanged();
        },

        scrollToCurrentMedia: function () {
            nextTick(() => {
                const currentElem = this.$el.querySelector(".search-result-item.current");
                if (currentElem) {
                    currentElem.scrollIntoView();
                }
            });
        },

        onSearchParamsChanged: function () {
            this.searchParams = packSearchParams(this.page, this.order);
            AppStatus.ChangeSearchParams(this.searchParams);
        },

        goToMedia: function (mid: number) {
            AppStatus.ClickOnMedia(mid, true);
        },

        updateSearchParams: function () {
            const params = unPackSearchParams(this.searchParams);
            this.page = params.page;
            this.order = params.order;
            this.updateLoadingFiller();
        },

        updateLoadingFiller: function () {
            const filler = [];

            for (let i = 0; i < this.pageSize; i++) {
                filler.push(i);
            }

            this.loadingFiller = filler;
        },

        clearSearch: function () {
            AppStatus.ClearSearch();
        },

        goAdvancedSearch: function () {
            AppStatus.GoToPage("search");
        },

        findCurrentMediaIndex: function (): number {
            for (let i = 0; i < this.pageItems.length; i++) {
                if (this.pageItems[i].id === this.currentMedia) {
                    return i;
                }
            }
            return -1;
        },

        onCurrentMediaChanged: function () {
            const i = this.findCurrentMediaIndex();
            PagesController.OnPageLoad(i, this.pageItems.length, 1, 3);
        },

        prevMedia: function () {
            const i = this.findCurrentMediaIndex();
            if (i !== -1 && i > 0) {
                this.goToMedia(this.pageItems[i - 1].id);
            } else if (i === -1 && this.pageItems.length > 0) {
                this.goToMedia(this.pageItems[0].id);
            } else {
                this.switchMediaOnLoad = "prev";
                this.refreshSeed();
            }
        },

        nextMedia: function () {
            const i = this.findCurrentMediaIndex();
            if (i !== -1 && i < this.pageItems.length - 1) {
                this.goToMedia(this.pageItems[i + 1].id);
            } else if (i === -1 && this.pageItems.length > 0) {
                this.goToMedia(this.pageItems[0].id);
            } else {
                this.switchMediaOnLoad = "next";
                this.refreshSeed();
            }
        },

        handleGlobalKey: function (event: KeyboardEvent): boolean {
            if (AuthController.Locked || !AppStatus.IsPageVisible() || !this.display || !event.key || event.ctrlKey) {
                return false;
            }

            if (event.key === "Home") {
                if (this.pageItems.length > 0) {
                    this.goToMedia(this.pageItems[0].id);
                }
                return true;
            }

            if (event.key === "End") {
                if (this.pageItems.length > 0) {
                    this.goToMedia(this.pageItems[this.pageItems.length - 1].id);
                }
                return true;
            }

            if (event.key === "ArrowLeft") {
                this.prevMedia();
                return true;
            }

            if (event.key === "ArrowRight") {
                this.nextMedia();
                return true;
            }

            if (event.key.toUpperCase() === "R") {
                this.refreshSeed();
                return true;
            }

            return false;
        },

        updateWindowWidth: function () {
            this.windowWidth = this.$el.getBoundingClientRect().width;
        },
    },
});
</script>
