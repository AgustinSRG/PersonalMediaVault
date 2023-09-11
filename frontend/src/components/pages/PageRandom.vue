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
                    <div class="search-result-title">{{ $t("Loading") }}...</div>
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
                    <button type="button" @click="load" class="btn btn-primary"><i class="fas fa-sync-alt"></i> {{ $t("Refresh") }}</button>
                </div>
            </div>

            <div v-if="!loading && total <= 0 && search && firstLoaded" class="search-results-msg-display">
                <div class="search-results-msg-icon"><i class="fas fa-search"></i></div>
                <div class="search-results-msg-text">
                    {{ $t("Could not find any result") }}
                </div>
                <div class="search-results-msg-btn">
                    <button type="button" @click="load" class="btn btn-primary"><i class="fas fa-sync-alt"></i> {{ $t("Refresh") }}</button>
                </div>
                <div class="search-results-msg-btn">
                    <button type="button" @click="clearSearch" class="btn btn-primary">
                        <i class="fas fa-times"></i> {{ $t("Clear search") }}
                    </button>
                </div>
                <div class="search-results-msg-btn">
                    <button type="button" @click="goAdvancedSearch" class="btn btn-primary">
                        <i class="fas fa-search"></i> {{ $t("Advanced search") }}
                    </button>
                </div>
            </div>

            <div v-if="!loading && total > 0" class="search-results-final-display">
                <a
                    v-for="(item, i) in pageItems"
                    :key="i"
                    class="search-result-item clickable"
                    :class="{ current: currentMedia == item.id }"
                    @click="goToMedia(item.id, $event)"
                    :href="getMediaURL(item.id)"
                    target="_blank"
                    rel="noopener noreferrer"
                >
                    <div class="search-result-thumb" :title="renderHintTitle(item, tagData)">
                        <div class="search-result-thumb-inner">
                            <div v-if="!item.thumbnail" class="no-thumb">
                                <i v-if="item.type === 1" class="fas fa-image"></i>
                                <i v-else-if="item.type === 2" class="fas fa-video"></i>
                                <i v-else-if="item.type === 3" class="fas fa-headphones"></i>
                                <i v-else class="fas fa-ban"></i>
                            </div>
                            <img
                                v-if="item.thumbnail"
                                :src="getThumbnail(item.thumbnail)"
                                :alt="item.title || $t('Untitled')"
                                loading="lazy"
                            />
                            <div class="search-result-thumb-tag" v-if="item.type === 2 || item.type === 3">
                                {{ renderTime(item.duration) }}
                            </div>
                        </div>
                    </div>
                    <div class="search-result-title">
                        {{ item.title || $t("Untitled") }}
                    </div>
                </a>
            </div>
        </div>
    </div>
</template>

<script lang="ts">
import { MediaListItem } from "@/api/models";
import { SearchAPI } from "@/api/api-search";
import { AlbumsController } from "@/control/albums";
import { AppEvents } from "@/control/app-events";
import { AppPreferences } from "@/control/app-preferences";
import { AppStatus } from "@/control/app-status";
import { AuthController } from "@/control/auth";
import { KeyboardManager } from "@/control/keyboard";
import { TagEntry, TagsController } from "@/control/tags";
import { clone } from "@/utils/objects";
import { GenerateURIQuery, GetAssetURL, Request } from "@/utils/request";
import { renderTimeSeconds } from "@/utils/time";
import { Timeouts } from "@/utils/timeout";
import { defineComponent, nextTick } from "vue";

