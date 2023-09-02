<template>
    <div :class="{ 'page-inner': !inModal, hidden: !display }">
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
                    <div class="media-tag-name">{{ getTagName(tag, tagData) }}</div>
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

        <div class="search-results" tabindex="-1">
            <div v-if="!loading && started && pageItems.length === 0" class="search-results-msg-display">
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

            <div v-if="!finished && pageItems.length >= PAGE_SIZE" class="search-continue-mark">
                <button type="button" class="btn btn-primary btn-mr" disabled>
                    <i class="fa fa-spinner fa-spin"></i> {{ $t("Searching") }}... ({{ cssProgress(progress) }})
                </button>
            </div>
        </div>
    </div>
</template>

<script lang="ts">
import { AlbumsAPI } from "@/api/api-albums";
import { MediaListItem } from "@/api/models";
import { SearchAPI } from "@/api/api-search";
import { AlbumsController } from "@/control/albums";
import { AppEvents } from "@/control/app-events";
import { AppStatus } from "@/control/app-status";
import { AuthController } from "@/control/auth";
import { KeyboardManager } from "@/control/keyboard";
import { MediaEntry } from "@/control/media";
import { TagEntry, TagsController } from "@/control/tags";
import { elementInView } from "@/utils/in-view";
import { filterToWords, matchSearchFilter, normalizeString } from "@/utils/normalize";
import { clone } from "@/utils/objects";
import { GenerateURIQuery, GetAssetURL, Request } from "@/utils/request";
import { renderTimeSeconds } from "@/utils/time";
import { Timeouts } from "@/utils/timeout";
import { defineComponent, nextTick } from "vue";

const PAGE_SIZE = 50;

