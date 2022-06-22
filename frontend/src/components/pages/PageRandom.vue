<template>
  <div class="page-inner" :class="{ hidden: !display }">
     <div v-if="total > 0" class="search-results-options">
        <div class="search-results-option">
          <button type="button" @click="load" class="btn btn-primary btn-sm">
            <i class="fas fa-shuffle"></i> {{ $t("Refresh") }}
          </button>
        </div>
        <div class="search-results-option text-right">
          <select
            class="form-control form-select form-control-full-width"
            v-model="pageSize"
            @change="onPageSizeChanged"
          >
            <option :value="10">10 {{ $t("items per page") }}</option>
            <option :value="20">20 {{ $t("items per page") }}</option>
            <option :value="30">30 {{ $t("items per page") }}</option>
            <option :value="40">40 {{ $t("items per page") }}</option>
            <option :value="50">50 {{ $t("items per page") }}</option>
            <option :value="60">60 {{ $t("items per page") }}</option>
            <option :value="70">70 {{ $t("items per page") }}</option>
            <option :value="80">80 {{ $t("items per page") }}</option>
            <option :value="90">90 {{ $t("items per page") }}</option>
            <option :value="100">100 {{ $t("items per page") }}</option>
          </select>
        </div>
      </div>

    <div class="search-results">
      <div v-if="loading" class="search-results-loading-display">
        <div v-for="f in loadingFiller" :key="f" class="search-result-item">
          <div class="search-result-thumb">
            <div class="search-result-loader">
              <i class="fa fa-spinner fa-spin"></i>
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
          <button type="button" @click="load" class="btn btn-primary">
            <i class="fas fa-sync-alt"></i> {{ $t("Refresh") }}
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
          </div>
          <div class="search-result-title">
            {{ item.title || $t("Untitled") }}
          </div>
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

export default defineComponent({
  name: "PageRandom",
  props: {
    display: Boolean,
  },
  data: function () {
    return {
      loading: false,

      pageSize: 50,
      order: "desc",
      searchParams: AppStatus.SearchParams,

      pageItems: [],
      total: 0,

      loadingFiller: [],
    };
  },
  methods: {
    load: function () {
      Timeouts.Abort("page-random-load");
      Request.Abort("page-random-load");

      if (!this.display) {
        return;
      }

      this.loading = true;

      if (AuthController.Locked) {
        return; // Vault is locked
      }

      Request.Pending(
        "page-random-load",
        SearchAPI.Random("", Date.now(), this.pageSize)
      )
        .onSuccess((result) => {
          this.pageItems = result.page_items;
          this.total = this.pageItems.length;
          this.loading = false;
        })
        .onRequestError((err) => {
          Request.ErrorHandler()
            .add(401, "*", () => {
              AppEvents.Emit("unauthorized", false);
            })
            .add("*", "*", () => {
              // Retry
              Timeouts.Set("page-random-load", 1500, this.$options.loadH);
            })
            .handle(err);
        })
        .onUnexpectedError((err) => {
          console.error(err);
          // Retry
          Timeouts.Set("page-random-load", 1500, this.$options.loadH);
        });
    },

    onPageSizeChanged: function () {
      this.updateLoadingFiller();
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
  },
  mounted: function () {
    this.$options.loadH = this.load.bind(this);
    this.$options.statusChangeH = this.onAppStatusChanged.bind(this);

    AppEvents.AddEventListener("auth-status-changed", this.$options.loadH);
    AppEvents.AddEventListener(
      "app-status-update",
      this.$options.statusChangeH
    );

    this.updateSearchParams();
    this.load();
  },
  beforeUnmount: function () {
    Timeouts.Abort("page-random-load");
    Request.Abort("page-random-load");
    AppEvents.RemoveEventListener("auth-status-changed", this.$options.loadH);
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