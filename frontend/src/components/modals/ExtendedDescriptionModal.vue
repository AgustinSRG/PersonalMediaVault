<template>
  <div class="modal-container modal-container-settings modal-container-top" :class="{ hidden: !display }" tabindex="-1" role="dialog" :aria-hidden="!display" @mousedown="close" @touchstart="close" @keydown="keyDownHandle">
    <div v-if="display" class="modal-dialog modal-xl modal-height-100-wf" role="document" @click="stopPropagationEvent" @mousedown="stopPropagationEvent" @touchstart="stopPropagationEvent">
      <div class="modal-header">
        <div class="modal-title">{{ $t("Extended description") }}</div>
        <button type="button" class="modal-close-btn" :title="$t('Close')" @click="close">
          <i class="fas fa-times"></i>
        </button>
      </div>

      <div class="modal-body no-padding">
        <LoadingOverlay v-if="loading"></LoadingOverlay>
        <div v-if="!loading && editing" class="modal-body-textarea-container">
          <textarea :disabled="busy" class="form-control form-textarea no-resize auto-focus" v-model="contentToChange" :placeholder="$t('Input your description here') + '...'"></textarea>
        </div>
        <div v-if="!loading && !editing" class="extended-description-container" v-html="renderContent(content)"></div>
      </div>

      <div class="modal-footer text-right">
        <button v-if="canWrite && !editing" type="button" @click="startEdit" :disabled="busy || loading" class="btn btn-primary btn-mr">
          <i class="fas fa-pencil-alt"></i> {{ $t("Edit") }}
        </button>
        <button v-if="canWrite && editing" type="button" @click="cancelEdit" :disabled="busy || loading" class="btn btn-primary btn-mr">
          <i class="fas fa-times"></i> {{ $t("Cancel") }}
        </button>
        <button v-if="canWrite && editing" type="button" @click="saveChanges" :disabled="busy || loading" class="btn btn-primary">
          <i class="fas fa-check"></i> {{ $t("Save changes") }}
        </button>
        <button v-if="!editing" type="button" @click="close" :disabled="busy" class="btn btn-primary">
          <i class="fas fa-check"></i> {{ $t("Done") }}
        </button>
      </div>
    </div>
  </div>
</template>

<script lang="ts">
import { AppEvents } from "@/control/app-events";
import { AppStatus } from "@/control/app-status";
import { AuthController } from "@/control/auth";
import { defineComponent, nextTick } from "vue";
import { useVModel } from "../../utils/v-model";
import { FocusTrap } from "../../utils/focus-trap";
import { MediaController } from "@/control/media";

import LoadingOverlay from "@/components/layout/LoadingOverlay.vue";
import { Timeouts } from "@/utils/timeout";
import { GetAssetURL, Request } from "@/utils/request";
import { MediaAPI } from "@/api/api-media";
import { escapeHTML } from "@/utils/html";

