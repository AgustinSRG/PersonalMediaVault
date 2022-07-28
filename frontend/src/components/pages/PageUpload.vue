<template>
  <div class="page-inner page-inner-padded" :class="{ hidden: !display }">
    <input
      type="file"
      class="file-hidden"
      @change="inputFileChanged"
      name="media-upload"
      multiple="multiple"
    />
    <div
      class="upload-box"
      :class="{ dragging: dragging }"
      tabindex="0"
      @click="clickToSelect"
      @dragover="dragOver"
      @dragenter="dragEnter"
      @dragstart="dragEnter"
      @dragend="dragLeave"
      @dragleave="dragLeave"
      @drop="onDrop"
      @keydown="clickOnEnter"
    >
      <div class="upload-box-hint">
        {{ $t("Drop file here or click to open the file selection dialog.") }}
      </div>
    </div>

    <div
      class="upload-table table-responsive"
      v-if="pendingToUpload.length > 0"
    >
      <table class="table table-vmiddle">
        <thead>
          <tr>
            <th class="text-left">{{ $t("File") }}</th>
            <th class="text-left">{{ $t("Size") }}</th>
            <th class="text-left">{{ $t("Status") }}</th>
            <th></th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="m in pendingToUpload" :key="m.id">
            <td class="bold">{{ m.name }}</td>
            <td>{{ renderSize(m.size) }}</td>
            <td>{{ renderStatus(m.status, m.progress, m.error) }}</td>
            <td class="text-right one-line">
              <button
                v-if="
                  m.status === 'pending' ||
                  m.status === 'uploading' ||
                  m.status === 'encrypting'
                "
                type="button"
                class="table-btn"
                :title="$t('Cancel upload')"
                @click="removeFile(m.id)"
              >
                <i class="fas fa-times"></i>
              </button>
              <button
                v-if="m.status === 'ready'"
                type="button"
                class="table-btn"
                :title="$t('View media')"
                @click="goToMedia(m)"
              >
                <i class="fas fa-eye"></i>
              </button>
              <button
                v-if="m.status === 'ready'"
                type="button"
                class="table-btn"
                :title="$t('Done')"
                @click="removeFile(m.id)"
              >
                <i class="fas fa-check"></i>
              </button>
              <button
                v-if="m.status === 'error'"
                type="button"
                class="table-btn"
                :title="$t('Try again')"
                @click="tryAgain(m)"
              >
                <i class="fas fa-rotate"></i>
              </button>
              <button
                v-if="m.status === 'error'"
                type="button"
                class="table-btn"
                :title="$t('Remove')"
                @click="removeFile(m.id)"
              >
                <i class="fas fa-times"></i>
              </button>
            </td>
          </tr>
        </tbody>
      </table>
      <div class="form-group" v-if="pendingToUpload.length > 0">
        <button type="button" class="btn btn-primary btn-sm" @click="clearList">
          <i class="fas fa-broom"></i> {{ $t("Clear list") }}
        </button>
      </div>
    </div>
  </div>
</template>

<script lang="ts">
import { MediaAPI } from "@/api/api-media";
import { AppEvents } from "@/control/app-events";
import { AppStatus } from "@/control/app-status";
import { Request } from "@/utils/request";
import { defineComponent } from "vue";

const MAX_PARALLEL_UPLOADS = 4;

