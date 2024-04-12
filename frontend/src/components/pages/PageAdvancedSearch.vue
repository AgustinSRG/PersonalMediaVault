<template>
    <div :class="{ 'page-inner': !inModal, 'page-in-modal': !!inModal, hidden: !display }" @scroll.passive="onPageScroll" tabindex="-1">
        <form class="adv-search-form" @submit="startSearch">
            <div class="form-group">
                <label>{{ $t("Title or description must contain") }}:</label>
                <input
                    type="text"
                    name="title-search"
                    autocomplete="off"
                    maxlength="255"
                    v-model="textSearch"
                    class="form-control form-control-full-width"
                    @input="markDirty"
                />
            </div>

            <div class="form-group">
                <label>{{ $t("Tags") }}:</label>
                <select class="form-control form-select form-control-full-width" v-model="tagMode" @change="markDirty">
                    <option :value="'all'">
                        {{ $t("Media must contain ALL of the selected tags") }}
                    </option>
                    <option :value="'any'">
                        {{ $t("Media must contain ANY of the selected tags") }}
                    </option>
                    <option :value="'none'">
                        {{ $t("Media must contain NONE of the selected tags") }}
                    </option>
                    <option :value="'untagged'">
                        {{ $t("Media must be untagged") }}
                    </option>
                </select>
            </div>
            <div class="form-group media-tags" v-if="tagMode !== 'untagged'">
                <div v-for="tag in tags" :key="tag" class="media-tag">
                    <div class="media-tag-name">{{ getTagName(tag, tagVersion) }}</div>
                    <button type="button" :title="$t('Remove tag')" class="media-tag-btn" @click="removeTag(tag)">
                        <i class="fas fa-times"></i>
                    </button>
                </div>
                <div class="media-tags-finder">
                    <input
                        type="text"
                        autocomplete="off"
                        maxlength="255"
                        v-model="tagToAdd"
                        @input="onTagAddChanged(false)"
                        @keydown="onTagAddKeyDown"
                        class="form-control auto-focus"
                        :placeholder="$t('Search for tags') + '...'"
                    />
                </div>
            </div>
            <div class="form-group" v-if="tagMode !== 'untagged' && matchingTags.length > 0">
                <button
                    v-for="mt in matchingTags"
                    :key="mt.id"
                    type="button"
                    class="btn btn-primary btn-sm btn-tag-mini"
                    @click="addMatchingTag(mt)"
                >
                    <i class="fas fa-plus"></i> {{ mt.name }}
                </button>
            </div>

            <div v-if="advancedSearch">
                <div class="form-group">
                    <label>{{ $t("Media type") }}:</label>
                    <select class="form-control form-select form-control-full-width" v-model="type" @change="markDirty">
                        <option :value="0">{{ $t("Any media") }}</option>
                        <option :value="1">{{ $t("Images") }}</option>
                        <option :value="2">{{ $t("Videos") }}</option>
                        <option :value="3">{{ $t("Audios") }}</option>
                    </select>
                </div>

                <div class="form-group">
                    <label>{{ $t("Album") }}:</label>
                    <select v-model="albumSearch" class="form-control form-select form-control-full-width" @change="markDirty">
                        <option :value="-1">--</option>
                        <option v-for="a in albums" :key="a.id" :value="a.id">
                            {{ a.name }}
                        </option>
                    </select>
                </div>

                <div class="form-group">
                    <label>{{ $t("Order") }}:</label>
                    <select class="form-control form-select form-control-full-width" v-model="order" @change="markDirty">
                        <option :value="'desc'">{{ $t("Show most recent") }}</option>
                        <option :value="'asc'">{{ $t("Show oldest") }}</option>
                    </select>
                </div>
            </div>

            <div class="">
                <button v-if="!loading" type="submit" class="btn btn-primary btn-mr">
                    <i class="fas fa-search"></i> {{ $t("Search") }}
                </button>
                <button v-if="loading" type="button" class="btn btn-primary btn-mr" disabled>
                    <i class="fa fa-spinner fa-spin"></i> {{ $t("Searching") }}... ({{ cssProgress(progress) }})
                </button>
                <button v-if="!advancedSearch" type="button" class="btn btn-primary btn-mr" @click="toggleAdvancedSearch">
                    <i class="fas fa-cog"></i> {{ $t("More options") }}
                </button>
                <button v-if="advancedSearch" type="button" class="btn btn-primary btn-mr" @click="toggleAdvancedSearch">
                    <i class="fas fa-cog"></i> {{ $t("Less options") }}
                </button>
                <button v-if="loading" type="button" class="btn btn-primary btn-mr" @click="cancel">
                    <i class="fas fa-times"></i> {{ $t("Cancel") }}
                </button>
            </div>
        </form>

        <div class="search-results">
            <div v-if="!loading && started && fullListLength === 0" class="search-results-msg-display">
                <div class="search-results-msg-icon"><i class="fas fa-search"></i></div>
                <div class="search-results-msg-text">
                    {{ $t("Could not find any result") }}
                </div>
                <div class="search-results-msg-btn">
                    <button type="button" @click="startSearch()" class="btn btn-primary">
                        <i class="fas fa-sync-alt"></i> {{ $t("Refresh") }}
                    </button>
                </div>
            </div>

            <div v-if="pageItems.length > 0" class="search-results-final-display">
                <div v-for="i in rowPaddingPreserveCols" :key="'pad-prev-' + i" class="search-result-item">
                    <div class="search-result-thumb">
                        <div class="search-result-thumb-inner">
                            <div class="search-result-loader">
                                <i class="fa fa-spinner fa-spin"></i>
                            </div>
                        </div>
                    </div>
                    <div v-if="displayTitles" class="search-result-title">{{ $t("Loading") }}...</div>
                </div>
                <div v-for="item in pageItems" :key="item.id" class="search-result-item" :class="{ current: currentMedia == item.id }">
                    <a
                        class="clickable"
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
                        <div v-if="displayTitles" class="search-result-title">
                            {{ item.title || $t("Untitled") }}
                        </div>
                    </a>
                </div>
                <div v-for="i in lastRowPadding" :key="'pad-last-' + i" class="search-result-item"></div>
            </div>

            <div v-if="!finished && fullListLength >= pageSize" class="search-continue-mark">
                <button type="button" class="btn btn-primary btn-mr" disabled>
                    <i class="fa fa-spinner fa-spin"></i> {{ $t("Searching") }}... ({{ cssProgress(progress) }})
                </button>
            </div>
        </div>
    </div>
