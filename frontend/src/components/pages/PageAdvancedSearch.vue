<template>
  <div class="page-inner" :class="{ hidden: !display }">
    <form class="adv-search-form" @submit="startSearch">
      <div class="form-group">
        <label>{{ $t("Title must contain") }}:</label>
        <input
          type="text"
          autoxomplete="off"
          maxlength="255"
          v-model="textSearch"
          class="form-control form-control-full-width"
        />
      </div>
      <div class="form-group">
        <label>{{ $t("Media type") }}:</label>
        <select
          class="form-control form-select form-control-full-width"
          v-model="type"
        >
          <option :value="0">{{ $t("Any media") }}</option>
          <option :value="1">{{ $t("Images") }}</option>
          <option :value="2">{{ $t("Videos") }}</option>
          <option :value="3">{{ $t("Audios") }}</option>
        </select>
      </div>

      <div class="form-group">
        <label>{{ $t("Order") }}:</label>
        <select
          class="form-control form-select form-control-full-width"
          v-model="order"
        >
          <option :value="'desc'">{{ $t("Show most recent") }}</option>
          <option :value="'asc'">{{ $t("Show oldest") }}</option>
        </select>
      </div>
      <div class="form-group">
        <label>{{ $t("Limit results") }}:</label>
        <select
          class="form-control form-select form-control-full-width"
          v-model="pageSize"
        >
          <option v-for="po in pageSizeOptions" :key="po" :value="po">
            {{ po }} {{ $t("results max") }}
          </option>
        </select>
      </div>

      <div class="form-group">
        <button
          v-if="!loading"
          type="submit"
          class="btn btn-primary btn-sm"
        >
          <i class="fas fa-search"></i> {{ $t("Search") }}
        </button>
        <button
          v-if="loading"
          type="button"
          class="btn btn-primary btn-sm"
          disabled
        >
          <i class="fa fa-spinner fa-spin"></i> {{ $t("Searching") }}... ({{
            cssProgress(progress)
          }})
        </button>
        <button
          v-if="loading"
          type="button"
          class="btn btn-primary btn-sm"
          @click="cancel"
        >
          <i class="fas fa-times"></i> {{ $t("Cancel") }}
        </button>
      </div>
    </form>

    <div class="search-results">
      <div
        v-if="!loading && pageItems.length > 0"
        class="search-results-final-display"
      >
        <div
          v-for="(item, i) in pageItems"
          :key="i"
          class="search-result-item clickable"
          tabindex="0"
          @keydown="clickOnEnter"
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
              <div
                class="search-result-thumb-tag"
                v-if="item.type === 2 || item.type === 3"
              >
                {{ renderTime(item.duration) }}
              </div>
            </div>
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
import { MediaEntry } from "@/control/media";
import { GetAssetURL, Request } from "@/utils/request";
import { renderTimeSeconds } from "@/utils/time-utils";
import { Timeouts } from "@/utils/timeout";
import { defineComponent } from "vue";

export default defineComponent({
  name: "PageAdvancedSearch",
  props: {
    display: Boolean,
  },
  data: function () {
    return {
      loading: false,

      pageSize: 50,
      order: "desc",
      textSearch: "",
      type: 0,

      pageItems: [],
      page: 0,
      totalPages: 0,
      progress: 0,

      finished: true,

      pageSizeOptions: [],
    };
  },
  methods: {
    load: function () {
      Timeouts.Abort("page-advsearch-load");
      Request.Abort("page-advsearch-load");

      if (!this.display || this.finished) {
        return;
      }

      this.loading = true;

      if (AuthController.Locked) {
        return; // Vault is locked
      }

      Request.Pending(
        "page-advsearch-load",
        SearchAPI.Search("", this.order, this.page, this.pageSize)
      )
        .onSuccess((result) => {
          this.filterElements(result.page_items);
          this.page = result.page_index;
          this.totalPages = result.page_count;
          this.progress =
            ((this.page + 1) / Math.max(1, this.totalPages)) * 100;
          if (this.pageItems.length >= this.pageSize) {
            this.loading = false;
            this.finished = true;
          } else if (this.page < this.totalPages - 1) {
            this.page++;
            this.load();
          } else {
            this.loading = false;
            this.finished = true;
          }
        })
        .onRequestError((err) => {
          Request.ErrorHandler()
            .add(401, "*", () => {
              AppEvents.Emit("unauthorized", false);
            })
            .add("*", "*", () => {
              // Retry
              Timeouts.Set("page-advsearch-load", 1500, this.$options.loadH);
            })
            .handle(err);
        })
        .onUnexpectedError((err) => {
          console.error(err);
          // Retry
          Timeouts.Set("page-advsearch-load", 1500, this.$options.loadH);
        });
    },

    filterElements: function (results: MediaEntry[]) {
      const filterText = this.textSearch.toLowerCase();
      const filterType = this.type;
      for (let e of results) {
        if (this.pageItems.length >= this.pageSize) {
          return;
        }

        if (filterText) {
          if (!e.title.toLowerCase().includes(filterText)) {
            continue;
          }
        }

        if (filterType) {
          if (e.type !== filterType) {
            continue;
          }
        }

        this.pageItems.push(e);
      }
    },

    startSearch: function (event) {
      if (event) {
        event.preventDefault();
      }
      this.loading = true;
      this.pageItems = [];
      this.page = 0;
      this.totalPages = 0;
      this.progress = 0;
      this.finished = false;
      this.load();
    },

    cancel: function () {
      Timeouts.Abort("page-advsearch-load");
      Request.Abort("page-advsearch-load");
      this.loading = false;
      this.finished = true;
    },

    resetSearch: function () {
      Timeouts.Abort("page-advsearch-load");
      Request.Abort("page-advsearch-load");
      this.pageItems = [];
      this.page = 0;
      this.totalPages = 0;
      this.progress = 0;
      this.loading = false;
      this.finished = true;
    },

    goToMedia: function (mid) {
      AppStatus.ClickOnMedia(mid, true);
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
  },
  mounted: function () {
    for (let i = 1; i <= 20; i++) {
      this.pageSizeOptions.push(5 * i);
    }

    this.$options.loadH = this.load.bind(this);
    this.$options.resetH = this.resetSearch.bind(this);

    AppEvents.AddEventListener("auth-status-changed", this.$options.loadH);
    AppEvents.AddEventListener("media-delete", this.$options.resetH);
    AppEvents.AddEventListener("media-meta-change", this.$options.resetH);
  },
  beforeUnmount: function () {
    Timeouts.Abort("page-advsearch-load");
    Request.Abort("page-advsearch-load");

    AppEvents.RemoveEventListener("auth-status-changed", this.$options.loadH);
    AppEvents.RemoveEventListener("media-delete", this.$options.resetH);
    AppEvents.RemoveEventListener("media-meta-change", this.$options.resetH);
  },
  watch: {
    display: function () {
      this.load();
    },
  },
});
</script>

<style>
.adv-search-form {
  padding: 1rem;
}
</style>