export default defineComponent({
  name: "PageUpload",
  props: {
    display: Boolean,
  },
  data: function () {
    return {
      dragging: false,
      nextId: 0,
      pendingToUpload: [],
      uploadingCount: 0,
    };
  },
  methods: {
    clickToSelect: function () {
      this.$el.querySelector(".file-hidden").click();
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

    addFile: function (file) {
      let id = this.nextId;
      this.nextId++;
      this.pendingToUpload.push({
        id: id,
        file: file,
        name: file.name,
        size: file.size,
        status: "pending",
        error: "",
        progress: 0,
        mid: -1,
        busy: false,
        lastRequest: 0,
      });
    },

    removeFile: function (id) {
      // Abort stuff here
      Request.Abort("upload-media-" + id);
      Request.Abort("check-media-encryption-" + id);

      // Remove from the array
      for (let i = 0; i < this.pendingToUpload.length; i++) {
        if (this.pendingToUpload[i].id === id) {
          if (this.pendingToUpload[i].status === "encrypting") {
            this.uploadingCount--;
          }

          this.pendingToUpload.splice(i, 1);
          return;
        }
      }
    },

    clearList: function () {
      for (let i = 0; i < this.pendingToUpload.length; i++) {
        const id = this.pendingToUpload[i].id;
        Request.Abort("upload-media-" + id);
        Request.Abort("check-media-encryption-" + id);
        if (this.pendingToUpload[i].status === "encrypting") {
          this.uploadingCount--;
        }
      }

      this.pendingToUpload = [];
    },

    tryAgain: function (m) {
      m.error = "";
      m.progress = 0;
      m.status = "pending";
    },

    goToMedia: function (m) {
      if (m.mid < 0) {
        return;
      }
      AppStatus.ClickOnMedia(m.mid, true);
    },

    renderStatus(status, p, err) {
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
        case "error":
          return this.$t("Error") + ": " + err;
        default:
          return "-";
      }
    },

    uploadMedia: function (m) {
      this.uploadingCount++;

      m.status = "uploading";
      m.progress = 0;

      Request.Pending(
        "upload-media-" + m.id,
        MediaAPI.UploadMedia(m.name, m.file)
      )
        .onUploadProgress((loaded, total) => {
          m.progress = Math.round(((loaded * 100) / total) * 100) / 100;
        })
        .onSuccess((data) => {
          m.mid = data.media_id;
          m.status = "encrypting";
          m.progress = 0;
        })
        .onCancel(() => {
          this.uploadingCount--;
        })
        .onRequestError((err) => {
          this.uploadingCount--;
          Request.ErrorHandler()
            .add(400, "*", () => {
              m.error = this.$t("Invalid media file provided");
              m.status = "error";
            })
            .add(401, "*", () => {
              m.error = this.$t("Access denied");
              m.status = "error";
              AppEvents.Emit("unauthorized");
            })
            .add(403, "*", () => {
              m.error = this.$t("Access denied");
              m.status = "error";
            })
            .add(500, "*", () => {
              m.error = this.$t("Internal server error");
              m.status = "error";
            })
            .add("*", "*", () => {
              m.error = this.$t("Could not connect to the server");
              m.status = "error";
            })
            .handle(err);
        })
        .onUnexpectedError((err) => {
          this.uploadingCount--;
          m.error = err.message;
          console.error(err);
          m.status = "error";
        });
    },

    checkEncryptionStatus: function (m) {
      if (m.busy) {
        return;
      }

      m.busy = true;
      m.lastRequest = Date.now();

      Request.Pending(
        "check-media-encryption-" + m.id,
        MediaAPI.GetMedia(m.mid)
      )
        .onSuccess((media) => {
          m.busy = false;
          if (media.ready) {
            m.status = "ready";
            this.uploadingCount--;
          } else {
            m.progress = media.ready_p;
          }
        })
        .onCancel(() => {
          m.busy = false;
        })
        .onRequestError((err) => {
          m.busy = false;
          Request.ErrorHandler()
            .add(401, "*", () => {
              AppEvents.Emit("unauthorized");
            })
            .add(404, "*", () => {
              m.error = this.$t("The media asset was deleted");
              m.status = "error";
              this.uploadingCount--;
            })
            .handle(err);
        })
        .onUnexpectedError((err) => {
          m.busy = false;
          console.error(err);
        });
    },

    tick: function () {
      for (let pending of this.pendingToUpload) {
        if (pending.status === "pending") {
          if (this.uploadingCount < MAX_PARALLEL_UPLOADS) {
            this.uploadMedia(pending);
          }
        } else if (pending.status === "encrypting") {
          if (!pending.busy && Date.now() - pending.lastRequest > 1000) {
            this.checkEncryptionStatus(pending);
          }
        }
      }
    },

    clickOnEnter: function (event) {
      if (event.key === "Enter") {
        event.preventDefault();
        event.stopPropagation();
        event.target.click();
      }
    },
  },
  mounted: function () {
    this.$options.timer = setInterval(this.tick.bind(this), 500);
  },
  beforeUnmount: function () {
    clearInterval(this.$options.timer);
  },
});
</script>

<style>
.file-hidden {
  display: none;
}

.upload-box {
  display: flex;

  width: 100%;
  height: 240px;

  border: dotted 2px var(--theme-border-color);

  align-items: center;
  justify-content: center;
  padding: 1rem;

  cursor: pointer;
}

.light-theme .upload-box:hover,
.light-theme .upload-box.dragging {
  background: rgba(0, 0, 0, 0.1);
}

.dark-theme .upload-box:hover,
.dark-theme .upload-box.dragging {
  background: rgba(255, 255, 255, 0.1);
}

.table.table-vmiddle td {
  vertical-align: middle;
}

.table-btn {
  display: inline-block;
  width: 32px;
  height: 32px;
  box-shadow: none;
  border: none;
  cursor: pointer;
  font-size: 16px;
  background: transparent;
  color: var(--theme-btn-color);
}

.table-btn:disabled {
  opacity: 0.7;
  cursor: default;
}

.table-btn:not(:disabled):hover {
  color: var(--theme-btn-hover-color);
}
</style>