<template>
    <div class="page-inner" :class="{ hidden: !display }">
        <div class="search-results" tabindex="-1">
            <div class="search-results-options">
                <div class="search-results-option">
                    <input
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

            <div v-if="loading" class="search-results-loading-display">
                <div v-for="f in pageSize" :key="f" class="search-result-item">
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

            <div v-if="!loading && total <= 0 && !filter && firstLoaded" class="search-results-msg-display">
                <div class="search-results-msg-icon">
                    <i class="fas fa-box-open"></i>
                </div>
                <div class="search-results-msg-text">
                    {{ $t("This vault does not have any albums yet") }}
                </div>
                <div class="search-results-msg-btn">
                    <button type="button" class="btn btn-primary" @click="refreshAlbums">
                        <i class="fas fa-sync-alt"></i> {{ $t("Refresh") }}
                    </button>
                </div>
            </div>

            <div v-if="!loading && total <= 0 && filter && firstLoaded" class="search-results-msg-display">
                <div class="search-results-msg-icon">
                    <i class="fas fa-box-open"></i>
                </div>
                <div class="search-results-msg-text">
                    {{ $t("Could not find any albums matching your filter") }}
                </div>
                <div class="search-results-msg-btn">
                    <button type="button" class="btn btn-primary" @click="clearFilter">
                        <i class="fas fa-times"></i> {{ $t("Clear filter") }}
                    </button>
                </div>
            </div>

            <div v-if="!loading && total > 0" class="search-results-final-display">
                <div v-for="item in pageItems" :key="item.id" class="search-result-item">
                    <a
                        class="clickable"
                        :href="getAlbumURL(item.id)"
                        target="_blank"
                        rel="noopener noreferrer"
                        @click="goToAlbum(item, $event)"
                        ><div
                            class="search-result-thumb"
                            :title="
                                (item.name || $t('Untitled album')) +
                                (item.lm ? '\n' + $t('Last modified') + ': ' + renderDate(item.lm) : '')
                            "
                        >
                            <div class="search-result-thumb-inner">
                                <div v-if="!item.thumbnail" class="no-thumb">
                                    <i class="fas fa-list-ol"></i>
                                </div>
                                <ThumbImage v-if="item.thumbnail" :src="getThumbnail(item.thumbnail)"></ThumbImage>
                                <div v-if="item.size == 0" class="thumb-bottom-right-tag" :title="$t('Album') + ' - ' + $t('Empty')">
                                    <i class="fas fa-list-ol"></i> {{ $t("Empty") }}
                                </div>
                                <div v-else-if="item.size == 1" class="thumb-bottom-right-tag" :title="$t('Album') + ' - 1 ' + $t('item')">
                                    <i class="fas fa-list-ol"></i> 1 {{ $t("item") }}
                                </div>
                                <div
                                    v-else-if="item.size > 1"
                                    class="thumb-bottom-right-tag"
                                    :title="$t('Album') + ' - ' + item.size + ' ' + $t('items')"
                                >
                                    <i class="fas fa-list-ol"></i> {{ item.size }} {{ $t("items") }}
                                </div>
                            </div>
                        </div>
                        <div v-if="displayTitles" class="search-result-title">
                            {{ item.name || $t("Untitled") }}
                        </div></a
                    >
                </div>

                <div v-for="i in lastRowPadding" :key="'pad-last-' + i" class="search-result-item"></div>
            </div>

            <PageMenu v-if="total > 0 && order !== 'rand'" :page="page" :pages="totalPages" :min="min" @goto="changePage"></PageMenu>

            <div v-if="total > 0 && order !== 'rand'" class="search-results-total">{{ $t("Total") }}: {{ total }}</div>
        </div>

        <AlbumCreateModal v-model:display="displayAlbumCreate" @new-album="onNewAlbum"></AlbumCreateModal>
    </div>
</template>

<script lang="ts">
import { AppEvents } from "@/control/app-events";
import { AppStatus, EVENT_NAME_APP_STATUS_CHANGED } from "@/control/app-status";
import { generateURIQuery, getAssetURL } from "@/utils/api";
import { makeNamedApiRequest, abortNamedApiRequest } from "@asanrom/request-browser";
import { setNamedTimeout, clearNamedTimeout } from "@/utils/named-timeouts";
import type { PropType } from "vue";
import { defineComponent, nextTick } from "vue";
import PageMenu from "@/components/utils/PageMenu.vue";
import { AuthController, EVENT_NAME_AUTH_CHANGED, EVENT_NAME_UNAUTHORIZED } from "@/control/auth";
import { EVENT_NAME_ALBUMS_CHANGED } from "@/control/albums";
import AlbumCreateModal from "../modals/AlbumCreateModal.vue";
import { filterToWords, matchSearchFilter, normalizeString } from "@/utils/normalize";
import { packSearchParams, unPackSearchParams } from "@/utils/search-params";
import type { AlbumListItem } from "@/api/models";
import { EVENT_NAME_RANDOM_PAGE_REFRESH, PagesController } from "@/control/pages";
import { getUniqueStringId } from "@/utils/unique-id";
import { apiAlbumsGetAlbums } from "@/api/api-albums";
import { isTouchDevice } from "@/utils/touch";
import { shuffleArray } from "@/utils/shuffle";
import ThumbImage from "../utils/ThumbImage.vue";

