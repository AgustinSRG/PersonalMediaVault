<template>
    <div class="page-inner" :class="{ hidden: !display }">
        <div class="search-results" tabindex="-1">
            <div class="search-results-options">
                <div class="search-results-option">
                    <input
                        type="text"
                        class="form-control form-control-full-width auto-focus"
                        autocomplete="off"
                        v-model="filter"
                        :placeholder="$t('Filter by name') + '...'"
                        @input="changeFilter"
                    />
                </div>
                <div class="search-results-option text-right">
                    <button v-if="canWrite" type="button" @click="createAlbum" class="btn btn-primary">
                        <i class="fas fa-plus"></i> {{ $t("Create album") }}
                    </button>
                </div>
            </div>

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

            <div v-if="!loading && total <= 0 && !filter && firstLoaded" class="search-results-msg-display">
                <div class="search-results-msg-icon">
                    <i class="fas fa-box-open"></i>
                </div>
                <div class="search-results-msg-text">
                    {{ $t("This vault does not have any albums yet") }}
                </div>
                <div class="search-results-msg-btn">
                    <button type="button" @click="refreshAlbums" class="btn btn-primary">
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
                    <button type="button" @click="clearFilter" class="btn btn-primary">
                        <i class="fas fa-times"></i> {{ $t("Clear filter") }}
                    </button>
                </div>
            </div>

            <div v-if="!loading && total > 0" class="search-results-final-display">
                <a
                    v-for="item in pageItems"
                    :key="item.id"
                    class="search-result-item clickable"
                    @click="goToAlbum(item, $event)"
                    :href="getAlbumURL(item.id)"
                    target="_blank"
                    rel="noopener noreferrer"
                >
                    <div
                        class="search-result-thumb"
                        :title="
                            (item.name || $t('Untitled album')) + (item.lm ? '\n' + $t('Last modified') + ': ' + renderDate(item.lm) : '')
                        "
                    >
                        <div class="search-result-thumb-inner">
                            <div v-if="!item.thumbnail" class="no-thumb">
                                <i class="fas fa-list-ol"></i>
                            </div>
                            <img
                                v-if="item.thumbnail"
                                :src="getThumbnail(item.thumbnail)"
                                :alt="item.title || $t('Untitled album')"
                                loading="lazy"
                            />
                            <div class="search-result-thumb-tag" :title="$t('Empty')" v-if="item.size == 0">({{ $t("Empty") }})</div>
                            <div class="search-result-thumb-tag" :title="'1' + $t('item')" v-else-if="item.size == 1">
                                1 {{ $t("item") }}
                            </div>
                            <div class="search-result-thumb-tag" :title="item.size + $t('items')" v-else-if="item.size > 1">
                                {{ item.size }} {{ $t("items") }}
                            </div>
                        </div>
                    </div>
                    <div class="search-result-title">
                        {{ item.name || $t("Untitled") }}
                    </div>
                </a>
            </div>

            <PageMenu v-if="total > 0" :page="page" :pages="totalPages" :min="min" @goto="changePage"></PageMenu>

            <div v-if="total > 0" class="search-results-total">{{ $t("Total") }}: {{ total }}</div>
        </div>

        <AlbumCreateModal v-model:display="displayAlbumCreate" @new-album="onNewAlbum"></AlbumCreateModal>
    </div>
</template>

<script lang="ts">
import { AppEvents } from "@/control/app-events";
import { AppStatus } from "@/control/app-status";
import { GenerateURIQuery, GetAssetURL, Request } from "@/utils/request";
import { Timeouts } from "@/utils/timeout";
import { defineComponent, nextTick } from "vue";

import PageMenu from "@/components/utils/PageMenu.vue";
import { AuthController, EVENT_NAME_UNAUTHORIZED } from "@/control/auth";
import { AlbumsAPI } from "@/api/api-albums";
import { AlbumsController } from "@/control/albums";
import { KeyboardManager } from "@/control/keyboard";

import AlbumCreateModal from "../modals/AlbumCreateModal.vue";
import { filterToWords, matchSearchFilter, normalizeString } from "@/utils/normalize";
import { EVENT_NAME_PAGE_SIZE_UPDATED, getPageMaxItems } from "@/control/app-preferences";
import { packSearchParams, unPackSearchParams } from "@/utils/search-params";
import { AlbumListItem } from "@/api/models";