</template>

<script lang="ts">
import { MediaListItem } from "@/api/models";
import { AlbumsController, EVENT_NAME_ALBUMS_LIST_UPDATE } from "@/control/albums";
import { AppEvents } from "@/control/app-events";
import { AppStatus, EVENT_NAME_APP_STATUS_CHANGED } from "@/control/app-status";
import { AuthController, EVENT_NAME_AUTH_CHANGED, EVENT_NAME_UNAUTHORIZED } from "@/control/auth";
import { EVENT_NAME_TAGS_UPDATE, TagsController } from "@/control/tags";
import { filterToWords, matchSearchFilter, normalizeString } from "@/utils/normalize";
import { generateURIQuery, getAssetURL } from "@/utils/api";
import { makeNamedApiRequest, abortNamedApiRequest } from "@asanrom/request-browser";
import { renderTimeSeconds } from "@/utils/time";
import { setNamedTimeout, clearNamedTimeout } from "@/utils/named-timeouts";
import { defineComponent, nextTick } from "vue";
import { useVModel } from "@/utils/v-model";
import { BigListScroller } from "@/utils/big-list-scroller";
import {
    EVENT_NAME_ADVANCED_SEARCH_GO_TOP,
    EVENT_NAME_MEDIA_DELETE,
    EVENT_NAME_MEDIA_METADATA_CHANGE,
    EVENT_NAME_PAGE_NAV_NEXT,
    EVENT_NAME_PAGE_NAV_PREV,
    PagesController,
} from "@/control/pages";
import { getUniqueStringId } from "@/utils/unique-id";
import { apiAlbumsGetAlbum } from "@/api/api-albums";
import { apiAdvancedSearch } from "@/api/api-search";
import { isTouchDevice } from "@/utils/touch";

