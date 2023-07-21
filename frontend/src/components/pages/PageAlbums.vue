<template>
  <div class="page-inner" :class="{ hidden: !display }">
    <div class="search-results" tabindex="-1">
      <div class="search-results-options">
        <div class="search-results-option">
          <input type="text" class="form-control form-control-full-width" autocomplete="off" v-model="filter" :placeholder="$t('Filter by name') + '...'" @input="changeFilter" />
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

      <div v-if="!loading && total <= 0 && !filter" class="search-results-msg-display">
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

      <div v-if="!loading && total <= 0 && filter" class="search-results-msg-display">
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
        <a v-for="item in pageItems" :key="item.id" class="search-result-item clickable" @click="goToAlbum(item, $event)" :href="getAlbumURL(item.id)" target="_blank" rel="noopener noreferrer">
          <div class="search-result-thumb" :title="(item.name || $t('Untitled album')) + (item.lm ? ('\n' + $t('Last modified') + ': ' + renderDate(item.lm)) : '')">
            <div class="search-result-thumb-inner">
              <div v-if="!item.thumbnail" class="no-thumb">
                <i class="fas fa-list-ol"></i>
              </div>
              <img v-if="item.thumbnail" :src="getThumbnail(item.thumbnail)" :alt="item.title || $t('Untitled album')" loading="lazy" />
              <div class="search-result-thumb-tag" :title="$t('Empty')" v-if="item.size == 0">
                ({{ $t("Empty") }})
              </div>
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

      <div v-if="total > 0" class="search-results-total">
        {{ $t("Total") }}: {{ total }}
      </div>

      <div v-if="total > 0" class="search-results-options">
        <div class="search-results-option">
          <select class="form-control form-select form-control-full-width" v-model="order" @change="onOrderChanged">
            <option :value="'desc'">{{ $t("Order by last modified date") }}</option>
            <option :value="'asc'">{{ $t("Order alphabetically") }}</option>
          </select>
        </div>
        <div class="search-results-option text-right">
          <select class="form-control form-select form-control-full-width" v-model="pageSize" @change="onPageSizeChanged">
            <option v-for="po in pageSizeOptions" :key="po" :value="po">
              {{ po }} {{ $t("items per page") }}
            </option>
          </select>
        </div>
      </div>
    </div>

    <AlbumCreateModal v-model:display="displayAlbumCreate" @new-album="onNewAlbum"></AlbumCreateModal>
  </div>
</template>

<script lang="ts">
import { AppEvents } from "@/control/app-events";
import { AppStatus } from "@/control/app-status";
import { GenerateURIQuery, GetAssetURL, Request } from "@/utils/request";
import { Timeouts } from "@/utils/timeout";
import { defineComponent } from "vue";

import PageMenu from "@/components/utils/PageMenu.vue";
import { AuthController } from "@/control/auth";
import { AlbumsAPI } from "@/api/api-albums";
import { AlbumEntry } from "@/control/albums";
import { KeyboardManager } from "@/control/keyboard";

