<template>
  <ModalDialogContainer ref="modalContainer" v-model:display="displayStatus">
    <div v-if="display" class="modal-dialog modal-lg" role="document">
      <div class="modal-header">
        <div class="modal-title">{{ $t("Administrate accounts") }}</div>
        <button type="button" class="modal-close-btn" :title="$t('Close')" @click="close">
          <i class="fas fa-times"></i>
        </button>
      </div>
      <div v-if="loading" class="modal-body">
        <p><i class="fa fa-spinner fa-spin"></i> {{ $t("Loading") }}...</p>
      </div>
      <div v-if="!loading" class="modal-body no-padding">

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
              <tr v-for="a in accounts" :key="a.username" class="tr-align-middle">
                <td class="bold">{{ a.username }}</td>
                <td v-if="!a.write">{{ $t("Read only") }}</td>
                <td v-if="a.write">{{ $t("Read / Write") }}</td>
                <td class="text-right">
                  <button type="button" class="btn btn-danger btn-xs" @click="askDeleteAccount(a.username)">
                    <i class="fas fa-trash-alt"></i> {{ $t("Delete") }}
                  </button>
                </td>
              </tr>
              <tr class="tr-align-middle">
                <td colspan="4" class="text-right">
                  <button type="button" @click="createAccount" :disabled="busy" class="btn btn-primary btn-xs">
                    <i class="fas fa-plus"></i> {{ $t("Create account") }}
                  </button>
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>

      <div class="modal-footer no-padding">
        <button type="button" @click="close" :disabled="busy" class="modal-footer-btn">
          <i class="fas fa-check"></i> {{ $t("Done") }}
        </button>
      </div>

    </div>

    <AccountDeleteModal ref="deleteModal" v-model:display="displayAccountDelete"></AccountDeleteModal>
    <AccountCreateModal v-model:display="displayAccountCreate" @account-created="load"></AccountCreateModal>
  </ModalDialogContainer>
</template>

<script lang="ts">
import { AdminAPI } from "@/api/api-admin";
import { AppEvents } from "@/control/app-events";
import { Request } from "@/utils/request";
import { Timeouts } from "@/utils/timeout";
import { defineComponent, nextTick } from "vue";
import { useVModel } from "../../utils/v-model";
import AccountDeleteModal from "../modals/AccountDeleteModal.vue";
import AccountCreateModal from "../modals/AccountCreateModal.vue";

export default defineComponent({
  components: {
    AccountDeleteModal,
    AccountCreateModal,
  },
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

      displayAccountDelete: false,
      displayAccountCreate: false,

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
      this.$refs.deleteModal.show({
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

    createAccount: function () {
      this.displayAccountCreate = true;
    },

    close: function () {
      this.displayAccountDelete = false;
      this.displayAccountCreate = false;
      this.$refs.modalContainer.close();
    },
  },
  mounted: function () {
    this.load();

    if (this.display) {
      this.error = "";
      nextTick(() => {
        this.$el.focus();
      });
      this.displayAccountDelete = false;
      this.displayAccountCreate = false;
    }
  },
  beforeUnmount: function () {
    Timeouts.Abort("admin-accounts");
    Request.Abort("admin-accounts");
  },
  watch: {
    display: function () {
      if (this.display) {
        this.error = "";
        nextTick(() => {
          this.$el.focus();
        });
        this.displayAccountDelete = false;
        this.displayAccountCreate = false;
        this.load();
      }
    },
  },
});
</script>
