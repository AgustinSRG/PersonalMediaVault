<template>
  <div
    class="modal-container modal-container-settings"
    :class="{ hidden: !display }"
    tabindex="-1"
    role="dialog"
    :aria-hidden="!display"
    @keydown="keyDownHandle"
  >
    <div
      class="modal-dialog modal-lg"
      role="document"
      @click="stopPropagationEvent"
    >
      <div class="modal-header">
        <div class="modal-title">{{ $t("Administrate accounts") }}</div>
        <button
          type="button"
          class="modal-close-btn"
          :title="$t('Close')"
          @click="close"
        >
          <i class="fas fa-times"></i>
        </button>
      </div>
      <div v-if="loading" class="modal-body">
        <p><i class="fa fa-spinner fa-spin"></i> {{ $t("Loading") }}...</p>
      </div>
      <div v-if="!loading" class="modal-body">
        <div class="form-group">
          <label>{{ $t("List of accounts") }}:</label>
          <div class="table-responsive">
            <table class="table">
              <thead>
                <tr>
                  <th class="text-left">{{ $t("Username") }}</th>
                  <th class="text-left">{{ $t("Account type") }}</th>
                  <th class="text-right"></th>
                </tr>
              </thead>
              <tbody>
                <tr v-for="a in accounts" :key="a.username">
                  <td class="bold">{{ a.username }}</td>
                  <td v-if="!a.write">{{ $t("Read only") }}</td>
                  <td v-if="a.write">{{ $t("Read / Write") }}</td>
                  <td class="text-right">
                    <button
                      type="button"
                      class="btn btn-danger btn-xs"
                      @click="askDeleteAccount(a.username)"
                    >
                      <i class="fas fa-trash-alt"></i> {{ $t("Delete") }}
                    </button>
                  </td>
                </tr>
              </tbody>
            </table>
          </div>
        </div>

        <form @submit="submit" class="border-top">
          <div class="form-group">
            <label>{{ $t("Account name") }}:</label>
            <input
              type="text"
              autocomplete="off"
              v-model="accountUsername"
              :disabled="busy"
              maxlength="255"
              class="form-control form-control-full-width"
            />
          </div>

          <div class="form-group">
            <label>{{ $t("Account password") }}:</label>
            <input
              type="password"
              autocomplete="off"
              v-model="accountPassword"
              :disabled="busy"
              maxlength="255"
              class="form-control form-control-full-width"
            />
          </div>

          <div class="form-group">
            <label>{{ $t("Account password") }} ({{ $t("Again") }}):</label>
            <input
              type="password"
              autocomplete="off"
              v-model="accountPassword2"
              :disabled="busy"
              maxlength="255"
              class="form-control form-control-full-width"
            />
          </div>

          <div class="form-group">
            <label>{{ $t("Account type") }}:</label>
            <select
              v-model="accountWrite"
              class="form-control form-select form-control-full-width"
            >
              <option :value="false">{{ $t("Read only") }}</option>
              <option :value="true">{{ $t("Read / Write") }}</option>
            </select>
          </div>

          <div class="form-group form-error">{{ error }}</div>

          <div class="form-group">
            <button
              type="submit"
              :disabled="busy"
              class="btn btn-primary"
            >
              <i class="fas fa-plus"></i> {{ $t("Create account") }}
            </button>
          </div>
        </form>
      </div>
    </div>
  </div>
</template>

<script lang="ts">
import { AdminAPI } from "@/api/api-admin";
import { AppEvents } from "@/control/app-events";
import { Request } from "@/utils/request";
import { Timeouts } from "@/utils/timeout";
import { defineComponent, nextTick } from "vue";
import { useVModel } from "../../utils/vmodel";
import { FocusTrap } from "../../utils/focus-trap";

export default defineComponent({
  name: "AccountsAdminModal",
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
      accounts: [],

      accountUsername: "",
      accountPassword: "",
      accountPassword2: "",
      accountWrite: false,

      loading: true,
      busy: false,
      error: "",
    };
  },
  methods: {
    load: function () {
      Timeouts.Abort("admin-accounts");
      Request.Abort("admin-accounts");

      if (!this.display) {
        return;
      }

      this.loading = true;

      Request.Pending("admin-accounts", AdminAPI.ListAccounts())
        .onSuccess((accounts) => {
          this.accounts = accounts;
          this.loading = false;
        })
        .onRequestError((err) => {
          Request.ErrorHandler()
            .add(401, "*", () => {
              AppEvents.Emit("unauthorized");
            })
            .add(403, "*", () => {
              this.displayStatus = false;
            })
            .add("*", "*", () => {
              // Retry
              Timeouts.Set("admin-accounts", 1500, this.load.bind(this));
            })
            .handle(err);
        })
        .onUnexpectedError((err) => {
          console.error(err);
          // Retry
          Timeouts.Set("admin-accounts", 1500, this.load.bind(this));
        });
    },

    askDeleteAccount: function (username: string) {
      AppEvents.Emit("account-del-confirmation", {
        name: username,
        callback: () => {
          this.deleteAccount(username);
        },
      });
    },

    deleteAccount: function (username: string) {
      if (this.busy) {
        return;
      }

      this.busy = true;
      this.error = "";

      Request.Do(AdminAPI.DeleteAccount(username))
        .onSuccess(() => {
          this.busy = false;
          AppEvents.Emit("snack", this.$t("Account deleted") + ": " + username);
          this.load();
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
              // Already deleted?
              this.busy = false;
              this.load();
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
          this.load();
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
    this.load();
  },
  beforeUnmount: function () {
    Timeouts.Abort("admin-accounts");
    Request.Abort("admin-accounts");
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
        nextTick(() => {
          this.$el.focus();
        });
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

<style>
</style>
