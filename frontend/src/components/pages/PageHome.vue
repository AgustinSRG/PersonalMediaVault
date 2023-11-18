<template>
    <div class="page-inner" :class="{ hidden: !display }">
        <div class="search-results auto-focus" tabindex="-1">
            <PageMenu v-if="total > 0" :page="page" :pages="totalPages" :min="min" @goto="changePage"></PageMenu>

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

            <div v-if="!loading && total <= 0 && firstLoaded" class="search-results-msg-display">
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
                    <div class="search-result-thumb" :title="renderHintTitle(item, tagVersion)">
                        <div class="search-result-thumb-inner">
                            <div v-if="!item.thumbnail" class="no-thumb">
                                <i v-if="item.type === 1" class="fas fa-image"></i>
                                <i v-else-if="item.type === 2" class="fas fa-video"></i>
                                <i v-else-if="item.type === 3" class="fas fa-headphones"></i>
                                <i v-else class="fas fa-ban"></i>
                            </div>
                            <img v-if="item.thumbnail" :src="getThumbnail(item.thumbnail)" :alt="$t('Thumbnail')" loading="lazy" />
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

            <PageMenu v-if="total > 0" :page="page" :pages="totalPages" :min="min" @goto="changePage"></PageMenu>

            <div v-if="total > 0" class="search-results-total">{{ $t("Total") }}: {{ total }}</div>
        </div>
    </div>
</template>

<script lang="ts">
import { SearchAPI } from "@/api/api-search";
import { AppEvents } from "@/control/app-events";
import { AppStatus } from "@/control/app-status";
import { AuthController, EVENT_NAME_UNAUTHORIZED } from "@/control/auth";
import { GenerateURIQuery, GetAssetURL, Request } from "@/utils/request";
import { setNamedTimeout, clearNamedTimeout } from "@/utils/named-timeouts";
import { defineComponent, nextTick } from "vue";

import PageMenu from "@/components/utils/PageMenu.vue";
import { renderTimeSeconds } from "@/utils/time";
import { KeyboardManager } from "@/control/keyboard";
import { MediaListItem } from "@/api/models";
import { TagsController } from "@/control/tags";
import { EVENT_NAME_PAGE_SIZE_UPDATED, getPageMaxItems } from "@/control/app-preferences";
import { packSearchParams, unPackSearchParams } from "@/utils/search-params";
import {
    EVENT_NAME_MEDIA_DELETE,
    EVENT_NAME_MEDIA_METADATA_CHANGE,
    EVENT_NAME_PAGE_NAV_NEXT,
    EVENT_NAME_PAGE_NAV_PREV,
    PagesController,
} from "@/control/pages";
import { getUniqueStringId } from "@/utils/unique-id";

