<template>
  <div class="page-inner" :class="{ hidden: !display }">
    <form class="adv-search-form" @submit="startSearch">
      <div class="form-group">
        <label>{{ $t("Title or description must contain") }}:</label>
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
        <label>{{ $t("Tags") }}:</label>
      </div>
      <div class="form-group media-tags">
        <label v-if="tags.length === 0">{{
          $t("There are no tags yet for this filter.")
        }}</label>
        <div v-for="tag in tags" :key="tag" class="media-tag">
          <div class="media-tag-name">{{ getTagName(tag, tagData) }}</div>
          <button
            type="button"
            :title="$t('Remove tag')"
            class="media-tag-btn"
            @click="removeTag(tag)"
          >
            <i class="fas fa-times"></i>
          </button>
        </div>
      </div>
      <div class="form-group">
        <input
          type="text"
          autocomplete="off"
          maxlength="255"
          v-model="tagToAdd"
          @input="onTagAddChanged"
          class="form-control"
          :placeholder="$t('Search for tags') + '...'"
        />
      </div>
      <div class="form-group" v-if="matchingTags.length > 0">
        <button
          v-for="mt in matchingTags"
          :key="mt.id"
          type="button"
          class="btn btn-primary btn-xs btn-tag-mini"
          @click="addMatchingTag(mt)"
        >
          <i class="fas fa-plus"></i> {{ mt.name }}
        </button>
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
          class="btn btn-primary btn-sm btn-mr"
        >
          <i class="fas fa-search"></i> {{ $t("Search") }}
        </button>
        <button
          v-if="loading"
          type="button"
          class="btn btn-primary btn-sm btn-mr"
          disabled
        >
          <i class="fa fa-spinner fa-spin"></i> {{ $t("Searching") }}... ({{
            cssProgress(progress)
          }})
        </button>
        <button
          v-if="loading"
          type="button"
          class="btn btn-primary btn-sm btn-mr"
          @click="cancel"
        >
          <i class="fas fa-times"></i> {{ $t("Cancel") }}
        </button>
      </div>
    </form>

    <div class="search-results">

      <div v-if="!loading && started && pageItems.length === 0" class="search-results-msg-display">
        <div class="search-results-msg-icon"><i class="fas fa-search"></i></div>
        <div class="search-results-msg-text">
          {{ $t("Could not find any result") }}
        </div>
        <div class="search-results-msg-btn">
          <button type="button" @click="startSearch()" class="btn btn-primary btn-sm">
            <i class="fas fa-sync-alt"></i> {{ $t("Refresh") }}
          </button>
        </div>
      </div>

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
import { TagsController } from "@/control/tags";
import { copyObject } from "@/utils/objects";
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

      started: false,
      finished: true,

      tagData: {},
      tags: [],
      tagToAdd: "",
      matchingTags: [],

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
      const filterTags = this.tags.slice();
      for (let e of results) {
        if (this.pageItems.length >= this.pageSize) {
          return;
        }

        if (filterText) {
          if (!e.title.toLowerCase().includes(filterText) && !e.description.toLowerCase().includes(filterText)) {
            continue;
          }
        }

        if (filterType) {
          if (e.type !== filterType) {
            continue;
          }
        }

        if (filterTags.length > 0) {
          let passesTags = true;
          for (let tag of filterTags) {
            if (!e.tags || !e.tags.includes(tag)) {
              passesTags = false;
              break;
            }
          }

          if (!passesTags) {
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
      this.started = true;
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
      this.started = false;
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

    updateTagData: function () {
      this.tagData = copyObject(TagsController.Tags);
      this.onTagAddChanged();
    },

    getTagName: function (tag, data) {
      if (data[tag + ""]) {
        return data[tag + ""].name;
      } else {
        return "???";
      }
    },

    removeTag: function (tag) {
      this.tags = this.tags.filter((t) => {
        return tag !== t;
      });
    },

    addMatchingTag: function (tag) {
      this.tags.push(tag.id);
    },

    onTagAddChanged: function () {
      if (this.$options.findTagTimeout) {
        return;
      }
      this.$options.findTagTimeout = setTimeout(() => {
        this.$options.findTagTimeout = null;
        this.findTags();
      }, 200);
    },

    findTags: function () {
      const tagFilter = this.tagToAdd
        .replace(/[\n\r]/g, " ")
        .trim()
        .replace(/[\s]/g, "_")
        .toLowerCase();
      this.matchingTags = Object.values(this.tagData)
        .map((a: any) => {
          if (!tagFilter) {
            return {
              id: a.id,
              name: a.name,
              starts: true,
              contains: true,
            };
          }
          const i = a.name.indexOf(tagFilter);
          return {
            id: a.id,
            name: a.name,
            starts: i === 0,
            contains: i >= 0,
          };
        })
        .filter((a) => {
          if (this.tags.indexOf(a.id) >= 0) {
            return false;
          }
          return a.starts || a.contains;
        })
        .sort((a, b) => {
          if (a.starts && !b.starts) {
            return -1;
          } else if (b.starts && !a.starts) {
            return 1;
          } else if (a.name < b.name) {
            return -1;
          } else {
            return 1;
          }
        })
        .slice(0, 10);
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

    this.$options.tagUpdateH = this.updateTagData.bind(this);

    AppEvents.AddEventListener("tags-update", this.$options.tagUpdateH);

    this.updateTagData();
  },
  beforeUnmount: function () {
    Timeouts.Abort("page-advsearch-load");
    Request.Abort("page-advsearch-load");

    AppEvents.RemoveEventListener("auth-status-changed", this.$options.loadH);
    AppEvents.RemoveEventListener("media-delete", this.$options.resetH);
    AppEvents.RemoveEventListener("media-meta-change", this.$options.resetH);

    AppEvents.RemoveEventListener("tags-update", this.$options.tagUpdateH);

    if (this.$options.findTagTimeout) {
      clearTimeout(this.$options.findTagTimeout);
    }
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