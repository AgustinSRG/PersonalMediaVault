<template>
  <div class="page-inner" :class="{ hidden: !display }">
    <div class="search-results">
      <div class="search-results-options">
        <div class="search-results-option">
          <input type="text" class="form-control form-control-full-width" autocomplete="off" v-model="filter" :placeholder="$t('Filter by name') + '...'" @input="changeFilter">
        </div>
        <div class="search-results-option text-right">
           <button type="button" @click="createAlbum" class="btn btn-primary btn-sm">
            <i class="fas fa-plus"></i> {{ $t("Create album") }}
          </button>
        </div>
      </div>

      <PageMenu
        v-if="total > 0"
        :page="page"
        :pages="totalPages"
        :min="min"
        @goto="changePage"
      ></PageMenu>

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

      <div
        v-if="!loading && total <= 0 && !filter"
        class="search-results-msg-display"
      >
        <div class="search-results-msg-icon">
          <i class="fas fa-box-open"></i>
        </div>
        <div class="search-results-msg-text">
          {{ $t("This vault does not have any albums yet") }}
        </div>
        <div class="search-results-msg-btn">
          <button type="button" @click="refreshAlbums" class="btn btn-primary btn-sm">
            <i class="fas fa-sync-alt"></i> {{ $t("Refresh") }}
          </button>
        </div>
      </div>

      <div
        v-if="!loading && total <= 0 && filter"
        class="search-results-msg-display"
      >
        <div class="search-results-msg-icon">
          <i class="fas fa-box-open"></i>
        </div>
        <div class="search-results-msg-text">
          {{ $t("Could not find any albums matching your filter") }}
        </div>
        <div class="search-results-msg-btn">
          <button
            type="button"
            @click="clearFilter"
            class="btn btn-primary btn-sm"
          >
            <i class="fas fa-times"></i> {{ $t("Clear filter") }}
          </button>
        </div>
      </div>

      <div v-if="!loading && total > 0" class="search-results-final-display">
        <div
          v-for="(item, i) in pageItems"
          :key="i"
          class="search-result-item clickable"
          tabindex="0"
          @click="goToAlbum(item)"
        >
          <div
            class="search-result-thumb"
            :title="item.name || $t('Untitled album')"
          >
            <div class="search-result-thumb-inner">
              <div v-if="!item.thumbnail" class="no-thumb">
                <i class="fas fa-list-ol"></i>
              </div>
              <img
                v-if="item.thumbnail"
                :src="getThumbnail(item.thumbnail)"
                :alt="item.title || $t('Untitled album')"
              />
            </div>
          </div>
          <div class="search-result-title">
            {{ item.name || $t("Untitled") }}
          </div>
        </div>
      </div>

      <PageMenu
        v-if="total > 0"
        :page="page"
        :pages="totalPages"
        :min="min"
        @goto="changePage"
      ></PageMenu>

      <div v-if="total > 0" class="search-results-total">
        {{ $t("Total") }}: {{ total }}
      </div>

      <div v-if="total > 0" class="search-results-options">
        <div class="search-results-option">
          <select
            class="form-control form-select form-control-full-width"
            v-model="order"
            @change="onOrderChanged"
          >
            <option :value="'desc'">{{ $t("Order alphabetically") }}</option>
            <option :value="'asc'">{{ $t("Show by creation order") }}</option>
          </select>
        </div>
        <div class="search-results-option text-right">
          <select
            class="form-control form-select form-control-full-width"
            v-model="pageSize"
            @change="onPageSizeChanged"
          >
            <option v-for="po in pageSizeOptions" :key="po" :value="po">
              {{ po }} {{ $t("items per page") }}
            </option>
          </select>
        </div>
      </div>
    </div>
  </div>
</template>

<script lang="ts">
import { AlbumsController } from "@/control/albums";
import { AppEvents } from "@/control/app-events";
import { AppStatus } from "@/control/app-status";
import { GetAssetURL, Request } from "@/utils/request";
import { Timeouts } from "@/utils/timeout";
import { defineComponent } from "vue";

