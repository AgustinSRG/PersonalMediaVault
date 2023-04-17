<template>
  <div class="modal-container modal-container-settings" :class="{ hidden: !display }" tabindex="-1" role="dialog" :aria-hidden="!display" @keydown="keyDownHandle">
    <div v-if="display" class="modal-dialog modal-xl" role="document" @click="stopPropagationEvent">
      <div class="modal-header">
        <div class="modal-title">{{ $t("Batch operation") }}</div>
        <button type="button" class="modal-close-btn" :title="$t('Close')" @click="close">
          <i class="fas fa-times"></i>
        </button>
      </div>
      <div class="modal-body">

        <div class="batch-op-group-search">

          <div class="form-group">
            <label>{{ $t("Title or description must contain") }}:</label>
            <input type="text" name="title-search" autocomplete="off" maxlength="255" v-model="textSearch" class="form-control form-control-full-width" />
          </div>

          <div class="form-group">
            <label>{{ $t("Media type") }}:</label>
            <select class="form-control form-select form-control-full-width" v-model="typeSearch">
              <option :value="0">{{ $t("Any media") }}</option>
              <option :value="1">{{ $t("Images") }}</option>
              <option :value="2">{{ $t("Videos") }}</option>
              <option :value="3">{{ $t("Audios") }}</option>
            </select>
          </div>

          <div class="form-group">
            <label>{{ $t("Album") }}:</label>
            <select v-model="albumSearch" class="form-control form-select form-control-full-width">
              <option :value="-1">--</option>
              <option v-for="a in albums" :key="a.id" :value="a.id">
                {{ a.name }}
              </option>
            </select>
          </div>

          <div class="form-group">
            <label>{{ $t("Tags") }}:</label>
          </div>
          <div class="form-group media-tags" v-if="tagModeSearch !== 'untagged'">
            <label v-if="tagsSearch.length === 0">{{
              $t("There are no tags yet for this filter.")
            }}</label>
            <div v-for="tag in tagsSearch" :key="tag" class="media-tag">
              <div class="media-tag-name">{{ getTagName(tag, tagData) }}</div>
              <button type="button" :title="$t('Remove tag')" class="media-tag-btn" @click="removeTagSearch(tag)">
                <i class="fas fa-times"></i>
              </button>
            </div>
          </div>
          <div class="form-group">
            <select class="form-control form-select form-control-full-width" v-model="tagModeSearch">
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
          <div class="form-group" v-if="tagModeSearch !== 'untagged'">
            <input type="text" autocomplete="off" maxlength="255" v-model="tagToAddSearch" @input="onTagSearchChanged(false)" @keydown="onTagSearchKeyDown" class="form-control" :placeholder="$t('Search for tags') + '...'" />
          </div>
          <div class="form-group" v-if="tagModeSearch !== 'untagged' && matchingTagsSearch.length > 0">
            <button v-for="mt in matchingTagsSearch" :key="mt.id" type="button" class="btn btn-primary btn-sm btn-tag-mini" @click="addMatchingTagSearch(mt)">
              <i class="fas fa-plus"></i> {{ mt.name }}
            </button>
          </div>


        </div>

        <hr>

        <div class="batch-op-group-action">
          <div class="form-group"></div>
          <div class="form-group">
            <label>{{ $t("Select and action to apply") }}:</label>
            <select class="form-control form-select form-control-full-width" v-model="action">
              <option :value="''">
                {{ $t("--- Select an action ---") }}
              </option>
              <option :value="'tag-add'">
                {{ $t("Add tags to the media assets") }}
              </option>
              <option :value="'tag-remove'">
                {{ $t("Remove tags from the media assets") }}
              </option>
              <option :value="'album-add'">
                {{ $t("Add media assets to album") }}
              </option>
              <option :value="'album-remove'">
                {{ $t("Remove media assets from album") }}
              </option>
              <option :value="'delete'">
                {{ $t("Delete media assets") }}
              </option>
            </select>
          </div>


          <div v-if="action === 'tag-add' || action === 'tag-remove'">
            <div class="form-group media-tags">
              <label v-if="tagsAction.length === 0">{{
                $t("There are no selected tags yet.")
              }}</label>
              <div v-for="tag in tagsAction" :key="tag" class="media-tag">
                <div class="media-tag-name">{{ tag }}</div>
                <button type="button" :title="$t('Remove tag')" class="media-tag-btn" @click="removeTagAction(tag)">
                  <i class="fas fa-times"></i>
                </button>
              </div>
            </div>

            <div class="form-group">
              <input type="text" autocomplete="off" maxlength="255" v-model="tagToAddAction" @input="onTagActionChanged(false)" @keydown="onTagActionKeyDown" class="form-control" :placeholder="$t('Search for tags') + '...'" />
            </div>
            <div class="form-group" v-if="matchingTagsAction.length > 0">
              <button v-for="mt in matchingTagsAction" :key="mt.id" type="button" class="btn btn-primary btn-sm btn-tag-mini" @click="addMatchingTagAction(mt)">
                <i class="fas fa-plus"></i> {{ mt.name }}
              </button>
            </div>
          </div>


          <div v-if="action === 'album-add' || action === 'album-remove'">

            <div class="form-group">
              <label>{{ $t("Album") }}:</label>
              <select v-model="albumToAdd" class="form-control form-select form-control-full-width">
                <option :value="-1">--</option>
                <option v-for="a in albums" :key="a.id" :value="a.id">
                  {{ a.name }}
                </option>
              </select>
            </div>

          </div>


        </div>

      </div>
      <div class="modal-footer  no-padding">
        <button type="button" class="modal-footer-btn" :disabled="displayProgress" @click="start">
          <i class="fas fa-check"></i> {{ $t("Apply") }}
        </button>
      </div>
    </div>

    <BatchOperationProgressModal v-if="displayProgress" v-model:display="displayProgress" :status="status" :progress="progress" :error="error" :actionCount="actionCount" @cancel="cancel" @confirm="confirm"></BatchOperationProgressModal>
  </div>
