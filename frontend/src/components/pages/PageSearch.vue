<template>
  <div class="page-inner" :class="{ hidden: !display }">
    <div class="search-results">
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

      <div v-if="!loading && total <= 0" class="search-results-msg-display">
        <div class="search-results-msg-icon"><i class="fas fa-search"></i></div>
        <div class="search-results-msg-text">
          {{ $t("Could not find any result") }}
        </div>
        <div class="search-results-msg-btn">
          <button type="button" @click="load" class="btn btn-primary btn-sm">
            <i class="fas fa-sync-alt"></i> {{ $t("Refresh") }}
          </button>
        </div>
        <div class="search-results-msg-btn">
          <button type="button" @click="clearSearch" class="btn btn-primary btn-sm">
            <i class="fas fa-times"></i> {{ $t("Clear Search") }}
          </button>
        </div>
      </div>

      <div v-if="!loading && total > 0" class="search-results-final-display">
        <div
          v-for="(item, i) in pageItems"
          :key="i"
          class="search-result-item clickable"
          tabindex="0"
          @click="goToMedia(item.id)"
        >
          <div
            class="search-result-thumb"
            :title="item.title || $t('Untitled')"
          >
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
              />
              <div class="search-result-thumb-tag" v-if="item.type === 2 || item.type === 3">{{ renderTime(item.duration) }}</div>
            </div>
          </div>
          <div class="search-result-title">
            {{ item.title || $t("Untitled") }}
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
            <option :value="'desc'">{{ $t("Show most recent") }}</option>
            <option :value="'asc'">{{ $t("Show oldest") }}</option>
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
import { SearchAPI } from "@/api/api-search";
import { AppEvents } from "@/control/app-events";
import { AppStatus } from "@/control/app-status";
import { AuthController } from "@/control/auth";
import { GetAssetURL, Request } from "@/utils/request";
import { Timeouts } from "@/utils/timeout";
import { defineComponent } from "vue";

import PageMenu from "@/components/utils/PageMenu.vue";
import { renderTimeSeconds } from "@/utils/time-utils";

export default defineComponent({
  name: "PageSearch",
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
    load: function () {
      Timeouts.Abort("page-search-load");
      Request.Abort("page-search-load");

      if (!this.display) {
        return;
      }

      this.loading = true;

      if (AuthController.Locked) {
        return; // Vault is locked
      }

      Request.Pending(
        "page-search-load",
        SearchAPI.Search(this.search, this.order, this.page, this.pageSize)
      )
        .onSuccess((result) => {
          this.pageItems = result.page_items;
          this.page = result.page_index;
          this.totalPages = result.page_count;
          this.total = result.total_count;
          this.loading = false;
        })
        .onRequestError((err) => {
          Request.ErrorHandler()
            .add(401, "*", () => {
              AppEvents.Emit("unauthorized", false);
            })
            .add("*", "*", () => {
              // Retry
              Timeouts.Set("page-search-load", 1500, this.$options.loadH);
            })
            .handle(err);
        })
        .onUnexpectedError((err) => {
          console.error(err);
          // Retry
          Timeouts.Set("page-search-load", 1500, this.$options.loadH);
        });
    },

    onOrderChanged: function () {
      this.page = 0;
      this.load();
      this.onSearchParamsChanged();
    },

    onPageSizeChanged: function () {
      this.updateLoadingFiller();
      this.page = 0;
      this.load();
      this.onSearchParamsChanged();
    },

    onAppStatusChanged: function () {
      if (AppStatus.CurrentSearch !== this.search) {
        this.search = AppStatus.CurrentSearch;
        this.page = 0;
        this.order = "desc";
        this.load();
        this.onSearchParamsChanged();
      }

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

    goToMedia: function (mid) {
      AppStatus.ClickOnMedia(mid);
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

    clearSearch: function () {
      AppStatus.GoToSearch("");
    },

    renderTime: function (s: number): string {
      return renderTimeSeconds(s);
    },
  },
  mounted: function () {
    this.$options.loadH = this.load.bind(this);
    this.$options.statusChangeH = this.onAppStatusChanged.bind(this);

    AppEvents.AddEventListener("auth-status-changed", this.$options.loadH);
    AppEvents.AddEventListener("media-meta-change", this.$options.loadH);
    AppEvents.AddEventListener("media-delete", this.$options.loadH);
    AppEvents.AddEventListener(
      "app-status-update",
      this.$options.statusChangeH
    );

    for (let i = 1; i <= 20; i++) {
      this.pageSizeOptions.push(5 * i);
    }

    this.updateSearchParams();
    this.load();
  },
  beforeUnmount: function () {
    Timeouts.Abort("page-search-load");
    Request.Abort("page-search-load");
    AppEvents.RemoveEventListener("auth-status-changed", this.$options.loadH);
    AppEvents.RemoveEventListener("media-meta-change", this.$options.loadH);
    AppEvents.RemoveEventListener("media-delete", this.$options.loadH);
    AppEvents.RemoveEventListener(
      "app-status-update",
      this.$options.statusChangeH
    );
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