const INITIAL_WINDOW_SIZE = 50;

export default defineComponent({
    name: "PageAdvancedSearch",
    emits: ["select-media", "update:pageScroll"],
    props: {
        display: Boolean,
        min: Boolean,
        inModal: Boolean,
        noAlbum: Number,
        pageScroll: Number,
        pageSize: Number,
        displayTitles: Boolean,

        rowSize: Number,
        rowSizeMin: Number,
        minItemsSize: Number,
        maxItemsSize: Number,
    },
    setup(props) {
        return {
            loadRequestId: getUniqueStringId(),
            dirtyTimeoutId: getUniqueStringId(),
            mediaIndexMap: new Map<number, number>(),
            listScroller: null as BigListScroller,
            findTagTimeout: null,
            continueCheckInterval: null,
            checkContainerTimer: null,
            pageScrollStatus: useVModel(props, "pageScroll"),
            windowResizeObserver: null as ResizeObserver,
        };
    },
    data: function () {
        return {
            loading: false,

            order: "desc" as "desc" | "asc",
            textSearch: "",
            type: 0,

            currentMedia: AppStatus.CurrentMedia,

            pageItems: [],
            page: 0,
            totalPages: 0,
            progress: 0,
            continueRef: null as number | null,

            fullListLength: 0,

            started: false,
            finished: true,

            advancedSearch: false,

            tagVersion: TagsController.TagsVersion,
            tags: [],
            tagToAdd: "",
            matchingTags: [],
            tagMode: "all",

            albums: [],
            albumSearch: -1,
            albumFilter: null,

            windowPosition: 0,

            windowWidth: 0,
        };
    },
    computed: {
        rowPaddingPreserveCols() {
            const containerWidth = this.windowWidth;

            const itemWidth = Math.max(
                this.minItemsSize,
                Math.min(
                    this.maxItemsSize,
                    this.min ? containerWidth / Math.max(1, this.rowSizeMin) : containerWidth / Math.max(1, this.rowSize),
                ),
            );

            const itemsFitInRow = Math.max(1, Math.floor(containerWidth / Math.max(1, itemWidth)));

            return this.windowPosition % itemsFitInRow;
        },
        lastRowPadding() {
            const containerWidth = this.windowWidth;

            const itemWidth = Math.max(
                this.minItemsSize,
                Math.min(
                    this.maxItemsSize,
                    this.min ? containerWidth / Math.max(1, this.rowSizeMin) : containerWidth / Math.max(1, this.rowSize),
                ),
            );

            const itemsFitInRow = Math.max(1, Math.floor(containerWidth / Math.max(1, itemWidth)));

            const lastWindowElement = this.windowPosition + this.pageItems.length - 1;

            return Math.max(0, itemsFitInRow - 1 - (lastWindowElement % itemsFitInRow));
        },
    },
    methods: {
        markDirty: function () {
            setNamedTimeout(this.dirtyTimeoutId, 330, () => {
                this.startSearch();
            });
        },

        autoFocus: function () {
            if (isTouchDevice()) {
                return;
            }
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

        scrollToCurrentMedia: function (): boolean {
            if (!this.mediaIndexMap.has(this.currentMedia)) {
                return false;
            }
            const index = this.mediaIndexMap.get(this.currentMedia);

            if (index < this.listScroller.windowPosition || index >= this.listScroller.windowPosition + this.listScroller.windowSize) {
                this.listScroller.moveWindowToElement(this.mediaIndexMap.get(this.currentMedia));
            }

            nextTick(() => {
                const currentElem = this.$el.querySelector(".search-result-item.current");
                if (currentElem) {
                    currentElem.scrollIntoView();
                }
            });

            return true;
        },

        load: function () {
            clearNamedTimeout(this.loadRequestId);
            abortNamedApiRequest(this.loadRequestId);

            if (!this.display || this.finished) {
                return;
            }

            this.loading = true;

            if (AuthController.Locked) {
                return; // Vault is locked
            }

            const pageSize = this.pageSize;

            makeNamedApiRequest(
                this.loadRequestId,
                apiAdvancedSearch(this.getTagMode(), this.getTagList(), this.order, this.continueRef, pageSize),
            )
                .onSuccess((result) => {
                    const completePageList = this.listScroller.list;
                    this.filterElements(result.items);
                    this.page = result.scanned;
                    this.totalPages = result.total_count;
                    this.progress = (Math.max(0, result.scanned) / Math.max(1, result.total_count)) * 100;
                    this.continueRef = result["continue"];
                    if (completePageList.length >= pageSize) {
                        // Done for now
                        this.loading = false;

                        if (this.page >= this.totalPages) {
                            this.finished = true;
                        }

                        if (!this.inModal) {
                            this.onCurrentMediaChanged();
                        }
                    } else if (result.scanned < result.total_count) {
                        // Maybe there are more items
                        this.load();
                    } else {
                        this.loading = false;
                        this.finished = true;
                        if (!this.inModal) {
                            this.onCurrentMediaChanged();
                        }
                    }
                })
                .onRequestError((err, handleErr) => {
                    handleErr(err, {
                        unauthorized: () => {
                            AppEvents.Emit(EVENT_NAME_UNAUTHORIZED);
                        },
                        temporalError: () => {
                            // Retry
                            setNamedTimeout(this.loadRequestId, 1500, this.load.bind(this));
                        },
                    });
                })
                .onUnexpectedError((err) => {
                    console.error(err);
                    // Retry
                    setNamedTimeout(this.loadRequestId, 1500, this.load.bind(this));
                });
        },

        toggleAdvancedSearch: function () {
            this.advancedSearch = !this.advancedSearch;
        },

        filterElements: function (results: MediaListItem[]) {
            TagsController.OnMediaListReceived(results);
            const filterText = normalizeString(this.textSearch).trim().toLowerCase();
            const filterTextWords = filterToWords(filterText);
            const filterType = this.type;
            const filterTags = this.tags.slice();
            const filterTagMode = this.tagMode;

            let backlistAlbum = new Set();

            if (this.noAlbum >= 0 && AlbumsController.CurrentAlbumData) {
                backlistAlbum = new Set(
                    AlbumsController.CurrentAlbumData.list.map((a) => {
                        return a.id;
                    }),
                );
            }

            const resultsToAdd = [];

            for (const e of results) {
                if (backlistAlbum.has(e.id)) {
                    continue;
                }

                if (this.mediaIndexMap.has(e.id)) {
                    continue;
                }

                if (this.albumFilter && !this.albumFilter.has(e.id)) {
                    continue;
                }

                if (filterText) {
                    if (
                        matchSearchFilter(e.title, filterText, filterTextWords) < 0 &&
                        matchSearchFilter(e.description, filterText, filterTextWords) < 0
                    ) {
                        continue;
                    }
                }

                if (filterType) {
                    if (e.type !== filterType) {
                        continue;
                    }
                }

                if (filterTagMode === "all") {
                    if (filterTags.length > 0) {
                        let passesTags = true;
                        for (const tag of filterTags) {
                            if (!e.tags || !e.tags.includes(tag)) {
                                passesTags = false;
                                break;
                            }
                        }

                        if (!passesTags) {
                            continue;
                        }
                    }
                } else if (filterTagMode === "none") {
                    if (filterTags.length > 0) {
                        let passesTags = true;
                        for (const tag of filterTags) {
                            if (e.tags && e.tags.includes(tag)) {
                                passesTags = false;
                                break;
                            }
                        }

                        if (!passesTags) {
                            continue;
                        }
                    }
                } else if (filterTagMode === "any") {
                    if (filterTags.length > 0) {
                        let passesTags = false;
                        for (const tag of filterTags) {
                            if (!e.tags || e.tags.includes(tag)) {
                                passesTags = true;
                                break;
                            }
                        }

                        if (!passesTags) {
                            continue;
                        }
                    }
                } else if (filterTagMode === "untagged") {
                    if (e.tags && e.tags.length > 0) {
                        continue;
                    }
                }

                this.mediaIndexMap.set(e.id, this.listScroller.list.length + resultsToAdd.length);

                resultsToAdd.push(e);
            }

            this.listScroller.addElements(resultsToAdd);
            this.fullListLength = this.listScroller.list.length;

            nextTick(() => {
                this.checkContainerHeight();
            });
        },

        startSearch: function (event?: Event) {
            if (event) {
                event.preventDefault();
            }
            clearNamedTimeout(this.dirtyTimeoutId);
            this.loading = true;
            this.listScroller.reset();
            this.fullListLength = 0;
            this.mediaIndexMap.clear();
            this.page = 0;
            this.continueRef = null;
            this.totalPages = 0;
            this.progress = 0;
            this.started = true;
            this.finished = false;
            this.albumFilter = null;
            if (this.albumSearch >= 0) {
                this.loadAlbumSearch();
            } else {
                this.load();
            }
        },

        loadAlbumSearch: function () {
            abortNamedApiRequest(this.loadRequestId);

            makeNamedApiRequest(this.loadRequestId, apiAlbumsGetAlbum(this.albumSearch))
                .onSuccess((result) => {
                    if (this.order === "asc") {
                        this.filterElements(
                            result.list.sort((a, b) => {
                                if (a.id < b.id) {
                                    return -1;
                                } else {
                                    return 1;
                                }
                            }),
                        );
                    } else {
                        this.filterElements(
                            result.list.sort((a, b) => {
                                if (a.id > b.id) {
                                    return -1;
                                } else {
                                    return 1;
                                }
                            }),
                        );
                    }

                    this.page = 1;
                    this.totalPages = 1;
                    this.progress = 100;
                    this.loading = false;
                    this.finished = true;
                    if (!this.inModal) {
                        this.onCurrentMediaChanged();
                    }
                })
                .onRequestError((err, handleErr) => {
                    handleErr(err, {
                        unauthorized: () => {
                            this.cancel();
                            AppEvents.Emit(EVENT_NAME_UNAUTHORIZED);
                        },
                        notFound: () => {
                            this.albumFilter = new Set();
                            this.load();
                        },
                        temporalError: () => {
                            setNamedTimeout(this.loadRequestId, 1500, this.loadAlbumSearch.bind(this));
                        },
                    });
                })
                .onUnexpectedError((err) => {
                    console.error(err);
                    this.cancel();
                });
        },

        cancel: function () {
            clearNamedTimeout(this.loadRequestId);
            abortNamedApiRequest(this.loadRequestId);
            this.loading = false;
            this.finished = true;
        },

        resetSearch: function () {
            clearNamedTimeout(this.loadRequestId);
            abortNamedApiRequest(this.loadRequestId);
            this.listScroller.reset();
            this.fullListLength = 0;
            this.mediaIndexMap.clear();
            this.page = 0;
            this.totalPages = 0;
            this.progress = 0;
            this.continueRef = null;
            this.loading = false;
            this.finished = true;
            this.started = false;
            this.startSearch();
        },

        updatePageSize: function () {
            this.resetSearch();
        },

        goToMedia: function (mid: number, e?: Event) {
            if (e) {
                e.preventDefault();
            }
            if (this.inModal) {
                this.$emit("select-media", mid, () => {
                    const fullList = this.listScroller.list;
                    const centerPosition = this.listScroller.getCenterPosition();

                    const mediaIndex = this.mediaIndexMap.get(mid);

                    if (mediaIndex !== undefined) {
                        fullList.splice(mediaIndex, 1);
                        this.mediaIndexMap.delete(mid);

                        this.listScroller.moveWindowToElement(centerPosition);

                        for (let i = mediaIndex; i < fullList.length; i++) {
                            this.mediaIndexMap.set(fullList[i].id, i);
                        }
                    }
                });
            } else {
                AppStatus.ClickOnMedia(mid, true);
            }
        },

        getMediaURL: function (mid: number): string {
            return (
                window.location.protocol +
                "//" +
                window.location.host +
                window.location.pathname +
                generateURIQuery({
                    media: mid + "",
                })
            );
        },

        getThumbnail(thumb: string) {
            return getAssetURL(thumb);
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

        cssProgress: function (p: number) {
            return Math.round(p) + "%";
        },

        updateTagData: function () {
            this.tagVersion = TagsController.TagsVersion;
            this.onTagAddChanged(false);
        },

        getTagMode: function (): "allof" | "anyof" | "noneof" {
            switch (this.tagMode) {
                case "any":
                    if (this.tags.length > 16) {
                        return "allof";
                    }
                    return "anyof";
                case "none":
                    return "noneof";
                default:
                    return "allof";
            }
        },

        getTagList: function (): string[] {
            if (this.tagMode === "untagged") {
                return [];
            }
            if (this.tagMode === "any" && this.tags.length > 16) {
                return [];
            }
            return this.tags
                .map((tag) => {
                    return this.getTagName(tag, this.tagVersion);
                })
                .slice(0, 16);
        },

        getTagName: function (tag: number, v: number) {
            return TagsController.GetTagName(tag, v);
        },

        removeTag: function (tag) {
            this.tags = this.tags.filter((t) => {
                return tag !== t;
            });
            this.markDirty();
            this.onTagAddChanged(true);
        },

        addMatchingTag: function (tag) {
            if (this.tags.indexOf(tag.id) >= 0) {
                return;
            }
            this.tags.push(tag.id);
            this.markDirty();
            this.onTagAddChanged(true);
        },

        onTagAddKeyDown: function (event: KeyboardEvent) {
            if (event.key === "Enter") {
                event.preventDefault();

                this.onTagAddChanged(true);

                if (this.matchingTags.length > 0) {
                    this.addMatchingTag(this.matchingTags[0]);
                    this.tagToAdd = "";
                    this.onTagAddChanged(true);
                }
            } else if (event.key === "Tab" && this.tagToAdd && !event.shiftKey) {
                this.onTagAddChanged(true);

                if (this.matchingTags.length > 0 && this.matchingTags[0].name !== this.tagToAdd) {
                    this.tagToAdd = this.matchingTags[0].name;
                    this.onTagAddChanged(true);
                    event.preventDefault();
                }
            }
        },

        onTagAddChanged: function (forced?: boolean) {
            if (forced) {
                if (this.findTagTimeout) {
                    clearTimeout(this.findTagTimeout);
                    this.findTagTimeout = null;
                }
                this.findTags();
            } else {
                if (this.findTagTimeout) {
                    return;
                }
                this.findTagTimeout = setTimeout(() => {
                    this.findTagTimeout = null;
                    this.findTags();
                }, 200);
            }
        },

        findTags: function () {
            const tagFilter = this.tagToAdd
                .replace(/[\n\r]/g, " ")
                .trim()
                .replace(/[\s]/g, "_")
                .toLowerCase();
            this.matchingTags = Array.from(TagsController.Tags.entries())
                .map((a) => {
                    if (!tagFilter) {
                        return {
                            id: a[0],
                            name: a[1],
                            starts: true,
                            contains: true,
                        };
                    }
                    const i = a[1].indexOf(tagFilter);
                    return {
                        id: a[0],
                        name: a[1],
                        starts: i === 0,
                        contains: i >= 0,
                    };
                })
                .filter((a) => {
                    if (this.tags.indexOf(a.id) >= 0) {
                        return false;
                    }
                    return a.starts || a.contains;
                })
                .sort((a, b) => {
                    if (a.starts && !b.starts) {
                        return -1;
                    } else if (b.starts && !a.starts) {
                        return 1;
                    } else if (a.name < b.name) {
                        return -1;
                    } else {
                        return 1;
                    }
                })
                .slice(0, 10);
        },

        onAppStatusChanged: function () {
            const changed = this.currentMedia !== AppStatus.CurrentMedia;
            this.currentMedia = AppStatus.CurrentMedia;
            if (!this.inModal) {
                if (changed) {
                    nextTick(() => {
                        if (!this.checkContainerHeight()) {
                            this.scrollToCurrentMedia();
                        }
                    });
                }
                this.onCurrentMediaChanged();
            }
        },

        findCurrentMediaIndex: function (): number {
            if (this.mediaIndexMap.has(this.currentMedia)) {
                return this.mediaIndexMap.get(this.currentMedia);
            } else {
                return -1;
            }
        },

        onCurrentMediaChanged: function () {
            if (!this.inModal) {
                const completePageList = this.listScroller.list;
                const i = this.findCurrentMediaIndex();
                PagesController.OnPageLoad(i, completePageList.length, 0, 1);
            }
        },

        nextMedia: function () {
            const completePageList = this.listScroller.list;
            const i = this.findCurrentMediaIndex();
            if (i !== -1 && i < completePageList.length - 1) {
                this.goToMedia(completePageList[i + 1].id);
            } else if (i === -1 && completePageList.length > 0) {
                this.goToMedia(completePageList[0].id);
            }
        },

        prevMedia: function () {
            const completePageList = this.listScroller.list;
            const i = this.findCurrentMediaIndex();
            if (i !== -1 && i > 0) {
                this.goToMedia(completePageList[i - 1].id);
            } else if (i === -1 && completePageList.length > 0) {
                this.goToMedia(completePageList[0].id);
            }
        },

        handleGlobalKey: function (event: KeyboardEvent): boolean {
            if (AuthController.Locked || !AppStatus.IsPageVisible() || !this.display || !event.key || event.ctrlKey) {
                return false;
            }

            if (this.inModal) {
                return false;
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

        checkContinueSearch: function () {
            if (this.finished || this.loading || this.fullListLength === 0) {
                return;
            }

            if (!this.listScroller.isAtTheEnd()) {
                return;
            }

            const con = this.getContainer();

            if (!con) {
                return;
            }

            const conBounds = con.getBoundingClientRect();

            const overflowLength = con.scrollHeight - conBounds.height;

            if (overflowLength < 1) {
                this.load();
                return;
            }

            const relScroll = con.scrollTop / overflowLength;

            if (relScroll < 0.8) {
                return;
            }

            this.load();
        },

        updateAlbums: function () {
            this.albums = AlbumsController.GetAlbumsListMin().sort((a, b) => {
                if (a.nameLowerCase < b.nameLowerCase) {
                    return -1;
                } else {
                    return 1;
                }
            });
        },

        onPageScroll: function (e: Event) {
            if (this.inModal) {
                return;
            }

            this.pageScrollStatus = (e.target as HTMLElement).scrollTop || 0;
            this.onScroll(e);
        },

        goTop: function () {
            if (!this.inModal) {
                this.listScroller.moveWindowToElement(0);
                nextTick(() => {
                    this.$el.scrollTop = 0;
                });
            } else {
                this.listScroller.moveWindowToElement(0);
            }
        },

        onScroll: function (e: Event) {
            this.listScroller.checkElementScroll(e.target as HTMLElement);
        },

        getContainer: function (): HTMLElement {
            if (this.inModal) {
                if (this.$el.parentElement && this.$el.parentElement.parentElement && this.$el.parentElement.parentElement.parentElement) {
                    return this.$el.parentElement.parentElement.parentElement;
                }
            } else {
                return this.$el;
            }
        },

        checkContainerHeight: function (): boolean {
            const cont = this.getContainer();

            if (!cont) {
                return false;
            }

            this.listScroller.checkElementScroll(cont);

            const centerPosition = this.listScroller.getCenterPosition();

            const el = this.$el.querySelector(".search-result-item");

            if (!el) {
                return false;
            }

            const changed = this.listScroller.checkScrollContainerHeight(cont, el);

            if (changed) {
                if (!this.scrollToCurrentMedia()) {
                    this.listScroller.moveWindowToElement(centerPosition);
                }
            }

            return changed;
        },

        updateWindowWidth: function () {
            const cont = this.getContainer();

            if (!cont) {
                return;
            }

            this.windowWidth = cont.getBoundingClientRect().width;
        },
    },
    mounted: function () {
        this.pageScrollStatus = 0;

        this.advancedSearch = false;

        this.$addKeyboardHandler(this.handleGlobalKey.bind(this), 20);

        this.$listenOnAppEvent(EVENT_NAME_AUTH_CHANGED, this.load.bind(this));

        this.$listenOnAppEvent(EVENT_NAME_MEDIA_DELETE, this.resetSearch.bind(this));
        this.$listenOnAppEvent(EVENT_NAME_MEDIA_METADATA_CHANGE, this.resetSearch.bind(this));

        this.$listenOnAppEvent(EVENT_NAME_APP_STATUS_CHANGED, this.onAppStatusChanged.bind(this));

        this.$listenOnAppEvent(EVENT_NAME_PAGE_NAV_NEXT, this.nextMedia.bind(this));

        this.$listenOnAppEvent(EVENT_NAME_PAGE_NAV_PREV, this.prevMedia.bind(this));

        this.continueCheckInterval = setInterval(this.checkContinueSearch.bind(this), 500);

        this.$listenOnAppEvent(EVENT_NAME_TAGS_UPDATE, this.updateTagData.bind(this));

        this.$listenOnAppEvent(EVENT_NAME_ADVANCED_SEARCH_GO_TOP, this.goTop.bind(this));

        this.updateAlbums();

        this.$listenOnAppEvent(EVENT_NAME_ALBUMS_LIST_UPDATE, this.updateAlbums.bind(this));

        this.updateTagData();

        this.listScroller = new BigListScroller(INITIAL_WINDOW_SIZE, {
            get: () => {
                return this.pageItems;
            },
            set: (l) => {
                this.pageItems = l;
            },
            onChange: () => {
                this.windowPosition = this.listScroller.windowPosition;
            },
        });

        this.checkContainerTimer = setInterval(this.checkContainerHeight.bind(this), 1000);

        this.startSearch();

        if (this.display) {
            this.autoFocus();
        }

        const container = this.getContainer();

        this.windowResizeObserver = new ResizeObserver(this.updateWindowWidth.bind(this));

        if (container) {
            this.windowWidth = container.getBoundingClientRect().width;
            this.windowResizeObserver.observe(container);
        } else {
            this.windowResizeObserver.observe(this.$el);
        }
    },
    beforeUnmount: function () {
        clearNamedTimeout(this.loadRequestId);
        abortNamedApiRequest(this.loadRequestId);

        clearNamedTimeout(this.dirtyTimeoutId);

        if (this.findTagTimeout) {
            clearTimeout(this.findTagTimeout);
        }

        clearInterval(this.continueCheckInterval);

        clearInterval(this.checkContainerTimer);

        if (!this.inModal) {
            PagesController.OnPageUnload();
        }

        this.windowResizeObserver.disconnect();
    },
    watch: {
        display: function () {
            this.load();
            if (this.display && this.inModal) {
                this.startSearch();
            } else if (this.inModal) {
                this.cancel();
            }
            if (this.display) {
                this.autoFocus();
            }
        },
        pageSize: function () {
            this.updatePageSize();
        },
    },
});
</script>