export default defineComponent({
    name: "PageAlbums",
    components: {
        PageMenu,
        AlbumCreateModal,
        ThumbImage,
    },
    props: {
        display: Boolean,
        min: Boolean,
        pageSize: Number,
        displayTitles: Boolean,

        inModal: Boolean,
        removeAlbumsFromList: Object as PropType<Set<number>>,

        rowSize: Number,
        rowSizeMin: Number,
        minItemsSize: Number,
        maxItemsSize: Number,
    },
    emits: ["select-album"],
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

            albumsList: [] as AlbumListItem[],

            filter: this.inModal ? PagesController.AlbumsPageSearch : "",

            order: "desc",
            searchParams: this.inModal ? "" : AppStatus.SearchParams,

            page: 0,
            total: 0,
            totalPages: 0,
            pageItems: [] as AlbumListItem[],

            canWrite: AuthController.CanWrite,

            displayAlbumCreate: false,

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

        this.$listenOnAppEvent(EVENT_NAME_AUTH_CHANGED, this.updateAuthInfo.bind(this));
        this.$listenOnAppEvent(EVENT_NAME_APP_STATUS_CHANGED, this.onAppStatusChanged.bind(this));

        this.$listenOnAppEvent(EVENT_NAME_ALBUMS_CHANGED, this.load.bind(this));

        this.$listenOnAppEvent(EVENT_NAME_RANDOM_PAGE_REFRESH, this.updateList.bind(this));

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
        this.windowResizeObserver.disconnect();
    },
    methods: {
        scrollToTop: function () {
            this.$el.scrollTop = 0;
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

        createAlbum: function () {
            this.displayAlbumCreate = true;
        },

        onNewAlbum: function (albumId: number) {
            AppStatus.ClickOnAlbum(albumId);
        },

        refreshAlbums: function () {
            this.load();
        },

        clearFilter: function () {
            this.filter = "";
            this.changeFilter();
        },

        changeFilter: function () {
            if (!this.inModal) {
                PagesController.AlbumsPageSearch = this.filter;
            }
            this.page = 0;
            this.updateList();
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

            makeNamedApiRequest(this.loadRequestId, apiAlbumsGetAlbums())
                .onSuccess((result) => {
                    this.albumsList = result;
                    clearNamedTimeout(this.loadRequestId);
                    this.loading = false;
                    this.firstLoaded = true;
                    this.updateList();
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

        updateList: function () {
            if (!this.display) {
                return;
            }

            const filter = normalizeString(this.filter + "")
                .trim()
                .toLowerCase();

            let albumsList = this.albumsList.map((a: AlbumListItem) => {
                return {
                    id: a.id,
                    name: a.name,
                    nameLowerCase: a.name.trim().toLowerCase(),
                    size: a.size,
                    thumbnail: a.thumbnail,
                    lm: a.lm,
                };
            });

            if (filter) {
                const filterWords = filterToWords(filter);
                albumsList = albumsList.filter((a) => {
                    return matchSearchFilter(a.name, filter, filterWords) >= 0;
                });
            }

            if (this.removeAlbumsFromList) {
                const blacklist = this.removeAlbumsFromList;
                albumsList = albumsList.filter((a) => {
                    return !blacklist.has(a.id);
                });
            }

            if (this.order === "asc") {
                albumsList = albumsList.sort((a, b) => {
                    if (a.nameLowerCase < b.nameLowerCase) {
                        return -1;
                    } else {
                        return 1;
                    }
                });
            } else if (this.order === "rand") {
                albumsList = shuffleArray(albumsList);
            } else {
                albumsList = albumsList.sort((a, b) => {
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

            this.total = albumsList.length;

            const pageSize = Math.max(1, this.pageSize);

            this.totalPages = Math.floor(this.total / pageSize);

            if (this.total % pageSize > 0) {
                this.totalPages++;
            }

            if (this.order === "rand") {
                this.page = 0;
            } else {
                this.page = Math.max(0, Math.min(this.page, this.totalPages - 1));
            }

            const skip = pageSize * this.page;

            this.pageItems = albumsList.slice(skip, skip + this.pageSize);

            this.onSearchParamsChanged();
        },

        onAppStatusChanged: function () {
            if (this.inModal) {
                return;
            }
            if (AppStatus.SearchParams !== this.searchParams) {
                this.searchParams = AppStatus.SearchParams;
                this.updateSearchParams();
                this.updateList();
            }
        },

        onSearchParamsChanged: function () {
            if (this.inModal) {
                return;
            }
            this.searchParams = packSearchParams(this.page, this.order);
            AppStatus.ChangeSearchParams(this.searchParams);
        },

        changePage: function (p) {
            this.page = p;
            this.onSearchParamsChanged();
            this.updateList();
        },

        updateSearchParams: function () {
            if (this.inModal) {
                return;
            }
            const params = unPackSearchParams(this.searchParams);
            this.page = params.page;
            this.order = params.order;
        },

        getThumbnail(thumb: string) {
            return getAssetURL(thumb);
        },

        goToAlbum: function (album: AlbumListItem, e: Event) {
            if (e) {
                e.preventDefault();
            }

            if (this.inModal) {
                this.$emit("select-album", album.id, () => {
                    this.updateList();
                });
                return;
            }

            AppStatus.ClickOnAlbum(album.id);
        },

        getAlbumURL: function (albumId: number): string {
            return (
                window.location.protocol +
                "//" +
                window.location.host +
                window.location.pathname +
                generateURIQuery({
                    album: albumId + "",
                })
            );
        },

        renderDate: function (ts: number): string {
            return new Date(ts).toLocaleString();
        },

        updateAuthInfo: function () {
            this.canWrite = AuthController.CanWrite;
            this.load();
        },

        updatePageSize: function () {
            this.page = 0;
            this.load();
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

            if (event.key === "+") {
                this.createAlbum();
                return true;
            }

            if (event.key.toUpperCase() === "R") {
                this.refreshAlbums();
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
