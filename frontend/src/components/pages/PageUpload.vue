<template>
  <div class="page-inner-padded" :class="{ 'page-inner': !inmodal, hidden: !display }">
    <div class="form-group">
      <button v-if="!optionsShown" @click="showOptions(true)" type="button" class="btn btn-primary btn-mr">
        <i class="fas fa-cog"></i> {{ $t("Show advanced options") }}
      </button>
      <button v-if="optionsShown" @click="showOptions(false)" type="button" class="btn btn-primary btn-mr">
        <i class="fas fa-cog"></i> {{ $t("Hide advanced options") }}
      </button>
      <button v-if="inmodal" @click="changeToSearch" type="button" class="btn btn-primary btn-mr">
        <i class="fas fa-search"></i> {{ $t("Search") }}
      </button>
    </div>
    <div class="upload-options-container" v-if="optionsShown">
      <div class="form-group">
        <label>{{ $t("Max number of uploads in parallel") }}:</label>
        <select v-model="maxParallelUploads" @change="updateMaxParallelUploads" class="form-control form-select">
          <option :value="1">1</option>
          <option :value="2">2</option>
          <option :value="4">4</option>
          <option :value="8">8</option>
        </select>
      </div>
      <div class="form-group">
        <label>{{ $t("Select an album to add the uploaded media into") }}:</label>
        <select v-model="album" :disabled="inmodal" class="form-control form-select">
          <option :value="-1">--</option>
          <option v-for="a in albums" :key="a.id" :value="a.id">
            {{ a.name }}
          </option>
        </select>
      </div>
      <div class="form-group" v-if="!inmodal">
        <button type="button" @click="createAlbum" class="btn btn-primary">
          <i class="fas fa-plus"></i> {{ $t("Create album") }}
        </button>
      </div>
      <div class="form-group">
        <label>{{ $t("Tags to automatically add to the uploaded media") }}:</label>
      </div>
      <div class="form-group media-tags">
        <label v-if="tags.length === 0">({{ $t("none") }})</label>
        <div v-for="tag in tags" :key="tag" class="media-tag">
          <div class="media-tag-name">{{ tag }}</div>
          <button type="button" :title="$t('Remove tag')" class="media-tag-btn" @click="removeTag(tag)">
            <i class="fas fa-times"></i>
          </button>
        </div>
      </div>
      <form @submit="addTag">
        <div class="form-group">
          <label>{{ $t("Tag to add") }}:</label>
          <input type="text" autocomplete="off" maxlength="255" v-model="tagToAdd" @input="onTagAddChanged(false)"
            class="form-control" />
        </div>
        <div class="form-group" v-if="matchingTags.length > 0">
          <button v-for="mt in matchingTags" :key="mt.id" type="button" class="btn btn-primary btn-sm btn-tag-mini"
            @click="addTagByName(mt.name)">
            <i class="fas fa-plus"></i> {{ mt.name }}
          </button>
        </div>
        <div class="form-group">
          <button type="submit" class="btn btn-primary" :disabled="!tagToAdd">
            <i class="fas fa-plus"></i> {{ $t("Add Tag") }}
          </button>
        </div>
      </form>
    </div>
    <input type="file" class="file-hidden" @change="inputFileChanged" name="media-upload" multiple />
    <div class="upload-box" :class="{ dragging: dragging }" tabindex="0" @click="clickToSelect" @dragover="dragOver"
      @dragenter="dragEnter" @dragstart="dragEnter" @dragend="dragLeave" @dragleave="dragLeave" @drop="onDrop"
      @keydown="clickOnEnter">
      <div class="upload-box-hint">
        {{ $t("Drop file here or click to open the file selection dialog.") }}
      </div>
    </div>

    <div class="upload-filter-menu">
      <a href="javascript:;" @click="updateSelectedState('pending')" class="upload-filter-menu-item"
        :class="{ selected: this.selectedState === 'pending' }">{{ $t("Pending") }} ({{ countPending }})</a>
      <a href="javascript:;" @click="updateSelectedState('ready')" class="upload-filter-menu-item"
        :class="{ selected: this.selectedState === 'ready' }">{{ $t("Ready") }} ({{ countReady }})</a>
      <a href="javascript:;" @click="updateSelectedState('error')" class="upload-filter-menu-item"
        :class="{ selected: this.selectedState === 'error' }">{{ $t("Error") }} ({{ countError }})</a>
    </div>

    <div class="upload-list">
      <div v-for="m in filteredEntries" :key="m.id" class="upload-list-item">
        <div class="upload-list-item-top">
          <div class="upload-list-item-file-name">
            <span v-if="m.status !== 'ready'" class="bold">{{ m.name }}</span>
            <a v-if="m.status === 'ready'" class="bold" @click="goToMedia(m, $event)" :href="getMediaURL(m.mid)"
              target="_blank" rel="noopener noreferrer">{{
                m.name }}</a>
          </div>
          <div class="upload-list-item-file-size">
            <span>{{ renderSize(m.size) }}</span>
          </div>
        </div>
        <div class="upload-list-item-bottom">
          <div class="upload-list-item-status">
            <div class="upload-list-item-status-bar">
              <div class="upload-list-item-status-bar-current"
                :class="{ error: m.status === 'error', success: m.status === 'ready' }"
                :style="{ width: cssProgress(m.status, m.progress) }"></div>
              <div class="upload-list-item-status-bar-text">{{ renderStatus(m.status, m.progress, m.error) }}</div>
            </div>
          </div>
          <div class="upload-list-item-right">
            <button v-if="
              m.status === 'pending' ||
              m.status === 'uploading' ||
              m.status === 'encrypting' ||
              m.status === 'tag'
            " type="button" class="table-btn" :title="$t('Cancel upload')" @click="removeFile(m.id)">
              <i class="fas fa-times"></i>
            </button>
            <button v-if="m.status === 'ready'" type="button" class="table-btn" :title="$t('View media')"
              @click="goToMedia(m)">
              <i class="fas fa-eye"></i>
            </button>
            <button v-if="m.status === 'ready'" type="button" class="table-btn" :title="$t('Done')"
              @click="removeFile(m.id)">
              <i class="fas fa-check"></i>
            </button>
            <button v-if="m.status === 'error'" type="button" class="table-btn" :title="$t('Try again')"
              @click="tryAgain(m)">
              <i class="fas fa-rotate"></i>
            </button>
            <button v-if="m.status === 'error'" type="button" class="table-btn" :title="$t('Remove')"
              @click="removeFile(m.id)">
              <i class="fas fa-times"></i>
            </button>
          </div>
        </div>
      </div>
    </div>

    <div class="upload-table table-responsive" v-if="pendingToUpload.length > 0">
      <div class="form-group" v-if="pendingToUpload.length > 0">
        <button type="button" class="btn btn-primary" @click="clearList">
          <i class="fas fa-broom"></i> {{ $t("Clear list") }}
        </button>
      </div>
      <div class="form-group" v-if="countCancellable > 0">
        <button type="button" class="btn btn-primary" @click="cancelAll">
          <i class="fas fa-times"></i> {{ $t("Cancel all uploads") }}
        </button>
      </div>
    </div>

    <AlbumCreateModal v-model:display="displayAlbumCreate" @new-album="onNewAlbum"></AlbumCreateModal>
  </div>
