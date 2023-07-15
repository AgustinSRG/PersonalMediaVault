<template>
  <div class="modal-container modal-container-settings" :class="{ hidden: !display }" tabindex="-1" role="dialog" :aria-hidden="!display" @mousedown="close" @touchstart="close" @keydown="keyDownHandle">
    <form v-if="display" @submit="submit" class="modal-dialog modal-md" role="document" @click="stopPropagationEvent" @mousedown="stopPropagationEvent" @touchstart="stopPropagationEvent">
      <div class="modal-header">
        <div class="modal-title">
          {{ $t("Delete media") }}
        </div>
        <button type="button" class="modal-close-btn" :title="$t('Close')" @click="close">
          <i class="fas fa-times"></i>
        </button>
      </div>
      <div class="modal-body">
        <div class="form-group">
          <label>{{
            $t(
              "Remember. If you delete the media by accident you would have to re-upload it. Make sure you actually want to delete it."
            )
          }}</label>
        </div>
        <div class="form-group">
          <label>{{ $t("Type 'confirm' for confirmation") }}:</label>
          <input type="text" name="confirmation" autocomplete="off" v-model="confirmation" :disabled="busy" maxlength="255" class="form-control form-control-full-width auto-focus" />
        </div>
        <div class="form-error">{{ error }}</div>
      </div>
      <div class="modal-footer no-padding">
        <button :disabled="busy" type="submit" class="modal-footer-btn">
          <i class="fas fa-trash-alt"></i> {{ $t("Delete media") }}
        </button>
      </div>
    </form>
  </div>
</template>

<script lang="ts">
import { MediaAPI } from "@/api/api-media";
import { AlbumsController } from "@/control/albums";
import { AppEvents } from "@/control/app-events";
import { AppStatus } from "@/control/app-status";
import { MediaController } from "@/control/media";
import { Request } from "@/utils/request";
import { defineComponent, nextTick } from "vue";
import { useVModel } from "../../utils/v-model";
import { FocusTrap } from "../../utils/focus-trap";

export default defineComponent({
  name: "MediaDeleteModal",
  emits: ["update:display"],
  props: {
    display: Boolean,
  },
  data: function () {
    return {
      currentMedia: -1,
      oldName: "",

      confirmation: "",

      busy: false,
      error: "",
    };
  },
  setup(props) {
    return {
      displayStatus: useVModel(props, "display"),
    };
  },
  methods: {
    show: function () {
      this.displayStatus = true;
    },

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

    onMediaUpdate: function () {
      this.currentMedia = AppStatus.CurrentMedia;
      if (MediaController.MediaData) {
        this.oldName = MediaController.MediaData.title;
      }
    },

    close: function () {
      if (this.busy) {
        return;
      }
      this.displayStatus = false;
    },

    stopPropagationEvent: function (e) {
      e.stopPropagation();
    },

    submit: function (e) {
      e.preventDefault();

      if (this.busy) {
        return;
      }

      if (this.confirmation.toLowerCase() !== "confirm") {
        this.error = this.$t(
          "You must type 'confirm' in order to confirm the deletion of the media"
        );
        return;
      }

      this.busy = true;
      this.error = "";

      const mediaId = this.currentMedia;

      Request.Do(MediaAPI.DeleteMedia(mediaId))
        .onSuccess(() => {
          AppEvents.Emit(
            "snack",
            this.$t("Media deleted") + ": " + this.oldName
          );
          this.busy = false;
          this.confirmation = "";
          this.close();
          AlbumsController.LoadCurrentAlbum();
          AppStatus.OnDeleteMedia();
        })
        .onCancel(() => {
          this.busy = false;
        })
        .onRequestError((err) => {
          this.busy = false;
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
          this.busy = false;
        });
    },

    keyDownHandle: function (e) {
      e.stopPropagation();
      if (e.key === "Escape") {
        this.close();
      }
    },
  },
  mounted: function () {
    this.$options.mediaUpdateH = this.onMediaUpdate.bind(this);
    AppEvents.AddEventListener("app-status-update", this.$options.mediaUpdateH);

    AppEvents.AddEventListener(
      "current-media-update",
      this.$options.mediaUpdateH
    );

    this.$options.showH = this.show.bind(this);
    AppEvents.AddEventListener("media-delete-request", this.$options.showH);

    this.$options.focusTrap = new FocusTrap(this.$el, this.close.bind(this));

    this.onMediaUpdate();

    if (this.display) {
      this.error = "";
      this.confirmation = "";
      this.$options.focusTrap.activate();
      this.autoFocus();
    }
  },
  beforeUnmount: function () {
    AppEvents.RemoveEventListener(
      "app-status-update",
      this.$options.mediaUpdateH
    );

    AppEvents.RemoveEventListener(
      "current-media-update",
      this.$options.mediaUpdateH
    );

    AppEvents.RemoveEventListener("media-delete-request", this.$options.showH);

    if (this.$options.focusTrap) {
      this.$options.focusTrap.destroy();
    }
  },
  watch: {
    display: function () {
      if (this.display) {
        this.error = "";
        this.confirmation = "";
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
