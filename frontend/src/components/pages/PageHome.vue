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
          <option :value="'desc'">{{$t('Show most recent')}}</option>
          <option :value="'asc'">{{$t('Show oldest')}}</option>
          </select>
        </div>
        <div class="search-results-option text-right">
          <select
            class="form-control form-select form-control-full-width"
            v-model="pageSize"
            @change="onPageSizeChanged"
          >
          <option :value="10">10 {{$t('items per page')}}</option>
         <option :value="20">20 {{$t('items per page')}}</option>
         <option :value="30">30 {{$t('items per page')}}</option>
         <option :value="40">40 {{$t('items per page')}}</option>
         <option :value="50">50 {{$t('items per page')}}</option>
         <option :value="60">60 {{$t('items per page')}}</option>
         <option :value="70">70 {{$t('items per page')}}</option>
         <option :value="80">80 {{$t('items per page')}}</option>
         <option :value="90">90 {{$t('items per page')}}</option>
          <option :value="100">100 {{$t('items per page')}}</option>
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

export default defineComponent({
  name: "PageHome",
  components: {
    PageMenu,
  },
  props: {
    display: Boolean,
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

      min: false,
    };
  },
  methods: {
    load: function () {
      Timeouts.Abort("page-home-load");
      Request.Abort("page-home-load");

      if (!this.display) {
        return;
      }

      this.loading = true;

      if (AuthController.Locked) {
        return; // Vault is locked
      }

      Request.Pending(
        "page-home-load",
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
              Timeouts.Set("page-home-load", 1500, this.$options.loadH);
            })
            .handle(err);
        })
        .onUnexpectedError((err) => {
          console.error(err);
          // Retry
          Timeouts.Set("page-home-load", 1500, this.$options.loadH);
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
    Timeouts.Abort("page-home-load");
    Request.Abort("page-home-load");
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
.search-results {
  display: flex;
  flex-direction: column;
}

.search-results-loading-display {
  display: flex;
  flex-wrap: wrap;
  justify-content: center;
}

.search-results-final-display {
  display: flex;
  flex-wrap: wrap;
  justify-content: center;
}

.search-result-item {
  width: 232px;
  padding: 24px;
}

.search-result-item.clickable {
  cursor: pointer;
}

.search-result-item.clickable:hover {
  opacity: 0.7;
}

.search-result-thumb {
  width: 184px;
  height: 184px;
  border-radius: 4px;
  overflow: hidden;
  background: rgba(255, 255, 255, 0.1);
  display: flex;
  justify-content: center;
  align-items: center;
}

.search-result-thumb .no-thumb {
  opacity: 0.7;
  font-size: 24px;
}

.search-result-thumb img {
  width: 100%;
  height: 100%;
}

.search-result-title {
  padding-top: 0.5rem;
  line-height: 22px;
  font-size: 14px;
  font-weight: bold;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.search-results-msg-display {
  padding: 1rem;
  display: flex;
  flex-direction: column;
  align-items: center;
  opacity: 0.7;
}

.search-results-msg-icon {
  font-size: 48px;
}

.search-results-msg-text {
  padding-top: 1rem;
  font-size: x-large;
}

.search-results-msg-btn {
  padding-top: 1rem;
}

.search-results-total {
  padding-top: 0.5rem;
  font-size: small;
  text-align: center;
}

.search-results-options {
  padding-top: 1rem;
  padding-bottom: 1rem;
  width: 100%;
  display: flex;
  flex-direction: row;
}

.search-results-option {
  width: 50%;
  padding: 0.5rem 1rem;
}
</style>