</template>

<script lang="ts">
import { AlbumsController } from "@/control/albums";
import { AppEvents } from "@/control/app-events";
import { AppStatus } from "@/control/app-status";
import { TagsController } from "@/control/tags";
import { UploadController, UploadEntryMin } from "@/control/upload";
import { copyObject } from "@/utils/objects";
import { GenerateURIQuery } from "@/utils/request";
import { parseTagName } from "@/utils/text";
import { defineComponent } from "vue";

import AlbumCreateModal from "../modals/AlbumCreateModal.vue";

const STATE_FILTER_PENDING = ['pending', 'uploading', 'encrypting', 'tag'];
const STATE_FILTER_READY = ['ready'];
const STATE_FILTER_ERROR = ['error'];

export default defineComponent({
  components: {
    AlbumCreateModal,
  },
  name: "PageUpload",
  emits: ['change-to-search', 'media-go'],
  props: {
    display: Boolean,
    inmodal: Boolean,
    fixedalbum: Number,
  },
  data: function () {
    return {
      dragging: false,
      pendingToUpload: [],
      countCancellable: 0,

      optionsShown: false,

      maxParallelUploads: UploadController.MaxParallelUploads,

      tags: [],
      tagToAdd: "",
      tagData: {},
      matchingTags: [],

      album: -1,
      albums: [],

      displayAlbumCreate: false,

      countPending: 0,
      countReady: 0,
      countError: 0,

      stateFilter: STATE_FILTER_PENDING.slice(),
      selectedState: 'pending',

      filteredEntries: [],
    };
  },
  methods: {
    clickToSelect: function () {
      this.$el.querySelector(".file-hidden").click();
    },

    updateSelectedState: function (s: string) {
      this.selectedState = s;
      switch (s) {
        case "ready":
          this.stateFilter = STATE_FILTER_READY.slice();
          break;
        case "error":
          this.stateFilter = STATE_FILTER_ERROR.slice();
          break;
        default:
          this.stateFilter = STATE_FILTER_PENDING.slice();
      }
      this.updateFilteredEntries();
    },

    updateFilteredEntries: function () {
      this.filteredEntries = this.pendingToUpload.filter((e: UploadEntryMin) => {
        return this.stateFilter.includes(e.status);
      });

      if (this.selectedState !== 'pending') {
        this.filteredEntries = this.filteredEntries.reverse();
      }
    },

    changeToSearch: function () {
      this.$emit("change-to-search");
    },

    createAlbum: function () {
      this.displayAlbumCreate = true;
    },

    onNewAlbum: function (albumId) {
      this.album = albumId;
    },

    updateMaxParallelUploads: function () {
      UploadController.MaxParallelUploads = this.maxParallelUploads;
    },

    inputFileChanged: function (e) {
      const data = e.target.files;
      if (data && data.length > 0) {
        for (let file of data) {
          this.addFile(file);
        }
      }
    },

    onDrop: function (e) {
      e.preventDefault();
      this.dragging = false;
      const data = e.dataTransfer.files;
      if (data && data.length > 0) {
        for (let file of data) {
          this.addFile(file);
        }
      }
    },

    dragOver: function (e) {
      e.preventDefault();
    },
    dragEnter: function (e) {
      e.preventDefault();
      this.dragging = true;
    },
    dragLeave: function (e) {
      e.preventDefault();
      this.dragging = false;
    },

    renderSize: function (bytes: number): string {
      if (bytes > 1024 * 1024 * 1024) {
        let gb = bytes / (1024 * 1024 * 1024);
        gb = Math.floor(gb * 100) / 100;
        return gb + " GB";
      } else if (bytes > 1024 * 1024) {
        let mb = bytes / (1024 * 1024);
        mb = Math.floor(mb * 100) / 100;
        return mb + " MB";
      } else if (bytes > 1024) {
        let kb = bytes / 1024;
        kb = Math.floor(kb * 100) / 100;
        return kb + " KB";
      } else {
        return bytes + " Bytes";
      }
    },

    addFile: function (file: File) {
      UploadController.AddFile(file, this.album, this.tags.slice());
      this.updateSelectedState("pending");
    },

    removeFile: function (id: number) {
      UploadController.RemoveFile(id);
    },

    clearList: function () {
      UploadController.ClearList();
    },

    cancelAll: function () {
      UploadController.CancelAll();
    },

    tryAgain: function (m: UploadEntryMin) {
      UploadController.TryAgain(m.id);
    },

    goToMedia: function (m: UploadEntryMin, e?: MouseEvent) {
      if (e) {
        e.preventDefault();
      }
      if (m.mid < 0) {
        return;
      }
      AppStatus.ClickOnMedia(m.mid, true);
      this.$emit("media-go");
    },

    renderStatus(status: string, p: number, err: string) {
      switch (status) {
        case "ready":
          return this.$t("Ready");
        case "pending":
          return this.$t("Pending");
        case "uploading":
          if (p > 0) {
            return this.$t("Uploading") + "... (" + p + "%)";
          } else {
            return this.$t("Uploading") + "...";
          }
        case "encrypting":
          if (p > 0) {
            return this.$t("Encrypting") + "... (" + p + "%)";
          } else {
            return this.$t("Encrypting") + "...";
          }
        case "tag":
          return (
            this.$t("Adding tags") +
            "... (" +
            this.$t("$N left").replace("$N", "" + p) +
            ")"
          );
        case "error":
          switch (err) {
            case "invalid-media":
              return (
                this.$t("Error") + ": " + this.$t("Invalid media file provided")
              );
            case "access-denied":
              return this.$t("Error") + ": " + this.$t("Access denied");
            case "deleted":
              return (
                this.$t("Error") + ": " + this.$t("The media asset was deleted")
              );
            case "no-internet":
              return (
                this.$t("Error") +
                ": " +
                this.$t("Could not connect to the server")
              );
            default:
              return this.$t("Error") + ": " + this.$t("Internal server error");
          }
        default:
          return "-";
      }
    },

    cssProgress: function (status: string, p: number) {
      switch (status) {
        case "uploading":
          return Math.round(p * 50 / 100) + "%";
        case "encrypting":
          return Math.round(50 + (p * 50 / 100)) + "%";
        case "ready":
        case "error":
        case "tag":
          return "100%";
        default:
          return "0";
      }
    },

    clickOnEnter: function (event) {
      if (event.key === "Enter") {
        event.preventDefault();
        event.stopPropagation();
        event.target.click();
      }
    },

    findTags: function () {
      const tagFilter = this.tagToAdd
        .replace(/[\n\r]/g, " ")
        .trim()
        .replace(/[\s]/g, "_")
        .toLowerCase();
      if (!tagFilter) {
        this.matchingTags = [];
        return;
      }
      this.matchingTags = Object.values(this.tagData)
        .map((a: any) => {
          const i = a.name.indexOf(tagFilter);
          return {
            id: a.id,
            name: a.name,
            starts: i === 0,
            contains: i >= 0,
          };
        })
        .filter((a) => {
          if (this.tags.indexOf(a.name) >= 0) {
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

    updateTagData: function () {
      this.tagData = copyObject(TagsController.Tags);
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

    removeTag: function (tag: string) {
      for (let i = 0; i < this.tags.length; i++) {
        if (this.tags[i] === tag) {
          this.tags.splice(i, 1);
          this.onTagAddChanged(true);
          break;
        }
      }
    },

    addTag: function (e) {
      if (e) {
        e.preventDefault();
      }
      this.addTagByName(this.tagToAdd);
    },

    addTagByName: function (tag: string) {
      tag = parseTagName(tag);
      this.removeTag(tag);
      this.tags.push(tag);
      this.onTagAddChanged(true);
    },

    showOptions: function (b: boolean) {
      this.optionsShown = b;
    },

    updateAlbums: function () {
      this.albums = AlbumsController.GetAlbumsListCopy().sort((a, b) => {
        if (a.nameLowerCase < b.nameLowerCase) {
          return -1;
        } else {
          return 1;
        }
      });

      if (this.inmodal) {
        this.album = this.fixedalbum;
      }
    },

    onPendingPush: function (m: UploadEntryMin) {
      this.pendingToUpload.push(m);
      this.updateCountCancellable(this.pendingToUpload);

      if (this.stateFilter.includes(m.status)) {
        if (this.selectedState === "pending") {
          this.filteredEntries.push(m);
        } else {
          this.filteredEntries.unshift(m);
        }
      }
    },

    onPendingRemove: function (i: number) {
      const removed = this.pendingToUpload.splice(i, 1)[0];
      this.updateCountCancellable(this.pendingToUpload);

      if (removed) {
        for (let j = 0; j < this.filteredEntries.length; j++) {
          if (this.filteredEntries[j].id === removed.id) {
            this.filteredEntries.splice(j, 1);
            break;
          }
        }
      }
    },

    onPendingClear: function () {
      this.pendingToUpload = UploadController.GetEntries();
      this.updateCountCancellable(this.pendingToUpload);
      this.updateFilteredEntries();
    },

    onPendingUpdate: function (i: number, m: UploadEntryMin) {
      let mustUpdate = (this.pendingToUpload[i].status !== m.status)
      this.pendingToUpload[i].status = m.status;
      this.pendingToUpload[i].error = m.error;
      this.pendingToUpload[i].progress = m.progress;
      this.pendingToUpload[i].mid = m.mid;
      if (mustUpdate) {
        this.updateCountCancellable(this.pendingToUpload);
      }

      let found = false;

      for (let j = 0; j < this.filteredEntries.length; j++) {
        if (this.filteredEntries[j].id === m.id) {
          found = true;
          if (mustUpdate && !this.stateFilter.includes(m.status)) {
            this.filteredEntries.splice(j, 1);
          } else {
            this.filteredEntries[j].status = m.status;
            this.filteredEntries[j].error = m.error;
            this.filteredEntries[j].progress = m.progress;
            this.filteredEntries[j].mid = m.mid;
          }
        }
      }

      if (!found && mustUpdate && this.stateFilter.includes(m.status)) {
        if (this.selectedState === "pending") {
          this.filteredEntries.push(m);
        } else {
          this.filteredEntries.unshift(m);
        }
      }
    },

    updateCountCancellable: function (list: UploadEntryMin[]) {
      let count = 0;
      let countPending = 0;
      let countReady = 0;
      let countError = 0;
      for (let l of list) {
        if (l.status !== "ready" && l.status !== "error") {
          count++;
        }

        if (STATE_FILTER_PENDING.includes(l.status)) {
          countPending++;
        } else if (STATE_FILTER_READY.includes(l.status)) {
          countReady++;
        } else if (STATE_FILTER_ERROR.includes(l.status)) {
          countError++;
        }
      }
      this.countCancellable = count;
      this.countPending = countPending;
      this.countReady = countReady;
      this.countError = countError;
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
  },
  mounted: function () {
    this.pendingToUpload = UploadController.GetEntries();
    this.updateCountCancellable(this.pendingToUpload);
    this.updateFilteredEntries();

    this.$options.onPendingPushH = this.onPendingPush.bind(this);
    this.$options.onPendingRemoveH = this.onPendingRemove.bind(this);
    this.$options.onPendingClearH = this.onPendingClear.bind(this);
    this.$options.onPendingUpdateH = this.onPendingUpdate.bind(this);

    AppEvents.AddEventListener(
      "upload-list-push",
      this.$options.onPendingPushH
    );
    AppEvents.AddEventListener(
      "upload-list-rm",
      this.$options.onPendingRemoveH
    );
    AppEvents.AddEventListener(
      "upload-list-clear",
      this.$options.onPendingClearH
    );
    AppEvents.AddEventListener(
      "upload-list-update",
      this.$options.onPendingUpdateH
    );

    this.updateTagData();
    this.$options.tagUpdateH = this.updateTagData.bind(this);
    AppEvents.AddEventListener("tags-update", this.$options.tagUpdateH);

    this.updateAlbums();
    this.$options.albumsUpdateH = this.updateAlbums.bind(this);
    AppEvents.AddEventListener("albums-update", this.$options.albumsUpdateH);
  },
  beforeUnmount: function () {
    AppEvents.RemoveEventListener("tags-update", this.$options.tagUpdateH);

    AppEvents.RemoveEventListener("albums-update", this.$options.albumsUpdateH);

    AppEvents.RemoveEventListener(
      "upload-list-push",
      this.$options.onPendingPushH
    );
    AppEvents.RemoveEventListener(
      "upload-list-rm",
      this.$options.onPendingRemoveH
    );
    AppEvents.RemoveEventListener(
      "upload-list-clear",
      this.$options.onPendingClearH
    );
    AppEvents.RemoveEventListener(
      "upload-list-update",
      this.$options.onPendingUpdateH
    );

    if (this.$options.findTagTimeout) {
      clearTimeout(this.$options.findTagTimeout);
    }
  },
});
</script>