</template>

<script lang="ts">
import { defineComponent, nextTick } from "vue";
import { useVModel } from "../../utils/v-model";
import { FocusTrap } from "../../utils/focus-trap";

import BatchOperationProgressModal from "./BatchOperationProgressModal.vue";
import { TagEntry, TagsController } from "@/control/tags";
import { AppEvents } from "@/control/app-events";
import { copyObject } from "@/utils/objects";
import { AlbumsController } from "@/control/albums";
import { Request } from "@/utils/request";
import { AlbumsAPI } from "@/api/api-albums";
import { SearchAPI } from "@/api/api-search";
import { MediaController, MediaEntry } from "@/control/media";
import { MediaAPI } from "@/api/api-media";
import { TagsAPI } from "@/api/api-tags";

const PAGE_SIZE = 50;

export default defineComponent({
  components: {
    BatchOperationProgressModal,
  },
  name: "BatchOperationModal",
  emits: ["update:display"],
  props: {
    display: Boolean,
  },
  setup(props) {
    return {
      displayStatus: useVModel(props, "display"),
    };
  },
  data: function () {
    return {
      busy: false,

      tagData: {},
      albums: [],

      matchingTagsSearch: [],
      tagToAddSearch: "",

      matchingTagsAction: [],
      tagToAddAction: "",

      textSearch: "",
      typeSearch: 0,
      albumSearch: -1,
      tagsSearch: [],

      tagModeSearch: "all",

      action: "",

      tagsAction: [],

      albumToAdd: -1,

      displayProgress: false,

      progress: 0,
      status: "",
      actionCount: 0,
      actionItems: [],
      error: "",
    };
  },
  methods: {
    autoFocus: function () {
      if (!this.display) {
        return;
      }
      nextTick(() => {
        const elem = this.$el.querySelector(".auto-focus");
        if (elem) {
          elem.focus();
        }
      });
    },

    close: function () {
      this.displayStatus = false;
    },

    stopPropagationEvent: function (e: Event) {
      e.stopPropagation();
    },

    keyDownHandle: function (e: KeyboardEvent) {
      e.stopPropagation();
      if (e.key === "Escape") {
        this.close();
      }
    },

    getTagName: function (tag, data) {
      if (data[tag + ""]) {
        return data[tag + ""].name;
      } else {
        return "???";
      }
    },

    onTagSearchChanged: function (forced?: boolean) {
      if (forced) {
        if (this.$options.findTagSearchTimeout) {
          clearTimeout(this.$options.findTagSearchTimeout);
          this.$options.findTagSearchTimeout = null;
        }
        this.findTagsSearch();
      } else {
        if (this.$options.findTagSearchTimeout) {
          return;
        }
        this.$options.findTagSearchTimeout = setTimeout(() => {
          this.$options.findTagSearchTimeout = null;
          this.findTagsSearch();
        }, 200);
      }
    },

    findTagsSearch: function () {
      const tagFilter = this.tagToAddSearch
        .replace(/[\n\r]/g, " ")
        .trim()
        .replace(/[\s]/g, "_")
        .toLowerCase();
      this.matchingTagsSearch = Object.values(this.tagData)
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
          if (this.tagsSearch.indexOf(a.id) >= 0) {
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


    onTagActionChanged: function (forced?: boolean) {
      if (forced) {
        if (this.$options.findTagActionTimeout) {
          clearTimeout(this.$options.findTagActionTimeout);
          this.$options.findTagActionTimeout = null;
        }
        this.findTagsAction();
      } else {
        if (this.$options.findTagActionTimeout) {
          return;
        }
        this.$options.findTagActionTimeout = setTimeout(() => {
          this.$options.findTagActionTimeout = null;
          this.findTagsAction();
        }, 200);
      }
    },

    findTagsAction: function () {
      const tagFilter = this.tagToAddAction
        .replace(/[\n\r]/g, " ")
        .trim()
        .replace(/[\s]/g, "_")
        .toLowerCase();
      this.matchingTagsAction = Object.values(this.tagData)
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
          if (this.tagsAction.indexOf(a.id) >= 0) {
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

    removeTagSearch: function (tag: number) {
      this.tagsSearch = this.tagsSearch.filter((t) => {
        return tag !== t;
      });
      this.onTagSearchChanged(true);
    },

    addMatchingTagSearch: function (tag: TagEntry) {
      if (this.tagsSearch.indexOf(tag.id) >= 0) {
        return;
      }
      this.tagsSearch.push(tag.id);
      this.onTagSearchChanged(true);
    },

    removeTagAction: function (tag: string) {
      this.tagsAction = this.tagsAction.filter((t) => {
        return tag !== t;
      });
      this.onTagActionChanged(true);
    },

    addMatchingTagAction: function (tag: TagEntry) {
      if (this.tagsAction.indexOf(tag.name) >= 0) {
        return;
      }
      this.tagsAction.push(tag.name);
      this.onTagActionChanged(true);
    },

    onTagSearchKeyDown: function (e: KeyboardEvent) {
      if (e.key === "Enter") {
        e.preventDefault();

        this.onTagSearchChanged(true);

        if (this.matchingTagsSearch.length > 0) {
          this.addMatchingTagSearch(this.matchingTagsSearch[0]);
          this.tagToAddSearch = "";
          this.onTagSearchChanged(true);
        }
      } else if (e.key === "Tab") {
        e.preventDefault();

        if (this.matchingTagsSearch.length > 0) {
          this.tagToAddSearch = this.matchingTagsSearch[0].name;
          this.onTagSearchChanged(true);
        }
      }
    },

    onTagActionKeyDown: function (e: KeyboardEvent) {
      if (e.key === "Enter") {
        e.preventDefault();

        this.addMatchingTagAction({ name: this.tagToAddAction, id: -1 });
        this.tagToAddAction = "";
        this.onTagActionChanged(true);
      } else if (e.key === "Tab") {
        e.preventDefault();

        this.onTagActionChanged(true);

        if (this.matchingTagsAction.length > 0) {
          this.tagToAddAction = this.matchingTagsAction[0].name;
          this.onTagActionChanged(true);
        }
      }
    },

    updateTagData: function () {
      this.tagData = copyObject(TagsController.Tags);
      this.onTagSearchChanged(false);
      this.onTagActionChanged(false);
    },

    updateAlbums: function () {
      this.albums = AlbumsController.GetAlbumsListCopy().sort((a, b) => {
        if (a.nameLowerCase < b.nameLowerCase) {
          return -1;
        } else {
          return 1;
        }
      });
    },

    start: function () {
      if (!this.action) {
        return;
      }

      switch (this.action) {
        case "tag-add":
        case "tag-remove":
          if (this.tagsAction.length === 0) {
            return;
          }
          break;
        case "album-add":
        case "album-remove":
          if (this.albumToAdd < 0) {
            return;
          }
          break;
      }

      this.displayProgress = true;
      this.status = "search";
      this.progress = 0;
      this.actionItems = [];

      if (this.albumSearch >= 0) {
        this.loadAlbumSearch();
      } else {
        this.searchNext(0);
      }
    },

    loadAlbumSearch: function () {
      Request.Abort("modal-batch-request");

      Request.Pending(
        "modal-batch-request",
        AlbumsAPI.GetAlbum(this.albumSearch)
      )
        .onSuccess((result) => {
          const albumList = new Set(result.list.map(m => m.id));
          this.searchNext(0, albumList);
        })
        .onRequestError((err) => {
          Request.ErrorHandler()
            .add(401, "*", () => {
              this.status = "error";
              this.error = this.$t("Access denied");
              AppEvents.Emit("unauthorized");
            })
            .add(403, "*", () => {
              this.status = "error";
              this.error = this.$t("Access denied");
            })
            .add(404, "*", () => {
              this.status = "error";
              this.error = this.$t("The selected album was not found");
            })
            .add(500, "*", () => {
              this.status = "error";
              this.error = this.$t("Internal server error");
            })
            .add("*", "*", () => {
              this.status = "error";
              this.error = this.$t("Could not connect to the server");
            })
            .handle(err);
        })
        .onUnexpectedError((err) => {
          console.error(err);
          this.status = "error";
          this.error = err.message;
        });
    },

    getFirstTag: function () {
      if (this.tagModeSearch === "all" && this.tagsSearch.length > 0) {
        return this.getTagName(this.tagsSearch[0], this.tagData);
      } else {
        return "";
      }
    },

    searchNext: function (page: number, albumList?: Set<number>) {
      Request.Abort("modal-batch-request");

      Request.Pending(
        "modal-batch-request",
        SearchAPI.Search(
          this.getFirstTag(),
          "asc",
          page,
          PAGE_SIZE
        )
      )
        .onSuccess((result) => {
          this.filterElements(result.page_items, albumList);

          this.progress = (page + 1) * 100 / (result.page_count || 1);

          if (page >= (result.page_count - 1)) {
            // Finished
            if (this.actionItems.length > 0) {
              this.status = this.action === "delete" ? "confirmation-delete" : "confirmation";
              this.actionCount = this.actionItems.length;
            } else {
              this.status = "error";
              this.error = this.$t("No items found matching the specified criteria");
            }
          } else {
            this.searchNext(page + 1, albumList);
          }
        })
        .onRequestError((err) => {
          Request.ErrorHandler()
            .add(401, "*", () => {
              this.status = "error";
              this.error = this.$t("Access denied");
              AppEvents.Emit("unauthorized");
            })
            .add(500, "*", () => {
              this.status = "error";
              this.error = this.$t("Internal server error");
            })
            .add("*", "*", () => {
              this.status = "error";
              this.error = this.$t("Could not connect to the server");
            })
            .handle(err);
        })
        .onUnexpectedError((err) => {
          console.error(err);
          this.status = "error";
          this.error = err.message;
        });
    },

    filterElements: function (results: MediaEntry[], filterAlbum: Set<number>) {
      const filterText = this.textSearch.toLowerCase();
      const filterType = this.typeSearch;
      const filterTags = this.tagsSearch.slice();
      const filterTagMode = this.tagModeSearch;

      for (let e of results) {
        if (filterAlbum && !filterAlbum.has(e.id)) {
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

        this.actionItems.push(e.id);
      }
    },

    cancel: function () {
      Request.Abort("modal-batch-request");
    },

    confirm: function () {
      this.status = "action";
      this.progress = 0;
      this.actionNext(0);
    },

    actionNext: function (i: number) {
      Request.Abort("modal-batch-request");

      if (i >= this.actionItems.length) {
        // Finish
        this.status = "success";

        AlbumsController.LoadCurrentAlbum();
        MediaController.Load();
        TagsController.Load();
        return;
      }

      const mediaId = this.actionItems[i];

      switch (this.action) {
        case "tag-add":
          this.actionAddTag(mediaId, this.tagsAction.slice(), i + 1)
          break;
        case "tag-remove":
          this.actionRemoveTag(mediaId, this.tagsAction.slice(), i + 1)
          break;
        case "album-add":
          this.actionAddAlbum(mediaId, i + 1)
          break;
        case "album-remove":
          this.actionRemoveAlbum(mediaId, i + 1)
          break;
        case "delete":
          this.actionDelete(mediaId, i + 1);
          break;
      }

      this.progress = (i + 1) * 100 / (this.actionItems.length || 1);
    },

    actionDelete: function (mid: number, next: number) {
      Request.Pending("modal-batch-request", MediaAPI.DeleteMedia(mid))
        .onSuccess(() => {
          this.actionNext(next);
        })
        .onRequestError((err) => {
          this.status = "error";
          Request.ErrorHandler()
            .add(401, "*", () => {
              this.error = this.$t("Access denied");
              AppEvents.Emit("unauthorized");
            })
            .add(403, "*", () => {
              this.error = this.$t("Access denied");
            })
            .add(404, "*", () => {
              this.error = this.$t("Not found");
            })
            .add(500, "*", () => {
              this.error = this.$t("Internal server error");
            })
            .add("*", "*", () => {
              this.error = this.$t("Could not connect to the server");
            })
            .handle(err);
        })
        .onUnexpectedError((err) => {
          this.error = err.message;
          console.error(err);
          this.status = "error";
        });
    },


    actionAddAlbum: function (mid: number, next: number) {
      Request.Pending("modal-batch-request", AlbumsAPI.AddMediaToAlbum(this.albumToAdd, mid))
        .onSuccess(() => {
          this.actionNext(next);
        })
        .onRequestError((err) => {
          this.status = "error";
          Request.ErrorHandler()
            .add(401, "*", () => {
              this.error = this.$t("Access denied");
              AppEvents.Emit("unauthorized");
            })
            .add(403, "*", () => {
              this.error = this.$t("Access denied");
            })
            .add(404, "*", () => {
              this.error = this.$t("Not found");
            })
            .add(500, "*", () => {
              this.error = this.$t("Internal server error");
            })
            .add("*", "*", () => {
              this.error = this.$t("Could not connect to the server");
            })
            .handle(err);
        })
        .onUnexpectedError((err) => {
          this.error = err.message;
          console.error(err);
          this.status = "error";
        });
    },

    actionRemoveAlbum: function (mid: number, next: number) {
      Request.Pending("modal-batch-request", AlbumsAPI.RemoveMediaFromAlbum(this.albumToAdd, mid))
        .onSuccess(() => {
          this.actionNext(next);
        })
        .onRequestError((err) => {
          this.status = "error";
          Request.ErrorHandler()
            .add(401, "*", () => {
              this.error = this.$t("Access denied");
              AppEvents.Emit("unauthorized");
            })
            .add(403, "*", () => {
              this.error = this.$t("Access denied");
            })
            .add(404, "*", () => {
              this.error = this.$t("Not found");
            })
            .add(500, "*", () => {
              this.error = this.$t("Internal server error");
            })
            .add("*", "*", () => {
              this.error = this.$t("Could not connect to the server");
            })
            .handle(err);
        })
        .onUnexpectedError((err) => {
          this.error = err.message;
          console.error(err);
          this.status = "error";
        });
    },


    actionAddTag: function (mid: number, tags: string[], next: number) {
      if (tags.length === 0) {
        this.actionNext(next);
        return;
      }

      Request.Pending("modal-batch-request", TagsAPI.TagMedia(mid, tags[0]))
        .onSuccess(() => {
          this.actionAddTag(mid, tags.slice(1), next);
        })
        .onRequestError((err) => {
          this.status = "error";
          Request.ErrorHandler()
            .add(401, "*", () => {
              this.error = this.$t("Access denied");
              AppEvents.Emit("unauthorized");
            })
            .add(403, "*", () => {
              this.error = this.$t("Access denied");
            })
            .add(404, "*", () => {
              this.error = this.$t("Not found");
            })
            .add(500, "*", () => {
              this.error = this.$t("Internal server error");
            })
            .add("*", "*", () => {
              this.error = this.$t("Could not connect to the server");
            })
            .handle(err);
        })
        .onUnexpectedError((err) => {
          this.error = err.message;
          console.error(err);
          this.status = "error";
        });
    },

    actionRemoveTag: function (mid: number, tags: string[], next: number) {
      if (tags.length === 0) {
        this.actionNext(next);
        return;
      }

      const tagId = TagsController.FindTag(tags[0]);

      if (tagId < 0) {
        // Tag not found
        this.actionRemoveTag(mid, tags.slice(1), next);
        return;
      }

      Request.Pending("modal-batch-request", TagsAPI.UntagMedia(mid, tagId))
        .onSuccess(() => {
          this.actionRemoveTag(mid, tags.slice(1), next);
        })
        .onRequestError((err) => {
          this.status = "error";
          Request.ErrorHandler()
            .add(401, "*", () => {
              this.error = this.$t("Access denied");
              AppEvents.Emit("unauthorized");
            })
            .add(403, "*", () => {
              this.error = this.$t("Access denied");
            })
            .add(404, "*", () => {
              this.error = this.$t("Not found");
            })
            .add(500, "*", () => {
              this.error = this.$t("Internal server error");
            })
            .add("*", "*", () => {
              this.error = this.$t("Could not connect to the server");
            })
            .handle(err);
        })
        .onUnexpectedError((err) => {
          this.error = err.message;
          console.error(err);
          this.status = "error";
        });
    },
  },
  mounted: function () {
    this.$options.focusTrap = new FocusTrap(this.$el, this.close.bind(this));

    this.$options.tagUpdateH = this.updateTagData.bind(this);

    AppEvents.AddEventListener("tags-update", this.$options.tagUpdateH);

    this.updateTagData();

    this.updateAlbums();
    this.$options.albumsUpdateH = this.updateAlbums.bind(this);
    AppEvents.AddEventListener("albums-update", this.$options.albumsUpdateH);

    if (this.display) {
      this.error = "";
      this.$options.focusTrap.activate();
      this.autoFocus();
    }
  },
  beforeUnmount: function () {
    Request.Abort("modal-batch-request");

    if (this.$options.focusTrap) {
      this.$options.focusTrap.destroy();
    }

    if (this.$options.findTagSearchTimeout) {
      clearTimeout(this.$options.findTagSearchTimeout);
      this.$options.findTagSearchTimeout = null;
    }

    if (this.$options.findTagActionTimeout) {
      clearTimeout(this.$options.findTagActionTimeout);
      this.$options.findTagActionTimeout = null;
    }

    AppEvents.RemoveEventListener("tags-update", this.$options.tagUpdateH);
    AppEvents.RemoveEventListener("albums-update", this.$options.albumsUpdateH);
  },
  watch: {
    display: function () {
      if (this.display) {
        this.error = "";
        if (this.$options.focusTrap) {
          this.$options.focusTrap.activate();
        }
        this.autoFocus();
      } else {
        if (this.$options.focusTrap) {
          this.$options.focusTrap.deactivate();
        }
      }
    },
  },
});
</script>

<style>
@import "@/style/content/batch-operation.css";
</style>
