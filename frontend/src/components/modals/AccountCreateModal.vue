<template>
  <div class="modal-container modal-container-settings" :class="{ hidden: !display }" tabindex="-1" role="dialog" :aria-hidden="!display" @keydown="keyDownHandle">
    <form v-if="display" @submit="submit" class="modal-dialog modal-md" role="document" @click="stopPropagationEvent">
      <div class="modal-header">
        <div class="modal-title">{{ $t("Create account") }}</div>
        <button type="button" class="modal-close-btn" :title="$t('Close')" @click="close">
          <i class="fas fa-times"></i>
        </button>
      </div>
      <div class="modal-body">
        <div class="form-group">
          <label>{{ $t("Account name") }}:</label>
          <input type="text" autocomplete="off" v-model="accountUsername" :disabled="busy" maxlength="255" class="form-control form-control-full-width auto-focus" />
        </div>

        <div class="form-group">
          <label>{{ $t("Account password") }}:</label>
          <input type="password" autocomplete="new-password" v-model="accountPassword" :disabled="busy" maxlength="255" class="form-control form-control-full-width" />
        </div>

        <div class="form-group">
          <label>{{ $t("Account password") }} ({{ $t("Again") }}):</label>
          <input type="password" autocomplete="new-password" v-model="accountPassword2" :disabled="busy" maxlength="255" class="form-control form-control-full-width" />
        </div>

        <div class="form-group">
          <label>{{ $t("Account type") }}:</label>
          <select v-model="accountWrite" class="form-control form-select form-control-full-width">
            <option :value="false">{{ $t("Read only") }}</option>
            <option :value="true">{{ $t("Read / Write") }}</option>
          </select>
        </div>

        <div class="form-group form-error">{{ error }}</div>

      </div>

      <div class="modal-footer no-padding">

        <button type="submit" :disabled="busy" class="modal-footer-btn">
          <i class="fas fa-plus"></i> {{ $t("Create account") }}
        </button>
      </div>
    </form>
  </div>
</template>

<script lang="ts">
import { AdminAPI } from "@/api/api-admin";
import { AppEvents } from "@/control/app-events";
import { Request } from "@/utils/request";
import { defineComponent, nextTick } from "vue";
import { useVModel } from "../../utils/v-model";
import { FocusTrap } from "../../utils/focus-trap";

export default defineComponent({
  components: {
  },
  name: "AccountCreateModal",
  emits: ["update:display", "account-created"],
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
      accountUsername: "",
      accountPassword: "",
      accountPassword2: "",
      accountWrite: false,

      busy: false,
      error: "",
    };
  },
  methods: {
    autoFocus: function () {
      nextTick(() => {
        const elem = this.$el.querySelector(".auto-focus");
        if (elem) {
          elem.focus();
          elem.select();
        }
      });
    },

    submit: function (e) {
      e.preventDefault();

      if (this.busy) {
        return;
      }

      if (this.accountPassword !== this.accountPassword2) {
        this.error = this.$t("The passwords do not match");
        return;
      }

      const username = this.accountUsername;
      const password = this.accountPassword;
      const write = this.accountWrite;

      this.busy = true;
      this.error = "";

      Request.Do(AdminAPI.CreateAccount(username, password, write))
        .onSuccess(() => {
          this.busy = false;
          AppEvents.Emit("snack", this.$t("Account created") + ": " + username);
          this.accountUsername = "";
          this.accountPassword = "";
          this.accountPassword2 = "";
          this.accountWrite = false;
          this.$emit("account-created");
          this.close();
        })
        .onCancel(() => {
          this.busy = false;
        })
        .onRequestError((err) => {
          this.busy = false;
          Request.ErrorHandler()
            .add(400, "USERNAME_INVALID", () => {
              this.error = this.$t("Invalid username provided");
            })
            .add(400, "USERNAME_IN_USE", () => {
              this.error = this.$t("The username is already in use");
            })
            .add(400, "PASSWORD_INVALID", () => {
              this.error = this.$t("Invalid password provided");
            })
            .add(400, "*", () => {
              this.error = this.$t("Bad request");
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

    if (this.display) {
      this.error = "";
      this.$options.focusTrap.activate();
      this.autoFocus();
    }
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
