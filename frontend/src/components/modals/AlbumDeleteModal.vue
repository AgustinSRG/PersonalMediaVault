<template>
  <div
    class="modal-container modal-container-settings"
    :class="{ hidden: !display }"
    tabindex="-1"
    role="dialog"
    :aria-hidden="!display"
    @click="close"
    @keydown="keyDownHandle"
  >
    <form
      @submit="submit"
      class="modal-dialog modal-md"
      role="document"
      @click="stopPropagationEvent"
    >
      <div class="modal-header">
        <div class="modal-title">
          {{ $t("Delete album") }}
        </div>
        <button
          type="button"
          class="modal-close-btn"
          :title="$t('Close')"
          @click="close"
        >
          <i class="fas fa-times"></i>
        </button>
      </div>
      <div class="modal-body">
        <div class="form-group">
          <label>{{
            $t(
              "Remember. If you delete the album by accident you would have to recreate it. Make sure you actually want to delete it."
            )
          }}</label>
        </div>
        <div class="form-group">
          <label>{{ $t("Type 'confirm' for confirmation") }}:</label>
          <input
            type="text"
            name="confirmation"
            autocomplete="off"
            v-model="confimation"
            :disabled="busy"
            maxlength="255"
            class="form-control form-control-full-width auto-focus"
          />
        </div>
        <div class="form-error">{{ error }}</div>
      </div>
      <div class="modal-footer">
        <button :disabled="busy" type="submit" class="modal-footer-btn">
          <i class="fas fa-trash-alt"></i> {{ $t("Delete album") }}
        </button>
      </div>
    </form>
  </div>
</template>

<script lang="ts">
import { AmbumsAPI } from "@/api/api-albums";
import { AlbumsController } from "@/control/albums";
import { AppEvents } from "@/control/app-events";
import { Request } from "@/utils/request";
import { defineComponent, nextTick } from "vue";
import { useVModel } from "../../utils/vmodel";
import { FocusTrap } from "../../utils/focus-trap";

export default defineComponent({
  name: "AlbumDeleteModal",
  emits: ["update:display"],
  props: {
    display: Boolean,
  },
  data: function () {
    return {
      currentAlbum: -1,
      oldName: "",

      confimation: "",

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
    autoFocus: function () {
      if (!this.display) {
        return;
      }
      const elem = this.$el.querySelector(".auto-focus");
      if (elem) {
        setTimeout(() => {
          elem.focus();
        }, 200);
      }
    },

    onAlbumUpdate: function () {
      this.currentAlbum = AlbumsController.CurrentAlbum;
      if (AlbumsController.CurrentAlbumData) {
        this.oldName = AlbumsController.CurrentAlbumData.name;
      }
    },

    close: function () {
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

      if (this.confimation.toLowerCase() !== "confirm") {
        this.error = this.$t(
          "You must type 'confirm' in order to confirm the deletion of the album"
        );
        return;
      }

      this.busy = true;
      this.error = "";

      const albumId = this.currentAlbum;

      Request.Do(AmbumsAPI.DeleteAlbum(albumId))
        .onSuccess(() => {
          AppEvents.Emit(
            "snack",
            this.$t("Album deleted") + ": " + this.oldName
          );
          this.busy = false;
          this.confimation = "";
          this.close();
          AlbumsController.OnChangedAlbum(albumId);
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
    this.$options.albumUpdateH = this.onAlbumUpdate.bind(this);
    AppEvents.AddEventListener(
      "current-album-update",
      this.$options.albumUpdateH
    );

    this.onAlbumUpdate();
    this.$options.focusTrap = new FocusTrap(this.$el, this.close.bind(this));
  },
  beforeUnmount: function () {
    AppEvents.RemoveEventListener(
      "current-album-update",
      this.$options.albumUpdateH
    );
    if (this.$options.focusTrap) {
      this.$options.focusTrap.destroy();
    }
  },
  watch: {
    display: function () {
      if (this.display) {
        this.error = "";
        this.confimation = "";
        if (this.$options.focusTrap) {
          this.$options.focusTrap.activate();
        }
        nextTick(() => {
          this.$el.focus();
        });
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
</style>
