<template>
  <div
    class="modal-container modal-container-settings"
    :class="{ hidden: !display }"
    tabindex="-1"
    role="dialog"
    :aria-hidden="!display"
    @click="close"
  >
    <form
       @submit="submit"
      class="modal-dialog modal-md"
      role="document"
      @click="stopPropagationEvent"
    >
      <div class="modal-header">
        <div class="modal-title">{{ $t("Change username") }}</div>
        <button type="button" class="modal-close-btn" :title="$t('Close')" @click="close">
          <i class="fas fa-times"></i>
        </button>
      </div>
      <div class="modal-body">
        <div class="form-group">
          <label>{{ $t("Current username") }}:</label>
          <input
            type="text"
            name="current-username"
            v-model="currentUsername"
            :disabled="busy"
            maxlength="255"
            readonly="readonly"
            class="form-control form-control-full-width"
          />
        </div>
        <div class="form-group">
          <label>{{ $t("New username") }}:</label>
          <input
            type="text"
            name="username"
            v-model="username"
            :disabled="busy"
            maxlength="255"
            class="form-control form-control-full-width auto-focus"
          />
        </div>
        <div class="form-group">
          <label>{{ $t("Password") }}:</label>
          <input
            type="password"
            name="password"
            v-model="password"
            :disabled="busy"
            maxlength="255"
            class="form-control form-control-full-width"
          />
        </div>
        <div class="form-error">{{ error }}</div>
      </div>
      <div class="modal-footer">
        <button type="submit" class="modal-footer-btn">
          <i class="fas fa-check"></i> {{ $t("Change username") }}
        </button>
      </div>
    </form>
  </div>
</template>

<script lang="ts">
import { AccountAPI } from "@/api/api-account";
import { AppEvents } from "@/control/app-events";
import { AuthController } from "@/control/auth";
import { Request } from "@/utils/request";
import { defineComponent } from "vue";
import { useVModel } from "../../utils/vmodel";

export default defineComponent({
  name: "ChangeUsernameModal",
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
      currentUsername: "",
      username: "",
      password: "",
      busy: false,
      error: "",
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

    submit: function (e) {
      e.preventDefault();

      if (this.busy) {
        return;
      }

      this.busy = true;
      this.error = "";

      Request.Do(AccountAPI.ChangeUsername(this.username, this.password))
        .onSuccess(() => {
          this.busy = false;
          AuthController.UpdateUsername(this.username);
          this.username = "";
          this.password = "";
          AppEvents.Emit("snack", this.$t("Vault username changed!"));
          this.close();
        })
        .onCancel(() => {
          this.busy = false;
        })
        .onRequestError((err) => {
          this.busy = false;
          Request.ErrorHandler()
            .add(400, "*", () => {
              this.error = this.$t("Invalid username provided");
            })
            .add(401, "*", () => {
              this.error = this.$t("Access denied");
              AppEvents.Emit("unauthorized");
            })
            .add(403, "*", () => {
              this.error = this.$t("Invalid password");
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

    close: function () {
      this.displayStatus = false;
    },

    stopPropagationEvent: function (e) {
      e.stopPropagation();
    },

    usernameUpdated: function () {
      this.currentUsername = AuthController.Username;
    },
  },
  mounted: function () {
    this.currentUsername = AuthController.Username;
    this.$options.usernameUpdatedH = this.usernameUpdated.bind(this);
    AppEvents.AddEventListener(
      "auth-status-changed",
      this.$options.usernameUpdatedH
    );
    this.autoFocus();
  },
  beforeUnmount: function () {
    AppEvents.RemoveEventListener(
      "auth-status-changed",
      this.$options.usernameUpdatedH
    );
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