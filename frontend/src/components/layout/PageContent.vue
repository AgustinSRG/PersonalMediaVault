<template>
  <div class="page-content" tabindex="-1">
    <div class="page-header">
      <button
        type="button"
        :title="$t('Expand')"
        class="page-header-btn page-expand-btn"
        @click="expandPage"
      >
        <i class="fas fa-chevron-left"></i>
      </button>
      <div class="page-title">
        <i :class="getIcon(page)"></i> {{ renderTitle(page, search) }}
      </div>
      <button
        type="button"
        :title="$t('Close')"
        class="page-header-btn page-close-btn"
        @click="closePage"
      >
        <i class="fas fa-times"></i>
      </button>
    </div>

    <PageHome
      v-if="isDisplayed && page === 'home'"
      :display="isDisplayed && page === 'home'"
      :min="min"
    ></PageHome>
    <PageSearch
      v-if="isDisplayed && page === 'search'"
      :display="isDisplayed && page === 'search'"
      :min="min"
    ></PageSearch>
    <PageUpload
      v-if="isDisplayed && page === 'upload'"
      :display="isDisplayed && page === 'upload'"
    ></PageUpload>
    <PageRandom
      v-if="isDisplayed && page === 'random'"
      :display="isDisplayed && page === 'random'"
    ></PageRandom>
    <PageAdvancedSearch
      v-if="isDisplayed && page === 'adv-search'"
      :display="isDisplayed && page === 'adv-search'"
      :inModal="false"
      :noAlbum="-1"
    ></PageAdvancedSearch>
    <PageAlbums
      v-if="isDisplayed && page === 'albums'"
      :display="isDisplayed && page === 'albums'"
      :min="min"
    ></PageAlbums>
  </div>
</template>

<script lang="ts">
import { AppEvents } from "@/control/app-events";
import { AppStatus } from "@/control/app-status";
import { defineComponent } from "vue";

import PageHome from "../pages/PageHome.vue";
import PageSearch from "../pages/PageSearch.vue";
import PageUpload from "../pages/PageUpload.vue";
import PageRandom from "../pages/PageRandom.vue";
import PageAlbums from "../pages/PageAlbums.vue";
import PageAdvancedSearch from "../pages/PageAdvancedSearch.vue";
import { AuthController } from "@/control/auth";
import { KeyboardManager } from "@/control/keyboard";

export default defineComponent({
  components: {
    PageHome,
    PageSearch,
    PageAlbums,
    PageUpload,
    PageRandom,
    PageAdvancedSearch,
  },
  name: "PageContent",
  emits: [],
  props: {
    min: Boolean,
  },
  data: function () {
    return {
      isDisplayed:
        (AppStatus.CurrentMedia < 0 || AppStatus.ListSplitMode) &&
        AppStatus.CurrentAlbum < 0,
      page: AppStatus.CurrentPage,
      search: AppStatus.CurrentSearch,
    };
  },
  methods: {
    updatePage: function () {
      this.page = AppStatus.CurrentPage;
      this.search = AppStatus.CurrentSearch;
      this.isDisplayed =
        (AppStatus.CurrentMedia < 0 || AppStatus.ListSplitMode) &&
        AppStatus.CurrentAlbum < 0;
    },

    expandPage: function () {
      AppStatus.ExpandPage();
    },

    closePage: function () {
      AppStatus.ClosePage();
    },

    renderTitle: function (p, s) {
      switch (p) {
        case "home":
          return this.$t("Home");
        case "search":
          return this.$t("Search results") + ": " + s;
        case "adv-search":
          return this.$t("Advanced search");
        case "upload":
          return this.$t("Upload media");
        case "albums":
          return this.$t("Albums list");
        case "random":
          return this.$t("Random results");
        default:
          return "";
      }
    },

    getIcon: function (p) {
      switch (p) {
        case "home":
          return "fas fa-home";
        case "search":
        case "adv-search":
          return "fas fa-search";
        case "upload":
          return "fas fa-upload";
        case "albums":
          return "fas fa-list";
        case "random":
          return "fas fa-shuffle";
        default:
          return "";
      }
    },

    handleGlobalKey: function (event: KeyboardEvent): boolean {
      if (
        AuthController.Locked ||
        !AppStatus.IsPageVisible() ||
        !event.key ||
        event.ctrlKey
      ) {
        return false;
      }

      if (event.key.toUpperCase() === "Q") {
        this.closePage();
        return true;
      }

      if (event.key.toUpperCase() === "BACKSPACE") {
        this.expandPage();
        return true;
      }

      return false;
    },
  },
  mounted: function () {
    this.$options.pageUpdater = this.updatePage.bind(this);

    AppEvents.AddEventListener("app-status-update", this.$options.pageUpdater);

    this.$options.handleGlobalKeyH = this.handleGlobalKey.bind(this);
    KeyboardManager.AddHandler(this.$options.handleGlobalKeyH, 10);
  },
  beforeUnmount: function () {
    AppEvents.RemoveEventListener(
      "app-status-update",
      this.$options.pageUpdater
    );
    KeyboardManager.RemoveHandler(this.$options.handleGlobalKeyH);
  },
});
</script>