export default defineComponent({
    name: "PageAdvancedSearch",
    emits: ["select-media"],
    props: {
        display: Boolean,
        inModal: Boolean,
        noAlbum: Number,
    },
    data: function () {
        return {
            loading: false,

            order: "desc",
            textSearch: "",
            type: 0,

            currentMedia: AppStatus.CurrentMedia,

            pageItems: [],
            page: 0,
            totalPages: 0,
            progress: 0,

            started: false,
            finished: true,

            advancedSearch: false,

            tagData: {},
            tags: [],
            tagToAdd: "",
            matchingTags: [],
            tagMode: "all",
            PAGE_SIZE: PAGE_SIZE,

            albums: [],
            albumSearch: -1,
            albumFilter: null,
        };
    },
    methods: {
        markDirty: function () {
            Timeouts.Set("page-adv-search-dirty", 330, () => {
                this.startSearch();
            });
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

        autoScroll: function () {
            if (!this.inModal) {
                nextTick(() => {
                    const currentElem = this.$el.querySelector(".search-result-item.current");
                    if (currentElem) {
                        currentElem.scrollIntoView();
                    }
                });
                this.onCurrentMediaChanged();
            }
        },

        load: function () {
            Timeouts.Abort("page-adv-search-load");
            Request.Abort("page-adv-search-load");

            if (!this.display || this.finished) {
                return;
            }

            this.loading = true;

            if (AuthController.Locked) {
                return; // Vault is locked
            }

            Request.Pending("page-adv-search-load", SearchAPI.Search(this.getFirstTag(), this.order, this.page, PAGE_SIZE))
                .onSuccess((result) => {
                    this.filterElements(result.page_items);
                    this.page = result.page_index + 1;
                    this.totalPages = result.page_count;
                    this.progress = (this.page / Math.max(1, this.totalPages)) * 100;
                    if (this.pageItems.length >= PAGE_SIZE) {
                        // Done for now
                        this.loading = false;

                        if (this.page >= this.totalPages) {
                            this.finished = true;
                        }

                        this.autoScroll();
                    } else if (this.page < this.totalPages) {
                        this.load();
                    } else {
                        this.loading = false;
                        this.finished = true;
                        this.autoScroll();
                    }
                })
                .onRequestError((err) => {
                    Request.ErrorHandler()
                        .add(401, "*", () => {
                            AppEvents.Emit("unauthorized", false);
                        })
                        .add("*", "*", () => {
                            // Retry
                            Timeouts.Set("page-adv-search-load", 1500, this._handles.loadH);
                        })
                        .handle(err);
                })
                .onUnexpectedError((err) => {
                    console.error(err);
                    // Retry
                    Timeouts.Set("page-adv-search-load", 1500, this._handles.loadH);
                });
        },

        toggleAdvancedSearch: function () {
            this.advancedSearch = !this.advancedSearch;
        },

        filterElements: function (results: MediaEntry[]) {
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

            for (let e of results) {
                if (backlistAlbum.has(e.id)) {
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
                        for (let tag of filterTags) {
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
                        for (let tag of filterTags) {
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
                        for (let tag of filterTags) {
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

                this.pageItems.push(e);
            }
        },

        startSearch: function (event?: Event) {
            if (event) {
                event.preventDefault();
            }
            Timeouts.Abort("page-adv-search-dirty");
            this.loading = true;
            this.pageItems = [];
            this.page = 0;
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
            Request.Abort("page-adv-search-load");

            Request.Pending("page-adv-search-load", AlbumsAPI.GetAlbum(this.albumSearch))
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
                    this.autoScroll();
                })
                .onRequestError((err) => {
                    Request.ErrorHandler()
                        .add(401, "*", () => {
                            this.cancel();
                            AppEvents.Emit("unauthorized");
                        })
                        .add(404, "*", () => {
                            this.albumFilter = new Set();
                            this.load();
                        })
                        .add("*", "*", () => {
                            Timeouts.Set("page-adv-search-load", 1500, this.loadAlbumSearch.bind(this));
                        })
                        .handle(err);
                })
                .onUnexpectedError((err) => {
                    console.error(err);
                    this.cancel();
                });
        },

        cancel: function () {
            Timeouts.Abort("page-adv-search-load");
            Request.Abort("page-adv-search-load");
            this.loading = false;
            this.finished = true;
        },

        resetSearch: function () {
            Timeouts.Abort("page-adv-search-load");
            Request.Abort("page-adv-search-load");
            this.pageItems = [];
            this.page = 0;
            this.totalPages = 0;
            this.progress = 0;
            this.loading = false;
            this.finished = true;
            this.started = false;
        },

        goToMedia: function (mid, e) {
            if (e) {
                e.preventDefault();
            }
            if (this.inModal) {
                this.$emit("select-media", mid, () => {
                    this.pageItems = this.pageItems.filter((i) => {
                        return mid !== i.id;
                    });
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
                GenerateURIQuery({
                    media: mid + "",
                })
            );
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

        cssProgress: function (p: number) {
            return Math.round(p) + "%";
        },

        updateTagData: function () {
            this.tagData = clone(TagsController.Tags);
            this.onTagAddChanged(false);
        },

        getFirstTag: function () {
            if (this.tagMode === "all" && this.tags.length > 0) {
                return this.getTagName(this.tags[0], this.tagData);
            } else {
                return "";
            }
        },

        getTagName: function (tag, data) {
            if (data[tag + ""]) {
                return data[tag + ""].name;
            } else {
                return "???";
            }
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
                if (this._handles.findTagTimeout) {
                    clearTimeout(this._handles.findTagTimeout);
                    this._handles.findTagTimeout = null;
                }
                this.findTags();
            } else {
                if (this._handles.findTagTimeout) {
                    return;
                }
                this._handles.findTagTimeout = setTimeout(() => {
                    this._handles.findTagTimeout = null;
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
            this.matchingTags = Object.values(this.tagData)
                .map((a: any) => {
                    if (!tagFilter) {
                        return {
                            id: a.id,
                            name: a.name,
                            starts: true,
                            contains: true,
                        };
                    }
                    const i = a.name.indexOf(tagFilter);
                    return {
                        id: a.id,
                        name: a.name,
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
            this.currentMedia = AppStatus.CurrentMedia;
            if (!this.inModal) {
                nextTick(() => {
                    const currentElem = this.$el.querySelector(".search-result-item.current");
                    if (currentElem) {
                        currentElem.scrollIntoView();
                    }
                });
                this.onCurrentMediaChanged();
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

        onCurrentMediaChanged: function () {
            if (!this.inModal) {
                const i = this.findCurrentMediaIndex();
                AlbumsController.OnPageLoad(i, this.pageItems.length, 0, 1);
            }
        },

        nextMedia: function () {
            const i = this.findCurrentMediaIndex();
            if (i !== -1 && i < this.pageItems.length - 1) {
                this.goToMedia(this.pageItems[i + 1].id);
            } else if (i === -1 && this.pageItems.length > 0) {
                this.goToMedia(this.pageItems[0].id);
            }
        },

        prevMedia: function () {
            const i = this.findCurrentMediaIndex();
            if (i !== -1 && i > 0) {
                this.goToMedia(this.pageItems[i - 1].id);
            } else if (i === -1 && this.pageItems.length > 0) {
                this.goToMedia(this.pageItems[0].id);
            }
        },

        handleGlobalKey: function (event: KeyboardEvent): boolean {
            if (AuthController.Locked || !AppStatus.IsPageVisible() || !this.display || !event.key || event.ctrlKey) {
                return false;
            }

            if (this.inModal) {
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

            return false;
        },

        renderHintTitle(item: MediaListItem, tags: { [id: string]: TagEntry }): string {
            let parts = [item.title || this.$t("Untitled")];

            if (item.tags.length > 0) {
                let tagNames = [];

                for (let tag of item.tags) {
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

        checkContinueSearch: function () {
            if (this.finished || this.loading || this.pageItems.length === 0) {
                return;
            }

            const elem = this.$el.querySelector(".search-continue-mark");

            if (!elem) {
                return;
            }

            if (!elementInView(elem)) {
                return;
            }

            this.load();
        },

        updateAlbums: function () {
            this.albums = AlbumsController.GetAlbumsListCopy().sort((a, b) => {
                if (a.nameLowerCase < b.nameLowerCase) {
                    return -1;
                } else {
                    return 1;
                }
            });
        },
    },
    mounted: function () {
        this._handles = Object.create(null);
        this.advancedSearch = false;
        this._handles.handleGlobalKeyH = this.handleGlobalKey.bind(this);
        KeyboardManager.AddHandler(this._handles.handleGlobalKeyH, 20);

        this._handles.loadH = this.load.bind(this);
        this._handles.resetH = this.resetSearch.bind(this);
        this._handles.statusChangeH = this.onAppStatusChanged.bind(this);

        AppEvents.AddEventListener("auth-status-changed", this._handles.loadH);
        AppEvents.AddEventListener("media-delete", this._handles.resetH);
        AppEvents.AddEventListener("media-meta-change", this._handles.resetH);

        AppEvents.AddEventListener("app-status-update", this._handles.statusChangeH);

        this._handles.nextMediaH = this.nextMedia.bind(this);
        AppEvents.AddEventListener("page-media-nav-next", this._handles.nextMediaH);

        this._handles.prevMediaH = this.prevMedia.bind(this);
        AppEvents.AddEventListener("page-media-nav-prev", this._handles.prevMediaH);

        this._handles.continueCheckInterval = setInterval(this.checkContinueSearch.bind(this), 500);

        this._handles.tagUpdateH = this.updateTagData.bind(this);
        AppEvents.AddEventListener("tags-update", this._handles.tagUpdateH);

        this.updateAlbums();
        this._handles.albumsUpdateH = this.updateAlbums.bind(this);
        AppEvents.AddEventListener("albums-update", this._handles.albumsUpdateH);

        this.updateTagData();

        this.startSearch();

        if (this.display) {
            this.autoFocus();
        }
    },
    beforeUnmount: function () {
        Timeouts.Abort("page-adv-search-load");
        Request.Abort("page-adv-search-load");

        Timeouts.Abort("page-adv-search-dirty");

        AppEvents.RemoveEventListener("auth-status-changed", this._handles.loadH);
        AppEvents.RemoveEventListener("media-delete", this._handles.resetH);
        AppEvents.RemoveEventListener("media-meta-change", this._handles.resetH);

        AppEvents.RemoveEventListener("app-status-update", this._handles.statusChangeH);

        AppEvents.RemoveEventListener("page-media-nav-next", this._handles.nextMediaH);
        AppEvents.RemoveEventListener("page-media-nav-prev", this._handles.prevMediaH);

        AppEvents.RemoveEventListener("tags-update", this._handles.tagUpdateH);
        AppEvents.RemoveEventListener("albums-update", this._handles.albumsUpdateH);

        if (this._handles.findTagTimeout) {
            clearTimeout(this._handles.findTagTimeout);
        }

        clearInterval(this._handles.continueCheckInterval);

        KeyboardManager.RemoveHandler(this._handles.handleGlobalKeyH);

        if (!this.inModal) {
            AlbumsController.OnPageUnload();
        }
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
    },
});
</script>