export default defineComponent({
    name: "PageAlbums",
    emits: [],
    components: {
        PageMenu,
        AlbumCreateModal,
    },
    props: {
        display: Boolean,
        min: Boolean,
    },
    data: function () {
        return {
            loading: false,
            firstLoaded: false,

            albumsList: [],

            filter: AlbumsController.AlbumsPageSearch,

            pageSize: getPageMaxItems(),
            order: "desc",
            searchParams: AppStatus.SearchParams,

            page: 0,
            total: 0,
            totalPages: 0,
            pageItems: [],

            loadingFiller: [],

            canWrite: AuthController.CanWrite,

            displayAlbumCreate: false,
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
            AlbumsController.AlbumsPageSearch = this.filter;
            this.page = 0;
            this.updateList();
        },

        load: function () {
            Timeouts.Abort("page-albums-load");
            Request.Abort("page-albums-load");

            if (!this.display) {
                return;
            }

            this.scrollToTop();

            Timeouts.Set("page-albums-load", 330, () => {
                this.loading = true;
            });

            if (AuthController.Locked) {
                return; // Vault is locked
            }

            Request.Pending("page-albums-load", AlbumsAPI.GetAlbums())
                .onSuccess((result) => {
                    this.albumsList = result;
                    Timeouts.Abort("page-albums-load");
                    this.loading = false;
                    this.firstLoaded = true;
                    this.updateList();
                })
                .onRequestError((err) => {
                    Request.ErrorHandler()
                        .add(401, "*", () => {
                            AppEvents.Emit(EVENT_NAME_UNAUTHORIZED);
                        })
                        .add("*", "*", () => {
                            // Retry
                            this.loading = true;
                            Timeouts.Set("page-albums-load", 1500, this._handles.loadH);
                        })
                        .handle(err);
                })
                .onUnexpectedError((err) => {
                    console.error(err);
                    // Retry
                    this.loading = true;
                    Timeouts.Set("page-albums-load", 1500, this._handles.loadH);
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

            if (this.order !== "asc") {
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
            } else {
                albumsList = albumsList.sort((a, b) => {
                    if (a.nameLowerCase < b.nameLowerCase) {
                        return -1;
                    } else {
                        return 1;
                    }
                });
            }

            this.total = albumsList.length;

            const pageSize = Math.max(1, this.pageSize);

            this.totalPages = Math.floor(this.total / pageSize);

            if (this.total % pageSize > 0) {
                this.totalPages++;
            }

            this.page = Math.max(0, Math.min(this.page, this.totalPages - 1));

            const skip = pageSize * this.page;

            this.pageItems = albumsList.slice(skip, skip + this.pageSize);
        },

        onAppStatusChanged: function () {
            if (AppStatus.SearchParams !== this.searchParams) {
                this.searchParams = AppStatus.SearchParams;
                this.updateSearchParams();
                this.updateList();
            }
        },

        onSearchParamsChanged: function () {
            this.searchParams = packSearchParams(this.page, this.order);
            AppStatus.ChangeSearchParams(this.searchParams);
        },

        changePage: function (p) {
            this.page = p;
            this.onSearchParamsChanged();
            this.updateList();
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

        goToAlbum: function (album, e) {
            if (e) {
                e.preventDefault();
            }
            AppStatus.ClickOnAlbum(album.id);
        },

        getAlbumURL: function (albumId: number): string {
            return (
                window.location.protocol +
                "//" +
                window.location.host +
                window.location.pathname +
                GenerateURIQuery({
                    album: albumId + "",
                })
            );
        },

        renderDate: function (ts: number): string {
            return new Date(ts).toLocaleString();
        },

        clickOnEnter: function (event) {
            if (event.key === "Enter") {
                event.preventDefault();
                event.stopPropagation();
                event.target.click();
            }
        },

        updateAuthInfo: function () {
            this.canWrite = AuthController.CanWrite;
        },

        updatePageSize: function () {
            this.pageSize = getPageMaxItems();
            this.updateLoadingFiller();
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
    },
    mounted: function () {
        this._handles = Object.create(null);
        this._handles.loadH = this.load.bind(this);
        this._handles.statusChangeH = this.onAppStatusChanged.bind(this);

        this._handles.handleGlobalKeyH = this.handleGlobalKey.bind(this);
        KeyboardManager.AddHandler(this._handles.handleGlobalKeyH, 20);

        AuthController.AddChangeEventListener(this._handles.loadH);
        AppStatus.AddEventListener(this._handles.statusChangeH);

        this._handles.authUpdateH = this.updateAuthInfo.bind(this);

        AuthController.AddChangeEventListener(this._handles.authUpdateH);

        AppEvents.AddEventListener("albums-list-change", this._handles.loadH);

        this._handles.updatePageSizeH = this.updatePageSize.bind(this);
        AppEvents.AddEventListener(EVENT_NAME_PAGE_SIZE_UPDATED, this._handles.updatePageSizeH);

        this.updateSearchParams();
        this.load();

        if (this.display) {
            this.autoFocus();
        }
    },
    beforeUnmount: function () {
        Timeouts.Abort("page-albums-load");
        Request.Abort("page-albums-load");
        AuthController.RemoveChangeEventListener(this._handles.loadH);
        AppStatus.RemoveEventListener(this._handles.statusChangeH);

        AppEvents.RemoveEventListener("albums-list-change", this._handles.loadH);

        AuthController.RemoveChangeEventListener(this._handles.authUpdateH);

        AppEvents.RemoveEventListener(EVENT_NAME_PAGE_SIZE_UPDATED, this._handles.updatePageSizeH);

        KeyboardManager.RemoveHandler(this._handles.handleGlobalKeyH);
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
