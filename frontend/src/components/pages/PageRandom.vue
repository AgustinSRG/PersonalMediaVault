<template>
    <div class="page-inner" :class="{ hidden: !display }">
        <div class="search-results auto-focus" tabindex="-1">
            <div v-if="loading" class="search-results-loading-display">
                <div v-for="f in loadingFiller" :key="f" class="search-result-item">
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

            <div v-if="!loading && total <= 0 && !search && firstLoaded" class="search-results-msg-display">
                <div class="search-results-msg-icon">
                    <i class="fas fa-box-open"></i>
                </div>
                <div class="search-results-msg-text">
                    {{ $t("The vault is empty") }}
                </div>
                <div class="search-results-msg-btn">
                    <button type="button" class="btn btn-primary" @click="load"><i class="fas fa-sync-alt"></i> {{ $t("Refresh") }}</button>
                </div>
            </div>

            <div v-if="!loading && total <= 0 && search && firstLoaded" class="search-results-msg-display">
                <div class="search-results-msg-icon"><i class="fas fa-search"></i></div>
                <div class="search-results-msg-text">
                    {{ $t("Could not find any result") }}
                </div>
                <div class="search-results-msg-btn">
                    <button type="button" class="btn btn-primary" @click="load"><i class="fas fa-sync-alt"></i> {{ $t("Refresh") }}</button>
                </div>
                <div class="search-results-msg-btn">
                    <button type="button" class="btn btn-primary" @click="clearSearch">
                        <i class="fas fa-times"></i> {{ $t("Clear search") }}
                    </button>
                </div>
                <div class="search-results-msg-btn">
                    <button type="button" class="btn btn-primary" @click="goAdvancedSearch">
                        <i class="fas fa-search"></i> {{ $t("Advanced search") }}
                    </button>
                </div>
            </div>

            <div v-if="!loading && total > 0" class="search-results-final-display">
                <div v-for="(item, i) in pageItems" :key="i" class="search-result-item" :class="{ current: currentMedia == item.id }">
                    <a
                        class="clickable"
                        :href="getMediaURL(item.id)"
                        target="_blank"
                        rel="noopener noreferrer"
                        @click="goToMedia(item.id, $event)"
                    >
                        <div class="search-result-thumb" :title="renderHintTitle(item, tagVersion)">
                            <div class="search-result-thumb-inner">
                                <div v-if="!item.thumbnail" class="no-thumb">
                                    <i v-if="item.type === 1" class="fas fa-image"></i>
                                    <i v-else-if="item.type === 2" class="fas fa-video"></i>
                                    <i v-else-if="item.type === 3" class="fas fa-headphones"></i>
                                    <i v-else class="fas fa-ban"></i>
                                </div>
                                <ThumbImage v-if="item.thumbnail" :src="getThumbnail(item.thumbnail)"></ThumbImage>
                                <DurationIndicator
                                    v-if="item.type === 2 || item.type === 3"
                                    :type="item.type"
                                    :duration="item.duration"
                                ></DurationIndicator>
                            </div>
                        </div>
                        <div v-if="displayTitles" class="search-result-title">
                            {{ item.title || $t("Untitled") }}
                        </div>
                    </a>
                </div>

                <div v-for="i in lastRowPadding" :key="'pad-last-' + i" class="search-result-item"></div>
            </div>
        </div>
    </div>
</template>

<script lang="ts">
import type { MediaListItem } from "@/api/models";
import { AppEvents } from "@/control/app-events";
import { AppStatus, EVENT_NAME_APP_STATUS_CHANGED } from "@/control/app-status";
import { AuthController, EVENT_NAME_AUTH_CHANGED, EVENT_NAME_UNAUTHORIZED } from "@/control/auth";
import { EVENT_NAME_TAGS_UPDATE, TagsController } from "@/control/tags";
import { getAssetURL, getFrontendUrl } from "@/utils/api";
import { makeNamedApiRequest, abortNamedApiRequest } from "@asanrom/request-browser";
import { setNamedTimeout, clearNamedTimeout } from "@/utils/named-timeouts";
import { defineComponent, nextTick } from "vue";
import { MediaController } from "@/control/media";
import { packSearchParams, unPackSearchParams } from "@/utils/search-params";
import {
    EVENT_NAME_MEDIA_DELETE,
    EVENT_NAME_MEDIA_METADATA_CHANGE,
    EVENT_NAME_PAGE_NAV_NEXT,
    EVENT_NAME_PAGE_NAV_PREV,
    EVENT_NAME_RANDOM_PAGE_REFRESH,
    PagesController,
} from "@/control/pages";
import { getUniqueStringId } from "@/utils/unique-id";
import { apiSearchRandom } from "@/api/api-search";
import ThumbImage from "../utils/ThumbImage.vue";
import DurationIndicator from "../utils/DurationIndicator.vue";

export default defineComponent({
    name: "PageRandom",
    components: {
        ThumbImage,
        DurationIndicator,
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

        this.$listenOnAppEvent(EVENT_NAME_TAGS_UPDATE, this.updateTagData.bind(this));

        this.$listenOnAppEvent(EVENT_NAME_RANDOM_PAGE_REFRESH, this.refreshSeed.bind(this));

        this.updateSearchParams();
        this.updateTagData();
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
                            AppEvents.Emit(EVENT_NAME_UNAUTHORIZED);
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

        goToMedia: function (mid: number, e?: Event) {
            if (e) {
                e.preventDefault();
            }
            AppStatus.ClickOnMedia(mid, true);
        },

        getMediaURL: function (mid: number): string {
            return getFrontendUrl({
                media: mid,
            });
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

        getThumbnail(thumb: string) {
            return getAssetURL(thumb);
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

        renderHintTitle(item: MediaListItem, tagVersion: number): string {
            const parts = [item.title || this.$t("Untitled")];

            if (item.tags.length > 0) {
                const tagNames = [];

                for (const tag of item.tags) {
                    tagNames.push(TagsController.GetTagName(tag, tagVersion));
                }

                parts.push(this.$t("Tags") + ": " + tagNames.join(", "));
            }

            return parts.join("\n");
        },

        updateTagData: function () {
            this.tagVersion = TagsController.TagsVersion;
        },

        updateWindowWidth: function () {
            this.windowWidth = this.$el.getBoundingClientRect().width;
        },
    },
});
</script>
