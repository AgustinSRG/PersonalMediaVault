<template>
  <div class="page-content">
    <div class="page-header">
      <button type="button" :title="$t('Expand')" class="page-header-btn page-expand-btn" @click="expandPage"><i class="fas fa-chevron-left"></i></button>
      <div class="page-title"><i :class="getIcon(page, search)"></i> {{renderTitle(page, search)}}</div>
      <button type="button" :title="$t('Close')" class="page-header-btn page-close-btn" @click="closePage"><i class="fas fa-times"></i></button>
    </div>

    <PageHome :display="page === 'home'"></PageHome>
    <PageUpload :display="page === 'upload'"></PageUpload>
    <PageRandom :display="page === 'random'"></PageRandom>
    <PageAlbums :display="page === 'albums'"></PageAlbums>
  </div>
</template>

<script lang="ts">
import { AppEvents } from "@/control/app-events";
import { AppStatus } from "@/control/app-status";
import { defineComponent } from "vue";

import PageHome from "../pages/PageHome.vue";
import PageUpload from "../pages/PageUpload.vue";
import PageRandom from "../pages/PageRandom.vue";
import PageAlbums from "../pages/PageAlbums.vue";

export default defineComponent({
  components: {
    PageHome,
    PageAlbums,
    PageUpload,
    PageRandom,
  },
  name: "PageContent",
  data: function () {
    return {
      page: AppStatus.CurrentPage,
      search: AppStatus.CurrentSearch,
    };
  },
  methods: {
    updatePage: function () {
      this.page = AppStatus.CurrentPage;
      this.search = AppStatus.CurrentSearch;
    },

    expandPage: function () {
      AppStatus.GoToPage(this.page);
    },

    closePage: function () {
      AppStatus.ClosePage();
    },

    renderTitle: function (p, s) {
      switch (p) {
        case "home":
          return s ? (this.$t('Search results') + ": " + s) : this.$t('Home');
        case "upload":
          return this.$t('Upload media');
        case 'albums':
          return this.$t('Albums list');
        case 'random':
          return this.$t('Random results');
        default:
          return "";
      }
    },

    getIcon: function (p, s) {
      switch (p) {
        case "home":
          return s ? 'fas fa-search' : 'fas fa-home';
        case "upload":
          return 'fas fa-upload';
        case 'albums':
          return 'fas fa-list';
        case 'random':
          return 'fas fa-shuffle';
        default:
          return "";
      }
    },
  },
  mounted: function () {
    this.$options.pageUpdater = this.updatePage.bind(this);

    AppEvents.AddEventListener("app-status-update", this.$options.pageUpdater);
  },
  beforeUnmount: function () {
    AppEvents.RemoveEventListener(
      "app-status-update",
      this.$options.pageUpdater
    );
  },
});
</script>

<style>
.page-content {
  position: absolute;
  top: 57px;
  height: calc(100% - 57px);
  left: 240px;
  width: calc(100% - 240px);
  overflow: auto;
}

.sidebar-hidden .page-content {
  left: 0;
  width: 100%;
}

/* Custom scroll bar */


/* width */

.page-content::-webkit-scrollbar {
    width: 5px;
    height: 3px;
}


/* Track */

.page-content::-webkit-scrollbar-track {
    background: #bdbdbd;
}


/* Handle */

.page-content::-webkit-scrollbar-thumb {
    background: #757575;
}

.layout-media-split .page-content,
.sidebar-hidden .layout-media-split .page-content {
  left: auto;
  right: 0;
  width: 400px;
  border-left: solid 1px rgba(255, 255, 255, 0.1);
}

.layout-album .page-content {
  display: none;
}

.layout-media .page-content {
  display: none;
}

.page-header {
  display: flex;
  flex-direction: row;
  align-items: center;

  border-bottom: solid 1px rgba(255, 255, 255, 0.1);
}

.page-title {
  font-size: 1.2rem;
  padding: 1rem;

  width: 100%;
  overflow: hidden;
  text-overflow: ellipsis;
}

.page-title i {
  margin-right: 0.5rem;
}

.page-header-btn {
  display: block;
  width: 48px;
  height: 48px;
  box-shadow: none;
  border: none;
  cursor: pointer;
  font-size: 24px;
  color: rgba(255, 255, 255, 0.75);
  background: transparent;
  outline: none;
}

.page-header-btn:disabled {
  opacity: 0.7;
  cursor: default;
}

.page-header-btn:hover {
  color: white;
}

.page-header-btn:disabled:hover {
  color: rgba(255, 255, 255, 0.75);
}

.page-header-btn:focus {
  outline: none;
}

.page-expand-btn,
.page-close-btn {
  display: none;
}


.layout-media-split .page-expand-btn,
.layout-media-split .page-close-btn {
  display: block;
}

.layout-media-split .page-title {
  width: calc(100% - 48px - 48px);
}

.page-inner {
  padding: 1rem;
}

.page-inner.hidden {
  display: none;
}

</style>