export default defineComponent({
  name: "PageAlbums",
  emits: ['album-create'],
  props: {
    display: Boolean,
  },
  data: function () {
    return {
      loading: AlbumsController.Loading,

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
    };
  },
  methods: {
    onAlbumsLoading: function (l) {
      this.loading = l;
    },

    createAlbum: function () {
      this.$emit("album-create")
    },

    refreshAlbums: function () {
      AlbumsController.Load();
    },

    clearFilter: function () {
      this.filter = "";
      this.changeFilter();
    },

    changeFilter: function () {
      this.page = 0;
      this.load();
    },

    load: function () {
      if (!this.display) {
        return;
      }

      var filter = (this.filter + "").toLowerCase();

      var albumsList = AlbumsController.GetAlbumsListCopy();

      if (filter) {
        albumsList = albumsList.filter((a) => {
          return !filter || a.nameLowerCase.indexOf(filter) >= 0;
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
      }

      this.total = albumsList.length;

      var pageSize = Math.max(1, this.pageSize);

      this.totalPages = Math.floor(this.total / pageSize);

      if (this.total % pageSize > 0) {
        this.totalPages++;
      }

      this.page = Math.max(0, Math.min(this.page, this.totalPages - 1));

      var skip = pageSize * this.page;

      this.pageItems = albumsList.slice(skip, skip + this.pageSize);
    },

    onPageSizeChanged: function () {
      this.updateLoadingFiller();
      this.page = 0;
      this.load();
      this.onSearchParamsChanged();
    },

    onOrderChanged: function () {
      this.page = 0;
      this.load();
      this.onSearchParamsChanged();
    },

    onAppStatusChanged: function () {
      if (AppStatus.SearchParams !== this.searchParams) {
        this.searchParams = AppStatus.SearchParams;
        this.updateSearchParams();
        this.load();
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
      this.load();
    },

    updateSearchParams: function () {
      const params = AppStatus.UnPackSearchParams(this.searchParams);
      this.page = params.page;
      this.pageSize = params.pageSize;
      this.order = params.order;
      this.updateLoadingFiller();
    },

    updateLoadingFiller: function () {
      var filler = [];

      for (var i = 0; i < this.pageSize; i++) {
        filler.push(i);
      }

      this.loadingFiller = filler;
    },

    getThumbnail(thumb: string) {
      return GetAssetURL(thumb);
    },

    goToAlbum: function (album) {
      AppStatus.ClickOnAlbum(
        album.id,
        album.list.length > 0 ? album.list[0] : -1
      );
    },
  },
  mounted: function () {
    this.$options.loadH = this.load.bind(this);
    this.$options.loadingH = this.onAlbumsLoading.bind(this);
    this.$options.statusChangeH = this.onAppStatusChanged.bind(this);

    AppEvents.AddEventListener("auth-status-changed", this.$options.loadH);
    AppEvents.AddEventListener(
      "app-status-update",
      this.$options.statusChangeH
    );

    AppEvents.AddEventListener("albums-loading", this.$options.loadingH);
    AppEvents.AddEventListener("albums-update", this.$options.loadH);

    for (let i = 1; i <= 20; i++) {
      this.pageSizeOptions.push(5 * i);
    }

    this.updateSearchParams();
    this.load();
  },
  beforeUnmount: function () {
    Timeouts.Abort("page-home-load");
    Request.Abort("page-home-load");
    AppEvents.RemoveEventListener("auth-status-changed", this.$options.loadH);
    AppEvents.RemoveEventListener(
      "app-status-update",
      this.$options.statusChangeH
    );

    AppEvents.RemoveEventListener("albums-loading", this.$options.loadingH);
    AppEvents.RemoveEventListener("albums-update", this.$options.loadH);
  },
  watch: {
    display: function () {
      this.load();
    },
  },
});
</script>

<style>
</style>