<template>
  <div
    class="modal-container modal-container-settings"
    :class="{ hidden: !display }"
    tabindex="-1"
    role="dialog"
    :aria-hidden="!display"
    @keydown="keyDownHandle"
  >
    <form
      @submit="submit"
      class="modal-dialog modal-md"
      role="document"
      @click="stopPropagationEvent"
    >
      <div class="modal-header">
        <div class="modal-title">{{ $t("Change password") }}</div>
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
          <label>{{ $t("Current password") }}:</label>
          <input
            type="password"
            name="password"
            v-model="currentPassword"
            :disabled="busy"
            maxlength="255"
            class="form-control form-control-full-width auto-focus"
          />
        </div>
        <div class="form-group">
          <label>{{ $t("New password") }}:</label>
          <input
            type="password"
            name="new-password"
            v-model="password"
            :disabled="busy"
            autocomplete="new-password"
            maxlength="255"
            class="form-control form-control-full-width"
          />
        </div>
        <div class="form-group">
          <label
            >{{ $t("New password") }} ({{
              $t("Repeat it for confirmation")
            }}):</label
          >
          <input
            type="password"
            name="new-password-repeat"
            v-model="password2"
            :disabled="busy"
            autocomplete="new-password"
            maxlength="255"
            class="form-control form-control-full-width"
          />
        </div>

        <div class="form-error">{{ error }}</div>
      </div>
      <div class="modal-footer">
        <button type="submit" class="modal-footer-btn">
          <i class="fas fa-check"></i> {{ $t("Change password") }}
        </button>
      </div>
    </form>
  </div>
</template>

<script lang="ts">
import { AccountAPI } from "@/api/api-account";
import { AppEvents } from "@/control/app-events";
import { Request } from "@/utils/request";
import { defineComponent, nextTick } from "vue";
import { useVModel } from "../../utils/vmodel";
import { FocusTrap } from "../../utils/focus-trap";

export default defineComponent({
  name: "ChangePasswordModal",
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
      currentPassword: "",
      password: "",
      password2: "",
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

      if (this.password !== this.password2) {
        this.error = this.$t("The passwords do not match");
        return;
      }

      this.busy = true;
      this.error = "";

      Request.Do(AccountAPI.ChangePassword(this.currentPassword, this.password))
        .onSuccess(() => {
          this.busy = false;
          this.currentPassword = "";
          this.password = "";
          this.password2 = "";
          AppEvents.Emit("snack", this.$t("Vault password changed!"));
          this.close();
        })
        .onCancel(() => {
          this.busy = false;
        })
        .onRequestError((err) => {
          this.busy = false;
          Request.ErrorHandler()
            .add(400, "*", () => {
              this.error = this.$t("Invalid password provided");
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

    keyDownHandle: function (e) {
      e.stopPropagation();
      if (e.key === "Escape") {
        this.close();
      }
    },
  },
  mounted: function () {
    this.$options.focusTrap = new FocusTrap(this.$el, this.close.bind(this));
  },
  beforeUnmount: function () {
    if (this.$options.focusTrap) {
      this.$options.focusTrap.destroy();
    }
  },
  watch: {
    display: function () {
      if (this.display) {
        this.error = "";
        this.currentPassword = "";
        this.password = "";
        this.password2 = "";
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