import AlbumCreateModal from "../modals/AlbumCreateModal.vue";
import { filterToWords, matchSearchFilter, normalizeString } from "@/utils/normalize";

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
            loading: true,

            albumsList: [],

            filter: "",

            pageSize: 50,
            order: "desc",
            searchParams: AppStatus.SearchParams,

            page: 0,
            total: 0,
            totalPages: 0,
            pageItems: [],

            loadingFiller: [],

            pageSizeOptions: [],

            canWrite: AuthController.CanWrite,

            displayAlbumCreate: false,
        };
    },
    methods: {
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
            this.page = 0;
            this.updateList();
        },

        load: function () {
            Timeouts.Abort("page-albums-load");
            Request.Abort("page-albums-load");

            if (!this.display) {
                return;
            }

            this.loading = true;

            if (AuthController.Locked) {
                return; // Vault is locked
            }

            Request.Pending("page-albums-load", AlbumsAPI.GetAlbums())
                .onSuccess((result) => {
                    this.albumsList = result;
                    this.loading = false;
                    this.updateList();
                })
                .onRequestError((err) => {
                    Request.ErrorHandler()
                        .add(401, "*", () => {
                            AppEvents.Emit("unauthorized", false);
                        })
                        .add("*", "*", () => {
                            // Retry
                            Timeouts.Set("page-albums-load", 1500, this.$options.loadH);
                        })
                        .handle(err);
                })
                .onUnexpectedError((err) => {
                    console.error(err);
                    // Retry
                    Timeouts.Set("page-albums-load", 1500, this.$options.loadH);
                });
        },

        updateList: function () {
            if (!this.display) {
                return;
            }

            let filter = normalizeString(this.filter + "").trim().toLowerCase();

            let albumsList = this.albumsList.map((a: AlbumEntry) => {
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
                albumsList = albumsList.filter(a => {
                    return matchSearchFilter(a.name, filter, filterWords)  >= 0;
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

        onPageSizeChanged: function () {
            this.updateLoadingFiller();
            this.page = 0;
            this.updateList();
            this.onSearchParamsChanged();
        },

        onOrderChanged: function () {
            this.page = 0;
            this.updateList();
            this.onSearchParamsChanged();
        },

        onAppStatusChanged: function () {
            if (AppStatus.SearchParams !== this.searchParams) {
                this.searchParams = AppStatus.SearchParams;
                this.updateSearchParams();
                this.updateList();
            }
        },

        onSearchParamsChanged: function () {
            this.searchParams = AppStatus.PackSearchParams(
                this.page,
                this.pageSize,
                this.order
            );
            AppStatus.ChangeSearchParams(this.searchParams);
        },

        changePage: function (p) {
            this.page = p;
            this.onSearchParamsChanged();
            this.updateList();
        },

        updateSearchParams: function () {
            const params = AppStatus.UnPackSearchParams(this.searchParams);
            this.page = params.page;
            this.pageSize = params.pageSize;
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
            return (new Date(ts)).toLocaleString();
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

        handleGlobalKey: function (event: KeyboardEvent): boolean {
            if (
                AuthController.Locked ||
        !AppStatus.IsPageVisible() ||
        !this.display ||
        !event.key ||
        event.ctrlKey
            ) {
                return false;
            }

            if (event.key === "PageDown") {
                if (this.page > 0) {
                    this.changePage(this.page - 1);
                }
                return true;
            }

            if (event.key === "PageUp") {
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
        this.$options.loadH = this.load.bind(this);
        this.$options.statusChangeH = this.onAppStatusChanged.bind(this);

        this.$options.handleGlobalKeyH = this.handleGlobalKey.bind(this);
        KeyboardManager.AddHandler(this.$options.handleGlobalKeyH, 20);

        AppEvents.AddEventListener("auth-status-changed", this.$options.loadH);
        AppEvents.AddEventListener(
            "app-status-update",
            this.$options.statusChangeH
        );

        this.$options.authUpdateH = this.updateAuthInfo.bind(this);

        AppEvents.AddEventListener(
            "auth-status-changed",
            this.$options.authUpdateH
        );

        AppEvents.AddEventListener("albums-list-change", this.$options.loadH);

        for (let i = 1; i <= 20; i++) {
            this.pageSizeOptions.push(5 * i);
        }

        this.updateSearchParams();
        this.load();
    },
    beforeUnmount: function () {
        Timeouts.Abort("page-albums-load");
        Request.Abort("page-albums-load");
        AppEvents.RemoveEventListener("auth-status-changed", this.$options.loadH);
        AppEvents.RemoveEventListener(
            "app-status-update",
            this.$options.statusChangeH
        );

        AppEvents.RemoveEventListener("albums-list-change", this.$options.loadH);

        AppEvents.RemoveEventListener(
            "auth-status-changed",
            this.$options.authUpdateH
        );

        KeyboardManager.RemoveHandler(this.$options.handleGlobalKeyH);
    },
    watch: {
        display: function () {
            this.load();
        },
    },
});
</script>