export default defineComponent({
    name: "PageHome",
    components: {
        PageMenu,
    },
    props: {
        display: Boolean,
        min: Boolean,
    },
    data: function () {
        return {
            search: AppStatus.CurrentSearch,

            loading: false,
            firstLoaded: false,

            pageSize: getPageMaxItems(),
            order: "desc",
            searchParams: AppStatus.SearchParams,

            currentMedia: AppStatus.CurrentMedia,

            page: 0,
            total: 0,
            totalPages: 0,
            pageItems: [],

            loadingFiller: [],

            pageSizeOptions: [],

            switchMediaOnLoad: "",

            tagVersion: TagsController.TagsVersion,
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
            clearNamedTimeout(this._handles.loadRequestId);
            Request.Abort(this._handles.loadRequestId);

            if (!this.display) {
                return;
            }

            this.scrollToTop();

            setNamedTimeout(this._handles.loadRequestId, 330, () => {
                this.loading = true;
            });

            if (AuthController.Locked) {
                return; // Vault is locked
            }

            Request.Pending(this._handles.loadRequestId, SearchAPI.Search("", this.order, this.page, this.pageSize))
                .onSuccess((result) => {
                    TagsController.OnMediaListReceived(result.page_items);
                    this.pageItems = result.page_items;
                    this.page = result.page_index;
                    this.totalPages = result.page_count;
                    this.total = result.total_count;
                    clearNamedTimeout(this._handles.loadRequestId);
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
                    if (this.page < 0) {
                        this.page = 0;
                        this.load();
                        return;
                    } else if (this.page >= this.totalPages && this.totalPages > 0) {
                        this.page = this.totalPages - 1;
                        this.load();
                        return;
                    }
                    this.scrollToCurrentMedia();
                    this.onCurrentMediaChanged();
                })
                .onRequestError((err) => {
                    Request.ErrorHandler()
                        .add(401, "*", () => {
                            AppEvents.Emit(EVENT_NAME_UNAUTHORIZED);
                        })
                        .add("*", "*", () => {
                            // Retry
                            this.loading = true;
                            setNamedTimeout(this._handles.loadRequestId, 1500, this._handles.loadH);
                        })
                        .handle(err);
                })
                .onUnexpectedError((err) => {
                    console.error(err);
                    // Retry
                    this.loading = true;
                    setNamedTimeout(this._handles.loadRequestId, 1500, this._handles.loadH);
                });
        },

        updatePageSize: function () {
            this.pageSize = getPageMaxItems();
            this.updateLoadingFiller();
            this.page = 0;
            this.load();
        },

        onAppStatusChanged: function () {
            const changed = this.currentMedia !== AppStatus.CurrentMedia;
            this.currentMedia = AppStatus.CurrentMedia;
            if (AppStatus.SearchParams !== this.searchParams) {
                this.searchParams = AppStatus.SearchParams;
                this.updateSearchParams();
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

        onCurrentMediaChanged: function () {
            const i = this.findCurrentMediaIndex();
            PagesController.OnPageLoad(i, this.pageItems.length, this.page, this.totalPages);
        },

        onSearchParamsChanged: function () {
            this.searchParams = packSearchParams(this.page, this.order);
            AppStatus.ChangeSearchParams(this.searchParams);
        },

        changePage: function (p) {
            this.page = p;
            this.onSearchParamsChanged();
            this.load();
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

        findCurrentMediaIndex: function (): number {
            for (let i = 0; i < this.pageItems.length; i++) {
                if (this.pageItems[i].id === this.currentMedia) {
                    return i;
                }
            }
            return -1;
        },

        nextMedia: function () {
            const i = this.findCurrentMediaIndex();
            if (i !== -1 && i < this.pageItems.length - 1) {
                this.goToMedia(this.pageItems[i + 1].id);
            } else if (i === -1 && this.pageItems.length > 0) {
                this.goToMedia(this.pageItems[0].id);
            } else if (i === this.pageItems.length - 1) {
                if (this.page < this.totalPages - 1) {
                    this.switchMediaOnLoad = "next";
                    this.changePage(this.page + 1);
                }
            }
        },

        prevMedia: function () {
            const i = this.findCurrentMediaIndex();
            if (i !== -1 && i > 0) {
                this.goToMedia(this.pageItems[i - 1].id);
            } else if (i === -1 && this.pageItems.length > 0) {
                this.goToMedia(this.pageItems[0].id);
            } else if (i === 0) {
                if (this.page > 0) {
                    this.switchMediaOnLoad = "prev";
                    this.changePage(this.page - 1);
                }
            }
        },

        handleGlobalKey: function (event: KeyboardEvent): boolean {
            if (AuthController.Locked || !AppStatus.IsPageVisible() || !this.display || !event.key || event.ctrlKey) {
                return false;
            }

            if (event.key === "PageUp" || (event.key === "ArrowLeft" && AppStatus.CurrentMedia < 0)) {
                if (this.page > 0) {
                    this.changePage(this.page - 1);
                }
                return true;
            }

            if (event.key === "PageDown" || (event.key === "ArrowRight" && AppStatus.CurrentMedia < 0)) {
                if (this.page < this.totalPages - 1) {
                    this.changePage(this.page + 1);
                }
                return true;
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
    },
    mounted: function () {
        this._handles = Object.create(null);
        this._handles.loadRequestId = getUniqueStringId();

        this._handles.loadH = this.load.bind(this);
        this._handles.statusChangeH = this.onAppStatusChanged.bind(this);

        this._handles.handleGlobalKeyH = this.handleGlobalKey.bind(this);
        KeyboardManager.AddHandler(this._handles.handleGlobalKeyH, 20);

        AuthController.AddChangeEventListener(this._handles.loadH);
        AppEvents.AddEventListener(EVENT_NAME_MEDIA_DELETE, this._handles.loadH);
        AppEvents.AddEventListener(EVENT_NAME_MEDIA_METADATA_CHANGE, this._handles.loadH);
        AppStatus.AddEventListener(this._handles.statusChangeH);

        this._handles.nextMediaH = this.nextMedia.bind(this);
        AppEvents.AddEventListener(EVENT_NAME_PAGE_NAV_NEXT, this._handles.nextMediaH);

        this._handles.prevMediaH = this.prevMedia.bind(this);
        AppEvents.AddEventListener(EVENT_NAME_PAGE_NAV_PREV, this._handles.prevMediaH);

        for (let i = 1; i <= 20; i++) {
            this.pageSizeOptions.push(5 * i);
        }

        this._handles.tagUpdateH = this.updateTagData.bind(this);
        TagsController.AddEventListener(this._handles.tagUpdateH);

        this._handles.updatePageSizeH = this.updatePageSize.bind(this);
        AppEvents.AddEventListener(EVENT_NAME_PAGE_SIZE_UPDATED, this._handles.updatePageSizeH);

        this.updateSearchParams();
        this.updateTagData();
        this.load();

        if (this.display) {
            this.autoFocus();
        }
    },
    beforeUnmount: function () {
        clearNamedTimeout(this._handles.loadRequestId);
        Request.Abort(this._handles.loadRequestId);
        AuthController.RemoveChangeEventListener(this._handles.loadH);
        AppEvents.RemoveEventListener(EVENT_NAME_MEDIA_METADATA_CHANGE, this._handles.loadH);
        AppEvents.RemoveEventListener(EVENT_NAME_MEDIA_DELETE, this._handles.loadH);
        AppStatus.RemoveEventListener(this._handles.statusChangeH);
        AppEvents.RemoveEventListener(EVENT_NAME_PAGE_NAV_NEXT, this._handles.nextMediaH);
        AppEvents.RemoveEventListener(EVENT_NAME_PAGE_NAV_PREV, this._handles.prevMediaH);
        TagsController.RemoveEventListener(this._handles.tagUpdateH);
        AppEvents.RemoveEventListener(EVENT_NAME_PAGE_SIZE_UPDATED, this._handles.updatePageSizeH);
        KeyboardManager.RemoveHandler(this._handles.handleGlobalKeyH);
        PagesController.OnPageUnload();
    },
    watch: {
        display: function () {
            this.load();
            this.switchMediaOnLoad = "";
            if (this.display) {
                this.autoFocus();
            }
        },
    },
});
</script>