export default defineComponent({
    name: "PageRandom",
    props: {
        display: Boolean,
    },
    data: function () {
        return {
            loading: false,
            firstLoaded: false,

            search: AppStatus.CurrentSearch,

            pageSize: AppPreferences.PageMaxItems,
            order: "desc",
            searchParams: AppStatus.SearchParams,

            currentMedia: AppStatus.CurrentMedia,

            pageItems: [],
            total: 0,

            loadingFiller: [],

            pageSizeOptions: [],

            switchMediaOnLoad: "",

            tagData: {},
        };
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

        load: function () {
            Timeouts.Abort("page-random-load");
            Request.Abort("page-random-load");

            if (!this.display) {
                return;
            }

            this.scrollToTop();

            Timeouts.Set("page-random-load", 330, () => {
                this.loading = true;
            });

            if (AuthController.Locked) {
                return; // Vault is locked
            }

            Request.Pending("page-random-load", SearchAPI.Random(this.search, Date.now(), this.pageSize))
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
                    Timeouts.Abort("page-random-load");
                    this.loading = false;
                    this.firstLoaded = true;
                    if (this.switchMediaOnLoad === "next") {
                        this.switchMediaOnLoad = "";
                        if (this.pageItems.length > 0) {
                            this.goToMedia(this.pageItems[0].id);
                        }
                    } else if (this.switchMediaOnLoad === "prev") {
                        this.switchMediaOnLoad = "";
                        if (this.pageItems.length > 0) {
                            this.goToMedia(this.pageItems[this.pageItems.length - 1].id);
                        }
                    }
                    nextTick(() => {
                        const currentElem = this.$el.querySelector(".search-result-item.current");
                        if (currentElem) {
                            currentElem.scrollIntoView();
                        }
                    });
                    this.onCurrentMediaChanged();
                })
                .onRequestError((err) => {
                    Request.ErrorHandler()
                        .add(401, "*", () => {
                            AppEvents.Emit("unauthorized", false);
                        })
                        .add("*", "*", () => {
                            // Retry
                            this.loading = true;
                            Timeouts.Set("page-random-load", 1500, this._handles.loadH);
                        })
                        .handle(err);
                })
                .onUnexpectedError((err) => {
                    console.error(err);
                    // Retry
                    this.loading = true;
                    Timeouts.Set("page-random-load", 1500, this._handles.loadH);
                });
        },

        updatePageSize: function () {
            this.pageSize = AppPreferences.PageMaxItems;
            this.updateLoadingFiller();
            this.load();
        },

        onAppStatusChanged: function () {
            this.currentMedia = AppStatus.CurrentMedia;
            if (AppStatus.CurrentSearch !== this.search) {
                this.search = AppStatus.CurrentSearch;
                this.load();
            }

            if (AppStatus.SearchParams !== this.searchParams) {
                this.searchParams = AppStatus.SearchParams;
                this.updateSearchParams();
                this.load();
            }

            nextTick(() => {
                const currentElem = this.$el.querySelector(".search-result-item.current");
                if (currentElem) {
                    currentElem.scrollIntoView();
                }
            });
            this.onCurrentMediaChanged();
        },

        onSearchParamsChanged: function () {
            this.searchParams = AppStatus.PackSearchParams(this.page, this.order);
            AppStatus.ChangeSearchParams(this.searchParams);
        },

        goToMedia: function (mid, e) {
            if (e) {
                e.preventDefault();
            }
            AppStatus.ClickOnMedia(mid, true);
        },

        getMediaURL: function (mid: number): string {
            return (
                window.location.protocol +
                "//" +
                window.location.host +
                window.location.pathname +
                GenerateURIQuery({
                    media: mid + "",
                })
            );
        },

        updateSearchParams: function () {
            const params = AppStatus.UnPackSearchParams(this.searchParams);
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
            return GetAssetURL(thumb);
        },

        renderTime: function (s: number): string {
            return renderTimeSeconds(s);
        },

        clickOnEnter: function (event) {
            if (event.key === "Enter") {
                event.preventDefault();
                event.stopPropagation();
                event.target.click();
            }
        },

        clearSearch: function () {
            AppStatus.ClearSearch();
        },

        goAdvancedSearch: function () {
            AppStatus.GoToPage("adv-search");
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
            AlbumsController.OnPageLoad(i, this.pageItems.length, 1, 3);
        },

        prevMedia: function () {
            const i = this.findCurrentMediaIndex();
            if (i !== -1 && i > 0) {
                this.goToMedia(this.pageItems[i - 1].id);
            } else if (i === -1 && this.pageItems.length > 0) {
                this.goToMedia(this.pageItems[0].id);
            } else {
                this.switchMediaOnLoad = "prev";
                this.load();
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
                this.load();
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
                this.load();
                return true;
            }

            return false;
        },

        renderHintTitle(item: MediaListItem, tags: { [id: string]: TagEntry }): string {
            const parts = [item.title || this.$t("Untitled")];

            if (item.tags.length > 0) {
                const tagNames = [];

                for (const tag of item.tags) {
                    if (tags[tag + ""]) {
                        tagNames.push(tags[tag + ""].name);
                    } else {
                        tagNames.push("???");
                    }
                }

                parts.push(this.$t("Tags") + ": " + tagNames.join(", "));
            }

            return parts.join("\n");
        },

        updateTagData: function () {
            this.tagData = clone(TagsController.Tags);
        },
    },
    mounted: function () {
        this._handles = Object.create(null);
        this._handles.loadH = this.load.bind(this);
        this._handles.statusChangeH = this.onAppStatusChanged.bind(this);

        this._handles.handleGlobalKeyH = this.handleGlobalKey.bind(this);
        KeyboardManager.AddHandler(this._handles.handleGlobalKeyH, 20);

        AppEvents.AddEventListener("auth-status-changed", this._handles.loadH);
        AppEvents.AddEventListener("media-meta-change", this._handles.loadH);
        AppEvents.AddEventListener("media-delete", this._handles.loadH);
        AppEvents.AddEventListener("app-status-update", this._handles.statusChangeH);

        this._handles.nextMediaH = this.nextMedia.bind(this);
        AppEvents.AddEventListener("page-media-nav-next", this._handles.nextMediaH);

        this._handles.prevMediaH = this.prevMedia.bind(this);
        AppEvents.AddEventListener("page-media-nav-prev", this._handles.prevMediaH);

        for (let i = 1; i <= 20; i++) {
            this.pageSizeOptions.push(5 * i);
        }

        this._handles.tagUpdateH = this.updateTagData.bind(this);
        AppEvents.AddEventListener("tags-update", this._handles.tagUpdateH);

        AppEvents.AddEventListener("random-page-refresh", this._handles.loadH);

        this._handles.updatePageSizeH = this.updatePageSize.bind(this);
        AppEvents.AddEventListener("page-size-pref-updated", this._handles.updatePageSizeH);

        this.updateSearchParams();
        this.updateTagData();
        this.load();

        if (this.display) {
            this.autoFocus();
        }
    },
    beforeUnmount: function () {
        Timeouts.Abort("page-random-load");
        Request.Abort("page-random-load");
        AppEvents.RemoveEventListener("auth-status-changed", this._handles.loadH);
        AppEvents.RemoveEventListener("media-meta-change", this._handles.loadH);
        AppEvents.RemoveEventListener("media-delete", this._handles.loadH);
        AppEvents.RemoveEventListener("app-status-update", this._handles.statusChangeH);
        AppEvents.RemoveEventListener("page-media-nav-next", this._handles.nextMediaH);
        AppEvents.RemoveEventListener("page-media-nav-prev", this._handles.prevMediaH);
        AppEvents.RemoveEventListener("tags-update", this._handles.tagUpdateH);
        AppEvents.RemoveEventListener("random-page-refresh", this._handles.loadH);
        AppEvents.RemoveEventListener("page-size-pref-updated", this._handles.updatePageSizeH);
        KeyboardManager.RemoveHandler(this._handles.handleGlobalKeyH);
        AlbumsController.OnPageUnload();
    },
    watch: {
        display: function () {
            this.load();
            if (this.display) {
                this.autoFocus();
            }
        },
    },
});
</script>