export default defineComponent({
  components: {
    LoadingOverlay,
  },
  name: "ExtendedDescriptionModal",
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
      mid: AppStatus.CurrentMedia,
      title: MediaController.MediaData ? MediaController.MediaData.title : "",

      editing: false,

      content: "",
      contentToChange: "",

      loading: true,
      busy: false,
      canWrite: AuthController.CanWrite,

      changed: false,
    };
  },
  methods: {
    load: function () {
      Timeouts.Abort("media-ext-desc-load");
      Request.Abort("media-ext-desc-load");

      if (!this.display) {
        return;
      }

      if (!MediaController.MediaData) {
        return;
      }

      const descFilePath = MediaController.MediaData.ext_desc_url;

      if (!descFilePath) {
        this.content = "";
        this.contentToChange = "";
        this.loading = false;
        this.editing = !!this.canWrite;
        this.autoFocus();
        return;
      }

      this.loading = true;

      Request.Pending("media-ext-desc-load", {
        method: "GET",
        url: GetAssetURL(descFilePath),
      }).onSuccess(extendedDescText => {
        this.content = extendedDescText;
        this.contentToChange = extendedDescText;
        this.loading = false;
        this.editing = this.canWrite && !this.content;
        this.autoFocus();
      }).onRequestError(err => {
        Request.ErrorHandler()
          .add(401, "*", () => {
            AppEvents.Emit("unauthorized", false);
          })
          .add(404, "*", () => {
            this.content = "";
            this.contentToChange = "";
            this.loading = false;
            this.editing = !!this.canWrite;
            this.autoFocus();
          })
          .add("*", "*", () => {
            // Retry
            Timeouts.Set("media-ext-desc-load", 1500, this.load.bind(this));
          })
          .handle(err);
      }).onUnexpectedError(err => {
        console.error(err);
        // Retry
        Timeouts.Set("media-ext-desc-load", 1500, this.load.bind(this));
      });
    },

    autoFocus: function () {
      if (!this.display) {
        return;
      }
      nextTick(() => {
        const elem = this.$el.querySelector(".auto-focus");
        if (elem) {
          elem.focus();
        } else {
          this.$el.focus();
        }
      });
    },

    updateAuthInfo: function () {
      this.canWrite = AuthController.CanWrite;
      if (!this.canWrite) {
        this.cancelEdit();
      }
    },

    close: function () {
      if (this.busy) {
        return;
      }
      this.displayStatus = false;
      if (this.changed) {
        MediaController.Load();
      }
    },

    stopPropagationEvent: function (e) {
      e.stopPropagation();
    },

    updateMediaData: function () {
      this.mid = AppStatus.CurrentMedia;
      this.title = MediaController.MediaData ? MediaController.MediaData.title : "";
      this.load();
    },

    keyDownHandle: function (e) {
      e.stopPropagation();
      if (e.key === "Escape") {
        this.close();
      }
    },

    startEdit: function () {
      this.editing = true;
    },

    cancelEdit: function () {
      this.contentToChange = this.content;
      this.editing = false;
    },

    renderContent: function (text: string): string {
      return text.split("\n\n").map(paragraph => {
        if (paragraph.startsWith("###")) {
          return "<h3>" + escapeHTML(paragraph.substring(3)).replace(/\n/g, "<br>") + "</h3>";
        } else if (paragraph.startsWith("##")) {
          return "<h2>" + escapeHTML(paragraph.substring(2)).replace(/\n/g, "<br>") + "</h2>";
        } else if (paragraph.startsWith("#")) {
          return "<h1>" + escapeHTML(paragraph.substring(1)).replace(/\n/g, "<br>") + "</h1>";
        } else {
          return "<p>" + escapeHTML(paragraph).replace(/\n/g, "<br>") + "</p>";
        } 
      }).join("");
    },

    saveChanges: function () {
      if (this.busy) {
        return;
      }

      this.busy = true;

      Request.Do(MediaAPI.SetExtendedDescription(this.mid, this.contentToChange))
        .onSuccess(() => {
          this.busy = false;
          AppEvents.Emit("snack", this.$t("Successfully saved extended description"));
          this.content = this.contentToChange;
          this.editing = false;
          this.changed = true;
        })
        .onRequestError((err) => {
          Request.ErrorHandler()
            .add(401, "*", () => {
              AppEvents.Emit("unauthorized");
            })
            .handle(err);
        })
        .onUnexpectedError((err) => {
          console.error(err);
        });
    },
  },
  mounted: function () {
    this.$options.focusTrap = new FocusTrap(this.$el, this.close.bind(this));

    this.$options.authUpdateH = this.updateAuthInfo.bind(this);

    AppEvents.AddEventListener(
      "auth-status-changed",
      this.$options.authUpdateH
    );

    this.$options.mediaUpdateH = this.updateMediaData.bind(this);

    AppEvents.AddEventListener(
      "current-media-update",
      this.$options.mediaUpdateH
    );

    if (this.display) {
      this.$options.focusTrap.activate();
      this.load();
    }
  },
  beforeUnmount: function () {
    AppEvents.RemoveEventListener(
      "auth-status-changed",
      this.$options.authUpdateH
    );
    AppEvents.RemoveEventListener(
      "current-media-update",
      this.$options.mediaUpdateH
    );

    if (this.$options.focusTrap) {
      this.$options.focusTrap.destroy();
    }
  },
  watch: {
    display: function () {
      if (this.display) {
        if (this.$options.focusTrap) {
          this.$options.focusTrap.activate();
        }
        this.load();
      } else {
        if (this.$options.focusTrap) {
          this.$options.focusTrap.deactivate();
        }
      }
    },
  },
});
</script>
