<template>
  <div :class="{ 'page-inner': !inmodal, hidden: !display }">
    <form class="adv-search-form" @submit="startSearch">
      <div class="form-group">
        <label>{{ $t("Title or description must contain") }}:</label>
        <input type="text" name="title-search" autocomplete="off" maxlength="255" :disabled="loading" v-model="textSearch" class="form-control form-control-full-width" />
      </div>

      <div v-if="advancedSearch">
        <div class="form-group">
          <label>{{ $t("Media type") }}:</label>
          <select class="form-control form-select form-control-full-width" :disabled="loading" v-model="type">
            <option :value="0">{{ $t("Any media") }}</option>
            <option :value="1">{{ $t("Images") }}</option>
            <option :value="2">{{ $t("Videos") }}</option>
            <option :value="3">{{ $t("Audios") }}</option>
          </select>
        </div>

        <div class="form-group">
          <label>{{ $t("Tags") }}:</label>
        </div>
        <div class="form-group media-tags" v-if="tagMode !== 'untagged'">
          <label v-if="tags.length === 0">{{
            $t("There are no tags yet for this filter.")
          }}</label>
          <div v-for="tag in tags" :key="tag" class="media-tag">
            <div class="media-tag-name">{{ getTagName(tag, tagData) }}</div>
            <button type="button" :title="$t('Remove tag')" class="media-tag-btn" :disabled="loading" @click="removeTag(tag)">
              <i class="fas fa-times"></i>
            </button>
          </div>
        </div>
        <div class="form-group">
          <select class="form-control form-select form-control-full-width" :disabled="loading" v-model="tagMode">
            <option :value="'all'">
              {{ $t("Media must contain ALL of the selected tags") }}
            </option>
            <option :value="'any'">
              {{ $t("Media must contain ANY of the selected tags") }}
            </option>
            <option :value="'none'">
              {{ $t("Media must contain NONE of the selected tags") }}
            </option>
            <option :value="'untagged'">
              {{ $t("Media must be untagged") }}
            </option>
          </select>
        </div>
        <div class="form-group" v-if="tagMode !== 'untagged'">
          <input type="text" autocomplete="off" maxlength="255" v-model="tagToAdd" :disabled="loading" @input="onTagAddChanged(false)" class="form-control" :placeholder="$t('Search for tags') + '...'" />
        </div>
        <div class="form-group" v-if="tagMode !== 'untagged' && matchingTags.length > 0">
          <button v-for="mt in matchingTags" :key="mt.id" type="button" :disabled="loading" class="btn btn-primary btn-sm btn-tag-mini" @click="addMatchingTag(mt)">
            <i class="fas fa-plus"></i> {{ mt.name }}
          </button>
        </div>

        <div class="form-group">
          <label>{{ $t("Order") }}:</label>
          <select class="form-control form-select form-control-full-width" :disabled="loading" v-model="order">
            <option :value="'desc'">{{ $t("Show most recent") }}</option>
            <option :value="'asc'">{{ $t("Show oldest") }}</option>
          </select>
        </div>
        <div class="form-group">
          <label>{{ $t("Limit results") }}:</label>
          <select class="form-control form-select form-control-full-width" :disabled="loading" v-model="pageSize">
            <option v-for="po in pageSizeOptions" :key="po" :value="po">
              {{ po }} {{ $t("results max") }}
            </option>
          </select>
        </div>
      </div>

      <div class="">
        <button v-if="!loading" type="submit" class="btn btn-primary btn-mr">
          <i class="fas fa-search"></i> {{ $t("Search") }}
        </button>
        <button v-if="loading" type="button" class="btn btn-primary btn-mr" disabled>
          <i class="fa fa-spinner fa-spin"></i> {{ $t("Searching") }}... ({{
            cssProgress(progress)
          }})
        </button>
        <button v-if="!advancedSearch" type="button" class="btn btn-primary btn-mr" @click="toggleAdvancedSearch">
          <i class="fas fa-cog"></i> {{ $t("More options") }}
        </button>
        <button v-if="advancedSearch" type="button" class="btn btn-primary btn-mr" @click="toggleAdvancedSearch">
          <i class="fas fa-cog"></i> {{ $t("Less options") }}
        </button>
        <button v-if="inmodal" type="button" class="btn btn-primary btn-mr" @click="changeToUpload">
          <i class="fas fa-upload"></i> {{ $t("Upload") }}
        </button>
        <button v-if="loading" type="button" class="btn btn-primary btn-mr" @click="cancel">
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
          <button type="button" @click="startSearch()" class="btn btn-primary">
            <i class="fas fa-sync-alt"></i> {{ $t("Refresh") }}
          </button>
        </div>
      </div>

      <div v-if="!loading && pageItems.length > 0" class="search-results-final-display">
        <a v-for="(item, i) in pageItems" :key="i" class="search-result-item clickable" :class="{ current: currentMedia == item.id }" @click="goToMedia(item.id, $event)" :href="getMediaURL(item.id)" target="_blank" rel="noopener noreferrer">
          <div class="search-result-thumb" :title="item.title || $t('Untitled')">
            <div class="search-result-thumb-inner">
              <div v-if="!item.thumbnail" class="no-thumb">
                <i v-if="item.type === 1" class="fas fa-image"></i>
                <i v-else-if="item.type === 2" class="fas fa-video"></i>
                <i v-else-if="item.type === 3" class="fas fa-headphones"></i>
                <i v-else class="fas fa-ban"></i>
              </div>
              <img v-if="item.thumbnail" :src="getThumbnail(item.thumbnail)" :alt="item.title || $t('Untitled')" />
              <div class="search-result-thumb-tag" v-if="item.type === 2 || item.type === 3">
                {{ renderTime(item.duration) }}
              </div>
            </div>
          </div>
          <div class="search-result-title">
            {{ item.title || $t("Untitled") }}
          </div>
        </a>
      </div>
    </div>
  </div>
