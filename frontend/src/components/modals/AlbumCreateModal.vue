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
        <div class="modal-title">
          {{ $t("Create new album") }}
        </div>
        <button type="button" class="modal-close-btn" :title="$t('Close')" @click="close">
          <i class="fas fa-times"></i>
        </button>
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
          <i class="fas fa-plus"></i> {{ $t("Create album") }}
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
  name: "AlbumCreateModal",
  emits: ["update:display"],
  props: {
    display: Boolean,
  },
  data: function () {
    return {
      name: "",

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

      if (AlbumsController.FindDuplicatedName(this.name)) {
        this.error = this.$t(
          "There is already another album with the same name"
        );
        return;
      }

      this.busy = true;
      this.error = "";

      Request.Do(AmbumsAPI.CreateAlbum(this.name))
        .onSuccess(() => {
          AppEvents.Emit("snack", this.$t("Album created") + ": " + this.name);
          this.busy = false;
          this.name = "";
          this.close();
          AlbumsController.Load();
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
    this.autoFocus();
  },
  watch: {
    display: function () {
      this.error = "";
      this.autoFocus();
    },
  },
});
</script>

<style>
</style>
