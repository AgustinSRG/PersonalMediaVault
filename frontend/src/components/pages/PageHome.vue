<template>
  <div class="page-inner" :class="{ hidden: !display }">
    <div class="search-results"></div>
  </div>
</template>

<script lang="ts">
import { SearchAPI } from "@/api/api-search";
import { AppEvents } from "@/control/app-events";
import { AppStatus } from "@/control/app-status";
import { AuthController } from "@/control/auth";
import { Request } from "@/utils/request";
import { Timeouts } from "@/utils/timeout";
import { defineComponent } from "vue";
import { generateMenuForPages } from "@/utils/manu-make";

export default defineComponent({
  name: "PageHome",
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

      pageMenu: [],
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
          this.updatePageMenu();
        })
        .onRequestError((err) => {
          Request.ErrorHandler()
            .add(401, "*", () => {
              AppEvents.Emit("unauthorized", false);
            })
            .add("*", "*", () => {
              // Retry
              Timeouts.Set("tags-load", 1500, this.$options.loadH);
            })
            .handle(err);
        })
        .onUnexpectedError((err) => {
          console.error(err);
          // Retry
          Timeouts.Set("tags-load", 1500, this.$options.loadH);
        });
    },

    onAppStatusChanged: function () {
      if (AppStatus.CurrentSearch !== this.search) {
        this.search = AppStatus.CurrentSearch;
        this.page = 0;
        this.order = "desc";
        this.pageMenu = [];
        this.load();
        this.onSearchParamsChanged();
      }

      if (AppStatus.SearchParams !== this.searchParams) {
        this.searchParams = AppStatus.SearchParams;
        this.updateSearchParams();
        this.pageMenu = [];
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

    updatePageMenu: function () {
      this.pageMenu = generateMenuForPages(this.page, this.totalPages);
    },

    updateSearchParams: function () {
      const params = AppStatus.UnPackSearchParams(this.searchParams);
      this.page = params.page;
      this.pageSize = params.pageSize;
      this.order = params.order;
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
</style>