</template>

<script lang="ts">
import { SearchAPI } from "@/api/api-search";
import { AlbumsController } from "@/control/albums";
import { AppEvents } from "@/control/app-events";
import { AppStatus } from "@/control/app-status";
import { AuthController } from "@/control/auth";
import { KeyboardManager } from "@/control/keyboard";
import { MediaEntry } from "@/control/media";
import { TagsController } from "@/control/tags";
import { copyObject } from "@/utils/objects";
import { GenerateURIQuery, GetAssetURL, Request } from "@/utils/request";
import { renderTimeSeconds } from "@/utils/time-utils";
import { Timeouts } from "@/utils/timeout";
import { defineComponent, nextTick } from "vue";

export default defineComponent({
  name: "PageAdvancedSearch",
  emits: ['select-media', 'change-to-upload'],
  props: {
    display: Boolean,
    inmodal: Boolean,
    noalbum: Number,
  },
  data: function () {
    return {
      loading: false,

      pageSize: 50,
      order: "desc",
      textSearch: "",
      type: 0,

      currentMedia: AppStatus.CurrentMedia,

      pageItems: [],
      page: 0,
      totalPages: 0,
      progress: 0,

      started: false,
      finished: true,

      advancedSearch: false,

      tagData: {},
      tags: [],
      tagToAdd: "",
      matchingTags: [],
      tagMode: "all",

      pageSizeOptions: [25, 50, 100, 150, 200, 250, 500],
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
        SearchAPI.Search(
          this.getFirstTag(),
          this.order,
          this.page,
          this.pageSize
        )
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
            if (!this.inmodal) {
              nextTick(() => {
                const currentElem = this.$el.querySelector(
                  ".search-result-item.current"
                );
                if (currentElem) {
                  currentElem.scrollIntoView();
                }
              });
              this.onCurrentMediaChanged();
            }
          } else if (this.page < this.totalPages - 1) {
            this.page++;
            this.load();
          } else {
            this.loading = false;
            this.finished = true;
            if (!this.inmodal) {
              nextTick(() => {
                const currentElem = this.$el.querySelector(
                  ".search-result-item.current"
                );
                if (currentElem) {
                  currentElem.scrollIntoView();
                }
              });
              this.onCurrentMediaChanged();
            }
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

    toggleAdvancedSearch: function () {
      this.advancedSearch = !this.advancedSearch;
    },

    changeToUpload: function () {
      this.$emit("change-to-upload");
    },

    filterElements: function (results: MediaEntry[]) {
      const filterText = this.textSearch.toLowerCase();
      const filterType = this.type;
      const filterTags = this.tags.slice();
      const filterTagMode = this.tagMode;

      let backlistAlbum = new Set();

      if (this.noalbum >= 0 && AlbumsController.CurrentAlbumData) {
        backlistAlbum = new Set(AlbumsController.CurrentAlbumData.list.map(a => {
          return a.id;
        }));
      }

      for (let e of results) {
        if (this.pageItems.length >= this.pageSize) {
          return;
        }

        if (backlistAlbum.has(e.id)) {
          continue;
        }

        if (filterText) {
          if (
            !e.title.toLowerCase().includes(filterText) &&
            !e.description.toLowerCase().includes(filterText)
          ) {
            continue;
          }
        }

        if (filterType) {
          if (e.type !== filterType) {
            continue;
          }
        }

        if (filterTagMode === "all") {
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
        } else if (filterTagMode === "none") {
          if (filterTags.length > 0) {
            let passesTags = true;
            for (let tag of filterTags) {
              if (e.tags && e.tags.includes(tag)) {
                passesTags = false;
                break;
              }
            }

            if (!passesTags) {
              continue;
            }
          }
        } else if (filterTagMode === "any") {
          if (filterTags.length > 0) {
            let passesTags = false;
            for (let tag of filterTags) {
              if (!e.tags || e.tags.includes(tag)) {
                passesTags = true;
                break;
              }
            }

            if (!passesTags) {
              continue;
            }
          }
        } else if (filterTagMode === "untagged") {
          if (e.tags && e.tags.length > 0) {
            continue;
          }
        }

        this.pageItems.push(e);
      }
    },

    startSearch: function (event?: Event) {
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

    goToMedia: function (mid, e) {
      if (e) {
        e.preventDefault();
      }
      if (this.inmodal) {
        this.$emit("select-media", mid, () => {
          this.pageItems = this.pageItems.filter(i => {
            return mid !== i.id;
          });
        });
      } else {
        AppStatus.ClickOnMedia(mid, true);
      }
    },

    getMediaURL: function (mid: number): string {
      return (
        window.location.protocol +
        "//" +
        window.location.host +
        window.location.pathname +
        GenerateURIQuery({
          media: mid + "",
        })
      );
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
      this.onTagAddChanged(false);
    },

    getFirstTag: function () {
      if (this.tagMode === "all" && this.tags.length > 0) {
        return this.getTagName(this.tags[0], this.tagData);
      } else {
        return "";
      }
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
      this.onTagAddChanged(true);
    },

    addMatchingTag: function (tag) {
      if (this.tags.indexOf(tag.id) >= 0) {
        return;
      }
      this.tags.push(tag.id);
      this.onTagAddChanged(true);
    },

    onTagAddChanged: function (forced: boolean) {
      if (forced) {
        if (this.$options.findTagTimeout) {
          clearTimeout(this.$options.findTagTimeout);
          this.$options.findTagTimeout = null;
        }
        this.findTags();
      } else {
        if (this.$options.findTagTimeout) {
          return;
        }
        this.$options.findTagTimeout = setTimeout(() => {
          this.$options.findTagTimeout = null;
          this.findTags();
        }, 200);
      }
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

    onAppStatusChanged: function () {
      this.currentMedia = AppStatus.CurrentMedia;
      if (!this.inmodal) {
        nextTick(() => {
          const currentElem = this.$el.querySelector(
            ".search-result-item.current"
          );
          if (currentElem) {
            currentElem.scrollIntoView();
          }
        });
        this.onCurrentMediaChanged();
      }
    },

    findCurrentMediaIndex: function (): number {
      for (let i = 0; i < this.pageItems.length; i++) {
        if (this.pageItems[i].id === this.currentMedia) {
          return i;
        }
      }
      return -1;
    },

    onCurrentMediaChanged: function () {
      if (!this.inmodal) {
        const i = this.findCurrentMediaIndex();
        AlbumsController.OnPageLoad(i, this.pageItems.length, 0, 1);
      }
    },

    nextMedia: function () {
      const i = this.findCurrentMediaIndex();
      if (i !== -1 && i < this.pageItems.length - 1) {
        this.goToMedia(this.pageItems[i + 1].id);
      } else if (i === -1 && this.pageItems.length > 0) {
        this.goToMedia(this.pageItems[0].id);
      }
    },

    prevMedia: function () {
      const i = this.findCurrentMediaIndex();
      if (i !== -1 && i > 0) {
        this.goToMedia(this.pageItems[i - 1].id);
      } else if (i === -1 && this.pageItems.length > 0) {
        this.goToMedia(this.pageItems[0].id);
      }
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

      if (this.inmodal) {
        return false;
      }

      if (event.key === "Home") {
        if (this.pageItems.length > 0) {
          this.goToMedia(this.pageItems[0].id);
        }
        return true;
      }

      if (event.key === "End") {
        if (this.pageItems.length > 0) {
          this.goToMedia(this.pageItems[this.pageItems.length - 1].id);
        }
        return true;
      }

      if (event.key === "ArrowLeft") {
        this.prevMedia();
        return true;
      }

      if (event.key === "ArrowRight") {
        this.nextMedia();
        return true;
      }

      return false;
    },
  },
  mounted: function () {
    this.advancedSearch = false;
    this.$options.handleGlobalKeyH = this.handleGlobalKey.bind(this);
    KeyboardManager.AddHandler(this.$options.handleGlobalKeyH, 20);

    this.$options.loadH = this.load.bind(this);
    this.$options.resetH = this.resetSearch.bind(this);
    this.$options.statusChangeH = this.onAppStatusChanged.bind(this);

    AppEvents.AddEventListener("auth-status-changed", this.$options.loadH);
    AppEvents.AddEventListener("media-delete", this.$options.resetH);
    AppEvents.AddEventListener("media-meta-change", this.$options.resetH);

    AppEvents.AddEventListener(
      "app-status-update",
      this.$options.statusChangeH
    );

    this.$options.nextMediaH = this.nextMedia.bind(this);
    AppEvents.AddEventListener("page-media-nav-next", this.$options.nextMediaH);

    this.$options.prevMediaH = this.prevMedia.bind(this);
    AppEvents.AddEventListener("page-media-nav-prev", this.$options.prevMediaH);

    this.$options.tagUpdateH = this.updateTagData.bind(this);

    AppEvents.AddEventListener("tags-update", this.$options.tagUpdateH);

    this.updateTagData();

    if (this.inmodal) {
      this.startSearch();
    }
  },
  beforeUnmount: function () {
    Timeouts.Abort("page-advsearch-load");
    Request.Abort("page-advsearch-load");

    AppEvents.RemoveEventListener("auth-status-changed", this.$options.loadH);
    AppEvents.RemoveEventListener("media-delete", this.$options.resetH);
    AppEvents.RemoveEventListener("media-meta-change", this.$options.resetH);

    AppEvents.RemoveEventListener(
      "app-status-update",
      this.$options.statusChangeH
    );

    AppEvents.RemoveEventListener("page-media-nav-next", this.$options.nextMediaH);
    AppEvents.RemoveEventListener("page-media-nav-prev", this.$options.prevMediaH);

    AppEvents.RemoveEventListener("tags-update", this.$options.tagUpdateH);

    if (this.$options.findTagTimeout) {
      clearTimeout(this.$options.findTagTimeout);
    }

    KeyboardManager.RemoveHandler(this.$options.handleGlobalKeyH);

    if (!this.inmodal) {
      AlbumsController.OnPageUnload();
    }
  },
  watch: {
    display: function () {
      this.load();
      if (this.display && this.inmodal) {
        this.startSearch();
      } else if (this.inmodal) {
        this.cancel();
      }
    },
  },
});
</script>
