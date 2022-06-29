<template>
  <div
    class="modal-container modal-container-settings"
    :class="{ hidden: !display }"
    tabindex="-1"
    role="dialog"
    :aria-hidden="!display"
    @click="close"
  >
    <form @submit="submit" class="modal-dialog modal-md" role="document" @click="stopPropagationEvent">
      <div class="modal-header">
        <div class="modal-title no-close">
          {{ $t("Rename album") }}
        </div>
      </div>
      <div class="modal-body">
        <div class="form-group">
          <label>{{ $t("Album name") }}:</label>
          <input
            type="text"
            name="album-name"
            autocomplete="off"
            v-model="name"
            :disabled="busy"
            maxlength="255"
            class="form-control form-control-full-width auto-focus"
          />
        </div>
        <div class="form-error">{{ error }}</div>
      </div>
      <div class="modal-footer">
        <button :disabled="busy" type="submit" class="modal-footer-btn">
          <i class="fas fa-pencil-alt"></i> {{ $t("Rename album") }}
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
import { defineComponent } from "vue";
import { useVModel } from "../../utils/vmodel";

export default defineComponent({
  name: "AlbumRenameModal",
  emits: ["update:display"],
  props: {
    display: Boolean,
  },
  data: function () {
    return {
      currentAlbum: -1,
      name: "",
      oldName: "",

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
        this.name = this.oldName;
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

      if (!this.name) {
        this.error = this.$t("Invalid album name provided");
        return;
      }

      if (this.name === this.oldName) {
        this.close();
        return;
      }

      if (AlbumsController.FindDuplicatedName(this.name)) {
        this.error = this.$t(
          "There is already another album with the same name"
        );
        return;
      }

      this.busy = true;
      this.error = "";

      const albumId = this.currentAlbum;

      Request.Do(AmbumsAPI.RenameAlbum(albumId, this.name))
        .onSuccess(() => {
          AppEvents.Emit("snack", this.$t("Album renamed") + ": " + this.name);
          this.busy = false;
          this.name = "";
          this.close();
          AlbumsController.OnChangedAlbum(albumId)
        })
        .onCancel(() => {
          this.busy = false;
        })
        .onRequestError((err) => {
          this.busy = false;
          Request.ErrorHandler()
            .add(400, "*", () => {
              this.error = this.$t("Invalid album name provided");
            })
            .add(401, "*", () => {
              this.error = this.$t("Access denied");
              AppEvents.Emit("unauthorized");
            })
            .add(403, "*", () => {
              this.error = this.$t("Access denied");
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
  },
  mounted: function () {
    this.$options.albumUpdateH = this.onAlbumUpdate.bind(this);
    AppEvents.AddEventListener(
      "current-album-update",
      this.$options.albumUpdateH
    );

    this.onAlbumUpdate();
    this.autoFocus();
  },
  beforeUnmount: function () {
    AppEvents.RemoveEventListener(
      "current-album-update",
      this.$options.albumUpdateH
    );
  },
  watch: {
    display: function () {
      this.error = "";
      this.name = this.oldName;
      this.autoFocus();
    },
  },
});
</script>

<style>
</style>
